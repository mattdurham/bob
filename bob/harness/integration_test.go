package harness

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mattdurham/bob/bob/extension"
	"github.com/mattdurham/bob/bob/sdk"
)

// runStreamCmd executes a cmd returned from SubmitMsg and returns the StreamDoneMsg.
// startStream now returns a tea.Batch (stream cmd + tick cmd), so we unwrap BatchMsg
// and find the sub-cmd that produces StreamDoneMsg.
func runStreamCmd(cmd tea.Cmd) (StreamDoneMsg, bool) {
	if cmd == nil {
		return StreamDoneMsg{}, false
	}
	msg := cmd()
	if done, ok := msg.(StreamDoneMsg); ok {
		return done, true
	}
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		return StreamDoneMsg{}, false
	}
	for _, subCmd := range batch {
		if subCmd == nil {
			continue
		}
		if done, ok := subCmd().(StreamDoneMsg); ok {
			return done, true
		}
	}
	return StreamDoneMsg{}, false
}

// TestIntegration_FullStreamingFlow exercises the full submit → stream → done flow.
// Because there is no real bubbletea program in tests, prog is nil, so TokenMsgs are
// not sent via prog.Send. The stream cmd returns a StreamDoneMsg directly.
func TestIntegration_FullStreamingFlow(t *testing.T) {
	lm := newMockLM("hello", " ", "world", "!", "\n")
	m := New(lm, "mock", nil)

	// Send SubmitMsg — should start streaming.
	m, cmd := callUpdate(m, SubmitMsg{Content: "what is 2+2?"})
	if !m.streaming {
		t.Fatal("expected streaming == true after SubmitMsg")
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd after SubmitMsg")
	}

	// Execute the cmd (unwrapping any BatchMsg from the tick being co-scheduled).
	doneMsg, ok := runStreamCmd(cmd)
	if !ok {
		t.Fatal("expected StreamDoneMsg from cmd")
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

	// Verify mock language model was called once.
	if lm.callCount != 1 {
		t.Errorf("expected callCount == 1, got %d", lm.callCount)
	}
}

// TestIntegration_UserMessageInHistory verifies the user message is recorded correctly.
func TestIntegration_UserMessageInHistory(t *testing.T) {
	lm := newMockLM("ok")
	m := New(lm, "mock", nil)

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
	lm := newMockLM("a", "b", "c")
	m := New(lm, "mock", nil)

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
	lm := newMockLM("hi")
	m := New(lm, "mock", nil)

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
	_, streamCmd := callUpdate(m, SubmitMsg{Content: "test"})
	if streamCmd == nil {
		t.Error("expected non-nil stream cmd")
	}
	// Execute cmd — no panic expected.
	_, ok := runStreamCmd(streamCmd)
	if !ok {
		t.Error("expected StreamDoneMsg from stream cmd")
	}
}

// TestIntegration_ExtensionHost_NoExtensions_Safe verifies a real host with no extensions loaded.
func TestIntegration_ExtensionHost_NoExtensions_Safe(t *testing.T) {
	lm := newMockLM("hi")
	h := extension.NewHost(nil)
	m := New(lm, "mock", h)

	// Init should not panic.
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return a non-nil Cmd")
	}

	// Submit and execute stream — no extensions loaded, no panics expected.
	_, streamCmd := callUpdate(m, SubmitMsg{Content: "test with host"})
	if streamCmd == nil {
		t.Fatal("expected non-nil stream cmd")
	}

	doneMsg, ok := runStreamCmd(streamCmd)
	if !ok {
		t.Fatal("expected StreamDoneMsg from stream cmd")
	}
	if doneMsg.Err != nil {
		t.Errorf("expected no error, got %v", doneMsg.Err)
	}
}
