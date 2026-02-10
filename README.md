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
- `state` - open, in_progress, completed, blocked, cancelled
- `priority` - low, medium, high
- `dependencies` - Task IDs this depends on
- `tags` - Categorization tags
- `type` - feature, bug, chore, refactor, docs, test
- `assignee` - Person or agent assigned to task
- `blocks` - Array of task IDs this task blocks
- `blockedBy` - Array of task IDs blocking this task
- `comments` - Array of comment objects
- `metadata` - Key-value pairs for arbitrary data
- `workflowState` - Workflow-specific state data
- `createdAt` - ISO timestamp when task was created
- `updatedAt` - ISO timestamp of last update
- `completedAt` - ISO timestamp when task was completed

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
- `workflow_list_workflows` - List available workflows
- `workflow_get_definition` - Get workflow definition
- `workflow_register` - Start new workflow session
- `workflow_get_guidance` - Get current step guidance
- `workflow_report_progress` - Advance to next step
- `workflow_get_status` - Get workflow status
- `workflow_get_session_status` - Get session-specific status
- `workflow_rejoin` - Rejoin workflow at specific step
- `workflow_reset` - Clear workflow state
- `workflow_list_agents` - List active agents
- `workflow_record_issues` - Record issues found during execution

### Task Management
- `task_create` - Create new task
- `task_get` - Get task by ID
- `task_list` - List all tasks with filters
- `task_update` - Update task properties
- `task_add_dependency` - Add task dependency
- `task_add_comment` - Add comment to task
- `task_get_ready` - Get ready-to-work tasks
- `task_delete` - Delete a task
- `task_set_workflow_state` - Set workflow state for task
- `task_get_workflow_state` - Get workflow state for task
- `task_delete_workflow_state_key` - Delete workflow state key

**Note:** When calling through MCP, tools are automatically prefixed with the server name (e.g., `bob.workflow_list_workflows` or `bob.task_create`).

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

## Workflow Skills

Bob provides workflow orchestration through Claude skills. These skills coordinate specialized subagents via the Task tool to guide you through complete development workflows.

### Available Workflow Skills

#### `/work` - Full Development Workflow

Complete feature development from idea to merged PR.

**Usage:** `/work "Add user authentication feature"`

**Workflow Phases:**
```
INIT → PROMPT → WORKTREE → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMMIT → MONITOR → COMPLETE
                                           ↑                         ↓
                                           └─────────────────────────┘
                                                (loop on issues)
```

**What It Does:**
- Researches existing patterns (spawns Explore agent)
- Creates implementation plan (spawns planner agent)
- Implements changes (spawns coder agents)
- Runs tests (spawns tester agent)
- Reviews code (spawns reviewer agent)
- Commits and creates PR
- Monitors CI/checks

**Loop-back Rules:**
- REVIEW → PLAN (major architectural issues)
- REVIEW → EXECUTE (minor implementation fixes)
- TEST → EXECUTE (test failures)
- MONITOR → REVIEW (CI failures - always review first!)

---

#### `/code-review` - Code Review Workflow

Review existing code, identify issues, fix them, and verify fixes.

**Usage:** `/code-review`

**Workflow Phases:**
```
INIT → PROMPT → WORKTREE → REVIEW → FIX → TEST → COMMIT → MONITOR → COMPLETE
                              ↑               ↓
                              └───────────────┘
                                (loop on issues)
```

**What It Does:**
- Reviews code for issues (spawns reviewer agent)
- Fixes identified problems (spawns coder agents)
- Runs tests to verify (spawns tester agent)
- Re-reviews after fixes (loops back to REVIEW)
- Commits fixes and creates PR

**Loop-back Rules:**
- REVIEW → FIX (issues found)
- TEST → REVIEW (re-verify after fixes)
- MONITOR → REVIEW (CI failures)

---

#### `/performance` - Performance Optimization Workflow

Benchmark, analyze bottlenecks, optimize, and verify improvements.

**Usage:** `/performance "Optimize API response times"`

**Workflow Phases:**
```
INIT → PROMPT → WORKTREE → BENCHMARK → ANALYZE → OPTIMIZE → VERIFY → COMMIT → MONITOR → COMPLETE
                                          ↑                       ↓
                                          └───────────────────────┘
                                            (loop if targets not met)
```

**What It Does:**
- Runs baseline benchmarks (spawns tester agent)
- Analyzes performance bottlenecks (spawns researcher agent)
- Implements optimizations (spawns coder agents)
- Verifies improvements meet targets (spawns tester agent)
- Commits optimizations with before/after metrics

**Loop-back Rules:**
- VERIFY → ANALYZE (performance targets not met)
- MONITOR → ANALYZE (CI performance tests fail)

