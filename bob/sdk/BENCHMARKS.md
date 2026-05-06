# bob/sdk — Benchmark Specifications

## Metric Targets

| Benchmark                          | Max ns/op | Max B/op | Max allocs/op | Notes                                    |
|------------------------------------|-----------|----------|---------------|------------------------------------------|
| BenchmarkEventMarshal              | 500       | 512      | 4             | Marshal Event with BeforeProviderRequest payload |
| BenchmarkEventUnmarshal            | 600       | 512      | 5             | Unmarshal same JSON back to Event        |
| BenchmarkHostCallRequestRoundTrip  | 1000      | 256      | 4             | Marshal + unmarshal HostCallRequest      |
| BenchmarkHostCallResponseRoundTrip | 800       | 256      | 3             | Marshal + unmarshal HostCallResponse     |
| BenchmarkToolMarshal               | 400       | 256      | 3             | Marshal Tool with raw JSON InputSchema   |

All targets measured on an Apple M-series CPU at `-benchtime=5s`. CI should run with `-benchmem` and fail if any benchmark exceeds 2x the ns/op target (allows for CI machine variance).

---

## Benchmark Specifications

### BenchmarkEventMarshal

**Scenario:** Measure the cost of marshalling a fully-populated `Event` whose payload is a `BeforeProviderRequestPayload` with two messages.

**Setup:**
```go
payload, _ := json.Marshal(sdk.BeforeProviderRequestPayload{
    Messages: []sdk.Message{
        {Role: sdk.RoleUser, Content: "hello"},
        {Role: sdk.RoleAssistant, Content: "hi there"},
    },
    Model: "claude-sonnet-4",
})
evt := sdk.Event{Type: sdk.EventBeforeProviderRequest, Payload: json.RawMessage(payload)}
```

**Measurement:** `json.Marshal(evt)` in the hot loop, with `b.ReportAllocs()`.

**Target:** < 500 ns/op, < 512 B/op, < 4 allocs/op.

---

### BenchmarkEventUnmarshal

**Scenario:** Measure the cost of unmarshalling the same JSON produced by BenchmarkEventMarshal back into an `Event`.

**Setup:** Pre-marshal the event outside the loop; pass the bytes to `json.Unmarshal` in the hot loop.

**Target:** < 600 ns/op, < 512 B/op, < 5 allocs/op.

---

### BenchmarkHostCallRequestRoundTrip

**Scenario:** Measure the combined cost of marshalling a `HostCallRequest` and then unmarshalling the result.

**Setup:**
```go
req := sdk.HostCallRequest{
    Method: sdk.MethodSubscribe,
    Params: json.RawMessage(`{"event":"session_start"}`),
}
```

**Measurement:** `json.Marshal(req)` followed by `json.Unmarshal(data, &req2)` in the hot loop.

**Target:** < 1000 ns/op, < 256 B/op, < 4 allocs/op.

---

### BenchmarkHostCallResponseRoundTrip

**Scenario:** Measure the combined cost of marshalling a `HostCallResponse` and then unmarshalling.

**Setup:**
```go
resp := sdk.HostCallResponse{Result: json.RawMessage(`{"ok":true}`)}
```

**Measurement:** Marshal + unmarshal in the hot loop.

**Target:** < 800 ns/op, < 256 B/op, < 3 allocs/op.

---

### BenchmarkToolMarshal

**Scenario:** Measure the cost of marshalling a `Tool` with a realistic JSON Schema in `InputSchema`.

**Setup:**
```go
tool := sdk.Tool{
    Name:        "search",
    Description: "Search the web for information",
    InputSchema: json.RawMessage(`{"type":"object","properties":{"q":{"type":"string"}},"required":["q"]}`),
}
```

**Measurement:** `json.Marshal(tool)` in the hot loop.

**Target:** < 400 ns/op, < 256 B/op, < 3 allocs/op.
