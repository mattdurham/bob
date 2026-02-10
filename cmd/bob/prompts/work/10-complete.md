# COMPLETE Phase

You are currently in the **COMPLETE** phase of the workflow.

## Your Goal
Finalize the workflow and clean up.

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
# Should show "MERGED"
```

### 2. Clean Up Worktree
```bash
# Remove worktree
git worktree remove <worktree-path>

# Delete local branch (safe if merged)
git branch -d <branch-name>

# Prune stale worktree references
git worktree prune
```

### 3. Update Main Branch
```bash
git checkout main
git pull origin main
```

### 4. Verify Change Landed
```bash
git log --oneline -5
# Should see your commit
```

## Summary Report

Tell the user:
```
✅ Workflow Complete!

Summary:
- Task: <original task description>
- PR: <URL>
- Merged: <timestamp>
- Commits: <count>
- Loops: <loop count>
- Issues Found: <total issues>
- Final Status: MERGED

Changes:
- <file1.go> - <what changed>
- <file2.go> - <what changed>

Worktree cleaned up: <path>
```

## DO NOT
- ❌ Do not skip cleanup
- ❌ Do not leave worktree lying around
- ❌ Do not forget to prune worktree references

## When You're Done
1. Show summary report to user
2. Workflow tracking is complete
3. No further progress reporting needed

## What's Next?
Ready for the next task! User can start a new workflow with a fresh worktree.
