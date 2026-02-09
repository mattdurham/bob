# VERIFY Phase

You are currently in the **VERIFY** phase of the performance workflow.

## Your Goal
Measure performance improvements and verify correctness.

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

### 1. Run Tests First
```bash
go test ./...
```
All tests MUST pass before benchmarking.

### 2. Run New Benchmarks
```bash
go test -bench=. -benchmem -count=5 ./... > bots/benchmark-optimized.txt
```

### 3. Compare Results
```bash
# Install benchstat if needed
go install golang.org/x/perf/cmd/benchstat@latest

# Compare baseline vs optimized
benchstat bots/benchmark-baseline.txt bots/benchmark-optimized.txt
```

### 4. Analyze Improvements
Update `bots/performance.md`:

```markdown
## Results

### Before vs After

| Metric | Baseline | Optimized | Improvement |
|--------|----------|-----------|-------------|
| ns/op | 1234 | 617 | 50% faster |
| B/op | 512 | 256 | 50% less memory |
| allocs/op | 10 | 5 | 50% fewer allocations |

### Benchstat Output
```
name    old time/op  new time/op  delta
Func-8  1.23µs ± 2%  0.62µs ± 1%  -49.59%  (p=0.000 n=5+5)

name    old alloc/op  new alloc/op  delta
Func-8    512B ± 0%     256B ± 0%  -50.00%  (p=0.000 n=5+5)
```

### Target Achievement
- [x] Target 1: Reduce time/op by 50% - ACHIEVED (49.6%)
- [x] Target 2: Reduce memory/op by 30% - ACHIEVED (50%)
- [x] Target 3: Reduce allocations/op by 40% - ACHIEVED (50%)
```

## DO NOT
- ❌ Do not accept if tests fail
- ❌ Do not proceed if performance regressed
- ❌ Do not skip comparison

## When You're Done

### If Targets Met:
1. Tell user: "Performance targets achieved! ✓"
2. Report progress:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "COMMIT",
       metadata: {
           "improvementPct": 50,
           "targetsAchieved": true
       }
   )
   ```

### If Targets NOT Met:
1. Tell user: "Targets not met, need more optimization"
2. Record issues:
   ```
   workflow_record_issues(
       worktreePath: "<worktree-path>",
       step: "VERIFY",
       issues: [{
           severity: "medium",
           description: "Performance target not achieved"
       }]
   )
   ```
3. Loop back:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "ANALYZE",
       metadata: {
           "loopReason": "targets not met",
           "iteration": 2
       }
   )
   ```