---

#### `/explore` - Codebase Exploration Workflow

Read-only exploration and documentation of codebase.

**Usage:** `/explore "Understand authentication flow"`

**Workflow Phases:**
```
INIT → PROMPT → DISCOVER → ANALYZE → DOCUMENT → COMPLETE
```

**What It Does:**
- Discovers relevant files and components (spawns Explore agent)
- Analyzes code and understands logic (spawns researcher agent)
- Creates comprehensive documentation (spawns documenter agent)
- No code changes (read-only workflow)

**No Loops:** This is a one-pass, read-only workflow.

---

### How Workflow Skills Work

**Skills are Orchestration Layers:**
- Skills don't implement features themselves
- Skills spawn Task tool subagents for actual work
- Skills enforce flow control rules (when to loop back)
- Skills use Bob MCP tools for state persistence

**Subagents Used:**
- **Explore** - Codebase discovery and search
- **planner** - Create implementation plans
- **coder** - Write and modify code
- **tester** - Run tests and benchmarks
- **reviewer** - Code review and quality checks
- **researcher** - Deep analysis and investigation

**State Management:**
- Skills use Bob MCP tools to persist workflow state
- Progress survives Claude CLI restarts
- Multiple users can collaborate on same workflow
- State stored in `~/.bob/state/`

---

## MCP Servers

When you run `make install-mcp-full`, Bob installs multiple MCP servers:

1. **Bob** - Workflow state persistence and task tracking
   - Tools: workflow_register, workflow_report_progress, task_create, task_update
   - Storage: ~/.bob/state/

2. **Filesystem** - Secure file operations in allowed directories
   - Allowed: ~/source, /tmp
   - Tools: read_file, write_file, search_files, etc.

3. **GitHub** - GitHub API integration (if available)
   - Tools: PR management, issue tracking, repository operations

4. **Sequential Thinking** - Enhanced reasoning (if available)
   - Advanced reasoning capabilities for complex decisions

All servers work together to provide comprehensive development support.

---

## Installation

### Quick Install (Everything)

```bash
# Install everything: Bob + Skills + LSP + MCP servers
make install-mcp-full
```

This installs:
- Bob MCP server
- Filesystem MCP server
- Workflow skills (/work, /code-review, /performance, /explore)
- Go LSP plugin (gopls)
- Additional MCP servers (GitHub, Sequential Thinking if available)

### Modular Installation

```bash
# Install only Bob and filesystem (basic)
make install-mcp

# Install only workflow skills
make install-skills

# Install only Go LSP plugin
make install-lsp

# Install only additional MCP servers
make install-mcp-servers
```

### After Installation

1. Restart Claude CLI to activate all features
2. Invoke workflow skills with slash commands: `/work`, `/code-review`, etc.
3. Skills will guide you through each workflow phase
4. Use Bob MCP tools for state persistence

---

## Skill vs MCP Tool Usage

### When to Use Skills

**Use workflow skills for:**
- Orchestrating complete workflows
- Guided development processes
- Team collaboration on structured tasks
- Enforcing best practices and review cycles

**Examples:**
- `/work "Add feature X"` - Complete development workflow
- `/code-review` - Systematic code review process
- `/performance "Optimize Y"` - Performance improvement cycle

### When to Use MCP Tools Directly

**Use Bob MCP tools for:**
- Custom workflows or automation scripts
- Programmatic access to state
- Building your own orchestration logic
- Integration with other tools

**Examples:**
- `bob.task_create()` - Create task for tracking
- `bob.workflow_register()` - Start custom workflow
- `bob.workflow_report_progress()` - Transition workflow phases

---

## Documentation

- **AGENTS.md** - Bob usage guide and workflow descriptions
- **CLAUDE.md** - Claude-specific configuration and setup
- **CODEX.md** - Codex-specific configuration
- **docs/SKILLS.md** - Detailed skill usage guide
- **docs/SUBAGENTS.md** - Subagent patterns and best practices

---

## Example: Using /work Skill

```
User: /work "Add JWT authentication to API"

Skill: Initializing work workflow...
       [Spawns Explore agent to research auth patterns]
       
       Found authentication patterns in auth/ directory.
       [Documents findings in bots/brainstorm.md]
       
       [Spawns planner agent]
       Created implementation plan in bots/plan.md
       
       [Spawns coder agent]
       Implementing JWT middleware...
       Adding token validation...
       
       [Spawns tester agent]
       Running tests... ✓ All pass
       
       [Spawns reviewer agent]
       Code review complete - no issues found
       
       Creating commit and PR...
       Monitoring CI checks... ✓ All pass
       
       Work workflow complete! PR ready to merge.
```

---
