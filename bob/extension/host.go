package extension

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/mattdurham/bob/bob/sdk"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// Extension wraps a loaded WASM module.
type Extension struct {
	name   string
	module api.Module
	store  *Store

	subMu         sync.RWMutex
	subscriptions map[sdk.EventType]bool
}

// Host manages a collection of WASM extensions.
type Host struct {
	mu         sync.RWMutex
	runtime    wazero.Runtime
	extensions []*Extension
	logFn      func(level int, msg string)

	// Registered tools keyed by name (for duplicate detection).
	registeredTools map[string]sdk.Tool

	// Callbacks set by the harness.
	OnSendMessage     func(msg sdk.Message)
	OnSetStatus       func(key, value string)
	OnRegisterTool    func(tool sdk.Tool) error
	OnRegisterCommand func(name string, desc string)
	OnNotify          func(text string)
	OnAbort           func()
	OnToolResult      func(toolCallID, result string, isError bool)
}

// NewHost creates a Host and installs the "env" host module into a fresh wazero runtime.
func NewHost(logFn func(level int, msg string)) *Host {
	if logFn == nil {
		logFn = func(_ int, msg string) { fmt.Fprintln(os.Stderr, msg) }
	}
	h := &Host{
		logFn:           logFn,
		registeredTools: make(map[string]sdk.Tool),
	}
	h.runtime = wazero.NewRuntime(context.Background())
	if err := h.installEnvModule(); err != nil {
		// If "env" fails to install, log and continue — extensions without
		// imports will still work.
		logFn(3, fmt.Sprintf("extension: install env module: %v", err))
	}
	return h
}

// installEnvModule registers the "env" host module that extensions import.
func (h *Host) installEnvModule() error {
	ctx := context.Background()
	_, err := h.runtime.NewHostModuleBuilder("env").
		NewFunctionBuilder().
		WithFunc(h.hostLogImpl).
		Export("host_log").
		NewFunctionBuilder().
		WithFunc(h.hostAllocImpl).
		Export("host_alloc").
		NewFunctionBuilder().
		WithFunc(h.hostFreeImpl).
		Export("host_free").
		NewFunctionBuilder().
		WithFunc(h.hostCallImpl).
		Export("host_call").
		Instantiate(ctx)
	return err
}

// hostLogImpl is the host_log import: host_log(level, ptr, length).
func (h *Host) hostLogImpl(ctx context.Context, m api.Module, level, ptr, length uint32) {
	mem := m.Memory()
	if mem == nil {
		return
	}
	bs, ok := mem.Read(ptr, length)
	if !ok {
		h.logFn(3, "extension: host_log: invalid memory read")
		return
	}
	h.logFn(int(level), string(bs))
}

// hostAllocImpl is the host_alloc import: unused in v1, always returns 0.
func (h *Host) hostAllocImpl(_ context.Context, _ api.Module, _ uint32) uint32 {
	return 0
}

// hostFreeImpl is the host_free import: no-op in v1.
func (h *Host) hostFreeImpl(_ context.Context, _ api.Module, _ uint32) {}

