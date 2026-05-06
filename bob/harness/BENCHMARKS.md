# harness — Benchmark Specifications

## Metric Targets

| Benchmark | Target | Description |
|---|---|---|
| `BenchmarkModel_Update_TokenMsg` | < 5 µs/op | Cost of processing a single `TokenMsg` through `Model.Update`, including `ChatView.AppendToken` and `refreshContent` |
| `BenchmarkChatView_RefreshContent_100msgs` | < 1 ms/op | `refreshContent` rebuild with 100 finalised messages |
| `BenchmarkChatView_RefreshContent_1000msgs` | < 10 ms/op | `refreshContent` rebuild with 1000 finalised messages |

---

## Benchmark Scenarios

### BenchmarkModel_Update_TokenMsg

**Setup:**
- Create a `Model` via `newTestModel()`.
- Set `m.streaming = true`.
- Run `b.ResetTimer()`.

**Operation:**
- Call `callUpdate(m, TokenMsg{Token: "x"})` in the benchmark loop.

**Rationale:** Token throughput is on the critical path for streaming UX. A slow `AppendToken` + `refreshContent` cycle causes visible lag.

---

### BenchmarkChatView_RefreshContent_100msgs

**Setup:**
- Create `NewChatView(120, 40)`.
- Call `AddUserMessage` and `FinalizeMessage` (alternating) 50 times to produce 100 finalised messages.
- Run `b.ResetTimer()`.

**Operation:**
- Call `c.refreshContent()` in the benchmark loop.

**Rationale:** `refreshContent` is called on every token during streaming. With a long history, this must remain sub-millisecond.

---

### BenchmarkChatView_RefreshContent_1000msgs

**Setup:**
- Create `NewChatView(120, 40)`.
- Call `AddUserMessage` and `FinalizeMessage` (alternating) 500 times to produce 1000 finalised messages.
- Run `b.ResetTimer()`.

**Operation:**
- Call `c.refreshContent()` in the benchmark loop.

**Rationale:** Validates that the O(n) rebuild does not degrade unacceptably for long sessions.

---

## Running Benchmarks

```bash
go test ./bob/harness/... -bench=. -benchtime=5s -benchmem
```

To run a single benchmark:

```bash
go test ./bob/harness/... -bench=BenchmarkModel_Update_TokenMsg -benchtime=10s -benchmem
```

To compare against a baseline (using benchstat):

```bash
go test ./bob/harness/... -bench=. -count=10 | tee new.txt
benchstat old.txt new.txt
```
