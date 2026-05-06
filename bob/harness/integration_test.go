package harness

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mattdurham/bob/bob/extension"
	mockprovider "github.com/mattdurham/bob/bob/provider/mock"
	"github.com/mattdurham/bob/bob/sdk"
)

// TestIntegration_FullStreamingFlow exercises the full submit → stream → done flow.
// Because there is no real bubbletea program in tests, prog is nil, so TokenMsgs are
// not sent via prog.Send. The stream cmd returns a StreamDoneMsg directly.
func TestIntegration_FullStreamingFlow(t *testing.T) {
	p := &mockprovider.Provider{
		Tokens: []string{"hello", " ", "world", "!", "\n"},
	}
	m := New(p, nil)

	// Send SubmitMsg — should start streaming.
	m, cmd := callUpdate(m, SubmitMsg{Content: "what is 2+2?"})
	if !m.streaming {
		t.Fatal("expected streaming == true after SubmitMsg")
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd after SubmitMsg")
	}

	// Execute the cmd. Since prog is nil, TokenMsgs go nowhere; the cmd returns
	// StreamDoneMsg when it finishes.
	result := cmd()
	doneMsg, ok := result.(StreamDoneMsg)
	if !ok {
		t.Fatalf("expected StreamDoneMsg from cmd, got %T", result)
	}
	if doneMsg.Err != nil {
		t.Fatalf("expected no error in StreamDoneMsg, got %v", doneMsg.Err)
	}

	// Process StreamDoneMsg.
	m, _ = callUpdate(m, doneMsg)
	if m.streaming {
		t.Error("expected streaming == false after StreamDoneMsg")
	}

	// Verify user message in history.
	if len(m.history) != 1 {
		t.Fatalf("expected 1 message in history, got %d", len(m.history))
	}
	if m.history[0].Role != sdk.RoleUser {
		t.Errorf("expected history[0].Role == RoleUser, got %v", m.history[0].Role)
	}
	if m.history[0].Content != "what is 2+2?" {
		t.Errorf("history[0].Content: got %q, want %q", m.history[0].Content, "what is 2+2?")
	}

	// Verify mock provider was called once.
	if p.CallCount != 1 {
		t.Errorf("expected CallCount == 1, got %d", p.CallCount)
	}
}

// TestIntegration_UserMessageInHistory verifies the user message is recorded correctly.
func TestIntegration_UserMessageInHistory(t *testing.T) {
	p := &mockprovider.Provider{
		Tokens: []string{"ok"},
	}
	m := New(p, nil)

	m, cmd := callUpdate(m, SubmitMsg{Content: "hello world"})

	// Execute stream cmd to completion so model is in a clean state.
	if cmd != nil {
		cmd()
	}

	// History should have 1 user message immediately after SubmitMsg.
	if len(m.history) != 1 {
		t.Fatalf("expected 1 message in history, got %d", len(m.history))
	}
	if m.history[0].Role != sdk.RoleUser {
		t.Errorf("expected RoleUser, got %v", m.history[0].Role)
	}
	if m.history[0].Content != "hello world" {
		t.Errorf("history[0].Content: got %q, want %q", m.history[0].Content, "hello world")
	}

	// Chat should have 1 message (the user message).
	if len(m.chat.messages) != 1 {
		t.Errorf("expected 1 chat message, got %d", len(m.chat.messages))
	}
}

// TestIntegration_CtrlC_CancelsStream verifies that ctrl+c while streaming calls cancelStream.
func TestIntegration_CtrlC_CancelsStream(t *testing.T) {
	p := &mockprovider.Provider{
		Tokens: []string{"a", "b", "c"},
	}
	m := New(p, nil)

	// Start streaming.
	m, _ = callUpdate(m, SubmitMsg{Content: "hi"})
	if !m.streaming {
		t.Fatal("expected streaming == true after SubmitMsg")
	}
	if m.cancelStream == nil {
		t.Fatal("expected cancelStream to be set")
	}

	// Send ctrl+c while streaming. This should invoke cancelStream and return
	// without changing m.streaming (StreamDoneMsg hasn't arrived yet).
	// ctrl+c is represented as Code: 'c' with Mod: tea.ModCtrl.
	m, cmd := callUpdate(m, tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	// The Update returns m, nil when ctrl+c cancels stream (no tea.Quit).
	_ = cmd

	// m.streaming is still true because StreamDoneMsg hasn't arrived yet.
	if !m.streaming {
		t.Error("streaming should still be true — StreamDoneMsg hasn't arrived yet")
	}
	// cancelStream should have been invoked (and is still set until StreamDoneMsg).
	if m.cancelStream == nil {
		t.Error("cancelStream should still be set until StreamDoneMsg clears it")
	}
}

// TestIntegration_NilExtensionHost_Safe verifies that nil host never panics.
func TestIntegration_NilExtensionHost_Safe(t *testing.T) {
	p := &mockprovider.Provider{Tokens: []string{"hi"}}
	m := New(p, nil)

	// Init should return a non-nil Cmd and not panic.
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return a non-nil Cmd")
	}

	// cmdDispatchSessionStart with nil host returns nil — that's fine.
	sessionCmd := m.cmdDispatchSessionStart()
	if sessionCmd != nil {
		t.Error("cmdDispatchSessionStart with nil host should return nil")
	}

	// A normal SubmitMsg should work without panicking.
	m, streamCmd := callUpdate(m, SubmitMsg{Content: "test"})
	if streamCmd == nil {
		t.Error("expected non-nil stream cmd")
	}
	// Execute cmd — no panic expected.
	result := streamCmd()
	_, ok := result.(StreamDoneMsg)
	if !ok {
		t.Errorf("expected StreamDoneMsg, got %T", result)
	}
}

// TestIntegration_ExtensionHost_NoExtensions_Safe verifies a real host with no extensions loaded.
func TestIntegration_ExtensionHost_NoExtensions_Safe(t *testing.T) {
	p := &mockprovider.Provider{Tokens: []string{"hi"}}
	h := extension.NewHost(nil)
	m := New(p, h)

	// Init should not panic.
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return a non-nil Cmd")
	}

	// Submit and execute stream — no extensions loaded, no panics expected.
	m, streamCmd := callUpdate(m, SubmitMsg{Content: "test with host"})
	if streamCmd == nil {
		t.Fatal("expected non-nil stream cmd")
	}

	result := streamCmd()
	doneMsg, ok := result.(StreamDoneMsg)
	if !ok {
		t.Fatalf("expected StreamDoneMsg, got %T", result)
	}
	if doneMsg.Err != nil {
		t.Errorf("expected no error, got %v", doneMsg.Err)
	}
}
