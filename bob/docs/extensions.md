# Extensions — WASM Extension Author API Reference

Bob extensions are WebAssembly modules loaded at startup. They receive lifecycle
events from the harness, can register tools and slash commands, and communicate
with the host via a synchronous JSON-RPC mechanism (`host_call`).

---

## Overview

- Extensions are `.wasm` files placed in `BOB_EXTENSIONS_DIR`.
- The host loads them with [wazero](https://github.com/tetratelabs/wazero) — a
  pure-Go, zero-dependency WebAssembly runtime.
- Extensions must be compiled to the **WASI** target. The reference toolchain is
  [TinyGo](https://tinygo.org/) (`tinygo build -target wasi`).
- All communication between host and extension happens through shared linear
  memory and four host import functions (see below).
- Extensions are isolated: each has its own WASM linear memory and its own
  in-process key-value store.

---

## Required Exports

Every `.wasm` extension **must** export exactly these four functions. The host
validates their presence at load time and refuses to load any module that is
missing one.

### `_init() int32`

Called once immediately after the module is instantiated, before any events are
dispatched.

- Use `_init` to subscribe to events (`host_call subscribe`), register tools,
  and register slash commands.
- Return `0` on success. Any non-zero return code is treated as a fatal
  initialisation error; the extension is unloaded.

### `_on_event(ptr int32, len int32) int32`

Called by the host for each event the extension has subscribed to.

- `ptr` — pointer into the extension's linear memory where the JSON-encoded
  `sdk.Event` has been written.
- `len` — byte length of the JSON payload.
- Return `0` if the extension has no response (the common case).
- Return a non-zero pointer to a JSON-encoded `EventResponse` (also in the
  extension's linear memory) to cancel or block the event, or to signal an
  error. The host calls `_free` on this pointer after reading the response.

### `_alloc(size int32) int32`

Allocate `size` bytes in WASM linear memory and return the pointer.

- The host calls `_alloc` to obtain memory before writing an event payload or
  a `host_call` response into the extension's address space.
- Must return a valid pointer, or `0` to signal out-of-memory (the host will
  treat the allocation as failed).
- TinyGo example: `return int32(uintptr(unsafe.Pointer(&make([]byte, size)[0])))`

### `_free(ptr int32)`

Free a previously allocated pointer.

- Called by the host after it has finished reading a pointer that was either
  passed to `_on_event` or returned from `_on_event`.
- In TinyGo with GC enabled this can be a no-op — the GC reclaims memory.

---

## Host Imports (module `"env"`)

Extensions declare these as `//go:wasmimport env <name>` in TinyGo (or via
their toolchain's equivalent mechanism). All four are provided by the host
module named `"env"`.

### `host_log(level uint32, ptr uint32, length uint32)`

Write a log message.

| Level | Meaning |
|-------|---------|
| 0     | debug   |
| 1     | info    |
| 2     | warn    |
| 3     | error   |

`ptr` + `length` describe a UTF-8 string in the extension's linear memory.

### `host_alloc(size uint32) uint32`

Reserved for future use. In v1 this always returns `0`. Extensions should not
call it.

### `host_free(ptr uint32)`

No-op in v1. Reserved for future use.

### `host_call(req_ptr uint32, req_len uint32, resp_ptr_ptr uint32, resp_len_ptr uint32) uint32`

Synchronous JSON-RPC call from extension to host.

- `req_ptr` / `req_len` — pointer and byte length of a JSON-encoded
  `HostCallRequest` in the extension's linear memory.
- `resp_ptr_ptr` — pointer to a `uint32` slot where the host writes the pointer
  to the response bytes (allocated via `_alloc` in the extension).
- `resp_len_ptr` — pointer to a `uint32` slot where the host writes the byte
  length of the response.
- Returns an error code (see Error Codes below).

Pass `0` for both `resp_ptr_ptr` and `resp_len_ptr` if no response is needed.

---

## host_call Method Reference

`HostCallRequest` JSON envelope:

```json
{"method": "<method_name>", "params": { ... }}
```

`HostCallResponse` JSON envelope (written into extension memory via `_alloc`):

```json
{"result": { ... }, "error": "<error string if any>"}
```

An empty `"error"` field (or its absence) means success.

---

### `subscribe`

Register interest in a lifecycle event. Must be called from `_init`.

```json
{"method": "subscribe", "params": {"event": "session_start"}}
```

| Field   | Type   | Description                          |
|---------|--------|--------------------------------------|
| `event` | string | One of the event type strings below. |

No response result.

---

### `register_tool`

Register a tool that the LLM may call. The host forwards tool calls back to the
extension as `on_tool_call` events.

```json
{
  "method": "register_tool",
  "params": {
    "name": "my_tool",
    "description": "Human-readable description for the LLM.",
    "input_schema": {
      "type": "object",
      "properties": {
        "query": {"type": "string", "description": "The search query"}
      },
      "required": ["query"]
    }
  }
}
```

| Field          | Type   | Description                                     |
|----------------|--------|-------------------------------------------------|
| `name`         | string | Unique tool name. Duplicate registration fails. |
| `description`  | string | Shown to the LLM to explain tool purpose.       |
| `input_schema` | object | JSON Schema object for the tool's input.        |

No response result. Returns an error if the tool name is already registered.

---

### `register_command`

Register a slash command visible in the TUI.

```json
{"method": "register_command", "params": {"name": "greet", "description": "Say hello"}}
```

| Field         | Type   | Description              |
|---------------|--------|--------------------------|
| `name`        | string | Command name (no slash). |
| `description` | string | Shown in `/help` output. |

No response result.

---

### `send_message`

Inject a message into the conversation as if typed by the user or generated by
the assistant. The harness will trigger a new provider request.

```json
{"method": "send_message", "params": {"role": "user", "content": "Hello!"}}
```

| Field     | Type   | Description                  |
|-----------|--------|------------------------------|
| `role`    | string | `"user"` or `"assistant"`.   |
| `content` | string | The message text.            |

No response result.

---

### `set_status`

Set a keyed value displayed in the status bar.

```json
{"method": "set_status", "params": {"key": "my_ext", "value": "active"}}
```

| Field   | Type   | Description                      |
|---------|--------|----------------------------------|
| `key`   | string | Identifier for the status entry. |
| `value` | string | Display value.                   |

No response result.

---

### `notify`

Append a notification line to the chat view (rendered as a system message).

```json
{"method": "notify", "params": {"text": "Background task complete."}}
```

| Field  | Type   | Description       |
|--------|--------|-------------------|
| `text` | string | Notification text.|

No response result.

---

### `tool_result`

Return the result of a tool call to the harness. Call this after processing an
`on_tool_call` event.

```json
{
  "method": "tool_result",
  "params": {
    "tool_call_id": "toolu_abc123",
    "result": "Paris",
    "is_error": false
  }
}
```

| Field          | Type    | Description                                         |
|----------------|---------|-----------------------------------------------------|
| `tool_call_id` | string  | The `tool_call_id` from the `on_tool_call` payload. |
| `result`       | string  | The tool output (text).                             |
| `is_error`     | boolean | `true` if the tool encountered an error.            |

No response result.

---

### `store_set`

Persist a string value in the extension's private key-value store. The store
survives for the lifetime of the process.

```json
{"method": "store_set", "params": {"key": "last_query", "value": "Paris"}}
```

| Field   | Type   | Description |
|---------|--------|-------------|
| `key`   | string | Store key.  |
| `value` | string | Store value.|

No response result.

---

### `store_get`

Retrieve a value from the extension's private key-value store.

```json
{"method": "store_get", "params": {"key": "last_query"}}
```

| Field | Type   | Description |
|-------|--------|-------------|
| `key` | string | Store key.  |

Response result on success:

```json
{"value": "Paris"}
```

Returns an error (`"not found"`) if the key does not exist.

---

### `abort`

Cancel the in-progress provider stream immediately. Equivalent to the user
pressing Ctrl+C.

```json
{"method": "abort", "params": {}}
```

No response result.

---

## Lifecycle Events

Events are dispatched to subscribed extensions via `_on_event`. The `sdk.Event`
JSON envelope is:

```json
{"type": "<event_type>", "payload": { ... }}
```

---

### `session_start`

Fired once when the TUI initialises a new session.

```json
{"reason": "new_session"}
```

| Field    | Type   | Description                    |
|----------|--------|--------------------------------|
| `reason` | string | Always `"new_session"` in v1.  |

---

### `before_agent_start`

Fired when the user submits a message, before any provider request is made.

```json
{"prompt": "What is the capital of France?", "system_prompt": ""}
```

| Field           | Type   | Description                       |
|-----------------|--------|-----------------------------------|
| `prompt`        | string | The raw user input.               |
| `system_prompt` | string | System prompt (may be empty).     |

---

### `before_provider_request`

Fired immediately before the request is sent to the LLM.

```json
{
  "messages": [{"role": "user", "content": "What is the capital?"}],
  "model": "claude-sonnet-4-5"
}
```

| Field      | Type            | Description                   |
|------------|-----------------|-------------------------------|
| `messages` | array of Message| Full conversation history.    |
| `model`    | string          | Active model identifier.      |

---

### `after_provider_response`

Fired after the provider stream completes successfully.

```json
{"usage": {"input_tokens": 0, "output_tokens": 0}}
```

| Field                    | Type | Description                       |
|--------------------------|------|-----------------------------------|
| `usage.input_tokens`     | int  | Input token count (when provided).|
| `usage.output_tokens`    | int  | Output token count (when provided).|

---

### `on_tool_call`

Fired when the LLM requests a tool call. The extension that registered the tool
should process this event and respond with `tool_result`.

```json
{
  "tool_call_id": "toolu_abc123",
  "tool_name": "my_tool",
  "input": {"query": "Paris"}
}
```

| Field          | Type   | Description                                 |
|----------------|--------|---------------------------------------------|
| `tool_call_id` | string | Opaque ID; must be returned in `tool_result`.|
| `tool_name`    | string | Name of the tool being called.              |
| `input`        | object | Raw JSON input matching the tool's schema.  |

---

### `on_tool_result`

Fired after a tool result has been submitted back to the provider.

```json
{
  "tool_call_id": "toolu_abc123",
  "result": "Paris",
  "is_error": false
}
```

| Field          | Type    | Description                  |
|----------------|---------|------------------------------|
| `tool_call_id` | string  | The corresponding call ID.   |
| `result`       | string  | The tool output.             |
| `is_error`     | boolean | Whether the tool errored.    |

---

### `message_start`

Fired when a new message begins streaming from the provider.

```json
{"role": "assistant"}
```

| Field  | Type   | Description                    |
|--------|--------|--------------------------------|
| `role` | string | `"user"` or `"assistant"`.     |

---

### `message_end`

Fired when a message has finished streaming.

```json
{"role": "assistant", "content": "The capital of France is Paris."}
```

| Field     | Type   | Description                         |
|-----------|--------|-------------------------------------|
| `role`    | string | `"user"` or `"assistant"`.          |
| `content` | string | Full content of the completed message.|

---

### `shutdown`

Fired when the harness is shutting down. Use this for cleanup.

```json
{"reason": "user_quit"}
```

| Field    | Type   | Description             |
|----------|--------|-------------------------|
| `reason` | string | Reason for shutdown.    |

---

## EventResponse

`_on_event` may return `0` (no response) or a pointer to a JSON-encoded
`EventResponse` object.

```json
{"cancel": false, "block": false, "error": ""}
```

| Field    | Type    | Semantics                                                          |
|----------|---------|--------------------------------------------------------------------|
| `cancel` | boolean | Cancel the current operation (e.g. abort a stream in progress).   |
| `block`  | boolean | Block the event from being processed further (reserved).          |
| `error`  | string  | Non-empty string signals an error; displayed as a notification.   |

All fields are optional and default to their zero values. Return `0` (null
pointer) from `_on_event` when no response is needed — this is the common case.

---

## Memory Management Protocol

Understanding the ownership contract prevents use-after-free and memory leaks.

### Event dispatch (`_on_event`)

1. Host calls `_alloc(len)` to allocate space in the extension's memory.
2. Host writes the event JSON at the returned pointer.
3. Host calls `_on_event(ptr, len)`.
4. Host calls `_free(ptr)` to release the event buffer.
5. If `_on_event` returns a non-zero pointer, the host reads the response JSON
   from that pointer, then calls `_free(resp_ptr)`.

**Extension owns the response buffer** it returns from `_on_event`. The host
frees it after reading.

### host_call response

1. Extension allocates `req` buffer via `_alloc`, writes the request JSON.
2. Extension calls `host_call(req_ptr, req_len, &resp_ptr, &resp_len)`.
3. Host allocates a response buffer inside the extension via `_alloc(resp_len)`.
4. Host writes response JSON and stores the pointer in `*resp_ptr_ptr` and
   length in `*resp_len_ptr`.
5. `host_call` returns.
6. **Extension is responsible for calling `_free(resp_ptr)`** after reading the
   response. The host does not free it.
7. Extension calls `_free(req_ptr)` to release the request buffer.

---

## Error Codes

Returned by `host_call` as a `uint32`.

| Constant     | Value | Meaning                                                |
|--------------|-------|--------------------------------------------------------|
| `ErrOK`      | `0`   | Success.                                               |
| `ErrGeneral` | `1`   | General error (check `HostCallResponse.error` for msg).|
| `ErrCancel`  | `2`   | Operation cancelled.                                   |

---

## Build and Install

### Prerequisites

- [TinyGo](https://tinygo.org/) 0.30+ (for WASI target support and `//go:wasmimport`)
- Standard Go module setup

### Build

```bash
# From the extension directory
tinygo build -o hello.wasm -target wasi .
```

### Install

```bash
# Copy to the extensions directory
cp hello.wasm "$BOB_EXTENSIONS_DIR/"
```

Bob scans `BOB_EXTENSIONS_DIR` at startup and loads all `.wasm` files found
there. Use `/reload` in the TUI to hot-reload extensions without restarting.

---

## Annotated Example

The `extensions/example/` directory contains a fully working extension. It
demonstrates the minimal pattern: subscribe in `_init`, handle an event in
`_on_event`, and make `host_call` calls via a helper.

Key points from `extensions/example/main.go`:

- The `//go:build tinygo` build tag ensures the file is only compiled with
  TinyGo.
- Host imports are declared with `//go:wasmimport env <name>` directives.
- Required exports are annotated with `//export <name>` directives.
- `_alloc` uses `unsafe.Pointer` on a Go slice to hand memory to the host.
- `_free` is a no-op because TinyGo's GC handles reclamation.
- `hostCallJSON` is a convenience wrapper that marshals params, writes them into
  WASM memory, calls `host_call`, and frees buffers on return.
- The extension subscribes only to `session_start` and calls `set_status` to
  update the status bar when the session begins.
