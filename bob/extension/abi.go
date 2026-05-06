package extension

// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

// This file documents the WASM ABI constants used by the extension host.
// The actual host export functions are methods on Host defined in host.go.

// Required extension exports (must be present in every .wasm extension):
//
//	_init() int32                        — called once on load; 0=success
//	_on_event(ptr int32, len int32) int32 — dispatch event; returns resp_ptr or 0
//	_alloc(size int32) int32             — allocate size bytes; returns ptr
//	_free(ptr int32)                     — free ptr
//
// Host imports (module "env"):
//
//	host_log(level, ptr, len)            — level: 0=debug 1=info 2=warn 3=error
//	host_alloc(size) int32               — unused in v1; always returns 0
//	host_free(ptr)                       — no-op in v1
//	host_call(req_ptr, req_len,          — synchronous RPC; 0=success
//	          resp_ptr_ptr, resp_len_ptr) int32
