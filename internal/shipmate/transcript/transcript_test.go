package transcript

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// captureExporter collects exported spans.
type captureExporter struct {
	spans []sdktrace.ReadOnlySpan
}

func (c *captureExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	c.spans = append(c.spans, spans...)
	return nil
}
func (c *captureExporter) Shutdown(_ context.Context) error { return nil }

// writeJSONL writes entries to a temp file and returns the path.
func writeJSONL(t *testing.T, entries []map[string]any) string {
	t.Helper()
	dir := t.TempDir()
	f, err := os.CreateTemp(dir, "transcript-*.jsonl")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	enc := json.NewEncoder(f)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}
	_ = f.Close()
	return f.Name()
}

// makeTurnEntries builds user+assistant JSONL entries for one conversation turn.
func makeTurnEntries(question, answer string, start, end time.Time) []map[string]any {
	return []map[string]any{
		{
			"type":      "user",
			"uuid":      "uuid-q",
			"timestamp": start.Format(time.RFC3339),
			"message": map[string]any{
				"content": question,
			},
		},
		{
			"type":      "assistant",
			"uuid":      "uuid-a",
			"timestamp": end.Format(time.RFC3339),
			"message": map[string]any{
				"content": []map[string]any{
					{"type": "text", "text": answer},
				},
			},
		},
	}
}

func stubExecSuccess(scores []scoreEntry) func(context.Context, string, ...string) *exec.Cmd {
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		data, _ := json.Marshal(scores)
		return exec.CommandContext(ctx, "echo", string(data))
	}
}

func stubExecError() func(context.Context, string, ...string) *exec.Cmd {
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "false")
	}
}

func stubExecBadJSON() func(context.Context, string, ...string) *exec.Cmd {
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "echo", "not-json")
	}
}

// ---- Parse tests ----

func TestParse_MissingFile(t *testing.T) {
	_, err := Parse(filepath.Join(t.TempDir(), "nonexistent.jsonl"))
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestParse_EmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.jsonl")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	turns, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse empty: %v", err)
	}
	if len(turns) != 0 {
		t.Errorf("expected 0 turns, got %d", len(turns))
	}
}

func TestParse_SingleTurn(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	p := writeJSONL(t, makeTurnEntries("hello world", "hi there", now, now.Add(2*time.Second)))
	turns, err := Parse(p)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected 1 turn, got %d", len(turns))
	}
	if turns[0].Question != "hello world" {
		t.Errorf("question: got %q, want %q", turns[0].Question, "hello world")
	}
	if turns[0].Answer != "hi there" {
		t.Errorf("answer: got %q, want %q", turns[0].Answer, "hi there")
	}
}

func TestParse_MultipleTurns(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	var entries []map[string]any
	entries = append(entries, makeTurnEntries("q1", "a1", now, now.Add(time.Second))...)
	entries = append(entries, makeTurnEntries("q2", "a2", now.Add(2*time.Second), now.Add(3*time.Second))...)
	turns, err := Parse(writeJSONL(t, entries))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(turns) != 2 {
		t.Fatalf("expected 2 turns, got %d", len(turns))
	}
}

func TestParse_SkipsMalformedLines(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	p := writeJSONL(t, makeTurnEntries("good question", "good answer", now, now.Add(time.Second)))
	f, _ := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0o644)
	_, _ = f.WriteString("not json\n")
	_ = f.Close()

	turns, err := Parse(p)
	if err != nil {
		t.Fatalf("Parse should not fail on malformed lines: %v", err)
	}
	if len(turns) != 1 {
		t.Errorf("expected 1 turn (malformed skipped), got %d", len(turns))
	}
}

func TestParse_SkipsToolResultUserMessages(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	entries := []map[string]any{
		{
			"type": "user", "uuid": "u1", "timestamp": now.Format(time.RFC3339),
			"message": map[string]any{
				"content": []map[string]any{{"type": "tool_result", "content": "ok"}},
			},
		},
		{
			"type": "user", "uuid": "u2", "timestamp": now.Add(time.Second).Format(time.RFC3339),
			"message": map[string]any{"content": "a real question"},
		},
		{
			"type": "assistant", "uuid": "a1", "timestamp": now.Add(2 * time.Second).Format(time.RFC3339),
			"message": map[string]any{
				"content": []map[string]any{{"type": "text", "text": "the answer"}},
			},
		},
	}
	turns, err := Parse(writeJSONL(t, entries))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected 1 turn (tool result skipped), got %d", len(turns))
	}
	if turns[0].Question != "a real question" {
		t.Errorf("question: got %q", turns[0].Question)
	}
}

// ---- ExportNew tests ----

func TestExportNew_EmptyTranscript(t *testing.T) {
	p := filepath.Join(t.TempDir(), "empty.jsonl")
	_ = os.WriteFile(p, nil, 0o644)
	cap := &captureExporter{}
	te := NewTurnExporter(p, "ses-1", cap)
	if err := te.ExportNew(context.Background()); err != nil {
		t.Fatalf("ExportNew empty: %v", err)
	}
	if len(cap.spans) != 0 {
		t.Errorf("expected 0 spans, got %d", len(cap.spans))
	}
}

