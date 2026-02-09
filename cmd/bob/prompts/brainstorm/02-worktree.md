# WORKTREE Phase

You are currently in the **WORKTREE** phase of the workflow.

## Your Goal
Verify the worktree is set up correctly before beginning work.

## What To Do

### 1. Verify Location
```bash
pwd
# Should be in ~/source/<repo>-worktrees/<name>
```

### 2. Verify Branch
```bash
git branch
# Should show your feature branch
```

### 3. Verify bots/ Directory
```bash
ls -la bots/
# Should exist (create if missing)
mkdir -p bots
```

### 4. Check Working Directory is Clean
```bash
git status
# Should show clean working directory
```

## DO NOT
- ❌ Do not start coding yet
- ❌ Do not create files outside of bots/ yet
- ❌ Do not automatically move to next phase

## When You're Done
Once verified:

1. Confirm to user: "Worktree verified, ready to begin brainstorming"
2. Report your progress:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "BRAINSTORM",
       metadata: { "verified": true }
   )
   ```

## Next Phase
After reporting progress, you'll move to **BRAINSTORM** phase to explore the problem space.
