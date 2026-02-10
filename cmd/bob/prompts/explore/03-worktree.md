# WORKTREE Phase

You are currently in the **WORKTREE** phase of the explore workflow.

## State Management

**IMPORTANT:** You can retrieve state from previous steps:
- **Retrieve state**: Call `workflow_get_guidance` to retrieve exploration goals from PROMPT phase
- The exploration goals will help you name the branch and worktree appropriately

## Your Goal
Create a git worktree for isolated exploration based on the goals from the PROMPT phase.

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

### 1. Get Exploration Goals from Previous Step

Retrieve the guidance to see the exploration goals from PROMPT:
```
workflow_get_guidance(worktreePath: "<current-worktree-path>")
```

Look for `explorationGoal` in the stored metadata.

### 2. Create Worktree

Based on the exploration goals, create an appropriately named worktree:

```bash
# Ensure you're in the main repo
cd ~/source/<repo-name>
git checkout main
git pull origin main

# Create worktree for exploration
mkdir -p ~/source/<repo-name>-worktrees
git worktree add -b explore/<description> ~/source/<repo-name>-worktrees/explore-<name>
```

Branch naming for exploration:
- `explore/<description>` (e.g., `explore/auth-system`, `explore/api-patterns`)

Worktree naming:
- Use descriptive names (e.g., `explore-auth`, `explore-api`)

### 3. Verify Worktree

```bash
cd ~/source/<repo-name>-worktrees/explore-<name>
pwd
git branch
```

### 4. Ensure bots/ Directory

```bash
mkdir -p bots
```

The `bots/` directory is for the exploration report (explore.md).

## DO NOT
- ❌ Do not skip worktree creation
- ❌ Do not work directly on main branch
- ❌ Do not start exploring yet
- ❌ Do not make any code changes (exploration is read-only)

## When You're Done
Once the worktree is created and you're in the worktree directory:

1. Confirm to user: "Worktree created at: <path> for exploration"
2. Report your progress:
   ```
   workflow_report_progress(
       worktreePath: "<full-worktree-path>",
       currentStep: "WORKTREE",
       metadata: { "branch": "<branch-name>", "worktreePath": "<full-path>" }
   )
   ```

## End Step

Ask bob what to do next based on the metadata you provided with bob_workflow_get_guidance.
