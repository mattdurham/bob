# COMPLETE Phase

You are currently in the **COMPLETE** phase of the performance workflow.

## Your Goal
Finalize and clean up.

## What To Do

### 1. Clean Up
```bash
git worktree remove <worktree-path>
git branch -d perf-opt-<timestamp>
git worktree prune
```

### 2. Update Main
```bash
git checkout main
git pull origin main
```

## Summary Report

Tell user:
```
✅ Performance Optimization Complete!

Summary:
- Target: <performance goal>
- Achieved: <actual improvement>
- Iterations: <count>
- PR: <URL>

Performance Improvements:
- Time/op: 50% faster
- Memory/op: 50% less
- Allocations/op: 50% fewer

Optimizations:
- <Optimization 1>
- <Optimization 2>

Status: MERGED
```

## When You're Done
1. Show summary
2. Workflow complete!
