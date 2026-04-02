package server

import (
	"context"
	"strings"
	"testing"

	"github.com/mattdurham/bob/internal/shipmate/recorder"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// mockRecorder captures calls to Record for assertions.
type mockRecorder struct {
	calls []recorder.RecordArgs
	err   error
}

func (m *mockRecorder) Record(ctx context.Context, args recorder.RecordArgs) error {
	m.calls = append(m.calls, args)
	return m.err
}

// staticSessionProvider returns a fixed session ID.
type staticSessionProvider struct {
	id string
}

func (s *staticSessionProvider) SessionID() string {
	return s.id
}

func TestShipmateRecordToolValidation(t *testing.T) {
	mock := &mockRecorder{}
	srv := New(mock, &staticSessionProvider{})

	result, _, err := srv.record(context.Background(), "", "agent-1", "text", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}
	text := contentText(result.Content[0])
	if !strings.Contains(text, "ERROR:") {
		t.Errorf("expected ERROR: in result, got %q", text)
	}
	if !result.IsError {
		t.Errorf("expected IsError=true for error result")
	}
	if len(mock.calls) != 0 {
		t.Errorf("recorder should not be called when name is empty")
	}
}

func TestShipmateRecordToolAgentValidation(t *testing.T) {
	mock := &mockRecorder{}
	srv := New(mock, &staticSessionProvider{})

	result, _, err := srv.record(context.Background(), "my-span", "", "text", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := contentText(result.Content[0])
	if !strings.Contains(text, "ERROR:") {
		t.Errorf("expected ERROR: in result, got %q", text)
	}
	if !result.IsError {
		t.Errorf("expected IsError=true for error result")
	}
	if len(mock.calls) != 0 {
		t.Errorf("recorder should not be called when agent is empty")
	}
}

func TestShipmateRecordToolCallsRecorder(t *testing.T) {
	mock := &mockRecorder{}
	session := &staticSessionProvider{id: "ses-123"}
	srv := New(mock, session)

	result, _, err := srv.record(context.Background(), "my-span", "coder-2", "doing work", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 recorder call, got %d", len(mock.calls))
	}
	call := mock.calls[0]
	if call.Name != "my-span" {
		t.Errorf("Name: got %q, want %q", call.Name, "my-span")
	}
	if call.Agent != "coder-2" {
		t.Errorf("Agent: got %q, want %q", call.Agent, "coder-2")
	}
	if call.Text != "doing work" {
		t.Errorf("Text: got %q, want %q", call.Text, "doing work")
	}
	if call.SessionID != "ses-123" {
		t.Errorf("SessionID: got %q, want %q", call.SessionID, "ses-123")
	}

	text := contentText(result.Content[0])
	if strings.Contains(text, "ERROR:") {
		t.Errorf("unexpected error in result: %q", text)
	}
}

func TestShipmateRecordToolAttributeForwarding(t *testing.T) {
	mock := &mockRecorder{}
	srv := New(mock, &staticSessionProvider{})

	attrs := map[string]string{
		"repo": "bob",
		"task": "1",
	}
	_, _, err := srv.record(context.Background(), "attr-span", "coder-2", "text", attrs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.calls) != 1 {
		t.Fatalf("expected 1 recorder call, got %d", len(mock.calls))
	}
	for k, want := range attrs {
		got := mock.calls[0].Attributes[k]
		if got != want {
			t.Errorf("attribute %s: got %q, want %q", k, got, want)
		}
	}
}

// contentText extracts the text from a mcp.Content value.
func contentText(c mcp.Content) string {
	if tc, ok := c.(*mcp.TextContent); ok {
		return tc.Text
	}
	return ""
}
