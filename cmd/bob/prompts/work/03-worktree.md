# WORKTREE Phase

You are currently in the **WORKTREE** phase of the workflow.

## State Management

**IMPORTANT:** You can retrieve state from previous steps:
- **Retrieve state**: Call `workflow_get_guidance` to retrieve task description and requirements from PROMPT phase
- The task description will help you name the branch and worktree appropriately

## Your Goal
Create a git worktree for isolated development based on the task from the PROMPT phase.

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

### 1. Get Task Description from Previous Step

Retrieve the guidance to see the task description from PROMPT:
```
workflow_get_guidance(worktreePath: "<current-worktree-path>")
```

Look for `taskDescription` in the stored metadata.

### 2. Create Worktree

Based on the task description, create an appropriately named worktree:

```bash
# Ensure you're in the main repo
cd ~/source/<repo-name>
git checkout main
git pull origin main

# Create worktree with descriptive name
mkdir -p ~/source/<repo-name>-worktrees
git worktree add -b <branch-name> ~/source/<repo-name>-worktrees/<worktree-name>
```

Branch naming:
- Features: `feature/<description>` (e.g., `feature/add-json-output`)
- Fixes: `fix/<description>` (e.g., `fix/auth-bug`)
- Refactors: `refactor/<description>`

Worktree naming:
- Use descriptive names matching the task (e.g., `add-json-output`, `fix-auth-bug`)

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

## When You're Done
Once the worktree is created and you're in the worktree directory:

1. Confirm to user: "Worktree created at: <path>"
2. Report your progress:
   ```
   workflow_report_progress(
       worktreePath: "<full-worktree-path>",
       currentStep: "WORKTREE",
       metadata: { "branch": "<branch-name>", "worktreePath": "<full-path>" }
   )
   ```


## End Step

Ask bob what to do next based on the metadata you provided with workflow_get_guidance.
