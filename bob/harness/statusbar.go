package harness

import (
	"fmt"
	"sort"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
	tea "charm.land/bubbletea/v2"
)

// StatusBar renders provider/model info and keyed status values.
type StatusBar struct {
	providerName string
	modelName    string
	totalTokens  int
	statuses     map[string]string
	width        int
}

var (
	statusBarStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#333333"))
)

// NewStatusBar creates a StatusBar.
func NewStatusBar(providerName, modelName string) StatusBar {
	return StatusBar{
		providerName: providerName,
		modelName:    modelName,
		statuses:     make(map[string]string),
	}
}

// SetWidth sets the display width for truncation.
func (s *StatusBar) SetWidth(w int) { s.width = w }

// Update handles StatusUpdateMsg and StreamDoneMsg (token count).
func (s StatusBar) Update(msg tea.Msg) (StatusBar, tea.Cmd) {
	switch m := msg.(type) {
	case StatusUpdateMsg:
		if s.statuses == nil {
			s.statuses = make(map[string]string)
		}
		s.statuses[m.Key] = m.Value
	}
	return s, nil
}

// AddTokens increments the total token counter.
func (s *StatusBar) AddTokens(n int) { s.totalTokens += n }

// View renders the status bar as a single line.
func (s StatusBar) View() string {
	var parts []string
	if s.providerName != "" {
		parts = append(parts, fmt.Sprintf("[%s]", s.providerName))
	}
	if s.modelName != "" {
		parts = append(parts, fmt.Sprintf("[%s]", s.modelName))
	}
	if s.totalTokens > 0 {
		parts = append(parts, fmt.Sprintf("[tokens: %d]", s.totalTokens))
	}

	// Append sorted statuses.
	keys := make([]string, 0, len(s.statuses))
	for k := range s.statuses {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("[%s: %s]", k, s.statuses[k]))
	}

	line := strings.Join(parts, " ")
	if s.width > 0 && len(line) > s.width {
		line = line[:s.width-1] + "…"
	}
	return statusBarStyle.Width(s.width).Render(line)
}
