package scorer

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// captureExporter is a test SpanExporter that captures all exported spans.
type captureExporter struct {
	spans []sdktrace.ReadOnlySpan
}

func (c *captureExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	c.spans = append(c.spans, spans...)
	return nil
}

func (c *captureExporter) Shutdown(_ context.Context) error { return nil }

// makeSpans builds a slice of ReadOnlySpan stubs with the given names using tracetest.
func makeSpans(names ...string) []sdktrace.ReadOnlySpan {
	stubs := make(tracetest.SpanStubs, 0, len(names))
	for _, name := range names {
		stubs = append(stubs, tracetest.SpanStub{Name: name})
	}
	return stubs.Snapshots()
}

// makeSpansWithAttr builds stubs that include a specific attribute.
func makeSpansWithAttr(name, key, val string) []sdktrace.ReadOnlySpan {
	stubs := tracetest.SpanStubs{
		tracetest.SpanStub{
			Name: name,
			Attributes: []attribute.KeyValue{
				attribute.String(key, val),
			},
		},
	}
	return stubs.Snapshots()
}

// stubExecSuccess returns a fake exec.CommandContext that produces a JSON scores response.
func stubExecSuccess(scores []scoreEntry) func(ctx context.Context, name string, args ...string) *exec.Cmd {
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		data, _ := json.Marshal(scores)
		// Use echo to produce the JSON output.
		return exec.CommandContext(ctx, "echo", string(data))
	}
}

// stubExecError returns a fake exec.CommandContext whose command exits non-zero.
func stubExecError() func(ctx context.Context, name string, args ...string) *exec.Cmd {
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "false")
	}
}

// stubExecBadJSON returns a fake exec.CommandContext that returns non-JSON output.
func stubExecBadJSON() func(ctx context.Context, name string, args ...string) *exec.Cmd {
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "echo", "not-json")
	}
}

// stubExecTimeout returns a fake exec.CommandContext that sleeps longer than the context allows.
func stubExecTimeout() func(ctx context.Context, name string, args ...string) *exec.Cmd {
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sleep", "60")
	}
}

func TestBufferingExporter_ExportSpansBuffers(t *testing.T) {
	upstream := &captureExporter{}
	buf := NewBufferingExporter(upstream)

	spans := makeSpans("Bash", "Read")
	if err := buf.ExportSpans(context.Background(), spans); err != nil {
		t.Fatalf("ExportSpans: %v", err)
	}
	// Buffered spans should NOT have reached the upstream yet.
	if len(upstream.spans) != 0 {
		t.Errorf("expected 0 spans forwarded to upstream before Score(), got %d", len(upstream.spans))
	}
	if len(buf.spans) != 2 {
		t.Errorf("expected 2 buffered spans, got %d", len(buf.spans))
	}
}

func TestBufferingExporter_ShutdownForwardsToUpstream(t *testing.T) {
	upstream := &captureExporter{}
	buf := NewBufferingExporter(upstream)

	spans := makeSpans("Bash")
	if err := buf.ExportSpans(context.Background(), spans); err != nil {
		t.Fatalf("ExportSpans: %v", err)
	}

	if err := buf.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	// Shutdown should flush buffered spans to the upstream.
	if len(upstream.spans) != 1 {
		t.Errorf("expected 1 span forwarded on Shutdown, got %d", len(upstream.spans))
	}
}

