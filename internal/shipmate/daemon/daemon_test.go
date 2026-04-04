package daemon

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mattdurham/bob/internal/shipmate/hook"
	"github.com/mattdurham/bob/internal/shipmate/recorder"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func newTestRecorder(t *testing.T) (*recorder.Recorder, *tracetest.InMemoryExporter) {
	t.Helper()
	exp := tracetest.NewInMemoryExporter()
	rec, err := recorder.New(exp)
	if err != nil {
		t.Fatalf("recorder.New: %v", err)
	}
	t.Cleanup(func() { rec.Shutdown(context.Background()) }) //nolint:errcheck
	return rec, exp
}

// sendCmd connects to sockPath and sends cmd as NDJSON.
func sendCmd(t *testing.T, sockPath string, cmd hook.Command) {
	t.Helper()
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		t.Fatalf("dial %s: %v", sockPath, err)
	}
	defer func() { _ = conn.Close() }()
	if err := json.NewEncoder(conn).Encode(cmd); err != nil {
		t.Fatalf("encode cmd: %v", err)
	}
}

// waitSocket polls until the socket file exists or a 2-second deadline passes.
func waitSocket(t *testing.T, sockPath string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(sockPath); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("socket %s did not appear within 2s", sockPath)
}

// pollSpans calls rec.ForceFlush in a loop until exp has at least want spans
// or a 2-second deadline passes. This replaces time.Sleep synchronization.
func pollSpans(t *testing.T, rec *recorder.Recorder, exp *tracetest.InMemoryExporter, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if err := rec.ForceFlush(context.Background()); err == nil {
			if len(exp.GetSpans()) >= want {
				return
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	// Final flush attempt; let the caller assert on the count.
	rec.ForceFlush(context.Background()) //nolint:errcheck
}

func TestServe_RecordCommandCreatesSpan(t *testing.T) {
	rec, exp := newTestRecorder(t)
	sockPath := filepath.Join(t.TempDir(), "test.sock")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() { Serve(ctx, sockPath, rec, nil) }() //nolint:errcheck
	waitSocket(t, sockPath)

	sendCmd(t, sockPath, hook.Command{
		Type:      "record",
		SessionID: "ses-d1",
		HookEvent: "Bash",
		Attrs:     map[string]string{"tool.command": "ls"},
	})

	// Poll until the span is visible or a 2-second deadline passes.
	pollSpans(t, rec, exp, 1)
	cancel()

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Name != "Bash" {
		t.Errorf("span name: got %q, want %q", spans[0].Name, "Bash")
	}
	attrs := spanAttrsMap(spans[0])
	if attrs["session.id"] != "ses-d1" {
		t.Errorf("session.id: got %q, want %q", attrs["session.id"], "ses-d1")
	}
	if attrs["tool.command"] != "ls" {
		t.Errorf("tool.command: got %q, want %q", attrs["tool.command"], "ls")
	}
}

func TestServe_MemoryCommandCreatesSpan(t *testing.T) {
	rec, exp := newTestRecorder(t)
	sockPath := filepath.Join(t.TempDir(), "test.sock")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() { Serve(ctx, sockPath, rec, nil) }() //nolint:errcheck
	waitSocket(t, sockPath)

	sendCmd(t, sockPath, hook.Command{
		Type:      "memory",
		SessionID: "ses-d2",
		Text:      "an important observation",
	})

	// Poll until the span is visible or a 2-second deadline passes.
	pollSpans(t, rec, exp, 1)
	cancel()

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Name != "memory" {
		t.Errorf("span name: got %q, want %q", spans[0].Name, "memory")
	}
	attrs := spanAttrsMap(spans[0])
	if attrs["shipmate.text"] != "an important observation" {
		t.Errorf("shipmate.text: got %q, want %q", attrs["shipmate.text"], "an important observation")
	}
}

func TestServe_StopCommandCausesReturn(t *testing.T) {
	rec, _ := newTestRecorder(t)
	sockPath := filepath.Join(t.TempDir(), "test.sock")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() { errCh <- Serve(ctx, sockPath, rec, nil) }()
	waitSocket(t, sockPath)

	sendCmd(t, sockPath, hook.Command{Type: "stop", SessionID: "ses-d3"})

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Serve returned non-nil error on stop: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Serve did not return within 3s after stop command")
	}
}

func TestServe_ConcurrentRecords(t *testing.T) {
	rec, exp := newTestRecorder(t)
	sockPath := filepath.Join(t.TempDir(), "test.sock")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() { Serve(ctx, sockPath, rec, nil) }() //nolint:errcheck
	waitSocket(t, sockPath)

	const n = 5
	for i := 0; i < n; i++ {
		sendCmd(t, sockPath, hook.Command{
			Type:      "record",
			SessionID: "ses-conc",
			HookEvent: "Write",
		})
	}

	// Poll until all expected spans are visible or a 2-second deadline passes.
	pollSpans(t, rec, exp, n)
	cancel()

	spans := exp.GetSpans()
	if len(spans) != n {
		t.Errorf("expected %d spans, got %d", n, len(spans))
	}
}

func TestServe_StaleSocketRemovedOnStart(t *testing.T) {
	rec, _ := newTestRecorder(t)
	sockPath := filepath.Join(t.TempDir(), "stale.sock")

	// Place a stale file at the socket path to simulate a crashed daemon.
	if err := os.WriteFile(sockPath, []byte("stale"), 0o600); err != nil {
		t.Fatalf("create stale file: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		Serve(ctx, sockPath, rec, nil) //nolint:errcheck
	}()

	// waitSocket succeeding means the stale file was removed and the socket bound.
	waitSocket(t, sockPath)
	cancel()
	// Wait for Serve to fully exit before TempDir cleanup removes the directory.
	<-done
}

func spanAttrsMap(span tracetest.SpanStub) map[string]string {
	m := make(map[string]string, len(span.Attributes))
	for _, kv := range span.Attributes {
		m[string(kv.Key)] = kv.Value.AsString()
	}
	return m
}

// Compile-time check: InMemoryExporter satisfies SpanExporter.
var _ sdktrace.SpanExporter = (*tracetest.InMemoryExporter)(nil)
