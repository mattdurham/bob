# BENCHMARK Phase

You are currently in the **BENCHMARK** phase of the performance workflow.

## Your Goal
Establish baseline performance metrics.

## Continuation Behavior

**IMPORTANT:** Do NOT ask continuation questions like:
- "Should I proceed?"
- "Ready to continue?"
- "Shall I move to the next step?"
- "Done. Continue?"

**AUTOMATICALLY PROCEED** after completing your tasks.

**ONLY ASK THE USER** when:
- Choosing between multiple approaches/solutions
- Clarifying unclear requirements
- Confirming potentially risky/destructive actions (deletes, force pushes, etc.)
- Making architectural or design decisions

## What To Do

### 1. Run Existing Benchmarks
```bash
# Run benchmarks
go test -bench=. -benchmem -count=5 ./... > bots/benchmark-baseline.txt

# Or for specific benchmarks
go test -bench=BenchmarkFunctionName -benchmem -count=5 ./path/to/package
```

### 2. Profile if Needed
```bash
# CPU profile
go test -bench=. -cpuprofile=cpu.prof

# Memory profile
go test -bench=. -memprofile=mem.prof

# View profiles
go tool pprof cpu.prof
```

### 3. Document Baseline
Write to `bots/performance.md`:

```markdown
# Performance Optimization: <Feature>

## Goal
[Performance target - e.g., "Reduce latency by 50%", "Improve throughput by 2x"]

## Baseline Metrics

### Benchmark Results
```
BenchmarkFunction-8    1000000    1234 ns/op    512 B/op    10 allocs/op
```

Key metrics:
- Operations/sec: X
- Time per op: Y ns
- Memory per op: Z bytes
- Allocations per op: N

### Profiling Insights
- CPU hotspots: [function names]
- Memory hotspots: [allocation sites]
- Bottlenecks: [identified issues]

## Performance Targets
- [ ] Target 1: Reduce time/op by 50%
- [ ] Target 2: Reduce memory/op by 30%
- [ ] Target 3: Reduce allocations/op by 40%
```

## DO NOT
- ❌ Do not optimize yet - just measure
- ❌ Do not skip baseline benchmarks
- ❌ Do not guess at bottlenecks

## When You're Done
After establishing baseline:

1. Tell user: "Baseline metrics captured"
2. Report progress:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "BENCHMARK",
       metadata: {
           "baselineOpsPerSec": X,
           "baselineNsPerOp": Y
       }
   )
   ```