// hostCallImpl is the host_call import.
// host_call(req_ptr, req_len, resp_ptr_ptr, resp_len_ptr) -> status
func (h *Host) hostCallImpl(ctx context.Context, m api.Module, reqPtr, reqLen, respPtrPtr, respLenPtr uint32) uint32 {
	mem := m.Memory()
	if mem == nil {
		return uint32(sdk.ErrGeneral)
	}

	// Read request JSON from WASM memory.
	reqBytes, ok := mem.Read(reqPtr, reqLen)
	if !ok {
		h.logFn(3, "extension: host_call: invalid request memory read")
		return uint32(sdk.ErrGeneral)
	}

	var req sdk.HostCallRequest
	if err := json.Unmarshal(reqBytes, &req); err != nil {
		h.logFn(3, fmt.Sprintf("extension: host_call: unmarshal request: %v", err))
		return uint32(sdk.ErrGeneral)
	}

	// Find the calling extension.
	ext := h.findExtensionByModule(m)

	// Dispatch to the router.
	resp := h.routeHostCall(ctx, m, ext, req)

	// Write response back into WASM memory if caller wants it.
	if respPtrPtr == 0 && respLenPtr == 0 {
		return uint32(sdk.ErrOK)
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		h.logFn(3, fmt.Sprintf("extension: host_call: marshal response: %v", err))
		return uint32(sdk.ErrGeneral)
	}

	// Allocate memory in the extension's WASM module for the response.
	allocFn := m.ExportedFunction("_alloc")
	if allocFn == nil {
		return uint32(sdk.ErrGeneral)
	}
	allocResult, err := allocFn.Call(ctx, uint64(len(respBytes)))
	if err != nil || len(allocResult) == 0 {
		h.logFn(3, fmt.Sprintf("extension: host_call: _alloc failed: %v", err))
		return uint32(sdk.ErrGeneral)
	}
	respPtr := uint32(allocResult[0])
	if respPtr == 0 {
		// Extension's _alloc returned 0 — can't write response.
		return uint32(sdk.ErrOK)
	}

	if !mem.Write(respPtr, respBytes) {
		h.logFn(3, "extension: host_call: write response to WASM memory failed")
		return uint32(sdk.ErrGeneral)
	}

	// Write respPtr and respLen into the caller-supplied pointer slots.
	if !mem.WriteUint32Le(respPtrPtr, respPtr) {
		return uint32(sdk.ErrGeneral)
	}
	if !mem.WriteUint32Le(respLenPtr, uint32(len(respBytes))) {
		return uint32(sdk.ErrGeneral)
	}

	return uint32(sdk.ErrOK)
}

// routeHostCall dispatches req to the appropriate handler and returns a response.
func (h *Host) routeHostCall(ctx context.Context, _ api.Module, ext *Extension, req sdk.HostCallRequest) sdk.HostCallResponse {
	switch req.Method {
	case sdk.MethodSubscribe:
		return h.handleSubscribe(ext, req)

	case sdk.MethodRegisterTool:
		return h.handleRegisterTool(req)

	case sdk.MethodRegisterCommand:
		return h.handleRegisterCommand(req)

	case sdk.MethodSendMessage:
		return h.handleSendMessage(req)

	case sdk.MethodSetStatus:
		return h.handleSetStatus(req)

	case sdk.MethodNotify:
		return h.handleNotify(req)

	case sdk.MethodToolResult:
		return h.handleToolResult(req)

	case sdk.MethodStoreSet:
		return h.handleStoreSet(ext, req)

	case sdk.MethodStoreGet:
		return h.handleStoreGet(ext, req)

	case sdk.MethodAbort:
		if h.OnAbort != nil {
			h.OnAbort()
		}
		return sdk.HostCallResponse{}

	default:
		return sdk.HostCallResponse{Error: fmt.Sprintf("unknown method: %s", req.Method)}
	}
}

