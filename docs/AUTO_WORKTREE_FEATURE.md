# Auto-Worktree Creation Feature

## Summary

Enhanced `workflow_register` to automatically create git worktrees when registering workflows on the main branch.

## Problem Solved

**Before:**
- Claude had to manually remember to create worktrees first
- Easy to accidentally start working on main branch
- Workflow registration failed if already on main

**After:**
- Bob automatically creates worktrees when needed
- Just provide `featureName` parameter
- No manual git commands required

## How It Works

### Enhanced workflow_register Tool

```json
{
  "tool": "workflow_register",
  "parameters": {
    "workflow": "brainstorm",
    "worktreePath": "/home/matt/source/bob",  // Can be main repo now!
    "featureName": "my-feature",  // Auto-creates worktree
    "taskDescription": "Add new feature"
  }
}
```

**Bob will:**
1. Detect you're on main branch
2. Auto-create worktree at `<repo>-worktrees/<featureName>`
3. Create branch `feature/<featureName>`
4. Create `.bob/` directory
5. Register workflow in the new worktree
6. Return the worktree path for Claude to `cd` into

### Return Value

```json
{
  "workflowId": "...",
  "workflow": "brainstorm",
  "worktreePath": "/home/matt/source/bob-worktrees/my-feature",
  "branch": "feature/my-feature",
  "createdWorktree": true,
  "message": "Created worktree at: /home/matt/source/bob-worktrees/my-feature\nBranch: feature/my-feature\nRun: cd /home/matt/source/bob-worktrees/my-feature"
}
```

## Use Cases

### Case 1: Auto-Create Worktree from Main

```bash
# Claude calls:
workflow_register(
  workflow: "brainstorm",
  worktreePath: "/home/matt/source/bob",  # Main repo
  featureName: "add-auth",  # Auto-creates worktree
  taskDescription: "Add authentication system"
)

# Bob creates:
# - /home/matt/source/bob-worktrees/add-auth/
# - Branch: feature/add-auth
# - .bob directory
```

### Case 2: Use Existing Worktree

```bash
# Claude calls:
workflow_register(
  workflow: "brainstorm",
  worktreePath: "/home/matt/source/bob-worktrees/existing-feature",
  # featureName is optional - Bob detects it's already a worktree
  taskDescription: "Continue working on feature"
)

# Bob uses existing worktree, no creation needed
```

### Case 3: Error Without featureName

```bash
# Claude calls:
workflow_register(
  workflow: "brainstorm",
  worktreePath: "/home/matt/source/bob",  # Main repo
  # Missing featureName!
  taskDescription: "Some task"
)

# Bob returns error:
# "cannot register workflow on main branch without featureName parameter"
```

## Implementation Details

### New StateManager Methods

**isMainRepo(path)**
- Detects if path is main repository vs worktree
- Checks if `.git` is directory (main) or file (worktree)
- Returns: (isMain bool, repoRoot string, error)

**createWorktree(repoPath, featureName)**
- Creates worktree at `<repo>-worktrees/<featureName>`
- Creates branch `feature/<featureName>`
- Checks out main and pulls latest
- Creates `.bob/` directory
- Returns: (worktreePath string, branchName string, error)

### Modified Methods

**Register(workflow, worktreePath, taskDescription, featureName, sessionID, agentID)**
- Added `featureName` parameter
- Calls `isMainRepo()` to detect repository type
- Auto-creates worktree if on main and featureName provided
- Errors if on main without featureName
- Uses existing worktree if already in one

## Workflow Integration

### Correct Order (Handled by Bob)

1. Claude calls `workflow_register` with featureName
2. Bob detects main branch
3. Bob creates worktree automatically
4. Bob registers workflow in new worktree
5. Claude `cd`s into new worktree
6. Follow workflow steps (INIT → WORKTREE → BRAINSTORM → etc.)

### Error Prevention

**Scenario: Trying to work on main**
```
❌ Before: Claude had to remember to create worktree
✅ After: Bob errors: "cannot register workflow on main without featureName"
```

**Scenario: Forgot feature name**
```
❌ Before: Would register workflow on main (bad!)
✅ After: Bob requires featureName parameter
```

## Migration Guide

### For Existing Workflows

Old code:
```bash
# Manual worktree creation
git worktree add ../bob-worktrees/my-feature -b feature/my-feature
cd ../bob-worktrees/my-feature

# Then register
workflow_register(
  worktreePath: "/path/to/worktree",
  ...
)
```

New code:
```bash
# Bob does it all!
workflow_register(
  worktreePath: "/home/matt/source/bob",  # Main repo
  featureName: "my-feature",
  ...
)

# Then cd to the returned worktreePath
```

### Backward Compatibility

✅ Existing worktree paths still work (no breaking changes)
✅ featureName is optional for worktrees
✅ Old Register() calls work if you add empty featureName: ""

## Testing

Tests cover:
- ✅ Auto-creation from main repo
- ✅ Error without featureName on main
- ✅ Using existing worktrees
- ✅ isMainRepo() detection logic
- ✅ Branch naming conventions
- ⚠️ Some old tests need updating for new signature

## Benefits

1. **Simpler API**: One call instead of multiple git commands
2. **Error Prevention**: Can't accidentally work on main
3. **Consistency**: Enforces naming conventions
4. **Automation**: Bob handles all git operations
5. **Clarity**: Clear error messages guide correct usage

## Future Enhancements

Potential improvements:
- Support custom branch prefixes (fix/, refactor/, etc.)
- Auto-detect feature name from task description
- Validate feature name format
- Support worktree cleanup after workflow completion
- Integration with PR creation

---

**🏴‍☠️ Bob now captains your worktrees!**
