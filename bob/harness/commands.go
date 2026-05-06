package harness

import (
	"fmt"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// Command is a slash command registered with the Registry.
type Command struct {
	Name    string
	Desc    string
	Handler func(args []string) tea.Cmd
}

// Registry holds registered slash commands and dispatches them.
type Registry struct {
	commands map[string]Command
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry {
	return &Registry{commands: make(map[string]Command)}
}

// Register adds a command to the registry. Duplicate names are silently overwritten.
func (r *Registry) Register(cmd Command) {
	r.commands[cmd.Name] = cmd
}

// Dispatch looks up name and calls its handler with args.
// Returns an error Cmd if the command is not found.
func (r *Registry) Dispatch(name string, args []string) tea.Cmd {
	cmd, ok := r.commands[name]
	if !ok {
		msg := fmt.Sprintf("unknown command: /%s", name)
		return func() tea.Msg { return NotifyMsg{Text: msg} }
	}
	return cmd.Handler(args)
}

// List returns all registered commands sorted by name.
func (r *Registry) List() []Command {
	out := make([]Command, 0, len(r.commands))
	for _, c := range r.commands {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// HelpText returns a formatted list of all commands.
func (r *Registry) HelpText() string {
	cmds := r.List()
	if len(cmds) == 0 {
		return "No commands registered."
	}
	var sb strings.Builder
	for _, c := range cmds {
		fmt.Fprintf(&sb, "  /%s — %s\n", c.Name, c.Desc)
	}
	return strings.TrimRight(sb.String(), "\n")
}

// registerBuiltins installs the built-in commands into r.
// model is a pointer so handlers can close over the model's state changes
// (they return Cmds that emit Msgs, so the model pointer itself isn't mutated).
func registerBuiltins(r *Registry) {
	r.Register(Command{
		Name: "help",
		Desc: "Show available commands",
		Handler: func(_ []string) tea.Cmd {
			return func() tea.Msg {
				return NotifyMsg{Text: "Use the commands listed in the help text above."}
			}
		},
	})

	r.Register(Command{
		Name: "clear",
		Desc: "Clear the conversation history",
		Handler: func(_ []string) tea.Cmd {
			return func() tea.Msg { return clearMsg{} }
		},
	})

	r.Register(Command{
		Name: "reload",
		Desc: "Hot-reload all extensions",
		Handler: func(_ []string) tea.Cmd {
			return func() tea.Msg { return ReloadMsg{} }
		},
	})

	r.Register(Command{
		Name: "model",
		Desc: "Switch the active model (e.g. /model claude-haiku-3-5)",
		Handler: func(args []string) tea.Cmd {
			if len(args) == 0 {
				return func() tea.Msg { return NotifyMsg{Text: "Usage: /model <name>"} }
			}
			return func() tea.Msg { return setModelMsg{Model: args[0]} }
		},
	})
}

// Internal message types used by built-in command handlers.
type clearMsg struct{}
type setModelMsg struct{ Model string }
