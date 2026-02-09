# bob - Brainstorm, Plan, Build

An MCP (Model Context Protocol) server that guides AI agents through structured development workflows and manages tasks using keyword-triggered workflow templates.

## Features

- **Multiple Workflow Types**: Keyword-based workflow selection (e.g., "brainstorm", "hotfix", "research")
- **Agent Self-Reporting**: Agent tracks itself and reports progress to MCP
- **Git Worktree Tracking**: Uses worktree path as unique workflow identifier
- **Step-by-Step Guidance**: Detailed prompts for each workflow phase
- **Loop Detection**: Automatically handles review→plan and monitor→plan loops
- **Persistent State**: JSON-based state tracking survives Claude restarts
- **Zero Dependencies**: Pure Go standard library

## Workflows

### `brainstorm` Workflow
Full development workflow with brainstorming phase:
```
INIT → WORKTREE → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMMIT → MONITOR → COMPLETE
           ↑                                              ↓           ↓
           └──────────────────[issues found]─────────────┴───────────┘
```

**Steps:**
1. **INIT** - Create git worktree
2. **WORKTREE** - Verify worktree setup
3. **BRAINSTORM** - Explore approaches, clarify requirements
4. **PLAN** - Create detailed implementation plan
5. **EXECUTE** - Implement with TDD
6. **TEST** - Run all tests and checks
7. **REVIEW** - Code review with subagent
8. **COMMIT** - Commit changes
9. **MONITOR** - Watch PR/CI, respond to feedback
10. **COMPLETE** - Merge and cleanup

**Loop Points:**
- REVIEW → PLAN (if issues found)
- MONITOR → PLAN (if CI fails)
- TEST → EXECUTE (if tests fail)

## Installation

### Option 1: Using Makefile (Recommended)

From repository root:
```bash
# Build and register MCP with Claude automatically
make install-workflow-tracker-mcp

# This will:
# 1. Build the workflow-tracker binary
# 2. Add it to ~/.claude/claude_desktop_config.json
# 3. Configure it to run with --serve flag

# To uninstall
make uninstall-workflow-tracker-mcp
```

**Requirements:** `jq` must be installed for automatic registration.

### Option 2: Manual Installation

```bash
# Build
cd cmd/workflow-tracker
go build -o workflow-tracker

# Manually add to ~/.claude/claude_desktop_config.json:
{
  "mcpServers": {
    "workflow-tracker": {
      "command": "/path/to/workflow-tracker",
      "args": ["--serve"]
    }
  }
}

# Or install globally
go install github.com/mattdurham/personal/cmd/workflow-tracker@latest
```

## Usage

### As MCP Server

Add to Claude's MCP configuration:
```json
{
  "mcpServers": {
    "workflow-tracker": {
      "command": "/path/to/workflow-tracker",
      "args": ["--serve"]
    }
  }
}
```

### Available MCP Tools

1. **workflow_list_workflows** - List all available workflow types
2. **workflow_get_definition** - Get full workflow definition by keyword
3. **workflow_register** - Register new workflow instance
4. **workflow_report_progress** - Report current step
5. **workflow_get_guidance** - Get detailed prompt for current step
6. **workflow_record_issues** - Record issues (triggers loop logic)
7. **workflow_get_status** - Get complete workflow history

## Agent Usage Example

```javascript
// 1. Agent lists available workflows
workflow_list_workflows()
// Returns: { workflows: [{ keyword: "brainstorm", name: "...", ... }] }

// 2. Agent registers a new workflow
workflow_register({
    workflow: "brainstorm",
    worktreePath: "~/source/personal-worktrees/add-json-output",
    taskDescription: "Add JSON output to codepath-dag"
})
// Returns: { workflowId: "personal/add-json-output", currentStep: "INIT", ... }

// 3. Agent gets guidance for current step
workflow_get_guidance({
    worktreePath: "~/source/personal-worktrees/add-json-output"
})
// Returns: { currentStep: "INIT", prompt: "# INIT Phase\n\n...", nextStep: "WORKTREE" }

// 4. Agent completes step and reports progress
workflow_report_progress({
    worktreePath: "~/source/personal-worktrees/add-json-output",
    currentStep: "WORKTREE",
    metadata: { branch: "feature/add-json-output" }
})
// Returns: { recorded: true, currentStep: "WORKTREE", ... }

// 5. Agent records issues during review
workflow_record_issues({
    worktreePath: "~/source/personal-worktrees/add-json-output",
    step: "REVIEW",
    issues: [
        {
            severity: "high",
            description: "Missing error handling",
            file: "main.go",
            line: 123
        }
    ]
})
// Returns: { recorded: true, shouldLoop: true, loopBackTo: "PLAN" }

// 6. Agent checks workflow status anytime
workflow_get_status({
    worktreePath: "~/source/personal-worktrees/add-json-output"
})
// Returns: full state including history, issues, loop count
```

