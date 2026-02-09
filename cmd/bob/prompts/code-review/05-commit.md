# COMMIT Phase

You are currently in the **COMMIT** phase of the code review workflow.

## Your Goal
Commit the review fixes.

## What To Do

### 1. Review Changes
```bash
git status
git diff
```

### 2. Stage Files
```bash
# Stage specific files
git add path/to/file1.go
git add path/to/file2_test.go
```

### 3. Commit with Clear Message
```bash
git commit -m "$(cat <<'EOF'
fix: address code review issues

Fixed issues found during code review:
- Issue 1: Description
- Issue 2: Description
- Issue 3: Description

Added tests for all bug fixes.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
EOF
)"
```

### 4. Verify Commit
```bash
git log -1 --stat
```

## DO NOT
- ❌ Do not commit secrets or .env files
- ❌ Do not use `git add .`
- ❌ Do not skip Co-Authored-By line

## When You're Done
After committing:

1. Tell user: "Fixes committed: <hash>"
2. Ask: "Ready to push and create PR?"
3. Wait for user approval
4. Report progress:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "COMMIT",
       metadata: {
           "committed": true,
           "commitHash": "<hash>"
       }
   )
   ```

