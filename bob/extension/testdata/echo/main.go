// Package main is a TinyGo echo extension for testing the extension host.
// Build with: tinygo build -o ../echo.wasm -target wasip1 .
package main

import (
	"encoding/json"
	"unsafe"
)

//go:wasmimport env host_log
func hostLog(level, ptr, length uint32)

//go:wasmimport env host_call
func hostCall(reqPtr, reqLen, respPtrPtr, respLenPtr uint32) uint32

// allocBuf is a simple bump-allocator buffer for small allocations.
var allocBuf [65536]byte
var allocOff uint32

//export _alloc
func alloc(size uint32) uint32 {
	if allocOff+size > uint32(len(allocBuf)) {
		allocOff = 0 // wrap around (simple bump)
	}
	ptr := allocOff
	allocOff += size
	return ptr
}

//export _free
func free(ptr uint32) {
	// No-op: bump allocator doesn't free individually.
}

//export _init
func initialize() int32 {
	msg := "echo extension initializing"
	ptr := alloc(uint32(len(msg)))
	copy(allocBuf[ptr:], msg)
	hostLog(1, ptr, uint32(len(msg)))

	return subscribeEvent("session_start")
}

//export _on_event
func onEvent(ptr, length uint32) uint32 {
	// Read event from memory.
	_ = unsafe.Slice(&allocBuf[0], len(allocBuf)) // ensure memory is live
	// Echo: return 0 (no response).
	return 0
}

func subscribeEvent(event string) int32 {
	type subscribeParams struct {
		Event string `json:"event"`
	}
	type req struct {
		Method string          `json:"method"`
		Params json.RawMessage `json:"params"`
	}

	params, err := json.Marshal(subscribeParams{Event: event})
	if err != nil {
		return 1
	}
	reqBytes, err := json.Marshal(req{Method: "subscribe", Params: params})
	if err != nil {
		return 1
	}

	reqPtr := alloc(uint32(len(reqBytes)))
	copy(allocBuf[reqPtr:], reqBytes)

	var respPtr, respLen uint32
	respPtrAddr := uint32(uintptr(unsafe.Pointer(&respPtr)))
	respLenAddr := uint32(uintptr(unsafe.Pointer(&respLen)))

	rc := hostCall(reqPtr, uint32(len(reqBytes)), respPtrAddr, respLenAddr)
	if respPtr != 0 {
		free(respPtr)
	}
	return int32(rc)
}

func main() {}
