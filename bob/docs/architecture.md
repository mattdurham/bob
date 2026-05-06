# Architecture Overview

Bob is a terminal chat interface for LLM coding assistance. It is built with
[Bubble Tea v2](https://github.com/charmbracelet/bubbletea) and extended via
WebAssembly modules loaded by a wazero-based extension host.

---

## Package Dependency Diagram

```
cmd/bob
  └── bob/harness ─────────────────────── bob/sdk
        ├── bob/provider/anthropic
        │     └── (anthropic-sdk-go)
        └── bob/extension ────────────── bob/sdk
              └── (wazero)

bob/provider/mock ──────────────────────── bob/provider
bob/provider/anthropic ─────────────────── bob/provider
```

`bob/sdk` is the only package imported by both `bob/harness` and
`bob/extension` — it contains all shared types (events, messages, tools) and
ABI constants. It has no internal dependencies.

`bob/provider` defines the `Provider` interface and `Request`/`StreamCallback`
types. The concrete implementations (`anthropic`, `mock`) live in sub-packages
and import only `bob/provider` and `bob/sdk`.

---

## Data Flow — User Input to Rendered Output

```
User types text
  → InputArea.Update (tea.KeyPressMsg "enter")
    → emits SubmitMsg{Content: "..."}
      → Model.Update (SubmitMsg)
        → Model.startStream(content)
          → appends to m.history
          → sets m.streaming = true
          → launches goroutine: provider.Stream(ctx, req, callback)
            → for each token: prog.Send(TokenMsg{Token: "..."})
              → Model.Update (TokenMsg)
                → ChatView.AppendToken(token)
                  → refreshContent() → vp.SetContent(...)
                    → Model.View() renders updated viewport
            → on completion: prog.Send(StreamDoneMsg{})
              → Model.Update (StreamDoneMsg)
                → ChatView.FinalizeMessage()
                → m.streaming = false
```

The provider goroutine communicates back to the Bubble Tea event loop
exclusively through `program.Send(msg)`. No shared mutable state is accessed
directly from the goroutine; all mutations go through `Model.Update`.

---

## Extension Lifecycle

```
Host.Load(path)
  → wazero: instantiate .wasm module
  → validateExports: check _init, _on_event, _alloc, _free
  → register Extension in Host.extensions
  → callInit: call _init()
      → extension calls host_call("subscribe", {"event": "session_start"})
      → extension calls host_call("register_tool", {...})
      → extension calls host_call("register_command", {...})
      → _init returns 0 (success)

Later, on user submit:
  Host.DispatchEvent(ctx, Event{Type: "before_agent_start", Payload: ...})
    → for each Extension with subscriptions["before_agent_start"]:
        dispatchToExtension(ctx, ext, evtJSON)
          → _alloc(len(evtJSON))        // allocate in WASM memory
          → mem.Write(ptr, evtJSON)     // copy event bytes
          → _on_event(ptr, len)         // call handler
          → _free(ptr)                  // release event buffer
          → if resp_ptr != 0:
              read response JSON
              _free(resp_ptr)
              unmarshal EventResponse
    → return []EventResponse
```

The host never calls into WASM on a background goroutine while the extension is
also executing. Dispatch is sequential per extension, and the WASM runtime
serialises concurrent calls to the same module.

---

## host_call Flow

```
Extension (WASM) calls host_call(req_ptr, req_len, resp_ptr_ptr, resp_len_ptr)
  → hostCallImpl (Go, runs inside wazero)
      → mem.Read(req_ptr, req_len)           // read JSON request
      → json.Unmarshal → HostCallRequest
      → findExtensionByModule(m)             // identify calling extension
      → routeHostCall(ctx, m, ext, req)
          switch req.Method:
          case "set_status":
              handleSetStatus(req)
                → json.Unmarshal params
                → Host.OnSetStatus(key, value)
                    → prog.Send(StatusUpdateMsg{Key: key, Value: value})
                        → Model.Update (StatusUpdateMsg)
                            → StatusBar.Update(msg)
                                → StatusBar.View() shows new value
          case "store_get":
              handleStoreGet(ext, req)
                → ext.store.Get(key) → value
                → json.Marshal {"value": value} → resp
          ...
      → json.Marshal HostCallResponse
      → _alloc(len(resp)) in extension memory
      → mem.Write(resp_ptr, respBytes)
      → mem.WriteUint32Le(resp_ptr_ptr, resp_ptr)
      → mem.WriteUint32Le(resp_len_ptr, resp_len)
      → return ErrOK
  ← extension reads response at *resp_ptr_ptr
  ← extension calls _free(resp_ptr)
```

The `OnSetStatus`, `OnNotify`, `OnSendMessage`, `OnAbort`, `OnToolResult`, and
`OnRegisterTool` callbacks are set by `Model.SetProgram` immediately after the
Bubble Tea program starts. They bridge the synchronous wazero call stack into
the Bubble Tea message loop via `program.Send`.

---

## Tool Call Flow

```
LLM response contains a tool_use block
  → (future: provider parses tool_use and emits ToolCallMsg)
  → Model.Update dispatches EventOnToolCall to extensions:
      Host.DispatchEvent(Event{
        Type: "on_tool_call",
        Payload: {tool_call_id, tool_name, input},
      })
        → extension._on_event receives the event
        → extension processes the call (e.g. runs a search)
        → extension calls host_call("tool_result", {
            tool_call_id: "toolu_abc123",
            result: "Paris",
            is_error: false,
          })
            → Host.OnToolResult(toolCallID, result, isError)
                → harness adds tool result to conversation
                → triggers next provider request
```

---

## Component Responsibilities

| Component       | Package          | Responsibility                                                      |
|-----------------|------------------|---------------------------------------------------------------------|
| Model           | `bob/harness`    | Root Bubble Tea model; coordinates all sub-components, owns history, manages streaming lifecycle. |
| ChatView        | `bob/harness`    | Scrollable viewport rendering the conversation; accumulates streaming tokens and finalises messages. |
| InputArea       | `bob/harness`    | Textarea with Enter-to-submit and `/command` detection; emits `SubmitMsg` or `CommandMsg`. |
| StatusBar       | `bob/harness`    | Single-line display showing provider, model, token count, and keyed extension statuses. |
| CommandRegistry | `bob/harness`    | Slash command dispatch table; built-ins: `/help`, `/clear`, `/reload`, `/model`. |
| Host            | `bob/extension`  | wazero runtime wrapper; loads `.wasm` files, validates exports, calls `_init`, dispatches events, routes `host_call` RPCs. |
| Extension       | `bob/extension`  | Per-module wrapper holding subscription set and private key-value store. |
| Store           | `bob/extension`  | Thread-safe in-process key-value store scoped to a single extension. |
| Provider        | `bob/provider`   | Interface abstracting any LLM streaming backend.                    |
| Anthropic       | `bob/provider/anthropic` | Concrete provider using the Anthropic Messages API.         |
| Mock            | `bob/provider/mock`      | Scripted provider for deterministic unit tests.             |
| Config          | `cmd/bob`        | Reads `ANTHROPIC_API_KEY`, `BOB_MODEL`, `BOB_PROVIDER`, `BOB_EXTENSIONS_DIR` from the environment. |
