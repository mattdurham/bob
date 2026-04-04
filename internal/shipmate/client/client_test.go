package client

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mattdurham/bob/internal/shipmate/hook"
)

// startEchoServer starts a Unix socket listener that reads one JSON object
// per connection and sends it to the returned channel.
func startEchoServer(t *testing.T) (string, <-chan hook.Command) {
	t.Helper()
	sockPath := filepath.Join(t.TempDir(), "test.sock")
	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	ch := make(chan hook.Command, 8)
	t.Cleanup(func() {
		_ = ln.Close()
		_ = os.Remove(sockPath)
	})
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer func() { _ = c.Close() }()
				var cmd hook.Command
				if err := json.NewDecoder(c).Decode(&cmd); err == nil {
					ch <- cmd
				}
			}(conn)
		}
	}()
	return sockPath, ch
}

func TestSend_DeliversRecordCommand(t *testing.T) {
	sockPath, ch := startEchoServer(t)

	cmd := hook.Command{
		Type:      "record",
		SessionID: "ses-1",
		HookEvent: "Bash",
		Attrs:     map[string]string{"tool.command": "go build"},
	}
	if err := Send(sockPath, cmd); err != nil {
		t.Fatalf("Send: %v", err)
	}

	select {
	case got := <-ch:
		if got.Type != "record" {
			t.Errorf("Type: got %q, want %q", got.Type, "record")
		}
		if got.SessionID != "ses-1" {
			t.Errorf("SessionID: got %q, want %q", got.SessionID, "ses-1")
		}
		if got.Attrs["tool.command"] != "go build" {
			t.Errorf("tool.command: got %q, want %q", got.Attrs["tool.command"], "go build")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for command")
	}
}

func TestSend_DeliversStopCommand(t *testing.T) {
	sockPath, ch := startEchoServer(t)

	cmd := hook.Command{Type: "stop", SessionID: "ses-2"}
	if err := Send(sockPath, cmd); err != nil {
		t.Fatalf("Send: %v", err)
	}

	select {
	case got := <-ch:
		if got.Type != "stop" {
			t.Errorf("Type: got %q, want %q", got.Type, "stop")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

func TestSend_DeliversMemoryCommand(t *testing.T) {
	sockPath, ch := startEchoServer(t)

	cmd := hook.Command{Type: "memory", SessionID: "ses-3", Text: "important insight"}
	if err := Send(sockPath, cmd); err != nil {
		t.Fatalf("Send: %v", err)
	}

	select {
	case got := <-ch:
		if got.Type != "memory" {
			t.Errorf("Type: got %q, want %q", got.Type, "memory")
		}
		if got.Text != "important insight" {
			t.Errorf("Text: got %q, want %q", got.Text, "important insight")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

func TestSend_ReturnsNilOnMissingSocket(t *testing.T) {
	// Socket doesn't exist — Send should exhaust retries and return nil (not error).
	// Override retry params for test speed.
	origMax := maxRetries
	origSleep := retrySleep
	t.Cleanup(func() {
		maxRetries = origMax
		retrySleep = origSleep
	})
	maxRetries = 2
	retrySleep = 1 * time.Millisecond

	sockPath := filepath.Join(t.TempDir(), "missing.sock")
	cmd := hook.Command{Type: "record", SessionID: "ses-x"}
	err := Send(sockPath, cmd)
	// Must return nil — hooks must exit 0 always.
	if err != nil {
		t.Errorf("Send returned non-nil error on missing socket: %v", err)
	}
}
