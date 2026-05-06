// Package harness implements the bubbletea TUI for the bob coding assistant.
package harness

// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

import "github.com/mattdurham/bob/bob/sdk"

// TokenMsg carries a single streamed token from the provider.
type TokenMsg struct{ Token string }

// StreamDoneMsg signals that a provider stream has finished.
type StreamDoneMsg struct{ Err error }

// ExtensionEventResultMsg carries results from dispatching an event to extensions.
type ExtensionEventResultMsg struct {
	Results []sdk.EventResponse
	Err     error
}

// ReloadMsg triggers a hot-reload of all loaded extensions.
type ReloadMsg struct{}

// NotifyMsg carries a notification message to display in the chat.
type NotifyMsg struct{ Text string }

// StatusUpdateMsg sets or updates a keyed value in the status bar.
type StatusUpdateMsg struct{ Key, Value string }

// SubmitMsg carries user-submitted input text.
type SubmitMsg struct{ Content string }

// CommandMsg carries a parsed slash command.
type CommandMsg struct {
	Name string
	Args []string
}

// abortStreamMsg is sent by the OnAbort callback to cancel the active stream
// through the bubbletea program, ensuring the live model's cancelStream is used.
type abortStreamMsg struct{}

// streamTickMsg fires periodically while streaming to update the working indicator.
type streamTickMsg struct{}
