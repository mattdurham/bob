package harness

import (
	"strings"

	lipgloss "charm.land/lipgloss/v2"
	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/viewport"
	"github.com/mattdurham/bob/bob/sdk"
)

// chatMessage is a finalised message in the chat history.
type chatMessage struct {
	role    sdk.Role
	content string
}

// ChatView renders the conversation history in a scrollable viewport.
type ChatView struct {
	vp       viewport.Model
	messages []chatMessage
	current  string // current in-progress assistant message
	width    int
	height   int
}

var (
	userStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FFFF"))
	assistantStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))
	systemStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("#888888"))
)

// NewChatView creates a ChatView with the given dimensions.
func NewChatView(width, height int) ChatView {
	vp := viewport.New()
	vp.SetWidth(width)
	vp.SetHeight(height)
	return ChatView{vp: vp, width: width, height: height}
}

// SetSize updates the viewport dimensions.
func (c *ChatView) SetSize(width, height int) {
	c.width = width
	c.height = height
	c.vp.SetWidth(width)
	c.vp.SetHeight(height)
	c.refreshContent()
}

// AppendToken adds a token to the in-progress assistant message and scrolls to bottom.
func (c *ChatView) AppendToken(token string) {
	c.current += token
	c.refreshContent()
	c.vp.GotoBottom()
}

// FinalizeMessage seals the in-progress message and adds it to the history.
func (c *ChatView) FinalizeMessage() {
	if c.current == "" {
		return
	}
	c.messages = append(c.messages, chatMessage{role: sdk.RoleAssistant, content: c.current})
	c.current = ""
	c.refreshContent()
}

// AddUserMessage prepends a user message to the history.
func (c *ChatView) AddUserMessage(content string) {
	c.messages = append(c.messages, chatMessage{role: sdk.RoleUser, content: content})
	c.refreshContent()
	c.vp.GotoBottom()
}

// AddNotification appends a system/notification line.
func (c *ChatView) AddNotification(text string) {
	c.messages = append(c.messages, chatMessage{role: "system", content: text})
	c.refreshContent()
	c.vp.GotoBottom()
}

// Clear resets the chat history.
func (c *ChatView) Clear() {
	c.messages = nil
	c.current = ""
	c.refreshContent()
}

// MessageCount returns the number of finalised messages.
func (c *ChatView) MessageCount() int { return len(c.messages) }

// Update handles viewport scrolling.
func (c ChatView) Update(msg tea.Msg) (ChatView, tea.Cmd) {
	var cmd tea.Cmd
	c.vp, cmd = c.vp.Update(msg)
	return c, cmd
}

// View renders the chat content.
func (c ChatView) View() string {
	return c.vp.View()
}

// refreshContent rebuilds the viewport content from messages.
func (c *ChatView) refreshContent() {
	var sb strings.Builder
	for _, m := range c.messages {
		renderMessage(&sb, m, c.width)
	}
	if c.current != "" {
		renderMessage(&sb, chatMessage{role: sdk.RoleAssistant, content: c.current}, c.width)
	}
	c.vp.SetContent(sb.String())
}

func renderMessage(sb *strings.Builder, m chatMessage, _ int) {
	switch m.role {
	case sdk.RoleUser:
		sb.WriteString(userStyle.Render("You: "))
		sb.WriteString(m.content)
	case sdk.RoleAssistant:
		sb.WriteString(assistantStyle.Render("Bob: "))
		sb.WriteString(m.content)
	default:
		sb.WriteString(systemStyle.Render("» " + m.content))
	}
	sb.WriteString("\n\n")
}
