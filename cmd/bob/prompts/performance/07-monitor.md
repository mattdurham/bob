# MONITOR Phase

You are currently in the **MONITOR** phase of the performance workflow.

## Your Goal
Push PR and monitor until merge.

## What To Do

### 1. Push
```bash
git push -u origin perf-opt-<timestamp>
```

### 2. Create PR with Benchmark Data
```bash
gh pr create --title "perf: optimize XYZ for 50% improvement" --body "$(cat <<'EOF'
## Summary
Performance optimization of XYZ component.

## Performance Improvements
| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| Time/op | 1234ns | 617ns | 50% faster |
| Memory/op | 512B | 256B | 50% less |
| Allocs/op | 10 | 5 | 50% fewer |

## Changes
- Used sync.Pool for buffer reuse
- Replaced O(n²) algorithm with O(n) using map lookup
- Added benchmarks to verify improvements

## Test Plan
- [x] All tests passing
- [x] Benchmarks show expected improvements
- [x] No correctness regressions

## Benchmark Details
See bots/performance.md for complete benchmark results.

🤖 Generated with Claude Code
EOF
)"
```

### 3. Monitor
Watch for:
- CI passing
- Benchmark regression checks
- Code review comments
- Performance validation

### 4. Auto-Merge
When approved and checks pass:
```bash
gh pr merge --auto --squash
```

## DO NOT
- ❌ Do not merge if benchmarks regress
- ❌ Do not ignore performance feedback

## When You're Done

### If Merged:
1. Tell user: "Performance improvements merged! ✓"
2. Report progress:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "COMPLETE",
       metadata: { "merged": true }
   )
   ```

### If Issues:
Loop back to OPTIMIZE.

## Next Phase
Move to **COMPLETE**.
