# MONITOR Phase

You are currently in the **MONITOR** phase of the code review workflow.

## Your Goal
Push PR and monitor until merge.

## What To Do

### 1. Push Branch
```bash
git push -u origin review-fix-<timestamp>
```

### 2. Create PR
```bash
gh pr create --title "fix: address code review issues" --body "$(cat <<'EOF'
## Summary
Fixed issues found during comprehensive code review:
- [Issue 1]
- [Issue 2]
- [Issue 3]

## Changes
- [File/component changed]
- Added tests for all bug fixes

## Test Plan
- [x] All existing tests pass
- [x] New tests added for bug fixes
- [x] Linting clean
- [x] Complexity check passed

🤖 Generated with Claude Code
EOF
)"
```

### 3. Monitor Actively
Every 2-3 minutes:
```bash
gh pr checks
gh pr view
gh pr status
```

Watch for:
- ❌ CI failures → fix and push
- 💬 Comments → respond
- ✅ Approvals → note when received

### 4. Auto-Merge When Ready
When all checks pass and approved:
```bash
gh pr merge --auto --squash
```

### 5. After Merge
```bash
git worktree remove <worktree-path>
git branch -d review-fix-<timestamp>
```

## DO NOT
- ❌ Do not stop monitoring after creating PR
- ❌ Do not merge without approval
- ❌ Do not ignore CI failures

## When You're Done

### If Merged:
1. Tell user: "PR merged successfully! ✓"
2. Report progress:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "COMPLETE",
       metadata: { "merged": true }
   )
   ```

### If CI Fails:
1. Loop back to FIX phase
2. Report progress:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "FIX",
       metadata: {
           "loopReason": "CI failure",
           "iteration": 3
       }
   )
   ```

## Next Phase
Move to **COMPLETE** after successful merge.
