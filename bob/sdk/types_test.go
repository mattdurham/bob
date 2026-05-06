package sdk_test

import (
	"encoding/json"
	"testing"

	"github.com/mattdurham/bob/bob/sdk"
)

func TestEventJSONRoundTrip(t *testing.T) {
	events := []struct {
		name    string
		evtType sdk.EventType
		payload any
	}{
		{
			name:    "session_start",
			evtType: sdk.EventSessionStart,
			payload: sdk.SessionStartPayload{Reason: "new_session"},
		},
		{
			name:    "before_agent_start",
			evtType: sdk.EventBeforeAgentStart,
			payload: sdk.BeforeAgentStartPayload{Prompt: "hello", SystemPrompt: "be helpful"},
		},
		{
			name:    "before_provider_request",
			evtType: sdk.EventBeforeProviderRequest,
			payload: sdk.BeforeProviderRequestPayload{
				Messages: []sdk.Message{{Role: sdk.RoleUser, Content: "hi"}},
				Model:    "claude-sonnet",
			},
		},
		{
			name:    "after_provider_response",
			evtType: sdk.EventAfterProviderResponse,
			payload: sdk.AfterProviderResponsePayload{Usage: sdk.UsageStats{InputTokens: 10, OutputTokens: 20}},
		},
		{
			name:    "on_tool_call",
			evtType: sdk.EventOnToolCall,
			payload: sdk.OnToolCallPayload{ToolCallID: "tc-1", ToolName: "search", Input: json.RawMessage(`{"q":"foo"}`)},
		},
		{
			name:    "on_tool_result",
			evtType: sdk.EventOnToolResult,
			payload: sdk.OnToolResultPayload{ToolCallID: "tc-1", Result: "bar", IsError: false},
		},
		{
			name:    "message_start",
			evtType: sdk.EventMessageStart,
			payload: sdk.MessageStartPayload{Role: "assistant"},
		},
		{
			name:    "message_end",
			evtType: sdk.EventMessageEnd,
			payload: sdk.MessageEndPayload{Role: "assistant", Content: "hello world"},
		},
		{
			name:    "shutdown",
			evtType: sdk.EventShutdown,
			payload: sdk.ShutdownPayload{Reason: "quit"},
		},
	}

	for _, tc := range events {
		t.Run(tc.name, func(t *testing.T) {
			payloadBytes, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatalf("marshal payload: %v", err)
			}
			evt := sdk.Event{
				Type:    tc.evtType,
				Payload: json.RawMessage(payloadBytes),
			}
			data, err := json.Marshal(evt)
			if err != nil {
				t.Fatalf("marshal event: %v", err)
			}
			var got sdk.Event
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("unmarshal event: %v", err)
			}
			if got.Type != tc.evtType {
				t.Errorf("type: got %q, want %q", got.Type, tc.evtType)
			}
			if string(got.Payload) != string(payloadBytes) {
				t.Errorf("payload: got %s, want %s", got.Payload, payloadBytes)
			}
		})
	}
}

func TestEventResponseRoundTrip(t *testing.T) {
	cases := []sdk.EventResponse{
		{Cancel: true},
		{Block: true},
		{Error: "something went wrong"},
		{},
	}
	for _, c := range cases {
		data, err := json.Marshal(c)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var got sdk.EventResponse
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got != c {
			t.Errorf("got %+v, want %+v", got, c)
		}
	}
}

func TestMessageRoundTrip(t *testing.T) {
	msg := sdk.Message{Role: sdk.RoleUser, Content: "hello"}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got sdk.Message
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Role != sdk.RoleUser || got.Content != "hello" {
		t.Errorf("got %+v", got)
	}
}

func TestToolInputSchemaPreservesRawJSON(t *testing.T) {
	raw := `{"type":"object","properties":{"q":{"type":"string"}}}`
	tool := sdk.Tool{
		Name:        "search",
		Description: "search the web",
		InputSchema: json.RawMessage(raw),
	}
	data, err := json.Marshal(tool)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got sdk.Tool
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if string(got.InputSchema) != raw {
		t.Errorf("input_schema: got %s, want %s", got.InputSchema, raw)
	}
}

func TestHostCallRoundTrip(t *testing.T) {
	req := sdk.HostCallRequest{
		Method: sdk.MethodSubscribe,
		Params: json.RawMessage(`{"event":"session_start"}`),
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal req: %v", err)
	}
	var gotReq sdk.HostCallRequest
	if err := json.Unmarshal(data, &gotReq); err != nil {
		t.Fatalf("unmarshal req: %v", err)
	}
	if gotReq.Method != sdk.MethodSubscribe {
		t.Errorf("method: got %q, want %q", gotReq.Method, sdk.MethodSubscribe)
	}
	if string(gotReq.Params) != `{"event":"session_start"}` {
		t.Errorf("params: got %s", gotReq.Params)
	}

	resp := sdk.HostCallResponse{
		Result: json.RawMessage(`{"ok":true}`),
	}
	data, err = json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal resp: %v", err)
	}
	var gotResp sdk.HostCallResponse
	if err := json.Unmarshal(data, &gotResp); err != nil {
		t.Fatalf("unmarshal resp: %v", err)
	}
	if string(gotResp.Result) != `{"ok":true}` {
		t.Errorf("result: got %s", gotResp.Result)
	}
}

func TestRoleConstants(t *testing.T) {
	if sdk.RoleUser != "user" {
		t.Errorf("RoleUser = %q, want \"user\"", sdk.RoleUser)
	}
	if sdk.RoleAssistant != "assistant" {
		t.Errorf("RoleAssistant = %q, want \"assistant\"", sdk.RoleAssistant)
	}
}

func TestEventTypeConstants(t *testing.T) {
	types := []sdk.EventType{
		sdk.EventSessionStart,
		sdk.EventBeforeAgentStart,
		sdk.EventBeforeProviderRequest,
		sdk.EventAfterProviderResponse,
		sdk.EventOnToolCall,
		sdk.EventOnToolResult,
		sdk.EventMessageStart,
		sdk.EventMessageEnd,
		sdk.EventShutdown,
	}
	if len(types) != 9 {
		t.Errorf("expected 9 event types, got %d", len(types))
	}
	for _, et := range types {
		if et == "" {
			t.Errorf("event type constant is empty string")
		}
	}
}

func TestMethodConstants(t *testing.T) {
	methods := []string{
		sdk.MethodSubscribe,
		sdk.MethodRegisterTool,
		sdk.MethodRegisterCommand,
		sdk.MethodSendMessage,
		sdk.MethodSetStatus,
		sdk.MethodNotify,
		sdk.MethodToolResult,
		sdk.MethodStoreSet,
		sdk.MethodStoreGet,
		sdk.MethodAbort,
	}
	if len(methods) != 10 {
		t.Errorf("expected 10 method constants, got %d", len(methods))
	}
	for _, m := range methods {
		if m == "" {
			t.Errorf("method constant is empty string")
		}
	}
}