func TestScore_EnrichesSpansWithClaudeOutput(t *testing.T) {
	upstream := &captureExporter{}
	buf := NewBufferingExporter(upstream)

	spans := makeSpans("Bash")
	if err := buf.ExportSpans(context.Background(), spans); err != nil {
		t.Fatalf("ExportSpans: %v", err)
	}

	// Inject a stub that returns a score for "Bash".
	origExec := execScorer
	execScorer = stubExecSuccess([]scoreEntry{{Name: "Bash", Score: "high", Reason: "efficient tool use"}})
	defer func() { execScorer = origExec }()

	if err := buf.Score(context.Background()); err != nil {
		t.Fatalf("Score: %v", err)
	}

	// Score should forward enriched spans to upstream.
	if len(upstream.spans) == 0 {
		t.Fatal("expected spans forwarded to upstream after Score()")
	}

	// Find an enriched span — it will have shipmate.score attribute.
	found := false
	for _, s := range upstream.spans {
		for _, kv := range s.Attributes() {
			if string(kv.Key) == "memory.score" {
				found = true
				if kv.Value.AsString() != "high" {
					t.Errorf("shipmate.score: got %q, want %q", kv.Value.AsString(), "high")
				}
			}
		}
	}
	if !found {
		t.Error("no span with shipmate.score attribute found after Score()")
	}
}

func TestScore_NonFatalOnClaudeError(t *testing.T) {
	upstream := &captureExporter{}
	buf := NewBufferingExporter(upstream)

	spans := makeSpans("Write")
	if err := buf.ExportSpans(context.Background(), spans); err != nil {
		t.Fatalf("ExportSpans: %v", err)
	}

	origExec := execScorer
	execScorer = stubExecError()
	defer func() { execScorer = origExec }()

	// Score must not return an error when claude fails.
	if err := buf.Score(context.Background()); err != nil {
		t.Errorf("Score returned error on claude failure: %v", err)
	}
}

func TestScore_NonFatalOnBadJSON(t *testing.T) {
	upstream := &captureExporter{}
	buf := NewBufferingExporter(upstream)

	spans := makeSpans("Read")
	if err := buf.ExportSpans(context.Background(), spans); err != nil {
		t.Fatalf("ExportSpans: %v", err)
	}

	origExec := execScorer
	execScorer = stubExecBadJSON()
	defer func() { execScorer = origExec }()

	if err := buf.Score(context.Background()); err != nil {
		t.Errorf("Score returned error on bad JSON: %v", err)
	}
}

func TestScore_NonFatalOnTimeout(t *testing.T) {
	upstream := &captureExporter{}
	buf := NewBufferingExporter(upstream)

	spans := makeSpans("Bash")
	if err := buf.ExportSpans(context.Background(), spans); err != nil {
		t.Fatalf("ExportSpans: %v", err)
	}

	origExec := execScorer
	execScorer = stubExecTimeout()
	defer func() { execScorer = origExec }()

	// Use a very short timeout so the test doesn't actually wait 30s.
	origTimeout := scorerTimeout
	scorerTimeout = 50 * time.Millisecond
	defer func() { scorerTimeout = origTimeout }()

	if err := buf.Score(context.Background()); err != nil {
		t.Errorf("Score returned error on timeout: %v", err)
	}
}

func TestScore_NoSpans_IsNoop(t *testing.T) {
	upstream := &captureExporter{}
	buf := NewBufferingExporter(upstream)
	// Do not add any spans.

	origExec := execScorer
	called := false
	execScorer = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		called = true
		return exec.CommandContext(ctx, "echo", "[]")
	}
	defer func() { execScorer = origExec }()

	if err := buf.Score(context.Background()); err != nil {
		t.Fatalf("Score: %v", err)
	}
	if called {
		t.Error("expected claude not to be called when no spans buffered")
	}
}

func TestScore_PromptContainsSpanNames(t *testing.T) {
	upstream := &captureExporter{}
	buf := NewBufferingExporter(upstream)

	spans := makeSpans("Bash", "Read")
	if err := buf.ExportSpans(context.Background(), spans); err != nil {
		t.Fatalf("ExportSpans: %v", err)
	}

	origExec := execScorer
	var capturedArgs []string
	execScorer = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		capturedArgs = args
		return exec.CommandContext(ctx, "echo", "[]")
	}
	defer func() { execScorer = origExec }()

	if err := buf.Score(context.Background()); err != nil {
		t.Fatalf("Score: %v", err)
	}

	// The prompt (last arg) should contain span names.
	if len(capturedArgs) == 0 {
		t.Fatal("no args captured")
	}
	prompt := strings.Join(capturedArgs, " ")
	if !strings.Contains(prompt, "Bash") {
		t.Errorf("prompt does not contain span name %q: %s", "Bash", prompt)
	}
	if !strings.Contains(prompt, "Read") {
		t.Errorf("prompt does not contain span name %q: %s", "Read", prompt)
	}
}

