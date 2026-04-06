// Package transcript parses Claude Code JSONL transcripts and exports conversation
// turns as OTEL spans. Each session gets one stable TraceID. Two export modes:
//   - ExportNew: exports only turns since the last export (unscored, for periodic previews)
//   - ExportAndScore: exports all turns with Claude quality scores (for session end)
package transcript

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// scorerTimeout is the deadline passed to `claude -p`. Package-level var for test override.
var scorerTimeout = 30 * time.Second

// execScorer is overridden in tests to inject fake claude output.
var execScorer = exec.CommandContext

// Turn represents one conversation turn.
type Turn struct {
	Question string
	Answer   string
	Start    time.Time
	End      time.Time
}

// scoreEntry is a JSON element returned by `claude -p`.
type scoreEntry struct {
	Name   string  `json:"name"`
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

// jsonlEntry is used for lenient line-by-line JSONL parsing.
type jsonlEntry struct {
	Type      string          `json:"type"`
	UUID      string          `json:"uuid"`
	Timestamp string          `json:"timestamp"`
	Message   json.RawMessage `json:"message"`
}

// messageBody holds the message field of a JSONL entry.
type messageBody struct {
	Content rawContent `json:"content"`
}

// rawContent is either a string (user question) or a JSON array (tool results /
// assistant blocks). We handle both via custom unmarshalling.
type rawContent struct {
	Str    string
	Blocks []contentBlock
	IsStr  bool
}

// contentBlock is one element of an assistant or tool-result content array.
type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (c *rawContent) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		c.Str = s
		c.IsStr = true
		return nil
	}
	var blocks []contentBlock
	if err := json.Unmarshal(data, &blocks); err != nil {
		return fmt.Errorf("content: neither string nor array")
	}
	c.Blocks = blocks
	return nil
}

// Parse reads the JSONL file at path and returns all complete turns.
// Malformed lines are skipped with a warning log. Returns an error only if the
// file cannot be opened.
func Parse(path string) ([]Turn, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("transcript: open %s: %w", path, err)
	}
	defer f.Close()

	var turns []Turn
	var pendingQ string
	var pendingStart time.Time

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		raw := scanner.Bytes()
		if len(raw) == 0 {
			continue
		}
		var entry jsonlEntry
		if err := json.Unmarshal(raw, &entry); err != nil {
			log.Printf("shipmate: transcript: skip malformed line: %v", err)
			continue
		}
		if entry.Type != "user" && entry.Type != "assistant" {
			continue
		}
		ts, err := time.Parse(time.RFC3339, entry.Timestamp)
		if err != nil {
			log.Printf("shipmate: transcript: skip bad timestamp %q: %v", entry.Timestamp, err)
			continue
		}
		switch entry.Type {
		case "user":
			var body messageBody
			if err := json.Unmarshal(entry.Message, &body); err != nil {
				log.Printf("shipmate: transcript: skip user entry: %v", err)
				continue
			}
			if !body.Content.IsStr {
				continue // tool result array — not a user question
			}
			q := strings.TrimSpace(body.Content.Str)
			if q == "" {
				continue
			}
			pendingQ = q
			pendingStart = ts
		case "assistant":
			if pendingQ == "" {
				continue
			}
			var body messageBody
			if err := json.Unmarshal(entry.Message, &body); err != nil {
				log.Printf("shipmate: transcript: skip assistant entry: %v", err)
				continue
			}
			var answer string
			for _, b := range body.Content.Blocks {
				if b.Type == "text" && b.Text != "" {
					answer = b.Text
					break
				}
			}
			turns = append(turns, Turn{
				Question: pendingQ,
				Answer:   answer,
				Start:    pendingStart,
				End:      ts,
			})
			pendingQ = ""
			pendingStart = time.Time{}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("shipmate: transcript: scanner: %v", err)
	}
	return turns, nil
}

// TurnExporter holds session state for exporting transcript turns as OTEL spans.
// A stable TraceID is generated once at construction so all spans across multiple
// periodic exports share the same trace.
type TurnExporter struct {
	path          string
	sessionID     string
	upstream      sdktrace.SpanExporter
	traceID       trace.TraceID
	sessionSpanID trace.SpanID
	lastExported  int
	mu            sync.Mutex
}

// NewTurnExporter creates a TurnExporter. Generates a stable random TraceID and
// stable session span SpanID once for the lifetime of this exporter.
func NewTurnExporter(path, sessionID string, upstream sdktrace.SpanExporter) *TurnExporter {
	var traceID trace.TraceID
	var spanID trace.SpanID
	if _, err := rand.Read(traceID[:]); err != nil {
		log.Printf("shipmate: transcript: generate traceID: %v (using zero)", err)
	}
	if _, err := rand.Read(spanID[:]); err != nil {
		log.Printf("shipmate: transcript: generate spanID: %v (using zero)", err)
	}
	return &TurnExporter{
		path:          path,
		sessionID:     sessionID,
		upstream:      upstream,
		traceID:       traceID,
		sessionSpanID: spanID,
	}
}

