# INIT Phase

You are currently in the **INIT** phase of the code review workflow.

## State Management

**IMPORTANT:** You can store and retrieve state throughout this workflow:
- **Store state**: Include any important information in the `metadata` field when calling `workflow_report_progress`
- **Retrieve state**: Call `workflow_get_guidance` at the start of any step to retrieve all stored state
- **State examples**: review findings, files to fix, issues found, test results, metrics

## Your Goal
Initialize the code review workflow - verify git repository and prepare to gather review scope.

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

### 1. Verify Git Repository
```bash
git status
```

Ensure you're in a git repository. If not, inform the user.

### 2. Check Current Branch
```bash
git branch
```

Note the current branch for reference.

### 3. Pull Latest Changes from Main
```bash
git checkout main
git pull origin main
```

Switch to main branch and pull latest changes to ensure you have the most recent code before starting.

## DO NOT
- ❌ Do not create a worktree yet - that happens in the WORKTREE step
- ❌ Do not ask the user what they want to review - that happens in the PROMPT step
- ❌ Do not start reviewing code

## When You're Done
Once you've verified the git repository:

1. Inform the user: "Ready to begin code review workflow"
2. Report your progress:
   ```
   workflow_report_progress(
       worktreePath: "<current-path>",
       currentStep: "INIT",
       metadata: { "repoVerified": true }
   )
   ```

