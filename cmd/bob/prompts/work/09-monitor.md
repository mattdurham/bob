# MONITOR Phase

You are currently in the **MONITOR** phase of the workflow.

## Your Goal
Push the branch, create a PR, and monitor CI/PR status, writing findings to `bots/monitor.md`.

## What To Do

### 1. Push Branch (with user permission)

Ask user permission first, then:
```
git push -u origin <branch-name>
```

### 2. Create Pull Request (with user permission)

Create PR with summary of changes.

### 3. Check PR Status

```
gh pr checks
gh pr view --json reviewThreads
gh pr status
```

### 4. Document Findings in bots/monitor.md

**If issues found:**
```
# MONITOR Phase Findings

## CI/Actions Status
- Test Suite: FAILED
- Failed checks: 2

## Review Comments
- Unresolved comments: 2

## Summary
Total Issues: 4

Recommendation: Loop back to REVIEW
```

**If all clear:**
```
# MONITOR Phase Findings

## CI/Actions Status
All checks passing

## Summary
Total Issues: 0

Ready to merge.
```

### 5. Report Progress

```
workflow_report_progress(
    worktreePath: "<worktree-path>",
    currentStep: "MONITOR"
)
```

Bob will read bots/monitor.md and route automatically.

## Important
- ALWAYS write findings to bots/monitor.md
- Bob reads the file and decides next step
- Issues found = loop back to REVIEW
- No issues = advance to COMPLETE

## End Step
Ask bob what to do next based on workflow_get_guidance.
