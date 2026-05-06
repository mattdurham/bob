// Package sdk defines shared types for the bob coding harness and its WASM extensions.
package sdk

import "encoding/json"

// EventType identifies a lifecycle event dispatched to extensions.
type EventType string

const (
	EventSessionStart          EventType = "session_start"
	EventBeforeAgentStart      EventType = "before_agent_start"
	EventBeforeProviderRequest EventType = "before_provider_request"
	EventAfterProviderResponse EventType = "after_provider_response"
	EventOnToolCall            EventType = "on_tool_call"
	EventOnToolResult          EventType = "on_tool_result"
	EventMessageStart          EventType = "message_start"
	EventMessageEnd            EventType = "message_end"
	EventShutdown              EventType = "shutdown"
)

// Event is dispatched to extensions via _on_event.
type Event struct {
	Type    EventType       `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// EventResponse is the optional JSON response from _on_event.
type EventResponse struct {
	Cancel bool   `json:"cancel,omitempty"`
	Block  bool   `json:"block,omitempty"`
	Error  string `json:"error,omitempty"`
}

// Payload types for each event.

// SessionStartPayload is the payload for EventSessionStart.
type SessionStartPayload struct {
	Reason string `json:"reason"`
}

// BeforeAgentStartPayload is the payload for EventBeforeAgentStart.
type BeforeAgentStartPayload struct {
	Prompt       string `json:"prompt"`
	SystemPrompt string `json:"system_prompt"`
}

// BeforeProviderRequestPayload is the payload for EventBeforeProviderRequest.
type BeforeProviderRequestPayload struct {
	Messages []Message `json:"messages"`
	Model    string    `json:"model"`
}

// AfterProviderResponsePayload is the payload for EventAfterProviderResponse.
type AfterProviderResponsePayload struct {
	Usage UsageStats `json:"usage"`
}

// OnToolCallPayload is the payload for EventOnToolCall.
type OnToolCallPayload struct {
	ToolCallID string          `json:"tool_call_id"`
	ToolName   string          `json:"tool_name"`
	Input      json.RawMessage `json:"input"`
}

// OnToolResultPayload is the payload for EventOnToolResult.
type OnToolResultPayload struct {
	ToolCallID string `json:"tool_call_id"`
	Result     string `json:"result"`
	IsError    bool   `json:"is_error"`
}

// MessageStartPayload is the payload for EventMessageStart.
type MessageStartPayload struct {
	Role string `json:"role"`
}

// MessageEndPayload is the payload for EventMessageEnd.
type MessageEndPayload struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ShutdownPayload is the payload for EventShutdown.
type ShutdownPayload struct {
	Reason string `json:"reason"`
}

// UsageStats holds token usage from a provider response.
type UsageStats struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Role is a message role (user or assistant).
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message is a chat message.
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// Tool is a function the LLM may call.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// HostCallRequest is the JSON payload sent by an extension via host_call.
type HostCallRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// HostCallResponse is the JSON response returned by the host via host_call.
type HostCallResponse struct {
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// host_call method constants.
const (
	MethodSubscribe        = "subscribe"
	MethodRegisterTool     = "register_tool"
	MethodRegisterCommand  = "register_command"
	MethodSendMessage      = "send_message"
	MethodSetStatus        = "set_status"
	MethodNotify           = "notify"
	MethodToolResult       = "tool_result"
	MethodStoreSet         = "store_set"
	MethodStoreGet         = "store_get"
	MethodAbort            = "abort"
)
