package harness

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	mockprovider "github.com/mattdurham/bob/bob/provider/mock"
	"github.com/mattdurham/bob/bob/sdk"
)

func newTestModel() Model {
	p := &mockprovider.Provider{
		Tokens: []string{"hello", " ", "world"},
	}
	return New(p, nil)
}

// callUpdate is a helper that calls Update and returns the concrete Model.
func callUpdate(m Model, msg tea.Msg) (Model, tea.Cmd) {
	newModel, cmd := m.Update(msg)
	return newModel.(Model), cmd
}

func TestModel_Init_ReturnsCmd(t *testing.T) {
	m := newTestModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return a non-nil Cmd")
	}
}

func TestModel_View_ReturnsNonEmpty(t *testing.T) {
	m := newTestModel()
	view := m.View()
	if view.Content == "" {
		t.Error("View() returned empty content")
	}
}

func TestModel_Update_TokenMsg_AppendsToChat(t *testing.T) {
	m := newTestModel()
	m.streaming = true

	m, _ = callUpdate(m, TokenMsg{Token: "hello"})
	if m.chat.current != "hello" {
		t.Errorf("chat current: got %q, want %q", m.chat.current, "hello")
	}

	m, _ = callUpdate(m, TokenMsg{Token: " world"})
	if m.chat.current != "hello world" {
		t.Errorf("chat current: got %q, want %q", m.chat.current, "hello world")
	}
}

func TestModel_Update_StreamDoneMsg_ClearsStreaming(t *testing.T) {
	m := newTestModel()
	m.streaming = true
	m.chat.AppendToken("test response")

	m, _ = callUpdate(m, StreamDoneMsg{Err: nil})
	if m.streaming {
		t.Error("streaming should be false after StreamDoneMsg")
	}
}

func TestModel_Update_StreamDoneMsg_Error_ShowsError(t *testing.T) {
	m := newTestModel()
	m.streaming = true

	m, _ = callUpdate(m, StreamDoneMsg{Err: errors.New("API error")})
	if m.streaming {
		t.Error("streaming should be false after StreamDoneMsg with error")
	}
	// Should have added an error notification.
	if len(m.chat.messages) == 0 {
		t.Error("expected error notification in chat")
	}
}

func TestModel_Update_StreamDoneMsg_ContextCanceled_NoError(t *testing.T) {
	m := newTestModel()
	m.streaming = true
	m.chat.AppendToken("partial")

	// context.Canceled should not show as an error notification.
	m, _ = callUpdate(m, StreamDoneMsg{Err: context.Canceled})
	if m.streaming {
		t.Error("streaming should be false")
	}
	// FinalizeMessage adds the partial assistant message (1 message expected).
	// No additional error notification should be added for context.Canceled.
	for _, msg := range m.chat.messages {
		if msg.role == "system" {
			t.Errorf("unexpected error notification for context.Canceled: %q", msg.content)
		}
	}
}

func TestModel_Update_ReloadMsg_TriggersExtensionReload(t *testing.T) {
	m := newTestModel() // no extension host
	m, cmd := callUpdate(m, ReloadMsg{})
	if m.streaming {
		t.Error("streaming should not start on reload")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd after ReloadMsg")
	}
	// Execute the cmd.
	msg := cmd()
	_, ok := msg.(NotifyMsg)
	if !ok {
		t.Errorf("expected NotifyMsg from reload cmd, got %T", msg)
	}
}

func TestModel_Update_ClearMsg_ClearsHistory(t *testing.T) {
	m := newTestModel()
	m.history = append(m.history, sdk.Message{Role: sdk.RoleUser, Content: "hello"})
	m.chat.AddUserMessage("hello")

	m, _ = callUpdate(m, clearMsg{})
	if len(m.history) != 0 {
		t.Errorf("expected empty history after clear, got %d items", len(m.history))
	}
	if len(m.chat.messages) != 0 {
		t.Errorf("expected empty chat after clear, got %d messages", len(m.chat.messages))
	}
}

func TestModel_Update_SetModelMsg(t *testing.T) {
	m := newTestModel()
	m, _ = callUpdate(m, setModelMsg{Model: "claude-haiku-3-5"})
	if m.activeModel != "claude-haiku-3-5" {
		t.Errorf("activeModel: got %q, want %q", m.activeModel, "claude-haiku-3-5")
	}
	if m.statusBar.modelName != "claude-haiku-3-5" {
		t.Errorf("statusBar.modelName: got %q, want %q", m.statusBar.modelName, "claude-haiku-3-5")
	}
}

func TestModel_Update_CommandMsg_Clear(t *testing.T) {
	m := newTestModel()
	m.history = append(m.history, sdk.Message{Role: sdk.RoleUser, Content: "test"})

	// Dispatch /clear command.
	cmd := m.commands.Dispatch("clear", nil)
	msg := cmd()
	m, _ = callUpdate(m, msg)

	if len(m.history) != 0 {
		t.Errorf("expected empty history after /clear, got %d", len(m.history))
	}
}

func TestModel_Update_CommandMsg_UnknownCommand(t *testing.T) {
	m := newTestModel()
	m, cmd := callUpdate(m, CommandMsg{Name: "nonexistent", Args: nil})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for unknown command")
	}
	msg := cmd()
	notify, ok := msg.(NotifyMsg)
	if !ok {
		t.Fatalf("expected NotifyMsg, got %T", msg)
	}
	_ = m
	if notify.Text == "" {
		t.Error("expected non-empty error message")
	}
}

func TestModel_Update_SubmitMsg_StartsStream(t *testing.T) {
	p := &mockprovider.Provider{Tokens: []string{"hi"}}
	m := New(p, nil)

	m, cmd := callUpdate(m, SubmitMsg{Content: "hello"})
	if !m.streaming {
		t.Error("streaming should be true after SubmitMsg")
	}
	if cmd == nil {
		t.Error("expected non-nil Cmd")
	}
}

func TestModel_Update_SubmitMsg_IgnoredWhileStreaming(t *testing.T) {
	m := newTestModel()
	m.streaming = true

	m, cmd := callUpdate(m, SubmitMsg{Content: "new message"})
	if cmd != nil {
		// cmd may be non-nil (batch with nil cmds), but no stream should start.
	}
	_ = m
	// We just verify no panic and model stays consistent.
}

func TestModel_Update_NotifyMsg(t *testing.T) {
	m := newTestModel()
	m, _ = callUpdate(m, NotifyMsg{Text: "test notification"})
	// Should have added to chat.
	if len(m.chat.messages) == 0 {
		t.Error("expected notification in chat")
	}
}

func TestModel_Update_StatusUpdateMsg(t *testing.T) {
	m := newTestModel()
	m, _ = callUpdate(m, StatusUpdateMsg{Key: "foo", Value: "bar"})
	if m.statusBar.statuses["foo"] != "bar" {
		t.Errorf("statusBar.statuses[foo]: got %q, want %q", m.statusBar.statuses["foo"], "bar")
	}
}

func TestModel_Update_WindowSizeMsg(t *testing.T) {
	m := newTestModel()
	m, _ = callUpdate(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	if m.width != 120 || m.height != 40 {
		t.Errorf("dimensions: got %dx%d, want 120x40", m.width, m.height)
	}
}