// ExportNew parses the transcript, exports only turns[lastExported:] as unscored
// spans to the upstream exporter, and advances lastExported. Also re-emits the
// session span with an updated end time. Noop if the transcript file is missing
// or there are no new turns.
//
// ExportNew must not be called concurrently; the daemon satisfies this via a
// single ticker goroutine.
func (te *TurnExporter) ExportNew(ctx context.Context) error {
	te.mu.Lock()
	last := te.lastExported
	te.mu.Unlock()

	turns, err := te.parseSafe()
	if err != nil || len(turns) == 0 {
		return nil
	}
	if len(turns) <= last {
		return nil // nothing new
	}

	newTurns := turns[last:]
	spans := te.buildSpans(turns, newTurns, nil)

	if err := te.upstream.ExportSpans(ctx, spans); err != nil {
		log.Printf("shipmate: transcript: ExportNew: export: %v", err)
		return nil // do not advance lastExported; retry on next tick
	}

	te.mu.Lock()
	te.lastExported = len(turns)
	te.mu.Unlock()
	return nil
}

// ExportAndScore parses the transcript, scores all turns via claude -p, and
// exports all spans (session + all turns) with score attributes to the upstream.
// Always re-exports all turns regardless of lastExported — intended for session end.
// Non-fatal: if claude fails or returns bad JSON, spans are exported unscored.
func (te *TurnExporter) ExportAndScore(ctx context.Context) error {
	turns, err := te.parseSafe()
	if err != nil || len(turns) == 0 {
		return nil
	}

	scores := te.callClaude(ctx, turns)
	spans := te.buildSpans(turns, turns, scores)

	if err := te.upstream.ExportSpans(ctx, spans); err != nil {
		log.Printf("shipmate: transcript: ExportAndScore: export: %v", err)
	}
	return nil
}

// Shutdown calls Shutdown on the upstream exporter.
func (te *TurnExporter) Shutdown(ctx context.Context) error {
	return te.upstream.Shutdown(ctx)
}

// parseSafe calls Parse and returns nil turns (not an error) when the file is missing.
func (te *TurnExporter) parseSafe() ([]Turn, error) {
	if te.path == "" {
		return nil, nil
	}
	turns, err := Parse(te.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		log.Printf("shipmate: transcript: parseSafe: %v", err)
		return nil, err
	}
	return turns, nil
}

// spanIDForTurn derives a deterministic SpanID from the sessionID and turn index so
// that the same turn always maps to the same SpanID across multiple export calls.
// This lets Tempo deduplicate periodic (unscored) and final (scored) exports of the
// same turn rather than treating them as distinct spans.
func spanIDForTurn(sessionID string, turnIndex int) trace.SpanID {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:turn:%d", sessionID, turnIndex)))
	var id trace.SpanID
	copy(id[:], h[:8])
	return id
}

