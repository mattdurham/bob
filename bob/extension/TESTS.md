# extension — Test Specifications

Documents existing tests and specifies missing tests worth adding.

---

## Existing Tests

### loader_test.go

#### TestValidateExports_AllPresent
**Scenario:** A WASM module exports all four required symbols.
**Setup:** Instantiate `minimalWASM` directly in a fresh wazero runtime.
**Assertions:**
- `validateExports` returns nil.

#### TestValidateExports_MissingFree
**Scenario:** A WASM module is missing the `_free` export.
**Setup:** Instantiate `missingFreeWASM` (has `_init`, `_on_event`, `_alloc` but no `_free`).
**Assertions:**
- `validateExports` returns a non-nil error.

#### TestCallInit_ReturnsZero
**Scenario:** `_init` returns 0 (success).
**Setup:** Instantiate `minimalWASM` directly in a fresh wazero runtime.
**Assertions:**
- `callInit` returns nil.

---

### host_test.go

#### TestHost_Load_MinimalWASM
**Scenario:** Successfully load a valid minimal WASM extension.
**Setup:** Write `minimalWASM` to a temp file; call `h.Load`.
**Assertions:**
- `Load` returns nil.
- `h.extensions` has length 1.

#### TestHost_Load_FileNotFound
**Scenario:** Path does not exist on disk.
**Setup:** Call `h.Load` with `/nonexistent/path/to/extension.wasm`.
**Assertions:**
- `Load` returns a non-nil error.

#### TestHost_Load_MissingExport
**Scenario:** WASM module is missing a required export (`_free`).
**Setup:** Write `missingFreeWASM` to a temp file; call `h.Load`.
**Assertions:**
- `Load` returns a non-nil error.

#### TestHost_DispatchEvent_NotSubscribed
**Scenario:** Extension is loaded but has not subscribed to the dispatched event.
**Setup:** Load `minimalWASM`; do not subscribe. Dispatch `sdk.EventSessionStart`.
**Assertions:**
- `DispatchEvent` returns nil error.
- Response slice is empty (length 0).

#### TestHost_DispatchEvent_Subscribed
**Scenario:** Extension is subscribed to the dispatched event.
**Setup:** Load `minimalWASM`; manually set `subscriptions[EventSessionStart] = true`. Dispatch `sdk.EventSessionStart`.
**Assertions:**
- `DispatchEvent` returns nil error.
- Response slice has length 1 (empty `EventResponse` because `_on_event` returns 0).

#### TestHost_Subscribe_ViaHostCall
**Scenario:** Extension subscribes via `host_call/subscribe`.
**Setup:** Load `minimalWASM`; call `handleSubscribe` directly with `{"event":"session_start"}`.
**Assertions:**
- Response has no error.
- `ext.subscriptions[EventSessionStart]` is true.

#### TestHost_Store_SetGet
**Scenario:** Extension stores and retrieves a value.
**Setup:** Load `minimalWASM`; call `handleStoreSet` then `handleStoreGet`.
**Assertions:**
- `handleStoreSet` returns no error.
- `handleStoreGet` returns no error and result JSON is `{"value":"bar"}`.

#### TestHost_Store_GetMiss
**Scenario:** Extension tries to get a key that was never set.
**Setup:** Load `minimalWASM`; call `handleStoreGet` for key `"missing"`.
**Assertions:**
- Response contains a non-empty error string.

#### TestHost_RegisterTool_DuplicateRejected
**Scenario:** Two extensions attempt to register a tool with the same name.
**Setup:** Load `minimalWASM`; call `handleRegisterTool` twice with the same tool JSON.
**Assertions:**
- First call returns no error.
- Second call returns a non-empty error string (`"tool already registered: search"`).

#### TestHost_Callbacks_SetStatus
**Scenario:** Extension calls `set_status`; harness callback is invoked.
**Setup:** Set `h.OnSetStatus`; load `minimalWASM`; route a `MethodSetStatus` host call.
**Assertions:**
- `OnSetStatus` callback receives `key="status"`, `value="ok"`.
- Response has no error.

#### TestHost_Reload
**Scenario:** Reload with the same path replaces extensions.
**Setup:** Load `minimalWASM`; call `Reload` with the same path.
**Assertions:**
- `Reload` returns nil.
- `h.extensions` still has length 1 (old module closed, new one loaded).

