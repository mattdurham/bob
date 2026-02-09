# 🏴‍☠️ Belayin' Pin Bob
## Captain of Your Agents

```
                                     |    |    |
                                    )_)  )_)  )_)
                                   )___))___))___)\
                                  )____)____)_____)\\
                                _____|____|____|____\\\__
                       ---------\                   /---------
                         ^^^^^ ^^^^^^^^^^^^^^^^^^^^^
                           ^^^^      ^^^^     ^^^    ^^
                                ^^^^      ^^^
```

**Belayin' Pin Bob** - Your trusty captain for orchestrating AI agent workflows!

> *"A belayin' pin is what keeps the ship's riggin' in order. Bob keeps your agents in line!"*

## What is Bob?

Bob is an MCP (Model Context Protocol) server that orchestrates AI agent workflows. Just like a ship's captain uses a belayin' pin to secure the ship's lines and rigging, Bob keeps your AI agent workflows organized, coordinated, and running smoothly.

## Features

- 🎯 **Workflow Orchestration** - Multi-step workflows with loop-back rules
- 📊 **Task Management** - Git-backed task tracking with dependencies
- 🔌 **MCP Server** - Claude integration via stdio protocol
- 💾 **Persistent State** - JSON state shared across all sessions
- 🔄 **Agent Self-Reporting** - Agents track and report their own progress
- 🏴‍☠️ **Captain of Your Agents** - Keep your AI workflows in line!

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/mattdurham/bob.git
cd bob

# Build Bob
make build
```

### Running Bob

Bob runs as an MCP server using stdio protocol for Claude integration:

```bash
# Run Bob as MCP server
cd cmd/bob
./bob --serve

# Or use make
make run
```

### MCP Configuration

Add Bob to your MCP configuration:

```json
{
  "mcpServers": {
    "bob": {
      "command": "/path/to/bob/cmd/bob/bob",
      "args": ["--serve"]
    }
  }
}
```

See [CLAUDE.md](CLAUDE.md) for detailed configuration instructions.

## Built-in Workflows

Bob includes four production-ready workflows:

### 1. brainstorm
Full development workflow with planning and iteration:
```
INIT → WORKTREE → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMMIT → MONITOR → COMPLETE
          ↑                                              ↓           ↓
          └──────────────────[issues found]─────────────┴───────────┘
```

### 2. code-review
Review, fix, test, and iterate until clean:
```
INIT → REVIEW → FIX → TEST → COMMIT → MONITOR → COMPLETE
        ↑        ↓     ↓
        └────────┴─────┘
```

### 3. performance
Benchmark, analyze, optimize, and verify:
```
INIT → BENCHMARK → ANALYZE → OPTIMIZE → VERIFY → COMMIT → MONITOR → COMPLETE
                      ↑          ↓         ↓
                      └──────────┴─────────┘
```

### 4. explore
Read-only codebase exploration:
```
DISCOVER → ANALYZE → DOCUMENT → COMPLETE
```

See [AGENTS.md](AGENTS.md) for detailed workflow documentation.

## Task Management

Bob manages tasks in `.bob/issues/` with git branch integration:

```bash
# Tasks are stored in git on the 'bob' or 'bob' branch
# Each task is a JSON file: .bob/issues/<id>.json
```

**Task Properties:**
- `id` - Unique identifier
- `title` - Task title
- `description` - Detailed description
- `state` - open, in_progress, completed, blocked
- `priority` - low, medium, high, critical
- `dependencies` - Task IDs this depends on
- `labels` - Categorization tags

## Custom Workflows

Create custom workflows in `.bob/workflows/*.json`:

```json
{
  "keyword": "my-workflow",
  "name": "My Custom Workflow",
  "description": "STEP1 → STEP2 → STEP3",
  "steps": [
    {
      "name": "STEP1",
      "description": "First step instructions"
    },
    {
      "name": "STEP2",
      "description": "Second step instructions"
    }
  ],
  "loopRules": [
    {
      "fromStep": "STEP2",
      "toStep": "STEP1",
      "condition": "needs_retry",
      "description": "Retry if needed"
    }
  ]
}
```

## Architecture

```
┌─────────────────┐
│ Claude Session 1│────┐
└─────────────────┘    │
                       │
┌─────────────────┐    │     ┌──────────────────┐
│ Claude Session 2│────┼────▶│ ~/.bob/state/    │
└─────────────────┘    │     │   state/         │
                       │     └──────────────────┘
┌─────────────────┐    │
│ Claude Session N│────┘
└─────────────────┘

   bob --serve          Shared JSON State Files
```

**How it works:**
- Each Claude session runs `bob --serve` as an MCP server
- All sessions write to `~/.bob/~/.bob/state/` (shared JSON state)
- Workflows and tasks persist across all sessions
- Updates from any session appear in all sessions

## Storage

Bob stores all state in `~/.bob/state/`:
- `state/` - JSON state with workflow and task state
- Shared across all Bob MCP server instances
- Updates appear immediately in all Claude sessions

## Development

```bash
# Install dependencies
make install-deps

# Run tests
make test

# Clean build artifacts
make clean
```

## Available MCP Tools

### Workflow Management
- `bob_workflow_list_workflows` - List available workflows
- `bob_workflow_get_definition` - Get workflow definition
- `bob_workflow_register` - Start new workflow session
- `bob_workflow_get_guidance` - Get current step guidance
- `bob_workflow_report_progress` - Advance to next step
- `bob_workflow_get_status` - Get workflow status
- `bob_workflow_get_session_status` - Get session-specific status

### Task Management
- `bob_task_create` - Create new task
- `bob_task_get` - Get task by ID
- `bob_task_list` - List all tasks with filters
- `bob_task_update` - Update task properties
- `bob_task_add_dependency` - Add task dependency
- `bob_task_add_comment` - Add comment to task
- `bob_task_get_ready` - Get ready-to-work tasks

## Development Principles

Bob follows these core principles:

1. **Workflows are loops** - Most workflows need iteration
2. **Review before fix** - MONITOR → REVIEW → FIX (not MONITOR → FIX)
3. **State is persistent** - All workflow state saved to JSON
4. **Git-based tasks** - Tasks stored in git for durability
5. **MCP-first** - Built for Claude integration
6. **Agent self-reporting** - Agents decide when to advance steps

## Contributing

Bob is your ship's captain - if you've got improvements to the riggin', send a pull request!

## License

MIT License - See LICENSE file

---

*🏴‍☠️ Fair winds and following seas! - Captain Bob*
