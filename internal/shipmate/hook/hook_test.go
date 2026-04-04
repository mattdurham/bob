package hook

import (
	"strings"
	"testing"
)

func TestParseHookInput_BasicRecord(t *testing.T) {
	input := `{
		"session_id": "ses-abc",
		"tool_name": "Bash",
		"tool_input": {"command": "go test ./..."},
		"tool_response": {"success": true}
	}`
	cmd, err := ParseHookInput(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseHookInput: %v", err)
	}
	if cmd.Type != "record" {
		t.Errorf("Type: got %q, want %q", cmd.Type, "record")
	}
	if cmd.SessionID != "ses-abc" {
		t.Errorf("SessionID: got %q, want %q", cmd.SessionID, "ses-abc")
	}
	if cmd.HookEvent != "Bash" {
		t.Errorf("HookEvent: got %q, want %q", cmd.HookEvent, "Bash")
	}
	if cmd.Attrs["tool.command"] != "go test ./..." {
		t.Errorf("tool.command: got %q, want %q", cmd.Attrs["tool.command"], "go test ./...")
	}
	if cmd.Attrs["tool.success"] != "true" {
		t.Errorf("tool.success: got %q, want %q", cmd.Attrs["tool.success"], "true")
	}
}

func TestParseHookInput_ReadToolExtractsFilePath(t *testing.T) {
	input := `{
		"session_id": "ses-xyz",
		"tool_name": "Read",
		"tool_input": {"file_path": "/tmp/foo.go"},
		"tool_response": {"success": true}
	}`
	cmd, err := ParseHookInput(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseHookInput: %v", err)
	}
	if cmd.Attrs["tool.file"] != "/tmp/foo.go" {
		t.Errorf("tool.file: got %q, want %q", cmd.Attrs["tool.file"], "/tmp/foo.go")
	}
}

func TestParseHookInput_ToolResponseFailure(t *testing.T) {
	input := `{
		"session_id": "ses-fail",
		"tool_name": "Write",
		"tool_input": {"file_path": "/tmp/bar.go"},
		"tool_response": {"success": false}
	}`
	cmd, err := ParseHookInput(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseHookInput: %v", err)
	}
	if cmd.Attrs["tool.success"] != "false" {
		t.Errorf("tool.success: got %q, want %q", cmd.Attrs["tool.success"], "false")
	}
}

func TestParseHookInput_EmptySessionID(t *testing.T) {
	input := `{
		"tool_name": "Bash",
		"tool_input": {},
		"tool_response": {}
	}`
	cmd, err := ParseHookInput(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseHookInput: %v", err)
	}
	if cmd.SessionID != "" {
		t.Errorf("expected empty SessionID, got %q", cmd.SessionID)
	}
}

func TestParseHookInput_InvalidJSON(t *testing.T) {
	_, err := ParseHookInput(strings.NewReader("{bad json"))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestParseHookInput_EmptyInput(t *testing.T) {
	_, err := ParseHookInput(strings.NewReader(""))
	if err == nil {
		t.Error("expected error for empty input, got nil")
	}
}

func TestParseHookInput_MissingToolResponse(t *testing.T) {
	// tool_response absent — tool.success should not appear in Attrs
	input := `{
		"session_id": "ses-x",
		"tool_name": "Glob",
		"tool_input": {}
	}`
	cmd, err := ParseHookInput(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseHookInput: %v", err)
	}
	if _, ok := cmd.Attrs["tool.success"]; ok {
		t.Error("expected no tool.success when tool_response absent")
	}
}

func TestParseHookInput_NilToolSuccessField(t *testing.T) {
	// tool_response present but success field absent
	input := `{
		"session_id": "ses-y",
		"tool_name": "Glob",
		"tool_input": {},
		"tool_response": {}
	}`
	cmd, err := ParseHookInput(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseHookInput: %v", err)
	}
	if _, ok := cmd.Attrs["tool.success"]; ok {
		t.Error("expected no tool.success when success field absent in tool_response")
	}
}
