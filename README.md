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

Bob is a workflow orchestration system for AI agents. Just like a ship's captain uses a belayin' pin to secure the ship's lines and rigging, Bob keeps your AI agent workflows organized, coordinated, and running smoothly.

## Features

- 🎯 **Workflow Orchestration** - Define and run complex multi-step workflows
- 🔄 **Loop Management** - Smart loop-back rules for iterative workflows
- 📊 **Task Tracking** - Manage tasks with dependencies and state
- 🔌 **MCP Server** - Claude Model Context Protocol integration (stdio mode)
- 💾 **Persistent State** - SQLite database for workflow and task state
- 📝 **Git Integration** - Task tracking with git branch management
- 🏴‍☠️ **Captain of Your Agents** - Keep your AI workflows in line!

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/mattdurham/bob.git
cd bob

# Build Bob
make build-all

# Or just build the backend
make build-backend
```

### Running Bob

```bash
# Run Bob as an MCP server (stdio mode for Claude)
cd cmd/bob
./bob --serve

# Or use make
make run
```

### Development

```bash
# Install dependencies
make install-deps

# Run tests
make test

# Clean build artifacts
make clean
```

## Workflows

Bob comes with several built-in workflows:

- **brainstorm** - Full development workflow with planning and iteration
- **code-review** - Review, fix, test, and iterate until clean
- **performance** - Benchmark, analyze, optimize, and verify
- **explore** - Read-only codebase exploration

### Custom Workflows

Create custom workflows in `.bob/workflows/*.json`:

```json
{
  "keyword": "my-workflow",
  "name": "My Custom Workflow",
  "description": "STEP1 → STEP2 → STEP3",
  "steps": [
    {
      "name": "STEP1",
      "description": "First step"
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

## Task Management

Bob manages tasks in `.bob/issues/` with git branch integration:

```bash
# Tasks are stored in git on the 'bob' branch
# Each task is a JSON file: .bob/issues/<id>.json
```

Task properties:
- **id** - Unique identifier
- **title** - Task title
- **description** - Detailed description
- **state** - Task state (pending, in_progress, completed)
- **priority** - Priority level
- **blocks** - Task IDs this blocks
- **blockedBy** - Task IDs blocking this

## MCP Integration

Bob implements the Model Context Protocol for Claude integration:

```json
{
  "mcpServers": {
    "bob": {
      "command": "/path/to/bob/cmd/bob/bob",
      "args": []
    }
  }
}
```

## Configuration

Bob stores state in `~/.bob/state/`:
- `db.sql` - SQLite database with workflow and task state
- Updates from all bob MCP servers appear here

## Architecture

```
bob/
├── cmd/
│   └── bob/                    # Main Bob application
│       ├── main.go             # Entry point
│       ├── mcp_server.go       # MCP protocol implementation
│       ├── task_manager.go     # Task management & git integration
│       ├── workflow_definition.go  # Workflow definitions
│       ├── state_manager.go    # State management
│       ├── database.go         # SQLite database layer
│       ├── guidance.go         # Claude guidance prompts
│       ├── workflows/          # Built-in workflow definitions
│       ├── prompts/            # Prompt templates
│       └── templates/          # Guidance templates
├── Makefile
└── README.md
```

## Development Principles

Bob follows these core principles:

1. **Workflows are loops** - Most workflows need iteration
2. **Review before fix** - MONITOR → REVIEW → FIX (not MONITOR → FIX)
3. **State is persistent** - All workflow state saved to SQLite
4. **Git-based tasks** - Tasks stored in git for durability
5. **MCP-first** - Built for Claude integration

## Contributing

Bob is your ship's captain - if you've got improvements to the riggin', send a pull request!

## License

MIT License - See LICENSE file

---

*🏴‍☠️ Fair winds and following seas! - Captain Bob*
