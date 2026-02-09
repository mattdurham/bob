# MONITOR Phase

You are currently in the **MONITOR** phase of the workflow.

## Your Goal
Push the branch, create a PR, and actively monitor it until merge.

## ⚠️ CRITICAL: Require User Permission

**DO NOT push or create PR automatically!**

Before proceeding with ANY of the steps below:
1. Tell user: "Ready to push and create PR?"
2. **WAIT for explicit user approval**
3. Only proceed after user says yes

This is a safety measure - never push to remote or create PRs without permission.

## What To Do

### 1. Push Branch
```bash
git push -u origin <branch-name>
```

### 2. Create Pull Request
```bash
gh pr create --title "<PR title>" --body "$(cat <<'EOF'
## Summary
- [Key change 1]
- [Key change 2]
- [Key change 3]

## Test Plan
- [x] Unit tests pass
- [x] Linting clean
- [x] Manual testing: [what you tested]

## Related Issues
Closes #[issue number if applicable]

🤖 Generated with Claude Code
EOF
)"
```

### 3. Get PR URL
```bash
gh pr view --web
```
Save the PR URL and share with user.

### 4. ACTIVELY MONITOR (Continuous Loop)
**This is critical - you must stay engaged until merge:**

#### Every 2-3 minutes:
```bash
# Check CI/Actions status
gh pr checks

# Check for comments
gh pr view

# Check PR status
gh pr status
```

#### Watch For:
- ❌ **CI failures** - investigate immediately
- 💬 **Comments** - respond to feedback
- ✅ **Approvals** - note when received
- ⚠️ **Change requests** - address feedback

### 5. Respond to Feedback

**If CI fails:**
1. Check logs: `gh pr checks --watch`
2. Identify the issue
3. Tell user: "CI failed: <reason>"
4. Record issues:
   ```
   workflow_record_issues(
       worktreePath: "<worktree-path>",
       step: "MONITOR",
       issues: [{ severity: "high", description: "CI failure: <details>" }]
   )
   ```
5. Loop back to PLAN/EXECUTE to fix
6. Push fix, continue monitoring

**If reviewer comments:**
1. Read comments: `gh pr view`
2. Address each comment
3. Make necessary changes
4. Commit and push
5. Reply to comments: `gh pr comment <number> --body "Fixed in <commit>"`
6. Continue monitoring

### 6. Auto-Merge When Ready
**Requirements for merge:**
- ✅ All CI checks passing (green)
- ✅ At least one approval
- ✅ No pending change requests
- ✅ No unresolved comments

**When all requirements met:**
```bash
gh pr merge --auto --squash
```

### 7. After Merge
```bash
# Verify merge
gh pr status

# Clean up
git worktree remove <worktree-path>
git branch -d <branch-name>
```

## DO NOT
- ❌ Do not stop monitoring after creating PR
- ❌ Do not merge without approvals
- ❌ Do not merge with failing checks
- ❌ Do not ignore comments or feedback
- ❌ Do not wait to be asked - proactively check status

## When You're Done

### If Merge Successful:
1. Tell user: "PR merged successfully! ✓"
2. Report progress to COMPLETE:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "COMPLETE",
       metadata: {
           "merged": true,
           "prUrl": "<url>"
       }
   )
   ```

### If CI Fails or Changes Needed:
1. Tell user: "Issues found during CI/review"
2. Record issues (use workflow_record_issues)
3. Loop back to PLAN:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "PLAN",
       metadata: {
           "loopReason": "CI failures / review feedback",
           "iteration": 3
       }
   )
   ```