## State Persistence

Workflow states are stored in `~/.claude/workflows/`:
```
~/.claude/workflows/
├── personal-add-json-output.json
├── personal-fix-auth-bug.json
└── ...
```

Each state file contains:
- Workflow type and step
- Progress history
- Issues found
- Loop count
- Metadata

## Adding New Workflows

1. Create prompt directory: `prompts/<workflow-keyword>/`
2. Add numbered prompt files: `01-step-name.md`, `02-step-name.md`, etc.
3. Register workflow in `workflow_definition.go`:
   ```go
   func MyWorkflow() *WorkflowDefinition {
       return &WorkflowDefinition{
           Keyword: "myworkflow",
           Name: "My Custom Workflow",
           Steps: []Step{ /* ... */ },
           LoopRules: []LoopRule{ /* ... */ },
       }
   }
   ```
4. Add to `GetWorkflowDefinition()` map

## Web UI

BPB includes a modern web interface for monitoring workflows and managing tasks across all Claude sessions.

### Quick Start

From repository root:
```bash
# Show all available commands
make ui-help

# Development mode (recommended for development)
make ui-dev

# Production mode
make ui
```

### Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│ Claude Session 1│────▶│                 │     │   Web Browser   │
├─────────────────┤     │   SQLite DB     │◀────│  React + Go UI  │
│ Claude Session 2│────▶│  ~/.bob/state/  │     │  localhost:3001 │
├─────────────────┤     │    db.sql       │     └─────────────────┘
│ Claude Session N│────▶│                 │
└─────────────────┘     └─────────────────┘
   bob --serve             Shared State         bob --web
   (MCP stdio)                                  (HTTP server)
```

**Key Features:**
- **Multi-Session Monitoring**: View workflows and tasks from all Claude sessions
- **Real-Time Updates**: Auto-refreshes every 5 seconds
- **Interactive Workflow Diagrams**: React Flow visualizations with active step highlighting
- **Task Management**: View, filter, and update tasks directly from the UI
- **Repository Filtering**: Filter by repository to focus on specific projects

### Available Make Commands

```bash
# Quick start
make ui                 # Build and run production mode
make ui-dev             # Development mode (Go API + React dev server)

# Build
make ui-build           # Build both Go backend and React frontend
make ui-build-go        # Build Go backend only
make ui-build-react     # Build React frontend only

# Development
make ui-dev             # Sequential startup (cleaner output)
make ui-dev-parallel    # Parallel startup (faster but noisier)

# Maintenance
make ui-install         # Install dependencies (Go + npm)
make ui-clean           # Clean build artifacts
make ui-test            # Run all tests
```

### Manual Usage

```bash
# Start MCP server (one per Claude session)
cd cmd/bob
./bob --serve

# Start web UI (run once, serves all sessions)
cd cmd/bob
./bob --web              # Defaults to port 9091
./bob --web --port 8080  # Custom port
```

### Web UI URLs

- **Production**: http://localhost:9091 (served by Go)
- **Development**: http://localhost:3001 (React dev server)
- **API**: http://localhost:9091/api (used by React app)

### API Endpoints

The Go backend provides REST API endpoints:

- `GET /api/workflows/definitions` - All workflow definitions
- `GET /api/workflows/running` - Currently active workflows
- `GET /api/tasks` - All tasks across all repositories
- `PATCH /api/tasks/:id` - Update task (state, assignee, priority)

### Database Schema

All data is stored in `~/.bob/state/db.sql`:

```sql
-- Active workflow instances
workflows (id, workflow, current_step, task_description, loop_count, ...)

-- Workflow progress history
workflow_progress (workflow_id, step, metadata, timestamp, ...)

-- Tasks and issues
tasks (id, repo_path, title, description, state, priority, ...)

-- Task comments
task_comments (task_id, author, text, timestamp, ...)
```

### React Components

- **TopBar**: Workflow selector and view mode toggle (All Workflows / Running)
- **Sidebar**: Repository filter, active workflows, and task list
- **WorkflowDiagram**: Interactive React Flow diagram with step highlighting
- **TaskPanel**: Slide-in panel for task details and editing

## Design Philosophy

- **Agent-Driven**: Agent controls its own progress, MCP tracks and guides
- **Declarative Workflows**: Workflows defined as data structures
- **Explicit State Transitions**: Agent must explicitly report progress
- **No Magic**: All state changes are explicit and traceable
- **Loop-Aware**: Built-in support for iterative development (review → plan → execute loops)
- **Shared State**: Single SQLite database enables multi-session coordination
- **Real-Time Visibility**: Web UI provides live monitoring across all sessions

## License

See repository root LICENSE file.
