// Package hook parses Claude Code hook stdin JSON into a Command for the shipmate daemon.
package hook

import (
	"encoding/json"
	"fmt"
	"io"
)

// Command is the NDJSON message sent over the Unix socket to the daemon.
// It is defined here so that both client and daemon import it without a cycle.
type Command struct {
	Type      string            `json:"type"` // "record" | "memory" | "stop"
	SessionID string            `json:"session_id"`
	HookEvent string            `json:"hook_event"` // tool_name from hook stdin
	Attrs     map[string]string `json:"attrs"`
	Text      string            `json:"text"` // for "memory" commands
}

// hookInput is the JSON structure Claude Code writes to hook stdin.
type hookInput struct {
	SessionID    string          `json:"session_id"`
	ToolName     string          `json:"tool_name"`
	ToolInput    json.RawMessage `json:"tool_input"`
	ToolResponse json.RawMessage `json:"tool_response"`
}

// toolInputFields contains fields we extract from tool_input.
type toolInputFields struct {
	FilePath string `json:"file_path"`
	Command  string `json:"command"`
}

// toolResponseFields contains fields we extract from tool_response.
type toolResponseFields struct {
	Success *bool `json:"success"`
}

// ParseHookInput reads a Claude Code hook payload from r and returns a Command
// with Type="record". Returns an error if r is empty or the JSON is malformed.
func ParseHookInput(r io.Reader) (Command, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return Command{}, fmt.Errorf("hook: read stdin: %w", err)
	}
	if len(data) == 0 {
		return Command{}, fmt.Errorf("hook: empty input")
	}

	var hi hookInput
	if err := json.Unmarshal(data, &hi); err != nil {
		return Command{}, fmt.Errorf("hook: parse JSON: %w", err)
	}

	attrs := make(map[string]string)

	// Extract known fields from tool_input.
	if len(hi.ToolInput) > 0 {
		var ti toolInputFields
		if err := json.Unmarshal(hi.ToolInput, &ti); err == nil {
			if ti.FilePath != "" {
				attrs["tool.file"] = ti.FilePath
			}
			if ti.Command != "" {
				attrs["tool.command"] = ti.Command
			}
		}
	}

	// Extract success from tool_response (only when the field is explicitly present).
	if len(hi.ToolResponse) > 0 {
		var tr toolResponseFields
		if err := json.Unmarshal(hi.ToolResponse, &tr); err == nil && tr.Success != nil {
			if *tr.Success {
				attrs["tool.success"] = "true"
			} else {
				attrs["tool.success"] = "false"
			}
		}
	}

	return Command{
		Type:      "record",
		SessionID: hi.SessionID,
		HookEvent: hi.ToolName,
		Attrs:     attrs,
	}, nil
}
