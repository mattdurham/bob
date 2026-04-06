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
	Type      string            `json:"type"` // "stop" — only meaningful command; others are logged and discarded
	SessionID string            `json:"session_id"`
	HookEvent string            `json:"hook_event"` // hook_event_name from hook stdin
	Attrs     map[string]string `json:"attrs"`
	Text      string            `json:"text"` // reserved; not used by any current command
}

// hookInput is the JSON structure Claude Code writes to hook stdin.
// Fields are common across all events; event-specific fields are captured
// via separate structs and merged into Attrs.
type hookInput struct {
	// Common fields present on all events.
	SessionID      string `json:"session_id"`
	HookEventName  string `json:"hook_event_name"`
	TranscriptPath string `json:"transcript_path"`
	Cwd            string `json:"cwd"`
	PermissionMode string `json:"permission_mode"`

	// Subagent identity (SubagentStart, SubagentStop, and tool events inside subagents).
	AgentID   string `json:"agent_id"`
	AgentType string `json:"agent_type"`

	// Tool events (PreToolUse, PostToolUse).
	ToolName    string          `json:"tool_name"`
	ToolUseID   string          `json:"tool_use_id"`
	ToolInput   json.RawMessage `json:"tool_input"`
	ToolResponse json.RawMessage `json:"tool_response"`

	// UserPromptSubmit.
	Prompt string `json:"prompt"`

	// SubagentStop.
	LastAssistantMessage  string `json:"last_assistant_message"`
	AgentTranscriptPath   string `json:"agent_transcript_path"`

	// TaskCreated / TaskCompleted.
	TaskID          string `json:"task_id"`
	TaskSubject     string `json:"task_subject"`
	TaskDescription string `json:"task_description"`
	TeammateName    string `json:"teammate_name"`
	TeamName        string `json:"team_name"`

	// SessionStart.
	Source string `json:"source"` // startup | resume | clear | compact
	Model  string `json:"model"`
}

// toolInputFields contains fields we extract from tool_input.
// Covers Bash, Edit, Write, Read, Grep, and other common tools.
type toolInputFields struct {
	// Bash
	Command        string `json:"command"`
	Description    string `json:"description"`
	RunInBackground bool   `json:"run_in_background"`

	// Edit / Write / Read
	FilePath  string `json:"file_path"`
	OldString string `json:"old_string"`
	NewString string `json:"new_string"`
	Content   string `json:"content"`

	// Grep
	Pattern string `json:"pattern"`
	Path    string `json:"path"`
	Glob    string `json:"glob"`

	// WebSearch / WebFetch
	Query string `json:"query"`
	URL   string `json:"url"`
}

// toolResponseFields contains fields we extract from tool_response.
type toolResponseFields struct {
	Success  *bool  `json:"success"`
	FilePath string `json:"filePath"` // Write response uses camelCase
	Error    string `json:"error"`
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

	// Common fields.
	setAttr(attrs, "hook.event", hi.HookEventName)
	setAttr(attrs, "cwd", hi.Cwd)
	setAttr(attrs, "permission_mode", hi.PermissionMode)
	setAttr(attrs, "transcript_path", hi.TranscriptPath)

	// Subagent identity.
	setAttr(attrs, "agent_id", hi.AgentID)
	setAttr(attrs, "agent_type", hi.AgentType)

	// Tool events.
	setAttr(attrs, "tool_use_id", hi.ToolUseID)

	if len(hi.ToolInput) > 0 {
		var ti toolInputFields
		if err := json.Unmarshal(hi.ToolInput, &ti); err == nil {
			setAttr(attrs, "tool.file", ti.FilePath)
			setAttr(attrs, "tool.command", ti.Command)
			setAttr(attrs, "tool.description", ti.Description)
			setAttr(attrs, "tool.pattern", ti.Pattern)
			setAttr(attrs, "tool.path", ti.Path)
			setAttr(attrs, "tool.glob", ti.Glob)
			setAttr(attrs, "tool.query", ti.Query)
			setAttr(attrs, "tool.url", ti.URL)
			// old_string/new_string can be large; cap at 256 chars.
			setAttrCapped(attrs, "tool.old_string", ti.OldString, 256)
			setAttrCapped(attrs, "tool.new_string", ti.NewString, 256)
			// content (Write) can be very large; cap at 256 chars.
			setAttrCapped(attrs, "tool.content", ti.Content, 256)
			if ti.RunInBackground {
				attrs["tool.run_in_background"] = "true"
			}
		}
	}

	if len(hi.ToolResponse) > 0 {
		var tr toolResponseFields
		if err := json.Unmarshal(hi.ToolResponse, &tr); err == nil {
			if tr.Success != nil {
				if *tr.Success {
					attrs["tool.success"] = "true"
				} else {
					attrs["tool.success"] = "false"
				}
			}
			setAttr(attrs, "tool.response_file", tr.FilePath)
			setAttr(attrs, "tool.error", tr.Error)
		}
	}

	// UserPromptSubmit.
	setAttrCapped(attrs, "prompt.text", hi.Prompt, 512)

	// SubagentStop.
	setAttrCapped(attrs, "agent.last_message", hi.LastAssistantMessage, 512)
	setAttr(attrs, "agent.transcript_path", hi.AgentTranscriptPath)

	// TaskCreated / TaskCompleted.
	setAttr(attrs, "task.id", hi.TaskID)
	setAttr(attrs, "task.subject", hi.TaskSubject)
	setAttrCapped(attrs, "task.description", hi.TaskDescription, 256)
	setAttr(attrs, "task.teammate", hi.TeammateName)
	setAttr(attrs, "task.team", hi.TeamName)

	// SessionStart.
	setAttr(attrs, "session.source", hi.Source)
	setAttr(attrs, "session.model", hi.Model)

	// Use hook_event_name as the span event name; fall back to tool_name for
	// tool events where hook_event_name may be absent in older Claude Code versions.
	eventName := hi.HookEventName
	if eventName == "" {
		eventName = hi.ToolName
	}

	return Command{
		Type:      "record",
		SessionID: hi.SessionID,
		HookEvent: eventName,
		Attrs:     attrs,
	}, nil
}

// setAttr sets key=value in attrs only when value is non-empty.
func setAttr(attrs map[string]string, key, value string) {
	if value != "" {
		attrs[key] = value
	}
}

// setAttrCapped sets key=value, truncating value to maxLen runes when non-empty.
func setAttrCapped(attrs map[string]string, key, value string, maxLen int) {
	if value == "" {
		return
	}
	runes := []rune(value)
	if len(runes) > maxLen {
		attrs[key] = string(runes[:maxLen]) + "…"
	} else {
		attrs[key] = value
	}
}
