# Workflow Resilience Features

## Overview

Bob now supports rejoining workflows at any step and resetting workflow state. This makes workflows more resilient to interruptions and allows flexible workflow management.

## New MCP Tools

### workflow_rejoin

Rejoin an existing workflow at a specified step with optional state reset.

**Parameters:**
- `worktreePath` (required): Absolute path to the git worktree
- `step` (optional): Step name to rejoin at (default: continue from current step)
- `taskDescription` (optional): Updated task description (preserves existing if omitted)
- `resetSubsequent` (optional): Reset progress for steps after rejoining step (default: "true", accepts: "true", "false")
- `sessionID` (optional): Session identifier
- `agentID` (optional): Agent identifier

**Example:**
```json
{
  "worktreePath": "/home/user/project",
  "step": "IMPLEMENT",
  "taskDescription": "Updated task context",
  "resetSubsequent": "true"
}
```

**Use Cases:**
- Resume a completed workflow at a specific step
- Change task description mid-workflow
- Restart from a checkpoint after fixing issues
- Jump to a different phase with new requirements

### workflow_reset

Clear workflow state for a worktree, optionally archiving the state before reset.

**Parameters:**
- `worktreePath` (required): Absolute path to the git worktree
- `archive` (optional): Archive state before clearing (default: "true", accepts: "true", "false")
- `sessionID` (optional): Session identifier
- `agentID` (optional): Agent identifier

**Example:**
```json
{
  "worktreePath": "/home/user/project",
  "archive": "true"
}
```

**Use Cases:**
- Start completely fresh after completing a workflow
- Clear stuck/corrupted workflow state
- Clean up before starting a new workflow type
- Reset while preserving history via archive

## State Structure Changes

### New State Fields

**WorkflowState** now includes:
```go
type WorkflowState struct {
    // ... existing fields ...
    RejoinHistory []RejoinEvent `json:"rejoinHistory,omitempty"`
    ResetHistory  []ResetEvent  `json:"resetHistory,omitempty"`
}
```

### Event Tracking

**RejoinEvent:**
```go
type RejoinEvent struct {
    Timestamp       time.Time `json:"timestamp"`
    FromStep        string    `json:"fromStep"`
    ToStep          string    `json:"toStep"`
    ResetSubsequent bool      `json:"resetSubsequent"`
    Reason          string    `json:"reason,omitempty"`
}
```

**ResetEvent:**
```go
type ResetEvent struct {
    Timestamp    time.Time `json:"timestamp"`
    PreviousStep string    `json:"previousStep"`
    Reason       string    `json:"reason,omitempty"`
}
```

## Behavior Details

### Rejoin with Reset

When `resetSubsequent=true`:
1. Sets current step to the specified step
2. Removes all progress history entries after the rejoin step
3. Clears issues from subsequent steps
4. Adds rejoin event to progress history
5. Updates task description if provided

### Rejoin without Reset

When `resetSubsequent=false`:
1. Sets current step to the specified step
2. Preserves all progress history
3. Keeps all recorded issues
4. Adds rejoin event to progress history
5. Updates task description if provided

### Reset with Archive

When `archive=true`:
1. Creates archive directory at `~/.bob/state/archive/`
2. Saves state to timestamped archive file
3. Adds reset event to state before archiving
4. Deletes the active state file

### Reset without Archive

When `archive=false`:
1. Deletes the active state file immediately
2. No history is preserved

## Error Handling

### Invalid Step Name
- Validates step name against workflow definition
- Returns error if step doesn't exist in workflow

### Non-Existent Workflow
- Returns error if workflow hasn't been registered
- Use `workflow_register` to create a new workflow

### Concurrent Access
- State is loaded and saved as a whole file, but writes are not guaranteed to be atomic across processes
- Concurrent writes from multiple processes are not supported; if they occur, the last completed write may overwrite previous state and can, in failure scenarios, result in a corrupted state file
- For single-process use (typical), state management is reliable

## Examples

### Example 1: Resume Completed Brainstorm Workflow

```bash
# Workflow completed, but need to revisit PLAN phase
workflow_rejoin(
  worktreePath: "/home/user/myproject",
  step: "PLAN",
  taskDescription: "Revise plan based on new requirements",
  resetSubsequent: "true"
)
```

### Example 2: Update Task Description Mid-Workflow

```bash
# Continue from current step with updated description
workflow_rejoin(
  worktreePath: "/home/user/myproject",
  taskDescription: "Updated: Now focusing on performance optimization",
  resetSubsequent: "false"
)
```

### Example 3: Start Fresh After Completion

```bash
# First reset to clear state (with archive)
workflow_reset(
  worktreePath: "/home/user/myproject",
  archive: "true"
)

# Then register new workflow
workflow_register(
  workflow: "code-review",
  worktreePath: "/home/user/myproject",
  taskDescription: "Review security fixes"
)
```

### Example 4: Jump to Testing Phase

```bash
# Skip ahead to TEST phase after manual implementation
workflow_rejoin(
  worktreePath: "/home/user/myproject",
  step: "TEST",
  resetSubsequent: "true"
)
```

## Testing

Comprehensive tests are available in `state_manager_rejoin_test.go`:

```bash
cd cmd/bob
go test -v -run "TestRejoin|TestReset"
```

Test coverage includes:
- ✅ Rejoin at same step with new description
- ✅ Rejoin at different step with reset
- ✅ Invalid step name validation
- ✅ Non-existent workflow handling
- ✅ Rejoin history tracking
- ✅ Reset with archive
- ✅ Reset without archive
- ✅ Session and agent ID support
- ✅ Timestamp validation

## Backward Compatibility

- ✅ All existing workflows continue working unchanged
- ✅ New tools are additive (no breaking changes)
- ✅ State format is backward compatible (new fields optional)
- ✅ Old state files work with new code
- ✅ No impact on existing MCP tools

## Audit Trail

All workflow rejoin and reset actions are tracked:

1. **Progress History**: Each rejoin adds an entry with metadata
2. **Rejoin History**: Dedicated array tracking all rejoin events
3. **Reset History**: Dedicated array tracking all reset events
4. **Timestamps**: All events include precise timestamps

This audit trail helps understand workflow evolution and troubleshoot issues.

## Best Practices

### When to Use Rejoin
- ✅ Resume after completing a workflow
- ✅ Fix mistakes in earlier phases
- ✅ Update requirements mid-workflow
- ✅ Jump to specific phase with context
- ✅ Recover from workflow interruptions

### When to Use Reset
- ✅ Start completely fresh
- ✅ Clear corrupted workflow state
- ✅ Clean up after testing
- ✅ Switch to different workflow type
- ✅ Archive completed work

### Recommendations
- Always archive when resetting (default behavior)
- Use `resetSubsequent=true` when jumping backwards
- Update task description when context changes
- Review rejoin history to understand workflow evolution
- Keep session/agent IDs consistent for multi-agent workflows

## Implementation Details

### Files Modified
- `cmd/bob/state_manager.go` - Added Rejoin() and Reset() methods
- `cmd/bob/mcp_server.go` - Registered workflow_rejoin and workflow_reset tools
- `cmd/bob/state_manager_rejoin_test.go` - Comprehensive test suite

### State Storage
- Active state: `~/.bob/state/<workflow-id>.json`
- Archives: `~/.bob/state/archive/<workflow-id>-archived-<timestamp>.json`

### Thread Safety
- Uses file-based state management
- Atomic read/write operations
- Latest write wins for concurrent access

---

**🏴‍☠️ Belayin' Pin Bob - Now Even More Resilient!**