func (h *Host) handleSubscribe(ext *Extension, req sdk.HostCallRequest) sdk.HostCallResponse {
	if ext == nil {
		return sdk.HostCallResponse{Error: "subscribe: unknown extension"}
	}
	var params struct {
		Event sdk.EventType `json:"event"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return sdk.HostCallResponse{Error: fmt.Sprintf("subscribe: %v", err)}
	}
	ext.subMu.Lock()
	ext.subscriptions[params.Event] = true
	ext.subMu.Unlock()
	return sdk.HostCallResponse{}
}

func (h *Host) handleRegisterTool(req sdk.HostCallRequest) sdk.HostCallResponse {
	var tool sdk.Tool
	if err := json.Unmarshal(req.Params, &tool); err != nil {
		return sdk.HostCallResponse{Error: fmt.Sprintf("register_tool: %v", err)}
	}
	h.mu.Lock()
	_, exists := h.registeredTools[tool.Name]
	if !exists {
		h.registeredTools[tool.Name] = tool
	}
	h.mu.Unlock()
	if exists {
		return sdk.HostCallResponse{Error: fmt.Sprintf("tool already registered: %s", tool.Name)}
	}
	if h.OnRegisterTool != nil {
		if err := h.OnRegisterTool(tool); err != nil {
			return sdk.HostCallResponse{Error: err.Error()}
		}
	}
	return sdk.HostCallResponse{}
}

func (h *Host) handleRegisterCommand(req sdk.HostCallRequest) sdk.HostCallResponse {
	var params struct {
		Name string `json:"name"`
		Desc string `json:"description"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return sdk.HostCallResponse{Error: fmt.Sprintf("register_command: %v", err)}
	}
	if h.OnRegisterCommand != nil {
		h.OnRegisterCommand(params.Name, params.Desc)
	}
	return sdk.HostCallResponse{}
}

func (h *Host) handleSendMessage(req sdk.HostCallRequest) sdk.HostCallResponse {
	var msg sdk.Message
	if err := json.Unmarshal(req.Params, &msg); err != nil {
		return sdk.HostCallResponse{Error: fmt.Sprintf("send_message: %v", err)}
	}
	if h.OnSendMessage != nil {
		h.OnSendMessage(msg)
	}
	return sdk.HostCallResponse{}
}

func (h *Host) handleSetStatus(req sdk.HostCallRequest) sdk.HostCallResponse {
	var params struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return sdk.HostCallResponse{Error: fmt.Sprintf("set_status: %v", err)}
	}
	if h.OnSetStatus != nil {
		h.OnSetStatus(params.Key, params.Value)
	}
	return sdk.HostCallResponse{}
}

func (h *Host) handleNotify(req sdk.HostCallRequest) sdk.HostCallResponse {
	var params struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return sdk.HostCallResponse{Error: fmt.Sprintf("notify: %v", err)}
	}
	if h.OnNotify != nil {
		h.OnNotify(params.Text)
	}
	return sdk.HostCallResponse{}
}

