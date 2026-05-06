# bob/sdk — Design Decisions

## 1. EventType is a string type, not an iota int
*Added: 2026-05-06*
**Decision:** `EventType` is defined as `type EventType string` with string constants, not as an `int` with iota.
**Rationale:** WASM extensions are independently compiled binaries. Using string values means the host and an extension can agree on an event name without sharing a compiled constant table. An iota-based int would require both sides to be compiled from the same version of this package (or a byte-for-byte identical layout). String constants survive independent compilation, separate module versions, and future additions without breaking existing extensions that simply ignore unknown types.
**Consequence:** Comparison is a string equality check (`O(n)` in string length), but event type strings are short and fixed, so this is negligible. Serialisation produces a human-readable wire format at no extra cost.

---

## 2. Event.Payload is json.RawMessage, not interface{}
*Added: 2026-05-06*
**Decision:** `Event.Payload` (and `OnToolCallPayload.Input`, `Tool.InputSchema`, `HostCallRequest.Params`, `HostCallResponse.Result`) are typed as `json.RawMessage` rather than `interface{}` or `any`.
**Rationale:** The host reads events from the WASM guest's linear memory as raw bytes and writes them back without needing to understand the payload structure. Using `json.RawMessage` allows the host to forward the bytes directly — zero extra allocations and no reflection. An `interface{}` field would require a round-trip through `encoding/json`'s generic map/slice representation (allocations, type assertions) for every event, even when the host does not inspect the payload.
**Consequence:** Callers must manually unmarshal `Payload` into the appropriate struct after switching on `Event.Type`. This is a deliberate trade-off: the host path stays zero-allocation; callers that need typed access do the unmarshal once on their side.

---

## 3. HostCallRequest/Response use json.RawMessage for params and result
*Added: 2026-05-06*
**Decision:** `HostCallRequest.Params` is `json.RawMessage` with `omitempty`; `HostCallResponse.Result` is `json.RawMessage` with `omitempty`.
**Rationale:** Each `host_call` method has a different parameter and result shape. Encoding them as `json.RawMessage` lets the extension SDK encode method-specific structs independently, then embed the bytes directly into the envelope without a second marshal/unmarshal cycle. The host similarly extracts `Params` bytes and routes them to the appropriate handler without parsing them at the envelope layer. This mirrors the same rationale as `Event.Payload` (see Note 2).
**Consequence:** There is no compile-time type safety between a method name and its params/result shape. The contract is documented in SPECS.md and enforced at runtime by the host handler for each method.

---

## 4. Role is a string type, not a bool or iota
*Added: 2026-05-06*
**Decision:** `Role` is defined as `type Role string` with constants `RoleUser = "user"` and `RoleAssistant = "assistant"`.
**Rationale:** The Anthropic Messages API uses `"user"` and `"assistant"` as role identifiers. Using matching string values means `Message` can be serialised directly into provider API calls without a mapping step. Future API roles (e.g. `"tool"` or `"system"`) can be added as constants without changing the type or breaking existing code.
**Consequence:** Invalid role strings are not caught at compile time. Extensions must validate role values at runtime if they care about correctness.

---

## 5. ABIVersion is an untyped int constant, not a typed version struct
*Added: 2026-05-06*
**Decision:** `ABIVersion = 1` is a bare untyped integer constant.
**Rationale:** The WASM ABI export `_abi_version()` must return a plain `i32`. Using an untyped constant lets it be assigned to any integer type without a cast, matching the WASM host binding pattern. A typed struct would require serialisation logic that has no place in a single-function WASM export.
**Consequence:** Semantic versioning is not represented here. Breaking changes require a new constant value and corresponding host-side gating logic; there is no minor/patch distinction at the ABI boundary.
