package harness

// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
)

// InputArea wraps a textarea and handles command detection.
type InputArea struct {
	ta         textarea.Model
	width      int
	lastWasEsc bool // true if the previous keypress was esc
}

// NewInputArea creates an InputArea with the given width.
func NewInputArea(width int) InputArea {
	ta := textarea.New()
	ta.SetWidth(width)
	ta.SetHeight(3)
	ta.Placeholder = "Type a message… (Enter to send, /command for commands)"
	ta.ShowLineNumbers = false
	// Override InsertNewline to only fire on shift+enter.
	ta.KeyMap.InsertNewline = key.NewBinding(key.WithKeys("shift+enter"))
	_ = ta.Focus()
	return InputArea{ta: ta, width: width}
}

// SetWidth updates the textarea width.
func (i *InputArea) SetWidth(w int) {
	i.width = w
	i.ta.SetWidth(w)
}

// Value returns the current textarea content.
func (i InputArea) Value() string { return i.ta.Value() }

// Reset clears the textarea.
func (i *InputArea) Reset() { i.ta.Reset() }

// Update handles keyboard input for the input area.
// It intercepts Enter (submit) and Esc (clear input or abort active stream on double-press)
// before forwarding remaining events to the textarea.
func (i InputArea) Update(msg tea.Msg) (InputArea, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyPressMsg:
		switch m.String() {
		case "enter":
			content := strings.TrimSpace(i.ta.Value())
			if content == "" {
				return i, nil
			}
			i.ta.Reset()
			if strings.HasPrefix(content, "/") {
				// Parse as command.
				parts := strings.Fields(content[1:])
				if len(parts) == 0 {
					return i, nil
				}
				name := parts[0]
				args := parts[1:]
				return i, func() tea.Msg { return CommandMsg{Name: name, Args: args} }
			}
			return i, func() tea.Msg { return SubmitMsg{Content: content} }

		case "esc":
			if i.lastWasEsc {
				i.lastWasEsc = false
				i.ta.Reset()
				return i, func() tea.Msg { return abortStreamMsg{} }
			}
			i.lastWasEsc = true
			i.ta.Reset()
			return i, nil

		default:
			i.lastWasEsc = false
		}
	}

	var cmd tea.Cmd
	i.ta, cmd = i.ta.Update(msg)
	return i, cmd
}

// View renders the input area.
func (i InputArea) View() string {
	return i.ta.View()
}