func TestExportNew_MissingPath_IsNoop(t *testing.T) {
	cap := &captureExporter{}
	te := NewTurnExporter(filepath.Join(t.TempDir(), "missing.jsonl"), "ses-m", cap)
	if err := te.ExportNew(context.Background()); err != nil {
		t.Errorf("ExportNew missing path should not error: %v", err)
	}
}

func TestExportNew_ExportsSessionAndTurnSpans(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	p := writeJSONL(t, makeTurnEntries("what is Go?", "a language", now, now.Add(2*time.Second)))
	cap := &captureExporter{}
	te := NewTurnExporter(p, "ses-2", cap)
	if err := te.ExportNew(context.Background()); err != nil {
		t.Fatalf("ExportNew: %v", err)
	}
	// 1 session + 1 turn = 2 spans.
	if len(cap.spans) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(cap.spans))
	}
	names := spanNames(cap.spans)
	if !names["session"] {
		t.Error("expected 'session' span")
	}
	if !names["what is Go?"] {
		t.Errorf("expected turn span 'what is Go?', got %v", names)
	}
}

func TestExportNew_StableTraceID(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	p := writeJSONL(t, makeTurnEntries("t1", "a1", now, now.Add(time.Second)))
	cap := &captureExporter{}
	te := NewTurnExporter(p, "ses-3", cap)

	if err := te.ExportNew(context.Background()); err != nil {
		t.Fatalf("ExportNew 1: %v", err)
	}
	// Append second turn.
	f, _ := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0o644)
	enc := json.NewEncoder(f)
	for _, e := range makeTurnEntries("t2", "a2", now.Add(2*time.Second), now.Add(3*time.Second)) {
		_ = enc.Encode(e)
	}
	_ = f.Close()

	if err := te.ExportNew(context.Background()); err != nil {
		t.Fatalf("ExportNew 2: %v", err)
	}

	if len(cap.spans) < 2 {
		t.Fatalf("need at least 2 spans, got %d", len(cap.spans))
	}
	traceID := cap.spans[0].SpanContext().TraceID()
	for i, s := range cap.spans {
		if s.SpanContext().TraceID() != traceID {
			t.Errorf("span[%d] TraceID differs from span[0]", i)
		}
	}
}

func TestExportNew_IncrementalExport(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	p := writeJSONL(t, makeTurnEntries("first turn", "first answer", now, now.Add(time.Second)))
	cap := &captureExporter{}
	te := NewTurnExporter(p, "ses-4", cap)

	if err := te.ExportNew(context.Background()); err != nil {
		t.Fatalf("ExportNew 1: %v", err)
	}
	afterFirst := len(cap.spans) // session + 1 turn = 2

	// Append second turn.
	f, _ := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0o644)
	enc := json.NewEncoder(f)
	for _, e := range makeTurnEntries("second turn", "second answer", now.Add(2*time.Second), now.Add(3*time.Second)) {
		_ = enc.Encode(e)
	}
	_ = f.Close()

	if err := te.ExportNew(context.Background()); err != nil {
		t.Fatalf("ExportNew 2: %v", err)
	}
	newSpans := len(cap.spans) - afterFirst
	// Should only emit session + 1 new turn (not the already-exported first turn).
	if newSpans != 2 {
		t.Errorf("second export: expected 2 new spans (session + new turn), got %d", newSpans)
	}
}

func TestExportNew_NoNewTurns_IsNoop(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	p := writeJSONL(t, makeTurnEntries("only turn", "only answer", now, now.Add(time.Second)))
	cap := &captureExporter{}
	te := NewTurnExporter(p, "ses-5", cap)

	if err := te.ExportNew(context.Background()); err != nil {
		t.Fatalf("ExportNew 1: %v", err)
	}
	afterFirst := len(cap.spans)
	if err := te.ExportNew(context.Background()); err != nil {
		t.Fatalf("ExportNew 2: %v", err)
	}
	if len(cap.spans) != afterFirst {
		t.Errorf("second call with no new turns added %d spans", len(cap.spans)-afterFirst)
	}
}

func TestExportNew_TurnSpanIsChildOfSession(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	p := writeJSONL(t, makeTurnEntries("parent test question", "answer", now, now.Add(time.Second)))
	cap := &captureExporter{}
	te := NewTurnExporter(p, "ses-6", cap)
	if err := te.ExportNew(context.Background()); err != nil {
		t.Fatalf("ExportNew: %v", err)
	}

	var sessionSpan, turnSpan sdktrace.ReadOnlySpan
	for _, s := range cap.spans {
		if s.Name() == "session" {
			sessionSpan = s
		} else {
			turnSpan = s
		}
	}
	if sessionSpan == nil || turnSpan == nil {
		t.Fatal("expected session and turn spans")
	}
	if turnSpan.Parent().SpanID() != sessionSpan.SpanContext().SpanID() {
		t.Errorf("turn parent SpanID %v != session SpanID %v",
			turnSpan.Parent().SpanID(), sessionSpan.SpanContext().SpanID())
	}
}

