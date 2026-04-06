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
	"github.com/mattdurham/bob/internal/shipmate/transcript"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// noopExporter is a SpanExporter that discards all spans.
type noopExporter struct{}

func (n *noopExporter) ExportSpans(_ context.Context, _ []sdktrace.ReadOnlySpan) error {
	return nil
}
func (n *noopExporter) Shutdown(_ context.Context) error { return nil }

// newNoopTurnExporter creates a TurnExporter backed by a no-op exporter for tests.
func newNoopTurnExporter(t *testing.T) *transcript.TurnExporter {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "transcript.jsonl")
	// Create empty transcript file.
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatalf("write empty transcript: %v", err)
	}
	return transcript.NewTurnExporter(path, "test-session", &noopExporter{})
}

// shortSockPath creates a short temporary socket path within os.TempDir() to
// avoid the 104-byte Unix socket path limit on macOS.
func shortSockPath(t *testing.T, name string) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "sm-")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return filepath.Join(dir, name)
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

// waitSocket polls until the socket file appears or a 2-second deadline passes.
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

func TestServe_StopCommandCausesReturn(t *testing.T) {
	te := newNoopTurnExporter(t)
	sockPath := shortSockPath(t, "t.sock")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() { errCh <- Serve(ctx, sockPath, te) }()
	waitSocket(t, sockPath)

	sendCmd(t, sockPath, hook.Command{Type: "stop", SessionID: "ses-d1"})

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Serve returned non-nil error on stop: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Serve did not return within 3s after stop")
	}
}

func TestServe_StaleSocketRemovedOnStart(t *testing.T) {
	te := newNoopTurnExporter(t)
	sockPath := shortSockPath(t, "stale.sock")

	if err := os.WriteFile(sockPath, []byte("stale"), 0o600); err != nil {
		t.Fatalf("create stale file: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		Serve(ctx, sockPath, te) //nolint:errcheck
	}()

	waitSocket(t, sockPath)
	cancel()
	<-done
}

func TestServe_CtxCancelCausesReturn(t *testing.T) {
	te := newNoopTurnExporter(t)
	sockPath := shortSockPath(t, "t.sock")

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() { errCh <- Serve(ctx, sockPath, te) }()
	waitSocket(t, sockPath)

	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Serve returned non-nil error on ctx cancel: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Serve did not return within 3s after ctx cancel")
	}
}
