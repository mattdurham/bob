# Usage Guide: bob MCP

An MCP server for **B**rainstorm, **P**lan, **B**uild workflows with integrated task management.

## Quick Start

### 1. Configure Claude with MCP

Add to your Claude MCP configuration (`~/.claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "bob": {
      "command": "/path/to/bob",
      "args": ["--serve"],
      "env": {
        "GITHUB_TOKEN": "ghp_your_token_here"
      }
    }
  }
}
```

### 2. Restart Claude

After adding the MCP server, restart Claude for it to take effect.

## Agent Workflow

### Starting a New Task

When a user asks Claude to do development work, follow this pattern:

#### Step 1: List Available Workflows
```javascript
// Ask MCP what workflows are available
workflow_list_workflows()

// Response:
{
  "workflows": [
    {
      "keyword": "brainstorm",
      "name": "Brainstorm Development Workflow",
      "description": "Full development workflow with brainstorming..."
    }
  ]
}
```

#### Step 2: Create Worktree (INIT Phase)
```bash
# Pull latest
git checkout main
git pull origin main

# Create worktree
git worktree add -b feature/my-feature ~/source/personal-worktrees/my-feature
cd ~/source/personal-worktrees/my-feature
mkdir -p bots
```

#### Step 3: Register Workflow
```javascript
workflow_register({
    workflow: "brainstorm",
    worktreePath: "/home/matt/source/personal-worktrees/my-feature",
    taskDescription: "Add JSON output to codepath-dag"
})

// Response:
{
  "workflowId": "personal/my-feature",
  "workflow": "brainstorm",
  "currentStep": "INIT",
  "steps": [/* array of all steps */],
  "registeredAt": "2024-02-07T22:30:00Z"
}
```

#### Step 4: Get Guidance for Current Step
```javascript
workflow_get_guidance({
    worktreePath: "/home/matt/source/personal-worktrees/my-feature"
})

// Response:
{
  "currentStep": "INIT",
  "prompt": "# INIT Phase\n\nYou are currently in the **INIT** phase...",
  "nextStep": "WORKTREE",
  "canLoopBack": []
}
```

The `prompt` field contains the full markdown guidance for the current step.

#### Step 5: Complete Step and Report Progress
```javascript
// After completing the INIT step (creating worktree)
workflow_report_progress({
    worktreePath: "/home/matt/source/personal-worktrees/my-feature",
    currentStep: "WORKTREE",
    metadata: {
        branch: "feature/my-feature"
    }
})

// Response:
{
  "recorded": true,
  "currentStep": "WORKTREE",
  "previousStep": "INIT",
  "loopCount": 0,
  "timestamp": "2024-02-07T22:31:00Z"
}
```

#### Step 6: Continue Through Workflow
Repeat steps 4-5 for each phase:
- WORKTREE → Verify setup
- BRAINSTORM → Explore approaches (write to bots/brainstorm.md)
- PLAN → Create detailed plan (write to bots/plan.md)
- EXECUTE → Implement with TDD
- TEST → Run all checks
- REVIEW → Code review with subagent
- COMMIT → Commit changes
- MONITOR → Watch PR/CI
- COMPLETE → Merge and cleanup

### Handling Issues (Loop Back)

#### When Review Finds Issues

```javascript
// During REVIEW phase, record issues
workflow_record_issues({
    worktreePath: "/home/matt/source/personal-worktrees/my-feature",
    step: "REVIEW",
    issues: [
        {
            severity: "high",
            description: "Missing error handling in parseJSON",
            file: "main.go",
            line: 245
        },
        {
            severity: "medium",
            description: "Untested edge case for empty input",
            file: "main.go",
            line: 230
        }
    ]
})

// Response:
{
  "recorded": true,
  "issueCount": 2,
  "shouldLoop": true,
  "loopBackTo": "PLAN",
  "totalIssues": 2
}
```

Then loop back:
```javascript
workflow_report_progress({
    worktreePath: "/home/matt/source/personal-worktrees/my-feature",
    currentStep: "PLAN",
    metadata: {
        loopReason: "review issues",
        iteration: 2
    }
})
```

### Checking Status Anytime

```javascript
workflow_get_status({
    worktreePath: "/home/matt/source/personal-worktrees/my-feature"
})

// Response:
{
  "workflowId": "personal/my-feature",
  "workflow": "brainstorm",
  "taskDescription": "Add JSON output to codepath-dag",
  "currentStep": "EXECUTE",
  "loopCount": 1,
  "issueCount": 2,
  "progressHistory": [
    { "step": "INIT", "timestamp": "..." },
    { "step": "WORKTREE", "timestamp": "..." },
    { "step": "BRAINSTORM", "timestamp": "..." },
    { "step": "PLAN", "timestamp": "..." },
    { "step": "REVIEW", "timestamp": "..." },
    { "step": "PLAN", "timestamp": "...", "metadata": {"loopReason": "review issues"} },
    { "step": "EXECUTE", "timestamp": "..." }
  ],
  "issues": [/* all issues recorded */],
  "startedAt": "2024-02-07T22:30:00Z",
  "updatedAt": "2024-02-07T22:45:00Z"
}
```

## Complete Example Session

```javascript
// 1. List workflows
workflow_list_workflows()

// 2. Create worktree (INIT)
// [execute git commands]

// 3. Register
workflow_register({
    workflow: "brainstorm",
    worktreePath: "/home/matt/source/personal-worktrees/add-json",
    taskDescription: "Add JSON output format"
})

// 4. Get guidance → Read prompt
workflow_get_guidance({ worktreePath: "/home/matt/source/personal-worktrees/add-json" })
// [Read and follow INIT guidance]

// 5. Report progress to WORKTREE
workflow_report_progress({
    worktreePath: "/home/matt/source/personal-worktrees/add-json",
    currentStep: "WORKTREE"
})

// 6. Get guidance → Read prompt
workflow_get_guidance({ worktreePath: "/home/matt/source/personal-worktrees/add-json" })
// [Read and follow WORKTREE guidance]

// 7. Report progress to BRAINSTORM
workflow_report_progress({
    worktreePath: "/home/matt/source/personal-worktrees/add-json",
    currentStep: "BRAINSTORM"
})

// ... continue through all steps ...

// Final step: COMPLETE
workflow_report_progress({
    worktreePath: "/home/matt/source/personal-worktrees/add-json",
    currentStep: "COMPLETE",
    metadata: { merged: true }
})
```

## Key Principles

1. **Agent Self-Reports**: The agent decides when to move to the next step and explicitly reports progress
2. **MCP Guides**: The MCP provides detailed prompts for each step but doesn't control progression
3. **No Automatic Advancement**: Each prompt explicitly says "DO NOT automatically move to next phase"
4. **Explicit State**: All state transitions are tracked in JSON files under `~/.claude/workflows/`
5. **Loop-Aware**: The system handles review→plan and monitor→plan loops automatically

## State Files

Workflow state is persisted in:
```
~/.claude/workflows/personal-add-json.json
```

Example state file:
```json
{
  "workflowId": "personal/add-json",
  "workflow": "brainstorm",
  "worktreePath": "/home/matt/source/personal-worktrees/add-json",
  "taskDescription": "Add JSON output format",
  "currentStep": "EXECUTE",
  "progressHistory": [
    { "step": "INIT", "timestamp": "2024-02-07T22:30:00Z" },
    { "step": "WORKTREE", "timestamp": "2024-02-07T22:31:00Z" }
  ],
  "issues": [],
  "loopCount": 0,
  "metadata": {},
  "startedAt": "2024-02-07T22:30:00Z",
  "updatedAt": "2024-02-07T22:35:00Z"
}
```

## Troubleshooting

### Workflow Not Found
If you get "workflow not found", you need to register it first:
```javascript
workflow_register({ workflow: "brainstorm", worktreePath: "...", taskDescription: "..." })
```

### Unknown Step
Make sure step names match exactly (all caps): `INIT`, `WORKTREE`, `BRAINSTORM`, etc.

### Prompt Not Found
Check that prompt files exist in `prompts/brainstorm/01-init.md` etc.
