//go:build tinygo

// Package main is the hello example extension for the bob coding harness.
//
// Build:
//
//	tinygo build -o hello.wasm -target wasi .
//
// Install:
//
//	cp hello.wasm "$BOB_EXTENSIONS_DIR/"
//
// What it does:
//   - Subscribes to session_start → logs a greeting and sets status "hello"="world"
package main

import (
	"encoding/json"
	"unsafe"
)

// ---- Host imports -----------------------------------------------------------

//go:wasmimport env host_log
func hostLog(level, ptr, length uint32)

//go:wasmimport env host_call
func hostCall(reqPtr, reqLen, respPtrPtr, respLenPtr uint32) uint32

// ---- Required exports -------------------------------------------------------

// _init is called once when the extension is loaded.
// It subscribes to lifecycle events.
//
//export _init
func _init() int32 {
	if rc := subscribe("session_start"); rc != 0 {
		return rc
	}
	return 0
}

// _on_event is called for every event the extension subscribed to.
// ptr points to a JSON-encoded sdk.Event in WASM memory; length is its byte length.
// Returns 0 (no response pointer) for all events in this example.
//
//export _on_event
func _on_event(ptr, length int32) int32 {
	data := readBytes(ptr, length)

	var evt struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(data, &evt); err != nil {
		logMsg(2, "hello: unmarshal event: "+err.Error())
		return 0
	}

	switch evt.Type {
	case "session_start":
		onSessionStart()
	}
	return 0
}

// _alloc allocates size bytes in WASM linear memory and returns the pointer.
//
//export _alloc
func _alloc(size int32) int32 {
	buf := make([]byte, size)
	return int32(uintptr(unsafe.Pointer(&buf[0])))
}

// _free frees a previously allocated pointer (no-op; GC handles it).
//
//export _free
func _free(_ int32) {}

// ---- Event handlers ---------------------------------------------------------

func onSessionStart() {
	logMsg(1, "hello extension: session started")
	setStatus("hello", "world")
}

// ---- Host call helpers ------------------------------------------------------

func subscribe(eventType string) int32 {
	type params struct {
		Event string `json:"event"`
	}
	return hostCallJSON("subscribe", params{Event: eventType})
}

func setStatus(key, value string) {
	type params struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	hostCallJSON("set_status", params{Key: key, Value: value})
}

func logMsg(level int, msg string) {
	b := []byte(msg)
	if len(b) == 0 {
		return
	}
	hostLog(uint32(level), uint32(uintptr(unsafe.Pointer(&b[0]))), uint32(len(b)))
}

// hostCallJSON marshals params to JSON and issues a host_call RPC.
// Returns 0 on success, non-zero on error.
func hostCallJSON(method string, params interface{}) int32 {
	type request struct {
		Method string      `json:"method"`
		Params interface{} `json:"params,omitempty"`
	}
	reqBytes, err := json.Marshal(request{Method: method, Params: params})
	if err != nil {
		logMsg(3, "hello: marshal host_call request: "+err.Error())
		return 1
	}

	// Copy request into WASM memory.
	reqPtr := _alloc(int32(len(reqBytes)))
	reqMem := (*[1 << 28]byte)(unsafe.Pointer(uintptr(reqPtr)))
	copy(reqMem[:len(reqBytes)], reqBytes)

	// Allocate out-params for response pointer and length.
	var respPtr uint32
	var respLen uint32
	rc := hostCall(
		uint32(reqPtr), uint32(len(reqBytes)),
		uint32(uintptr(unsafe.Pointer(&respPtr))),
		uint32(uintptr(unsafe.Pointer(&respLen))),
	)
	_free(reqPtr)

	if rc != 0 {
		return int32(rc)
	}
	if respPtr != 0 && respLen > 0 {
		_free(int32(respPtr))
	}
	return 0
}

// ---- Memory helpers ---------------------------------------------------------

// readBytes reads length bytes from WASM memory starting at ptr.
func readBytes(ptr, length int32) []byte {
	if length <= 0 {
		return nil
	}
	mem := (*[1 << 28]byte)(unsafe.Pointer(uintptr(ptr)))
	out := make([]byte, length)
	copy(out, mem[:length])
	return out
}

// main is required by TinyGo WASI target but never called.
func main() {}