func (h *Host) handleToolResult(req sdk.HostCallRequest) sdk.HostCallResponse {
	var params struct {
		ToolCallID string `json:"tool_call_id"`
		Result     string `json:"result"`
		IsError    bool   `json:"is_error"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return sdk.HostCallResponse{Error: fmt.Sprintf("tool_result: %v", err)}
	}
	if h.OnToolResult != nil {
		h.OnToolResult(params.ToolCallID, params.Result, params.IsError)
	}
	return sdk.HostCallResponse{}
}

func (h *Host) handleStoreSet(ext *Extension, req sdk.HostCallRequest) sdk.HostCallResponse {
	if ext == nil {
		return sdk.HostCallResponse{Error: "store_set: unknown extension"}
	}
	var params struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return sdk.HostCallResponse{Error: fmt.Sprintf("store_set: %v", err)}
	}
	ext.store.Set(params.Key, params.Value)
	return sdk.HostCallResponse{}
}

func (h *Host) handleStoreGet(ext *Extension, req sdk.HostCallRequest) sdk.HostCallResponse {
	if ext == nil {
		return sdk.HostCallResponse{Error: "store_get: unknown extension"}
	}
	var params struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return sdk.HostCallResponse{Error: fmt.Sprintf("store_get: %v", err)}
	}
	v, ok := ext.store.Get(params.Key)
	if !ok {
		return sdk.HostCallResponse{Error: "not found"}
	}
	result, _ := json.Marshal(map[string]string{"value": v})
	return sdk.HostCallResponse{Result: json.RawMessage(result)}
}

// findExtensionByModule returns the Extension whose module has the given name.
func (h *Host) findExtensionByModule(m api.Module) *Extension {
	h.mu.RLock()
	defer h.mu.RUnlock()
	name := m.Name()
	for _, ext := range h.extensions {
		if ext.module.Name() == name {
			return ext
		}
	}
	return nil
}

// Load reads the WASM file at path, validates it, calls _init, and registers it.
func (h *Host) Load(ctx context.Context, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("load extension %s: %w", path, err)
	}

	// Use file basename as the module name to avoid conflicts.
	modName := moduleNameFromPath(path)

	mod, err := h.runtime.InstantiateWithConfig(ctx, data,
		wazero.NewModuleConfig().WithName(modName).WithStartFunctions())
	if err != nil {
		return fmt.Errorf("instantiate extension %s: %w", path, err)
	}

	if err := validateExports(mod); err != nil {
		_ = mod.Close(ctx)
		return fmt.Errorf("validate extension %s: %w", path, err)
	}

	ext := &Extension{
		name:          modName,
		module:        mod,
		subscriptions: make(map[sdk.EventType]bool),
		store:         NewStore(),
	}

	// Register ext before calling _init so host_call works.
	h.mu.Lock()
	h.extensions = append(h.extensions, ext)
	h.mu.Unlock()

	if err := callInit(ctx, mod); err != nil {
		// Remove on _init failure.
		h.mu.Lock()
		h.removeExtension(ext)
		h.mu.Unlock()
		_ = mod.Close(ctx)
		return fmt.Errorf("init extension %s: %w", path, err)
	}

	return nil
}

// DispatchEvent dispatches evt to all subscribed extensions and returns their responses.
// A WASM trap (error from _on_event) is logged and does not stop dispatch to other extensions.
func (h *Host) DispatchEvent(ctx context.Context, evt sdk.Event) ([]sdk.EventResponse, error) {
	evtJSON, err := json.Marshal(evt)
	if err != nil {
		return nil, fmt.Errorf("marshal event: %w", err)
	}

	h.mu.RLock()
	exts := make([]*Extension, len(h.extensions))
	copy(exts, h.extensions)
	h.mu.RUnlock()

	var responses []sdk.EventResponse
	for _, ext := range exts {
		ext.subMu.RLock()
		subscribed := ext.subscriptions[evt.Type]
		ext.subMu.RUnlock()
		if !subscribed {
			continue
		}
		resp, dispErr := h.dispatchToExtension(ctx, ext, evtJSON)
		if dispErr != nil {
			h.logFn(2, fmt.Sprintf("extension %s: dispatch error: %v", ext.name, dispErr))
			continue
		}
		responses = append(responses, resp)
	}
	return responses, nil
}

// dispatchToExtension calls _on_event on a single extension.
func (h *Host) dispatchToExtension(ctx context.Context, ext *Extension, evtJSON []byte) (sdk.EventResponse, error) {
	mod := ext.module
	mem := mod.Memory()
	if mem == nil {
		return sdk.EventResponse{}, fmt.Errorf("extension %s has no memory", ext.name)
	}

	// Allocate memory in the extension for the event JSON.
	allocFn := mod.ExportedFunction("_alloc")
	if allocFn == nil {
		return sdk.EventResponse{}, fmt.Errorf("extension %s missing _alloc", ext.name)
	}
	allocResult, err := allocFn.Call(ctx, uint64(len(evtJSON)))
	if err != nil {
		return sdk.EventResponse{}, fmt.Errorf("extension %s _alloc: %w", ext.name, err)
	}
	if len(allocResult) == 0 {
		return sdk.EventResponse{}, fmt.Errorf("extension %s _alloc returned no results", ext.name)
	}
	evtPtr := uint32(allocResult[0])

	if evtPtr == 0 {
		// Extension's _alloc returned 0 — can't safely deliver event.
		return sdk.EventResponse{}, nil
	}

	if !mem.Write(evtPtr, evtJSON) {
		return sdk.EventResponse{}, fmt.Errorf("extension %s: write event to memory failed", ext.name)
	}

	// Call _on_event(ptr, len) → resp_ptr.
	onEvent := mod.ExportedFunction("_on_event")
	if onEvent == nil {
		return sdk.EventResponse{}, fmt.Errorf("extension %s missing _on_event", ext.name)
	}
	results, err := onEvent.Call(ctx, uint64(evtPtr), uint64(len(evtJSON)))
	if err != nil {
		return sdk.EventResponse{}, fmt.Errorf("extension %s _on_event trap: %w", ext.name, err)
	}

	// Free the event memory.
	if freeFn := mod.ExportedFunction("_free"); freeFn != nil {
		_, _ = freeFn.Call(ctx, uint64(evtPtr))
	}

	if len(results) == 0 || results[0] == 0 {
		return sdk.EventResponse{}, nil
	}

	// Read response JSON from WASM memory.
	respPtr := uint32(results[0])

	// Read length: we don't know the length directly, so read until null byte or
	// use a size prefix. Per the ABI, the extension must ensure the response is
	// valid JSON. We read up to 64KB and find the JSON boundary.
	// Simpler: the extension stores resp JSON starting at respPtr; scan for end.
	// We read a reasonable max (64KB) and try to unmarshal.
	respBytes := readNullTerminatedOrJSON(mem, respPtr)

	freeFn := mod.ExportedFunction("_free")
	if freeFn != nil {
		_, _ = freeFn.Call(ctx, uint64(respPtr))
	}

	var resp sdk.EventResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		h.logFn(2, fmt.Sprintf("extension %s: unmarshal _on_event response: %v", ext.name, err))
		return sdk.EventResponse{}, nil
	}
	return resp, nil
}

// readNullTerminatedOrJSON reads bytes from WASM memory starting at ptr,
// trying to find a complete JSON object. Returns at most 64KB.
// It correctly handles braces inside JSON string values.
func readNullTerminatedOrJSON(mem api.Memory, ptr uint32) []byte {
	const maxLen = 65536
	// Read max available.
	avail, ok := mem.Read(ptr, maxLen)
	if !ok {
		// Try smaller reads if at end of memory.
		for size := uint32(maxLen / 2); size >= 1; size /= 2 {
			avail, ok = mem.Read(ptr, size)
			if ok {
				break
			}
		}
	}
	if len(avail) == 0 {
		return nil
	}
	// Find the end of the JSON object, skipping characters inside string literals.
	var depth int
	inString := false
	escaped := false
	for i, b := range avail {
		if escaped {
			escaped = false
			continue
		}
		if b == '\\' && inString {
			escaped = true
			continue
		}
		if b == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		switch b {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return avail[:i+1]
			}
		}
	}
	return avail
}

// Reload closes all extensions and reloads them from the given paths.
func (h *Host) Reload(ctx context.Context, paths []string) error {
	h.mu.Lock()
	old := h.extensions
	h.extensions = nil
	h.mu.Unlock()

	for _, ext := range old {
		if err := ext.module.Close(ctx); err != nil {
			h.logFn(2, fmt.Sprintf("extension %s: close: %v", ext.name, err))
		}
	}

	for _, path := range paths {
		if err := h.Load(ctx, path); err != nil {
			h.logFn(3, fmt.Sprintf("reload: %v", err))
		}
	}
	return nil
}

// Close shuts down all extensions and the wazero runtime.
func (h *Host) Close(ctx context.Context) error {
	return h.runtime.Close(ctx)
}

// removeExtension removes ext from h.extensions. Caller must hold h.mu.Lock().
func (h *Host) removeExtension(target *Extension) {
	filtered := make([]*Extension, 0, len(h.extensions))
	for _, ext := range h.extensions {
		if ext != target {
			filtered = append(filtered, ext)
		}
	}
	h.extensions = filtered
}

// moduleNameFromPath derives a unique module name from a file path.
func moduleNameFromPath(path string) string {
	// Use the full path to avoid name collisions.
	return path
}
