# bob/sdk — Interface Contracts and Behavioral Invariants

## ABIVersion

- `ABIVersion = 1` (int, untyped constant)
- Extensions that export `_abi_version() i32` must return this value when strict version checking is enabled.
- Strict version checking is optional in v1; the host may accept extensions that omit the export.

---

## EventType Constants

`EventType` is a `string` typedef. All values are stable across ABI versions — the host and extensions
compare them as plain strings; no numeric mapping exists.

| Constant                      | Wire value                  | When dispatched                                              |
|-------------------------------|-----------------------------|--------------------------------------------------------------|
| `EventSessionStart`           | `"session_start"`           | A new Claude session begins                                   |
| `EventBeforeAgentStart`       | `"before_agent_start"`      | An agent is about to start (prompt + system prompt available) |
| `EventBeforeProviderRequest`  | `"before_provider_request"` | The host is about to call the LLM provider                   |
| `EventAfterProviderResponse`  | `"after_provider_response"` | The LLM provider has returned a response                     |
| `EventOnToolCall`             | `"on_tool_call"`            | The LLM has emitted a tool-call                              |
| `EventOnToolResult`           | `"on_tool_result"`          | A tool result has been produced                              |
| `EventMessageStart`           | `"message_start"`           | A new message stream begins                                  |
| `EventMessageEnd`             | `"message_end"`             | A message stream has completed                               |
| `EventShutdown`               | `"shutdown"`                | The host is shutting down; extensions should flush state     |

**Invariants:**
- The set of `EventType` string values must not change between ABI versions without a version bump.
- An unknown `EventType` must be silently ignored by extensions (forward-compatibility).
- There are exactly 9 defined event types.

---

## Event / EventResponse JSON Contract

### Event

```json
{ "type": "<EventType>", "payload": <raw-JSON-or-null> }
```

- `type` — required; one of the `EventType` string constants.
- `payload` — a `json.RawMessage`; always valid JSON (object, array, or null). Never an empty byte slice.
  The payload shape is determined by the `type` field and must match the corresponding payload struct.

### EventResponse

```json
{ "cancel": true, "block": true, "error": "message" }
```

All fields are `omitempty`. An empty `{}` response (or a nil response) is valid and means "no action".

| Field    | Type   | Meaning                                                            |
|----------|--------|--------------------------------------------------------------------|
| `cancel` | bool   | Request the host to cancel the current operation                   |
| `block`  | bool   | Request the host to block/suppress the current output or action    |
| `error`  | string | Report an extension error to the host; host decides how to surface |

**Invariants:**
- Extensions may set any combination of fields.
- The host must inspect all three fields; they are not mutually exclusive.
- `omitempty` means false booleans and empty strings are never serialized on the wire.

---

## Payload Types (9 total)

Each payload is deserialised from `Event.Payload` after matching on `Event.Type`.

### SessionStartPayload (`EventSessionStart`)

| Field    | Type   | Description               |
|----------|--------|---------------------------|
| `reason` | string | Why the session was started (e.g. `"new_session"`) |

### BeforeAgentStartPayload (`EventBeforeAgentStart`)

| Field           | Type   | Description                      |
|-----------------|--------|----------------------------------|
| `prompt`        | string | The user-facing agent prompt     |
| `system_prompt` | string | The system prompt for this agent |

### BeforeProviderRequestPayload (`EventBeforeProviderRequest`)

| Field      | Type       | Description                           |
|------------|------------|---------------------------------------|
| `messages` | []Message  | Full message history sent to provider |
| `model`    | string     | Model identifier string               |

### AfterProviderResponsePayload (`EventAfterProviderResponse`)

| Field   | Type       | Description                  |
|---------|------------|------------------------------|
| `usage` | UsageStats | Token usage from this call   |

### OnToolCallPayload (`EventOnToolCall`)

| Field         | Type            | Description                                |
|---------------|-----------------|--------------------------------------------|
| `tool_call_id`| string          | Unique identifier for this tool invocation |
| `tool_name`   | string          | Name of the tool being called              |
| `input`       | json.RawMessage | Raw JSON input arguments (never nil)       |

### OnToolResultPayload (`EventOnToolResult`)

| Field         | Type   | Description                                 |
|---------------|--------|---------------------------------------------|
| `tool_call_id`| string | Matches the corresponding `OnToolCallPayload` |
| `result`      | string | Text result from the tool                  |
| `is_error`    | bool   | True if the tool returned an error result   |

### MessageStartPayload (`EventMessageStart`)

| Field  | Type   | Description                         |
|--------|--------|-------------------------------------|
| `role` | string | Role of the message being started   |

### MessageEndPayload (`EventMessageEnd`)

| Field     | Type   | Description                              |
|-----------|--------|------------------------------------------|
| `role`    | string | Role of the completed message            |
| `content` | string | Full accumulated text content            |

### ShutdownPayload (`EventShutdown`)

