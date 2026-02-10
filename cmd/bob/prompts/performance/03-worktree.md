# WORKTREE Phase

You are currently in the **WORKTREE** phase of the performance workflow.

## State Management

**IMPORTANT:** You can retrieve state from previous steps:
- **Retrieve state**: Call `workflow_get_guidance` to retrieve performance goals from PROMPT phase
- The performance goals will help you name the branch and worktree appropriately

## Your Goal
Create a git worktree for isolated performance work based on the goals from the PROMPT phase.

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

### 1. Get Performance Goals from Previous Step

Retrieve the guidance to see the performance goals from PROMPT:
```
workflow_get_guidance(worktreePath: "<current-worktree-path>")
```

Look for `performanceGoal` in the stored metadata.

### 2. Create Worktree

Based on the performance goals, create an appropriately named worktree:

```bash
# Ensure you're in the main repo
cd ~/source/<repo-name>
git checkout main
git pull origin main

# Create worktree for performance work
mkdir -p ~/source/<repo-name>-worktrees
git worktree add -b perf/<description> ~/source/<repo-name>-worktrees/perf-<name>
```

Branch naming for performance:
- `perf/<description>` (e.g., `perf/optimize-queries`, `perf/reduce-memory`)

Worktree naming:
- Use descriptive names (e.g., `perf-queries`, `perf-memory`)

### 3. Verify Worktree

```bash
cd ~/source/<repo-name>-worktrees/perf-<name>
pwd
git branch
```

### 4. Ensure bots/ Directory

```bash
mkdir -p bots
```

The `bots/` directory is for performance reports and benchmarks.

## DO NOT
- ❌ Do not skip worktree creation
- ❌ Do not work directly on main branch
- ❌ Do not start benchmarking yet

## When You're Done
Once the worktree is created and you're in the worktree directory:

1. Confirm to user: "Worktree created at: <path> for performance work"
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
