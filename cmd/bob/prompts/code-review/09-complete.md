# COMPLETE Phase

You are currently in the **COMPLETE** phase of the code review workflow.

## Your Goal
Finalize and clean up.

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

### 1. Verify Merge
```bash
gh pr view <number>
```

### 2. Clean Up
```bash
git worktree remove <worktree-path>
git branch -d review-fix-<timestamp>
git worktree prune
```

### 3. Update Main
```bash
git checkout main
git pull origin main
```

## Summary Report

Tell user:
```
✅ Code Review Workflow Complete!

Summary:
- Issues Found: <count>
- Issues Fixed: <count>
- Tests Added: <count>
- Review Iterations: <count>
- PR: <URL>
- Status: MERGED

All code review issues have been addressed and merged!
```

## When You're Done
1. Show summary
2. Workflow complete
3. Ready for next task!