func TestScore_SpanAttributesPresentInPrompt(t *testing.T) {
	upstream := &captureExporter{}
	buf := NewBufferingExporter(upstream)

	spans := makeSpansWithAttr("Bash", "tool.command", "go test ./...")
	if err := buf.ExportSpans(context.Background(), spans); err != nil {
		t.Fatalf("ExportSpans: %v", err)
	}

	origExec := execScorer
	var capturedArgs []string
	execScorer = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		capturedArgs = args
		return exec.CommandContext(ctx, "echo", "[]")
	}
	defer func() { execScorer = origExec }()

	if err := buf.Score(context.Background()); err != nil {
		t.Fatalf("Score: %v", err)
	}

	prompt := strings.Join(capturedArgs, " ")
	if !strings.Contains(prompt, "go test ./...") {
		t.Errorf("prompt does not contain attribute value %q: %s", "go test ./...", prompt)
	}
}

func TestNewBufferingExporter_UpstreamShutdownCalled(t *testing.T) {
	shutdown := false
	upstream := &shutdownTrackingExporter{onShutdown: func() { shutdown = true }}
	buf := NewBufferingExporter(upstream)
	if err := buf.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	if !shutdown {
		t.Error("upstream Shutdown was not called")
	}
}

// shutdownTrackingExporter tracks whether Shutdown was called.
type shutdownTrackingExporter struct {
	onShutdown func()
}

func (s *shutdownTrackingExporter) ExportSpans(_ context.Context, _ []sdktrace.ReadOnlySpan) error {
	return nil
}

func (s *shutdownTrackingExporter) Shutdown(_ context.Context) error {
	s.onShutdown()
	return nil
}

func TestScore_MultipleScoresMatchedByName(t *testing.T) {
	upstream := &captureExporter{}
	buf := NewBufferingExporter(upstream)

	spans := makeSpans("Bash", "Read", "Write")
	if err := buf.ExportSpans(context.Background(), spans); err != nil {
		t.Fatalf("ExportSpans: %v", err)
	}

	origExec := execScorer
	execScorer = stubExecSuccess([]scoreEntry{
		{Name: "Bash", Score: "high", Reason: "fast"},
		{Name: "Read", Score: "low", Reason: "slow"},
	})
	defer func() { execScorer = origExec }()

	if err := buf.Score(context.Background()); err != nil {
		t.Fatalf("Score: %v", err)
	}

	scoresByName := map[string]string{}
	for _, s := range upstream.spans {
		for _, kv := range s.Attributes() {
			if string(kv.Key) == "memory.score" {
				scoresByName[s.Name()] = kv.Value.AsString()
			}
		}
	}

	if scoresByName["Bash"] != "high" {
		t.Errorf("Bash score: got %q, want %q", scoresByName["Bash"], "high")
	}
	if scoresByName["Read"] != "low" {
		t.Errorf("Read score: got %q, want %q", scoresByName["Read"], "low")
	}
	// Write had no matching score entry — it should still be exported (unenriched).
	found := false
	for _, s := range upstream.spans {
		if s.Name() == "Write" {
			found = true
		}
	}
	if !found {
		t.Error("Write span not found in upstream export after Score()")
	}
}

// Compile-time assertion: BufferingExporter implements SpanExporter.
var _ sdktrace.SpanExporter = (*BufferingExporter)(nil)

// Verify scoreEntry is an exported type (used in tests via package-internal access).
var _ = fmt.Sprintf("%T", scoreEntry{})
