# INIT Phase

You are currently in the **INIT** phase of the performance optimization workflow.

## State Management

**IMPORTANT:** You can store and retrieve state throughout this workflow:
- **Store state**: Include any important information in the `metadata` field when calling `workflow_report_progress`
- **Retrieve state**: Call `workflow_get_guidance` at the start of any step to retrieve all stored state
- **State examples**: baseline metrics, bottleneck locations, optimization targets, benchmark results, performance gains

## Your Goal
Create a git worktree for performance work.

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

### 1. Pull Latest
```bash
git checkout main
git pull origin main
```

### 2. Create Worktree
```bash
mkdir -p ~/source/<repo-name>-worktrees
git worktree add -b perf-opt-<timestamp> ~/source/<repo-name>-worktrees/perf-opt
```

### 3. Setup
```bash
cd ~/source/<repo-name>-worktrees/perf-opt
mkdir -p bots
```

## DO NOT
- ❌ Do not skip worktree
- ❌ Do not start optimizing yet

## When You're Done
1. Confirm: "Worktree created for performance optimization"
2. Report progress:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "INIT",
       metadata: { "worktreeCreated": true }
   )
   ```

