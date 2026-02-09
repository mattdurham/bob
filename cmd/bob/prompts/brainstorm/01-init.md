# INIT Phase

You are currently in the **INIT** phase of the workflow.

## State Management

**IMPORTANT:** You can store and retrieve state throughout this workflow:
- **Store state**: Include any important information in the `metadata` field when calling `workflow_report_progress`
- **Retrieve state**: Call `workflow_get_guidance` at the start of any step to retrieve all stored state
- **State examples**: file paths, configuration values, decisions made, issues found, metrics collected

## Your Goal
Create a git worktree for isolated development.

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
git worktree add -b <branch-name> ~/source/<repo-name>-worktrees/<worktree-name>
```

Branch naming:
- Features: `feature/<description>` (e.g., `feature/add-json-output`)
- Fixes: `fix/<description>` (e.g., `fix/auth-bug`)
- Refactors: `refactor/<description>`

Worktree naming:
- Use descriptive names (e.g., `add-json-output`, `fix-auth-bug`)
- Matches the feature/fix being worked on

### 3. Verify Worktree
```bash
cd ~/source/<repo-name>-worktrees/<worktree-name>
pwd
git branch
```

### 4. Ensure bots/ Directory
```bash
mkdir -p bots
```

The `bots/` directory is for working files like brainstorm.md, plan.md, review.md.

## DO NOT
- ❌ Do not skip worktree creation - ALWAYS work in worktrees
- ❌ Do not work directly on main branch
- ❌ Do not start coding yet
- ❌ Do not automatically move to next phase

## When You're Done
Once the worktree is created and you're in the worktree directory:

1. Confirm to the user: "Worktree created at: <path>"
2. Report your progress:
   ```
   workflow_report_progress(
       worktreePath: "<full-worktree-path>",
       currentStep: "INIT",
       metadata: { "branch": "<branch-name>" }
   )
   ```


## End Step

Ask bob what to do next based on the metadata you provided with bob_workflow_get_guidance.
