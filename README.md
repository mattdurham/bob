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
- 🔌 **MCP Servers** - Bob workflow orchestrator + filesystem operations
- 💾 **Persistent State** - JSON state shared across all sessions
- 🔄 **Agent Self-Reporting** - Agents track and report their own progress
- 🌐 **Web UI** - Browser-based dashboard for viewing workflows and tasks
- 📁 **Filesystem Access** - Secure file operations in allowed directories
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

Bob can run in two modes:

**MCP Server Mode** (for Claude integration):
```bash
# Run Bob as MCP server
cd cmd/bob
./bob --serve

# Or use make
make run
```

**Web UI Mode** (for viewing workflows in browser):
```bash
# Start web UI server (default: http://127.0.0.1:8080)
./bob --ui

# Custom port
./bob --ui --port 8081

# Custom host (use with caution)
./bob --ui --host 0.0.0.0 --port 8080
```

Then open your browser to http://127.0.0.1:8080 (or your configured port)

### MCP Configuration

Bob provides two MCP servers (workflow orchestrator + filesystem operations). Install both with one command:

```bash
make install-mcp
```

This installs:
- **Bob** - Workflow orchestration and task management
- **Filesystem server** - Secure file operations (restricted to `$HOME/source` and `/tmp`)

The command automatically detects and configures both Claude and Codex (if their CLIs are available).

**Manual Configuration:**

For Claude, add both servers to your MCP configuration:
```json
{
  "mcpServers": {
    "bob": {
      "command": "$HOME/.bob/bob",
      "args": ["--serve"]
    },
    "filesystem": {
      "command": "mcp-filesystem-server",
      "args": ["$HOME/source", "/tmp"]
    }
  }
}
```

For Codex, use:
```bash
codex mcp add bob -- ~/.bob/bob --serve
codex mcp add filesystem -- mcp-filesystem-server "$HOME/source" /tmp
```

**Platform-Specific Documentation:**
- [CLAUDE.md](CLAUDE.md) - Claude configuration and usage
- [CODEX.md](CODEX.md) - Codex configuration and usage

## Built-in Workflows

Bob includes four production-ready workflows:

### 1. work
Full development workflow with planning and iteration:
```
INIT → PROMPT → WORKTREE → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMMIT → MONITOR → COMPLETE
                   ↑                                                 ↓           ↓
                   └──────────────────[issues found]────────────────┴───────────┘
```

### 2. code-review
Review, fix, test, and iterate until clean:
```
INIT → PROMPT → WORKTREE → REVIEW → FIX → TEST → COMMIT → MONITOR → COMPLETE
                             ↑        ↓     ↓
                             └────────┴─────┘
```

### 3. performance
Benchmark, analyze, optimize, and verify:
```
INIT → PROMPT → WORKTREE → BENCHMARK → ANALYZE → OPTIMIZE → VERIFY → COMMIT → MONITOR → COMPLETE
                               ↑          ↓         ↓
                               └──────────┴─────────┘
```

### 4. explore
Read-only codebase exploration:
```
INIT → PROMPT → WORKTREE → DISCOVER → ANALYZE → DOCUMENT → COMPLETE
```

See [AGENTS.md](AGENTS.md) for detailed workflow documentation.

## Web UI

Bob includes a built-in web UI for monitoring workflows and tasks:

### Features
- **Dashboard** - View all active workflows at a glance
- **Workflow Details** - See complete workflow history and progress
- **Progress Timeline** - Visual timeline of workflow steps
- **Issue Tracking** - View issues found during workflow execution
- **Metadata Display** - See all collected workflow metadata

### Usage

Start the web UI server:
```bash
./bob --ui
```

Open http://127.0.0.1:8080 in your browser to see:
- Active workflows with current step
- Workflow statistics (loops, issues, updates)
- Detailed workflow progress history
- Task counts (coming soon)

### Security

The UI defaults to `127.0.0.1` (localhost only) for security. Only use `--host 0.0.0.0` on trusted networks.

### Implementation

The web UI:
- Runs independently of the MCP server (separate process)
- Reads workflow state from `~/.bob/state/`
- Uses embedded Go templates (no external files needed)
- Single binary with all assets included
- Refreshes on page load (no WebSocket yet)

## Task Management

Bob manages tasks in `.bob/issues/` with git branch integration:

```bash
# Tasks are stored in git on the 'bob' branch
# Each task is a JSON file: .bob/issues/<id>.json
```

**Task Properties:**
- `id` - Unique identifier
- `title` - Task title
- `description` - Detailed description
- `state` - open, in_progress, completed, blocked
- `priority` - low, medium, high, critical
- `dependencies` - Task IDs this depends on
- `tags` - Categorization tags

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
- All sessions write to `~/.bob/state/` (shared JSON state)
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

### Filesystem Operations (mark3labs/mcp-filesystem-server)
- `filesystem_read_file` - Read file contents
- `filesystem_write_file` - Write or create files
- `filesystem_list_directory` - List directory contents
- `filesystem_create_directory` - Create directories
- `filesystem_search_files` - Search by filename pattern
- `filesystem_search_within_files` - Search file contents
- `filesystem_get_file_info` - Get file metadata
- `filesystem_copy_file` - Copy files
- `filesystem_move_file` - Move/rename files
- `filesystem_delete_file` - Delete files
- `filesystem_tree` - Get directory tree structure
- `filesystem_read_multiple_files` - Read multiple files at once

**Security**: Filesystem access restricted to `$HOME/source` and `/tmp`

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

# Developing Bob

Note all development work in Bob must go through a bob workflow session.
