package extension

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mattdurham/bob/bob/sdk"
)

// writeWASM writes bytes to a temp file and returns the path.
func writeWASM(t *testing.T, name string, data []byte) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write test wasm: %v", err)
	}
	return path
}

func TestHost_Load_MinimalWASM(t *testing.T) {
	ctx := context.Background()
	h := NewHost(nil)
	defer h.Close(ctx)

	path := writeWASM(t, "minimal.wasm", minimalWASM)
	if err := h.Load(ctx, path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(h.extensions) != 1 {
		t.Errorf("expected 1 extension, got %d", len(h.extensions))
	}
}

func TestHost_Load_FileNotFound(t *testing.T) {
	ctx := context.Background()
	h := NewHost(nil)
	defer h.Close(ctx)

	err := h.Load(ctx, "/nonexistent/path/to/extension.wasm")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestHost_Load_MissingExport(t *testing.T) {
	ctx := context.Background()
	h := NewHost(nil)
	defer h.Close(ctx)

	path := writeWASM(t, "missing-free.wasm", missingFreeWASM)
	err := h.Load(ctx, path)
	if err == nil {
		t.Fatal("expected error for missing _free export, got nil")
	}
}

func TestHost_DispatchEvent_NotSubscribed(t *testing.T) {
	ctx := context.Background()

	var onEventCalled bool
	h := NewHost(nil)
	defer h.Close(ctx)

	path := writeWASM(t, "minimal.wasm", minimalWASM)
	if err := h.Load(ctx, path); err != nil {
		t.Fatalf("Load: %v", err)
	}

	// minimalWASM _init does not subscribe to anything.
	// So dispatching session_start should NOT call _on_event.
	// We verify by checking that no responses come back and no errors.
	_ = onEventCalled

	evt := sdk.Event{Type: sdk.EventSessionStart}
	responses, err := h.DispatchEvent(ctx, evt)
	if err != nil {
		t.Fatalf("DispatchEvent: %v", err)
	}
	if len(responses) != 0 {
		t.Errorf("expected 0 responses (unsubscribed), got %d", len(responses))
	}
}

func TestHost_DispatchEvent_Subscribed(t *testing.T) {
	ctx := context.Background()
	h := NewHost(nil)
	defer h.Close(ctx)

	path := writeWASM(t, "minimal.wasm", minimalWASM)
	if err := h.Load(ctx, path); err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Manually subscribe the extension to session_start.
	h.extensions[0].subscriptions[sdk.EventSessionStart] = true

	evt := sdk.Event{Type: sdk.EventSessionStart}
	responses, err := h.DispatchEvent(ctx, evt)
	if err != nil {
		t.Fatalf("DispatchEvent: %v", err)
	}
	// _on_event returns 0 (no response ptr), so we get an empty EventResponse.
	if len(responses) != 1 {
		t.Errorf("expected 1 response (subscribed), got %d", len(responses))
	}
}

func TestHost_Subscribe_ViaHostCall(t *testing.T) {
	ctx := context.Background()
	h := NewHost(nil)
	defer h.Close(ctx)

	path := writeWASM(t, "minimal.wasm", minimalWASM)
	if err := h.Load(ctx, path); err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Simulate what an extension would do via host_call: subscribe to session_start.
	ext := h.extensions[0]
	req := sdk.HostCallRequest{
		Method: sdk.MethodSubscribe,
		Params: []byte(`{"event":"session_start"}`),
	}
	resp := h.handleSubscribe(ext, req)
	if resp.Error != "" {
		t.Fatalf("handleSubscribe: %s", resp.Error)
	}
	if !ext.subscriptions[sdk.EventSessionStart] {
		t.Error("expected session_start to be subscribed")
	}
}

func TestHost_Store_SetGet(t *testing.T) {
	ctx := context.Background()
	h := NewHost(nil)
	defer h.Close(ctx)

	path := writeWASM(t, "minimal.wasm", minimalWASM)
	if err := h.Load(ctx, path); err != nil {
		t.Fatalf("Load: %v", err)
	}

	ext := h.extensions[0]

	// Test store_set.
	setResp := h.handleStoreSet(ext, sdk.HostCallRequest{
		Method: sdk.MethodStoreSet,
		Params: []byte(`{"key":"foo","value":"bar"}`),
	})
	if setResp.Error != "" {
		t.Fatalf("store_set: %s", setResp.Error)
	}

	// Test store_get.
	getResp := h.handleStoreGet(ext, sdk.HostCallRequest{
		Method: sdk.MethodStoreGet,
		Params: []byte(`{"key":"foo"}`),
	})
	if getResp.Error != "" {
		t.Fatalf("store_get: %s", getResp.Error)
	}
	if string(getResp.Result) != `{"value":"bar"}` {
		t.Errorf("store_get result: got %s, want %s", getResp.Result, `{"value":"bar"}`)
	}
}

func TestHost_Store_GetMiss(t *testing.T) {
	ctx := context.Background()
	h := NewHost(nil)
	defer h.Close(ctx)

	path := writeWASM(t, "minimal.wasm", minimalWASM)
	if err := h.Load(ctx, path); err != nil {
		t.Fatalf("Load: %v", err)
	}

	ext := h.extensions[0]
	resp := h.handleStoreGet(ext, sdk.HostCallRequest{
		Method: sdk.MethodStoreGet,
		Params: []byte(`{"key":"missing"}`),
	})
	if resp.Error == "" {
		t.Error("expected error for missing key, got none")
	}
}

func TestHost_RegisterTool_DuplicateRejected(t *testing.T) {
	ctx := context.Background()
	h := NewHost(nil)
	defer h.Close(ctx)

	path := writeWASM(t, "minimal.wasm", minimalWASM)
	if err := h.Load(ctx, path); err != nil {
		t.Fatalf("Load: %v", err)
	}

	toolJSON := []byte(`{"name":"search","description":"search the web","input_schema":{}}`)

	resp1 := h.handleRegisterTool(sdk.HostCallRequest{
		Method: sdk.MethodRegisterTool,
		Params: toolJSON,
	})
	if resp1.Error != "" {
		t.Fatalf("first register_tool: %s", resp1.Error)
	}

	resp2 := h.handleRegisterTool(sdk.HostCallRequest{
		Method: sdk.MethodRegisterTool,
		Params: toolJSON,
	})
	if resp2.Error == "" {
		t.Fatal("expected error for duplicate tool registration, got none")
	}
}

func TestHost_Callbacks_SetStatus(t *testing.T) {
	ctx := context.Background()
	h := NewHost(nil)
	defer h.Close(ctx)

	var gotKey, gotValue string
	h.OnSetStatus = func(k, v string) {
		gotKey = k
		gotValue = v
	}

	path := writeWASM(t, "minimal.wasm", minimalWASM)
	if err := h.Load(ctx, path); err != nil {
		t.Fatalf("Load: %v", err)
	}

	ext := h.extensions[0]
	resp := h.routeHostCall(ctx, ext.module, ext, sdk.HostCallRequest{
		Method: sdk.MethodSetStatus,
		Params: []byte(`{"key":"status","value":"ok"}`),
	})
	if resp.Error != "" {
		t.Fatalf("set_status: %s", resp.Error)
	}
	if gotKey != "status" || gotValue != "ok" {
		t.Errorf("OnSetStatus: got key=%q value=%q, want key=%q value=%q", gotKey, gotValue, "status", "ok")
	}
}

func TestHost_Reload(t *testing.T) {
	ctx := context.Background()
	h := NewHost(nil)
	defer h.Close(ctx)

	path := writeWASM(t, "minimal.wasm", minimalWASM)
	if err := h.Load(ctx, path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(h.extensions) != 1 {
		t.Fatalf("before reload: expected 1 extension, got %d", len(h.extensions))
	}

	// Reload with the same path.
	if err := h.Reload(ctx, []string{path}); err != nil {
		t.Fatalf("Reload: %v", err)
	}
	if len(h.extensions) != 1 {
		t.Errorf("after reload: expected 1 extension, got %d", len(h.extensions))
	}
}

func TestHost_Multiple_Extensions(t *testing.T) {
	ctx := context.Background()
	h := NewHost(nil)
	defer h.Close(ctx)

	// Load the same WASM twice under different names.
	path1 := writeWASM(t, "ext1.wasm", minimalWASM)
	path2 := writeWASM(t, "ext2.wasm", minimalWASM)

	if err := h.Load(ctx, path1); err != nil {
		t.Fatalf("Load ext1: %v", err)
	}
	if err := h.Load(ctx, path2); err != nil {
		t.Fatalf("Load ext2: %v", err)
	}
	if len(h.extensions) != 2 {
		t.Errorf("expected 2 extensions, got %d", len(h.extensions))
	}

	// Subscribe both to session_start.
	h.extensions[0].subscriptions[sdk.EventSessionStart] = true
	h.extensions[1].subscriptions[sdk.EventSessionStart] = true

	evt := sdk.Event{Type: sdk.EventSessionStart}
	responses, err := h.DispatchEvent(ctx, evt)
	if err != nil {
		t.Fatalf("DispatchEvent: %v", err)
	}
	if len(responses) != 2 {
		t.Errorf("expected 2 responses, got %d", len(responses))
	}
}

func TestHost_EchoWASM_SkipIfMissing(t *testing.T) {
	// This test uses the real echo.wasm built from testdata/echo/main.go with TinyGo.
	// Skip if the file doesn't exist.
	echoPath := filepath.Join("testdata", "echo.wasm")
	if _, err := os.Stat(echoPath); os.IsNotExist(err) {
		t.Skip("echo.wasm not found (build with: tinygo build -o testdata/echo.wasm -target wasi ./testdata/echo/)")
	}

	ctx := context.Background()
	h := NewHost(nil)
	defer h.Close(ctx)

	if err := h.Load(ctx, echoPath); err != nil {
		t.Fatalf("Load echo.wasm: %v", err)
	}
	if len(h.extensions) != 1 {
		t.Fatalf("expected 1 extension, got %d", len(h.extensions))
	}
}
