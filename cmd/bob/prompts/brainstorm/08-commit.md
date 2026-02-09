# COMMIT Phase

You are currently in the **COMMIT** phase of the workflow.

## Your Goal
Commit your changes with a clear commit message.

## What To Do

### 1. Review What's Changed
```bash
git status
git diff
```

### 2. Stage Files
```bash
# Stage specific files (preferred)
git add path/to/file1.go
git add path/to/file2.go

# Verify staged files
git status
```

**IMPORTANT:**
- Do NOT use `git add .` or `git add -A` (might include secrets)
- Stage files explicitly by name
- Never commit .env files or credentials

### 3. Write Commit Message
Follow conventional commit format:

```bash
git commit -m "$(cat <<'EOF'
<type>: <short description>

<detailed explanation if needed>

- Key change 1
- Key change 2
- Key change 3

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
EOF
)"
```

**Commit types:**
- `feat:` - New feature
- `fix:` - Bug fix
- `refactor:` - Code refactoring
- `test:` - Adding tests
- `docs:` - Documentation changes
- `chore:` - Maintenance tasks

**Example:**
```
feat: add JSON output format to codepath-dag

Add --format flag to support JSON output in addition to the default
text format. This enables easier programmatic consumption of the DAG.

- Add --format flag (json|text)
- Implement JSON serializer
- Update tests for both formats

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

### 4. Verify Commit
```bash
git log -1 --stat
```

## DO NOT
- ❌ Do not commit secrets, credentials, or .env files
- ❌ Do not use `git add .` (might include unwanted files)
- ❌ Do not skip the Co-Authored-By line
- ❌ Do not commit if tests are failing
- ❌ Do not automatically push yet

## When You're Done
After committing:

1. Tell user: "Changes committed: <commit-hash>"
2. Show commit message
3. Ask user: "Ready to push and create PR?"
4. Wait for user approval before proceeding
5. Report progress to MONITOR:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "MONITOR",
       metadata: {
           "committed": true,
           "commitHash": "<hash>"
       }
   )
   ```

## Next Phase
After user approves, push the branch and move to **MONITOR** phase to watch PR/CI.
