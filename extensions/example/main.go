//go:build wasip1

// Package main is the hello example extension for the bob coding harness.
//
// Build:
//
//	GOOS=wasip1 GOARCH=wasm go build -o hello.wasm .
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

// ---- Memory management ------------------------------------------------------

// pinned keeps allocations alive so the GC does not collect them before the
// host reads them. Entries are removed when the host calls _free.
var pinned = map[uintptr][]byte{}

//go:wasmexport _alloc
func extensionAlloc(size int32) int32 {
	if size <= 0 {
		return 0
	}
	buf := make([]byte, size)
	ptr := uintptr(unsafe.Pointer(&buf[0]))
	pinned[ptr] = buf
	return int32(ptr)
}

//go:wasmexport _free
func extensionFree(ptr int32) {
	delete(pinned, uintptr(ptr))
}

// ---- Required exports -------------------------------------------------------

//go:wasmexport _init
func extensionInit() int32 {
	if rc := subscribe("session_start"); rc != 0 {
		return rc
	}
	return 0
}

//go:wasmexport _on_event
func extensionOnEvent(ptr, length int32) int32 {
	data := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr))), length)

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

func hostCallJSON(method string, params any) int32 {
	type request struct {
		Method string `json:"method"`
		Params any    `json:"params,omitempty"`
	}
	reqBytes, err := json.Marshal(request{Method: method, Params: params})
	if err != nil {
		logMsg(3, "hello: marshal host_call request: "+err.Error())
		return 1
	}

	reqBuf := make([]byte, len(reqBytes))
	copy(reqBuf, reqBytes)
	reqPtr := uintptr(unsafe.Pointer(&reqBuf[0]))

	var respPtr, respLen uint32
	rc := hostCall(
		uint32(reqPtr), uint32(len(reqBuf)),
		uint32(uintptr(unsafe.Pointer(&respPtr))),
		uint32(uintptr(unsafe.Pointer(&respLen))),
	)
	if respPtr != 0 {
		delete(pinned, uintptr(respPtr))
	}
	return int32(rc)
}

func main() {}