#### TestHost_Multiple_Extensions
**Scenario:** Two extensions loaded and both subscribed; both receive events.
**Setup:** Load `minimalWASM` twice under different names; subscribe both to `EventSessionStart`; dispatch.
**Assertions:**
- `h.extensions` has length 2.
- `DispatchEvent` returns a response slice of length 2.

#### TestHost_EchoWASM_SkipIfMissing
**Scenario:** Integration test with a real TinyGo-compiled echo extension.
**Setup:** Skip if `testdata/echo.wasm` does not exist. Load the file.
**Assertions:**
- `Load` returns nil.
- `h.extensions` has length 1.

---

## Missing Tests Worth Adding

### Concurrent DispatchEvent
**Scenario:** Multiple goroutines call `DispatchEvent` simultaneously on the same `Host`.
**Why:** The Go race detector should see no data races in subscription reads, extension slice copy, or host callbacks. Validates that `subMu` and the `h.mu.RLock + copy` pattern are sufficient.
**Setup:** Load one or more extensions subscribed to an event; launch N goroutines each calling `DispatchEvent` in a loop; run with `-race`.
**Assertions:**
- No data race detected.
- All goroutines complete without error.

### Reload While Dispatching
**Scenario:** `Reload` is called concurrently with `DispatchEvent`.
**Why:** `Reload` holds `h.mu.Lock()` while swapping `h.extensions`; `DispatchEvent` copies the slice under `h.mu.RLock()`. A race here would be dangerous.
**Setup:** Subscribe an extension; start a goroutine dispatching in a loop; call `Reload` repeatedly from another goroutine; run with `-race`.
**Assertions:**
- No data race detected.
- Neither `DispatchEvent` nor `Reload` panics.

### _alloc Returns 0 — _on_event Not Called
**Scenario:** Extension's `_alloc` returns 0 for the event buffer.
**Why:** Documents and enforces the C2 invariant (SPECS.md §5).
**Setup:** Craft or use a WASM module whose `_alloc` always returns 0. Subscribe it to an event. Dispatch.
**Assertions:**
- `DispatchEvent` returns no error.
- Response slice is empty (the extension is skipped, not counted).

### _init Returns Non-Zero Error Code
**Scenario:** Extension's `_init` returns a non-zero status.
**Why:** Validates that the host rejects the extension cleanly.
**Setup:** Craft a WASM module whose `_init` returns 1.
**Assertions:**
- `Load` returns a non-nil error containing `"_init returned error code 1"`.
- `h.extensions` has length 0.

### _on_event WASM Trap Does Not Stop Dispatch
**Scenario:** One extension traps in `_on_event`; subsequent extensions still receive the event.
**Why:** Validates the trap-isolation contract (SPECS.md §4 rule 3).
**Setup:** Load two extensions: first traps in `_on_event`, second returns normally. Subscribe both. Dispatch.
**Assertions:**
- First extension's trap is logged.
- Second extension's response is included in the returned slice.
- `DispatchEvent` returns nil error.

### Store Isolation Between Extensions
**Scenario:** Extension A sets a key; Extension B cannot read it.
**Why:** Validates the per-extension store invariant (SPECS.md §9).
**Setup:** Load two extensions; set key on ext A's store; attempt `handleStoreGet` on ext B's store for the same key.
**Assertions:**
- `handleStoreGet` on ext B returns a "not found" error.

### host_call with No Response Pointers
**Scenario:** Extension calls `host_call` with `resp_ptr_ptr=0` and `resp_len_ptr=0`.
**Why:** The host should return `ErrOK` without attempting to write response data.
**Setup:** Craft a `host_call` invocation with zero pointer slots.
**Assertions:**
- Return code is `ErrOK` (0).
- No WASM memory write occurs.

### readNullTerminatedOrJSON — Handles Nested Objects
**Scenario:** The response JSON contains nested objects and string values with `{` and `}`.
**Why:** Validates that the brace-depth scanner handles string-literal escaping correctly.
**Setup:** Call `readNullTerminatedOrJSON` directly with hand-crafted byte slices.
**Assertions:**
- Returns exactly the outer JSON object bytes.
- Does not include trailing bytes beyond the closing `}`.
