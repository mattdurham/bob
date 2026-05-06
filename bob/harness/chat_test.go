package harness

import (
	"strings"
	"testing"

	"github.com/mattdurham/bob/bob/sdk"
)

func TestChatView_AppendToken(t *testing.T) {
	c := NewChatView(80, 20)
	c.AppendToken("hello")
	c.AppendToken(" world")
	if c.current != "hello world" {
		t.Errorf("current: got %q, want %q", c.current, "hello world")
	}
}

func TestChatView_FinalizeMessage_SetsRole(t *testing.T) {
	c := NewChatView(80, 20)
	c.AppendToken("response")
	c.FinalizeMessage()

	if len(c.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(c.messages))
	}
	if c.messages[0].role != sdk.RoleAssistant {
		t.Errorf("role: got %v, want %v", c.messages[0].role, sdk.RoleAssistant)
	}
	if c.messages[0].content != "response" {
		t.Errorf("content: got %q", c.messages[0].content)
	}
	if c.current != "" {
		t.Errorf("current should be empty after finalize, got %q", c.current)
	}
}

func TestChatView_FinalizeMessage_Empty_NoOp(t *testing.T) {
	c := NewChatView(80, 20)
	c.FinalizeMessage()
	if len(c.messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(c.messages))
	}
}

func TestChatView_AddUserMessage(t *testing.T) {
	c := NewChatView(80, 20)
	c.AddUserMessage("hello from user")

	if len(c.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(c.messages))
	}
	if c.messages[0].role != sdk.RoleUser {
		t.Errorf("role: got %v, want user", c.messages[0].role)
	}
	if c.messages[0].content != "hello from user" {
		t.Errorf("content: got %q", c.messages[0].content)
	}
}

func TestChatView_MessageOrder(t *testing.T) {
	c := NewChatView(80, 20)
	c.AddUserMessage("first")
	c.AppendToken("second")
	c.FinalizeMessage()

	if len(c.messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(c.messages))
	}
	if c.messages[0].role != sdk.RoleUser {
		t.Errorf("first message should be user")
	}
	if c.messages[1].role != sdk.RoleAssistant {
		t.Errorf("second message should be assistant")
	}
}

func TestChatView_Clear(t *testing.T) {
	c := NewChatView(80, 20)
	c.AddUserMessage("hi")
	c.AppendToken("hello")
	c.Clear()

	if len(c.messages) != 0 {
		t.Errorf("expected 0 messages after clear, got %d", len(c.messages))
	}
	if c.current != "" {
		t.Errorf("expected empty current after clear")
	}
}

func TestChatView_View_NonEmpty(t *testing.T) {
	c := NewChatView(80, 20)
	c.AddUserMessage("hello")
	c.AppendToken("world")
	view := c.View()
	// The view should contain the content somehow.
	_ = view // Just verify it doesn't panic.
}

func TestChatView_RenderUserMessage_ContainsContent(t *testing.T) {
	c := NewChatView(80, 20)
	c.AddUserMessage("unique-test-content")
	c.refreshContent()
	content := c.vp.View()
	if !strings.Contains(content, "unique-test-content") {
		t.Errorf("viewport content should contain user message, got: %q", content)
	}
}