// buildSpans creates ReadOnlySpan stubs for the session span and the given subset
// of turns. allTurns is used to determine the session span's time boundaries.
// subset is the slice of turns to emit as child spans.
// scores maps turn question (truncated to 128 chars) to scoreEntry for attribute enrichment.
func (te *TurnExporter) buildSpans(allTurns, subset []Turn, scores map[string]scoreEntry) []sdktrace.ReadOnlySpan {
	if len(allTurns) == 0 {
		return nil
	}

	sessionStart := allTurns[0].Start
	sessionEnd := allTurns[len(allTurns)-1].End

	// Construct a session SpanContext with our stable IDs.
	sessionSC := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    te.traceID,
		SpanID:     te.sessionSpanID,
		TraceFlags: trace.FlagsSampled,
		Remote:     false,
	})

	stubs := make(tracetest.SpanStubs, 0, 1+len(subset))

	// Session span.
	sessionAttrs := []attribute.KeyValue{
		attribute.String("session.id", te.sessionID),
	}
	stubs = append(stubs, tracetest.SpanStub{
		Name:        "session",
		SpanContext: sessionSC,
		StartTime:   sessionStart,
		EndTime:     sessionEnd,
		Attributes:  sessionAttrs,
	})

	// Build an index from subset turn to its position in allTurns so we can
	// derive a stable SpanID. subset is always a contiguous suffix of allTurns,
	// so we find the offset once by pointer-comparing the first element.
	subsetOffset := 0
	if len(subset) > 0 && len(allTurns) > 0 {
		for i := range allTurns {
			if allTurns[i].Start == subset[0].Start && allTurns[i].Question == subset[0].Question {
				subsetOffset = i
				break
			}
		}
	}

	// Turn spans.
	for i, turn := range subset {
		name := truncate(turn.Question, 128)
		spanID := spanIDForTurn(te.sessionID, subsetOffset+i)
		turnSC := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID:    te.traceID,
			SpanID:     spanID,
			TraceFlags: trace.FlagsSampled,
		})
		// Parent link: reference the session span.
		parentLink := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID:    te.traceID,
			SpanID:     te.sessionSpanID,
			TraceFlags: trace.FlagsSampled,
		})

		attrs := []attribute.KeyValue{
			attribute.String("session.id", te.sessionID),
		}
		if len(turn.Answer) > 0 {
			attrs = append(attrs, attribute.String("answer.preview", truncate(turn.Answer, 256)))
		}
		if scores != nil {
			if se, ok := scores[name]; ok {
				attrs = append(attrs, attribute.Float64("memory.score", se.Score))
				if se.Reason != "" {
					attrs = append(attrs, attribute.String("memory.score.reason", se.Reason))
				}
			}
		}

		stubs = append(stubs, tracetest.SpanStub{
			Name:        name,
			SpanContext: turnSC,
			Parent:      parentLink,
			StartTime:   turn.Start,
			EndTime:     turn.End,
			Attributes:  attrs,
		})
	}

	return stubs.Snapshots()
}

// callClaude invokes `claude -p` with a prompt summarising the turns and returns
// parsed score entries keyed by truncated turn name. Non-fatal: logs and returns nil on error.
func (te *TurnExporter) callClaude(ctx context.Context, turns []Turn) map[string]scoreEntry {
	if len(turns) == 0 {
		return nil
	}
	prompt := buildPrompt(turns)
	scoreCtx, cancel := context.WithTimeout(ctx, scorerTimeout)
	defer cancel()

	cmd := execScorer(scoreCtx, "claude", "-p", prompt)
	out, err := cmd.Output()
	if err != nil {
		log.Printf("shipmate: transcript: claude -p: %v", err)
		return nil
	}

	cleaned := stripMarkdownFences(string(out))
	var entries []scoreEntry
	if err := json.Unmarshal([]byte(cleaned), &entries); err != nil {
		log.Printf("shipmate: transcript: parse claude response: %v (output: %s)", err, truncate(string(out), 200))
		return nil
	}
	log.Printf("shipmate: transcript: scored %d turns", len(entries))
	byName := make(map[string]scoreEntry, len(entries))
	for _, e := range entries {
		byName[truncate(e.Name, 128)] = e
	}
	return byName
}

// buildPrompt constructs the text prompt for `claude -p` from conversation turns.
func buildPrompt(turns []Turn) string {
	var sb strings.Builder
	sb.WriteString("You are reviewing a Claude Code session transcript. " +
		"Below is a list of conversation turns (user question + assistant response). " +
		"For each turn, respond with a JSON array of objects with fields: " +
		"\"name\" (the first 128 chars of the question, matching exactly), " +
		"\"score\" (a float from -1.0 to 1.0, exclusive — never use exactly 1.0 or -1.0, up to 3 decimal places), " +
		"and \"reason\" (a brief explanation).\n\n" +
		"Scoring guidance:\n" +
		"  0.9: excellent response — correct, concise, actionable\n" +
		"  0.7: good response with minor issues\n" +
		"  0.4: acceptable but has problems (verbose, off-target, etc.)\n" +
		"  0.0: neutral — hard to evaluate\n" +
		" -0.4: poor — incorrect or misleading\n" +
		" -0.7: bad — likely harmful or wrong\n" +
		" -0.9: very bad — dangerous or fundamentally wrong\n\n" +
		"Return ONLY the JSON array, no other text.\n\nTurns:\n")
	for i, t := range turns {
		fmt.Fprintf(&sb, "%d. question=%q answer_preview=%q\n",
			i+1, truncate(t.Question, 128), truncate(t.Answer, 256))
	}
	return sb.String()
}

// stripMarkdownFences removes ```json / ``` wrappers that claude sometimes adds.
func stripMarkdownFences(s string) string {
	s = strings.TrimSpace(s)
	if after, ok := strings.CutPrefix(s, "```json"); ok {
		s = after
	} else if after, ok := strings.CutPrefix(s, "```"); ok {
		s = after
	}
	s, _ = strings.CutSuffix(strings.TrimSpace(s), "```")
	return strings.TrimSpace(s)
}

// truncate shortens s to at most n runes.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}
