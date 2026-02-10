# WORKTREE Phase

You are currently in the **WORKTREE** phase of the code-review workflow.

## State Management

**IMPORTANT:** You can retrieve state from previous steps:
- **Retrieve state**: Call `workflow_get_guidance` to retrieve review scope from PROMPT phase
- The review scope will help you name the branch and worktree appropriately

## Your Goal
Create a git worktree for isolated code review based on the scope from the PROMPT phase.

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

### 1. Get Review Scope from Previous Step

Retrieve the guidance to see the review scope from PROMPT:
```
workflow_get_guidance(worktreePath: "<current-worktree-path>")
```

Look for `reviewScope` in the stored metadata.

### 2. Create Worktree

Based on the review scope, create an appropriately named worktree:

```bash
# Ensure you're in the main repo
cd ~/source/<repo-name>
git checkout main
git pull origin main

# Create worktree for code review
mkdir -p ~/source/<repo-name>-worktrees
git worktree add -b review/<description> ~/source/<repo-name>-worktrees/review-<name>
```

Branch naming for reviews:
- `review/<description>` (e.g., `review/auth-fixes`, `review/performance`)

Worktree naming:
- Use descriptive names (e.g., `review-auth`, `review-perf`)

### 3. Verify Worktree

```bash
cd ~/source/<repo-name>-worktrees/review-<name>
pwd
git branch
```

### 4. Ensure bots/ Directory

```bash
mkdir -p bots
```

The `bots/` directory is for review findings (review.md).

## DO NOT
- ❌ Do not skip worktree creation
- ❌ Do not work directly on main branch
- ❌ Do not start reviewing yet

## When You're Done
Once the worktree is created and you're in the worktree directory:

1. Confirm to user: "Worktree created at: <path> for code review"
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
