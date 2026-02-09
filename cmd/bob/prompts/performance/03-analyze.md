# ANALYZE Phase

You are currently in the **ANALYZE** phase of the performance workflow.

## Your Goal
Identify specific performance bottlenecks and optimization opportunities.

## What To Do

### 1. Analyze Profiles
```bash
# CPU profile
go tool pprof -top cpu.prof
go tool pprof -list=FunctionName cpu.prof

# Memory profile
go tool pprof -alloc_space mem.prof
go tool pprof -list=FunctionName mem.prof
```

### 2. Identify Bottlenecks
Look for:
- **CPU hotspots**: Functions consuming most CPU time
- **Memory allocations**: Excessive allocations or large objects
- **Inefficient algorithms**: O(n²) where O(n) is possible
- **Unnecessary work**: Redundant computations
- **Lock contention**: Synchronization bottlenecks

### 3. Prioritize Optimizations
Rank by impact:
1. **High impact**: Addresses main bottleneck, big performance gain
2. **Medium impact**: Noticeable improvement, moderate effort
3. **Low impact**: Small gains, easy wins

### 4. Document Analysis
Update `bots/performance.md`:

```markdown
## Analysis

### Bottlenecks Identified

#### 1. Function XYZ - HIGH PRIORITY
- **Issue**: Allocates 1MB per call in hot path
- **Impact**: 60% of total allocations
- **Location**: file.go:123
- **Optimization Strategy**: Use sync.Pool for reuse

#### 2. Algorithm ABC - MEDIUM PRIORITY
- **Issue**: O(n²) complexity
- **Impact**: Slows down with large inputs
- **Location**: file.go:456
- **Optimization Strategy**: Use map for O(1) lookup

### Optimization Plan
1. Fix bottleneck #1 (expected 50% improvement)
2. Fix bottleneck #2 (expected 30% improvement)
3. Re-benchmark and verify
```

## DO NOT
- ❌ Do not guess - use profiling data
- ❌ Do not optimize without analysis
- ❌ Do not skip prioritization

## When You're Done
After thorough analysis:

1. Tell user: "Found X bottlenecks, ready to optimize"
2. Report progress:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "OPTIMIZE",
       metadata: {
           "bottlenecksFound": X,
           "expectedImprovement": "50%"
       }
   )
   ```

