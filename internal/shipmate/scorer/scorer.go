// Package scorer provides a buffering SpanExporter that enriches spans with
// a quality score by calling `claude -p` at session end.
//
// BufferingExporter wraps an upstream SpanExporter. ExportSpans buffers all
// received spans in memory without forwarding them. Score() calls claude with
// a summary of the buffered spans, parses a JSON score response, annotates
// each span with the matching score, and forwards all spans to the upstream.
//
// All errors from Score() are non-fatal: if claude is unavailable, times out,
// or returns invalid JSON, Score() logs the error and returns nil. Spans are
// always forwarded to the upstream regardless.
package scorer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// scorerTimeout is the deadline passed to the `claude -p` subprocess.
// Package-level var so tests can override it.
var scorerTimeout = 30 * time.Second

// execScorer is a package-level var so tests can inject a fake executor.
var execScorer = exec.CommandContext

// scoreEntry is the JSON element returned by `claude -p` for each span.
type scoreEntry struct {
	Name   string  `json:"name"`
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

// BufferingExporter implements sdktrace.SpanExporter.
// ExportSpans stores spans in memory; Score() enriches and forwards them.
type BufferingExporter struct {
	mu       sync.Mutex
	spans    []sdktrace.ReadOnlySpan
	upstream sdktrace.SpanExporter
}

// NewBufferingExporter wraps upstream, buffering all spans until Score() is called.
func NewBufferingExporter(upstream sdktrace.SpanExporter) *BufferingExporter {
	return &BufferingExporter{upstream: upstream}
}

// ExportSpans appends spans to the internal buffer without forwarding them.
func (b *BufferingExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.spans = append(b.spans, spans...)
	return nil
}

// Shutdown forwards any buffered spans to the upstream and then shuts it down.
// It does not call Score(). Call Score() explicitly before Shutdown if scoring
// is desired.
//
// A fresh bounded context is used for the upstream export so that a
// partially-expired or already-cancelled caller ctx does not prevent spans
// from reaching the upstream.
func (b *BufferingExporter) Shutdown(ctx context.Context) error {
	b.mu.Lock()
	spans := b.spans
	b.spans = nil
	b.mu.Unlock()

	if len(spans) > 0 {
		exportCtx, exportCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer exportCancel()
		if err := b.upstream.ExportSpans(exportCtx, spans); err != nil {
			log.Printf("shipmate: scorer: flush on shutdown: %v", err)
		}
	}
	return b.upstream.Shutdown(ctx)
}

// Score calls `claude -p` with a prompt summarising the buffered spans, parses
// the JSON score response, and forwards all spans (enriched where a matching
// score entry was returned) to the upstream exporter.
//
// Errors from the claude subprocess (non-zero exit, timeout, bad JSON) are
// logged and do not prevent spans from being forwarded. Score always returns nil.
func (b *BufferingExporter) Score(ctx context.Context) error {
	b.mu.Lock()
	spans := b.spans
	b.spans = nil
	b.mu.Unlock()

	if len(spans) == 0 {
		return nil
	}

	scores := b.callClaude(ctx, spans)
	enriched := enrichSpans(spans, scores)
	if err := b.upstream.ExportSpans(ctx, enriched); err != nil {
		log.Printf("shipmate: scorer: export enriched spans: %v", err)
	}
	return nil
}

// callClaude invokes `claude -p` with a span summary prompt and returns the
// parsed score entries. On any error it logs and returns nil.
func (b *BufferingExporter) callClaude(ctx context.Context, spans []sdktrace.ReadOnlySpan) []scoreEntry {
	prompt := buildPrompt(spans)

	// Use context.Background() so scorer always gets the full scorerTimeout budget,
	// regardless of how much of the parent context has already elapsed.
	scoreCtx, cancel := context.WithTimeout(context.Background(), scorerTimeout)
	defer cancel()

	cmd := execScorer(scoreCtx, "claude", "-p", prompt)
	out, err := cmd.Output()
	if err != nil {
		log.Printf("shipmate: scorer: claude -p: %v", err)
		return nil
	}

	cleaned := stripMarkdownFences(string(out))
	var entries []scoreEntry
	if err := json.Unmarshal([]byte(cleaned), &entries); err != nil {
		log.Printf("shipmate: scorer: parse claude response: %v (output: %s)", err, truncate(string(out), 200))
		return nil
	}
	log.Printf("shipmate: scorer: scored %d spans", len(entries))
	for _, e := range entries {
		log.Printf("shipmate: scorer: span=%q score=%.2f reason=%s", e.Name, e.Score, e.Reason)
	}
	return entries
}

// buildPrompt constructs the text prompt sent to `claude -p`.
func buildPrompt(spans []sdktrace.ReadOnlySpan) string {
	var sb strings.Builder
	sb.WriteString("You are reviewing a session trace for a Claude Code agent. " +
		"Below is a list of spans (tool calls and events) recorded during the session. " +
		"For each span, respond with a JSON array of objects with fields: " +
		"\"name\" (matching the span name), " +
		"\"score\" (a float from -1.0 to 1.0, exclusive — never use exactly 1.0 or -1.0, up to 3 decimal places), " +
		"and \"reason\" (a brief explanation).\n\n" +
		"Scoring guidance:\n" +
		"  0.9: near-perfect — correct tool, precise targeting, minimal blast radius\n" +
		"  0.7: good — works well with minor inefficiencies\n" +
		"  0.4: acceptable — correct outcome but notable issues (wrong tool, fragile approach)\n" +
		"  0.0: neutral — no clear signal either way\n" +
		" -0.4: poor — unnecessary risk, wrong tool, or misleading approach\n" +
		" -0.7: bad — likely to cause problems or significantly wasteful\n" +
		" -0.9: very bad — dangerous, destructive, or fundamentally wrong approach\n\n" +
		"Return ONLY the JSON array, no other text.\n\nSpans:\n")

	for i, s := range spans {
		fmt.Fprintf(&sb, "%d. name=%q", i+1, s.Name())
		for _, kv := range s.Attributes() {
			fmt.Fprintf(&sb, " %s=%q", kv.Key, kv.Value.AsString())
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// enrichSpans creates new SpanStubs with score attributes added where a
// matching score entry exists, and returns them as ReadOnlySpans.
func enrichSpans(spans []sdktrace.ReadOnlySpan, scores []scoreEntry) []sdktrace.ReadOnlySpan {
	// Build lookup by span name. Last score entry wins for duplicates.
	byName := make(map[string]scoreEntry, len(scores))
	for _, se := range scores {
		byName[se.Name] = se
	}

	stubs := tracetest.SpanStubsFromReadOnlySpans(spans)
	for i, stub := range stubs {
		se, ok := byName[stub.Name]
		if !ok {
			continue
		}
		extra := []attribute.KeyValue{
			attribute.Float64("memory.score", se.Score),
		}
		if se.Reason != "" {
			extra = append(extra, attribute.String("memory.score.reason", se.Reason))
		}
		stubs[i].Attributes = append(stub.Attributes, extra...)
	}
	return stubs.Snapshots()
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

// truncate shortens s to at most n bytes for safe log output.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