func TestExportNew_TurnNameTruncatedAt128(t *testing.T) {
	question := ""
	for i := 0; i < 200; i++ {
		question += "x"
	}
	now := time.Now().UTC().Truncate(time.Second)
	p := writeJSONL(t, makeTurnEntries(question, "answer", now, now.Add(time.Second)))
	cap := &captureExporter{}
	te := NewTurnExporter(p, "ses-7", cap)
	if err := te.ExportNew(context.Background()); err != nil {
		t.Fatalf("ExportNew: %v", err)
	}
	for _, s := range cap.spans {
		if s.Name() != "session" && len([]rune(s.Name())) > 128 {
			t.Errorf("turn span name length %d > 128", len([]rune(s.Name())))
		}
	}
}

// ---- ExportAndScore tests ----

func TestExportAndScore_EmptyTranscript(t *testing.T) {
	p := filepath.Join(t.TempDir(), "empty.jsonl")
	_ = os.WriteFile(p, nil, 0o644)
	cap := &captureExporter{}
	te := NewTurnExporter(p, "ses-es1", cap)
	if err := te.ExportAndScore(context.Background()); err != nil {
		t.Fatalf("ExportAndScore empty: %v", err)
	}
	if len(cap.spans) != 0 {
		t.Errorf("expected 0 spans, got %d", len(cap.spans))
	}
}

func TestExportAndScore_ExportsWithScores(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	p := writeJSONL(t, makeTurnEntries("scored question", "scored answer", now, now.Add(time.Second)))

	origExec := execScorer
	execScorer = stubExecSuccess([]scoreEntry{{Name: "scored question", Score: 0.8, Reason: "good"}})
	defer func() { execScorer = origExec }()

	cap := &captureExporter{}
	te := NewTurnExporter(p, "ses-es2", cap)
	if err := te.ExportAndScore(context.Background()); err != nil {
		t.Fatalf("ExportAndScore: %v", err)
	}

	found := false
	for _, s := range cap.spans {
		for _, kv := range s.Attributes() {
			if string(kv.Key) == "memory.score" {
				found = true
				if kv.Value.AsFloat64() != 0.8 {
					t.Errorf("memory.score: got %v, want 0.8", kv.Value.AsFloat64())
				}
			}
		}
	}
	if !found {
		t.Error("no memory.score attribute found")
	}
}

func TestExportAndScore_NonFatalOnClaudeError(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	p := writeJSONL(t, makeTurnEntries("q", "a", now, now.Add(time.Second)))

	origExec := execScorer
	execScorer = stubExecError()
	defer func() { execScorer = origExec }()

	cap := &captureExporter{}
	te := NewTurnExporter(p, "ses-es3", cap)
	if err := te.ExportAndScore(context.Background()); err != nil {
		t.Errorf("ExportAndScore should not error on claude failure: %v", err)
	}
	if len(cap.spans) == 0 {
		t.Error("spans should still be exported when claude fails")
	}
}

func TestExportAndScore_NonFatalOnBadJSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	p := writeJSONL(t, makeTurnEntries("q", "a", now, now.Add(time.Second)))

	origExec := execScorer
	execScorer = stubExecBadJSON()
	defer func() { execScorer = origExec }()

	cap := &captureExporter{}
	te := NewTurnExporter(p, "ses-es4", cap)
	if err := te.ExportAndScore(context.Background()); err != nil {
		t.Errorf("ExportAndScore should not error on bad JSON: %v", err)
	}
}

func TestExportAndScore_ReExportsAllTurns(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	p := writeJSONL(t, makeTurnEntries("first", "answer1", now, now.Add(time.Second)))

	origExec := execScorer
	execScorer = stubExecSuccess([]scoreEntry{})
	defer func() { execScorer = origExec }()

	cap := &captureExporter{}
	te := NewTurnExporter(p, "ses-es5", cap)

	// Simulate ExportNew already advancing lastExported.
	if err := te.ExportNew(context.Background()); err != nil {
		t.Fatalf("ExportNew: %v", err)
	}
	afterNew := len(cap.spans)

	if err := te.ExportAndScore(context.Background()); err != nil {
		t.Fatalf("ExportAndScore: %v", err)
	}
	// ExportAndScore must re-emit session + all turns regardless of lastExported.
	added := len(cap.spans) - afterNew
	if added < 2 {
		t.Errorf("ExportAndScore should re-export all turns, got %d new spans", added)
	}
}

// ---- Helpers ----

func spanNames(spans []sdktrace.ReadOnlySpan) map[string]bool {
	m := make(map[string]bool, len(spans))
	for _, s := range spans {
		m[s.Name()] = true
	}
	return m
}