| Field    | Type   | Description                      |
|----------|--------|----------------------------------|
| `reason` | string | Human-readable shutdown reason   |

---

## Supporting Types

### UsageStats

| Field           | Type | Description          |
|-----------------|------|----------------------|
| `input_tokens`  | int  | Prompt tokens used   |
| `output_tokens` | int  | Completion tokens used |

### Role Constants

`Role` is a `string` typedef. Values are stable across ABI versions.

| Constant        | Wire value    |
|-----------------|---------------|
| `RoleUser`      | `"user"`      |
| `RoleAssistant` | `"assistant"` |

**Invariant:** Role string values must not change; extensions may hard-code them.

### Message

| Field     | Type   | Description                  |
|-----------|--------|------------------------------|
| `role`    | Role   | `"user"` or `"assistant"`    |
| `content` | string | Text content of the message  |

### Tool

| Field          | Type            | Description                                            |
|----------------|-----------------|--------------------------------------------------------|
| `name`         | string          | Tool identifier                                        |
| `description`  | string          | Human-readable description                             |
| `input_schema` | json.RawMessage | JSON Schema object describing the tool's input; forwarded verbatim to the LLM provider |

**Invariant:** `InputSchema` is preserved as raw bytes through marshal/unmarshal; the sdk never parses it.

---

## HostCallRequest / HostCallResponse Contract

Extensions invoke host capabilities by encoding a `HostCallRequest` as JSON and calling the
`host_call` WASM import; the host returns a `HostCallResponse` encoded as JSON.

### HostCallRequest

```json
{ "method": "<MethodName>", "params": <raw-JSON-or-omitted> }
```

| Field    | Type            | Description                                     |
|----------|-----------------|-------------------------------------------------|
| `method` | string          | One of the `Method*` constants                  |
| `params` | json.RawMessage | Method-specific parameters; omitted when absent |

### HostCallResponse

```json
{ "result": <raw-JSON-or-omitted>, "error": "message-or-omitted" }
```

| Field    | Type            | Description                                       |
|----------|-----------------|---------------------------------------------------|
| `result` | json.RawMessage | Method-specific return value; omitted on error    |
| `error`  | string          | Non-empty when the host encountered an error      |

**Invariant:** Exactly one of `result` or `error` will be non-zero in a well-formed response.
Extensions should check `error` before using `result`.

---

## host_call Method Constants (10 total)

| Constant               | Wire value          | Purpose                                                        |
|------------------------|---------------------|----------------------------------------------------------------|
| `MethodSubscribe`      | `"subscribe"`       | Subscribe to one or more event types                           |
| `MethodRegisterTool`   | `"register_tool"`   | Advertise a tool that the extension can handle                 |
| `MethodRegisterCommand`| `"register_command"`| Register a slash command the extension provides               |
| `MethodSendMessage`    | `"send_message"`    | Inject a message into the conversation                        |
| `MethodSetStatus`      | `"set_status"`      | Update the extension's status string shown in the host UI     |
| `MethodNotify`         | `"notify"`          | Send a notification to the host/user                          |
| `MethodToolResult`     | `"tool_result"`     | Return the result of a tool call the extension handled        |
| `MethodStoreSet`       | `"store_set"`       | Persist a key-value pair in the host's extension store        |
| `MethodStoreGet`       | `"store_get"`       | Retrieve a value from the host's extension store              |
| `MethodAbort`          | `"abort"`           | Signal the host to abort the current agent operation          |

**Invariant:** Method strings must not change between ABI versions without a version bump.
There are exactly 10 defined method constants.

---

## Error Codes

Returned as `int32` from the `host_call` WASM import (not the JSON layer).

| Constant     | Value | Meaning                                        |
|--------------|-------|------------------------------------------------|
| `ErrOK`      | `0`   | Success                                        |
| `ErrGeneral` | `1`   | Unspecified host-side error; check JSON error field |
| `ErrCancel`  | `2`   | The operation was cancelled by the host        |

**Invariant:** Error code values must not change across ABI versions.

---

## Key Package Invariants

1. `EventType` values are stable string constants across ABI versions; numeric iota is never used.
2. `Event.Payload` is always valid JSON or explicitly `null`; it is never an empty byte slice on the wire.
3. `Tool.InputSchema` is stored and forwarded as `json.RawMessage`; the sdk never unmarshals it.
4. `HostCallRequest.Params` and `HostCallResponse.Result` are `json.RawMessage` so the sdk can forward them to/from WASM memory without any intermediate allocation.
5. `Role` string values (`"user"`, `"assistant"`) are stable and may be compared with `==`.
6. All `omitempty` fields in `EventResponse` and `HostCallRequest`/`HostCallResponse` are intentional; the wire format stays minimal.
7. `ABIVersion = 1` is the sole ABI version; future versions increment this constant and may add new event types or methods.
8. Error codes are `int32` and exist only at the WASM boundary; they do not appear in the JSON layer.
