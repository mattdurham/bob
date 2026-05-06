# extension — Benchmarks

Documents benchmark specifications and metric targets for the `extension` package.

---

## Metric Targets

| Benchmark | Target | Notes |
|-----------|--------|-------|
| `BenchmarkDispatchEvent` | < 10 µs/op per extension | Measured with a subscribed minimal WASM extension; includes JSON marshal of event, `_alloc` call, memory write, `_on_event` call, `_free` call |
| `BenchmarkLoad` | < 50 ms/op | Full WASM instantiation of a minimal module, including `validateExports` and `callInit`; dominated by wazero compilation cost |
| `BenchmarkHostCall` | < 5 µs/op | Single `host_call` round-trip (JSON decode of request + dispatch to handler + JSON encode of response); measured via `routeHostCall` directly without WASM overhead |

---

## Benchmark Specifications

### BenchmarkDispatchEvent

**Scenario:** Measure the per-call overhead of `Host.DispatchEvent` with one subscribed extension.

**Setup:**
1. Create a `Host` with `NewHost(nil)`.
2. Load `minimalWASM` and subscribe it to `sdk.EventSessionStart`.
3. Construct a `sdk.Event{Type: sdk.EventSessionStart}`.
4. Reset the timer after setup.

**Loop body:**
```go
responses, err := h.DispatchEvent(ctx, evt)
```

**Assertions:**
- No error on any iteration.
- `len(responses) == 1` on each iteration (minimalWASM `_on_event` returns 0, producing an empty response that is still counted).

**Metric target:** < 10 µs/op per loaded extension.

**Notes:** The dominant cost is the wazero function call overhead and the JSON marshal of the event. Real extensions with complex logic will be slower; the target covers the host-side overhead only.

---

### BenchmarkLoad

**Scenario:** Measure the cost of loading a WASM module from bytes (instantiation + validation + init).

**Setup:**
1. Write `minimalWASM` to a temp file once before the benchmark loop.
2. Create a new `Host` per iteration (or reuse if wazero supports reloading the same bytes — prefer a fresh runtime to measure full cold-path cost).
3. Reset the timer after file write.

**Loop body:**
```go
h := NewHost(nil)
if err := h.Load(ctx, path); err != nil { b.Fatal(err) }
h.Close(ctx)
```

**Metric target:** < 50 ms/op (wazero JIT compilation of the module is included).

**Notes:** The target is deliberately loose because WASM compilation latency depends on module size. The `minimalWASM` binary is ~100 bytes; a real extension will be larger. This benchmark tracks regression in host-side overhead, not extension code complexity.

---

### BenchmarkHostCall

**Scenario:** Measure the round-trip cost of a single `routeHostCall` dispatch (no WASM FFI — tests Go-side routing and JSON handling).

**Setup:**
1. Create a `Host` with a no-op `OnSetStatus` callback.
2. Load `minimalWASM` to get an `ext`.
3. Prepare a `sdk.HostCallRequest` for `MethodSetStatus`.
4. Reset the timer.

**Loop body:**
```go
resp := h.routeHostCall(ctx, ext.module, ext, req)
_ = resp
```

**Metric target:** < 5 µs/op (JSON unmarshal of params + handler logic + response allocation).

**Notes:** This benchmark isolates the Go host routing layer. WASM trap overhead, memory copies, and `_alloc`/`_free` calls are excluded. It is primarily useful for detecting regressions in the JSON decoding hot path.

---

## Running Benchmarks

```bash
# All benchmarks in the package
go test -bench=. -benchmem ./bob/extension/...

# Specific benchmark
go test -bench=BenchmarkDispatchEvent -benchmem -benchtime=5s ./bob/extension/...

# With race detector (expect ~5× slowdown)
go test -bench=. -benchmem -race ./bob/extension/...

# CPU profile
go test -bench=BenchmarkDispatchEvent -benchmem -cpuprofile=cpu.out ./bob/extension/...
go tool pprof cpu.out
```

---

## Profiling Guidance

If `BenchmarkDispatchEvent` exceeds the target:

1. Check `json.Marshal(evt)` allocation — consider a pre-serialised event cache if the same event type is dispatched frequently.
2. Check `_alloc` call overhead — wazero function call setup.
3. Check `mem.Write` — linear memory copy cost scales with event JSON size.

If `BenchmarkLoad` exceeds the target:

1. wazero compiles WASM to native code on first instantiation; subsequent loads of the same bytes reuse the compilation cache. Ensure tests are not defeating the cache by writing fresh bytes each iteration.
2. Profile `runtime.InstantiateWithConfig` — most time will be in wazero's compiler.
