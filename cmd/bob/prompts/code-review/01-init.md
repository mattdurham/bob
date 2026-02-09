# INIT Phase

You are currently in the **INIT** phase of the code review workflow.

## State Management

**IMPORTANT:** You can store and retrieve state throughout this workflow:
- **Store state**: Include any important information in the `metadata` field when calling `workflow_report_progress`
- **Retrieve state**: Call `workflow_get_guidance` at the start of any step to retrieve all stored state
- **State examples**: review findings, files to fix, issues found, test results, metrics

## Your Goal
Create a git worktree for code review and fixes.

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

### 1. Pull Latest Changes
```bash
git checkout main
git pull origin main
```

### 2. Create Worktree
```bash
mkdir -p ~/source/<repo-name>-worktrees
git worktree add -b review-fix-<timestamp> ~/source/<repo-name>-worktrees/review-fix
```

### 3. Verify Worktree
```bash
cd ~/source/<repo-name>-worktrees/review-fix
mkdir -p bots
```

## DO NOT
- ❌ Do not skip worktree creation
- ❌ Do not start reviewing yet
- ❌ Do not automatically move to next phase

## When You're Done
Once worktree is created:

1. Confirm to user: "Worktree created for code review"
2. Report progress:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "REVIEW",
       metadata: { "worktreeCreated": true }
   )
   ```

