# Claude Configuration for Bob

This repository uses **Belayin' Pin Bob** for workflow orchestration via MCP (Model Context Protocol).

## MCP Server Configuration

Add Bob to your MCP configuration:

```json
{
  "mcpServers": {
    "bob": {
      "command": "/home/matt/source/bob/cmd/bob/bob",
      "args": ["--serve"],
      "env": {}
    }
  }
}
```

**Or if Bob is in your PATH:**

```json
{
  "mcpServers": {
    "bob": {
      "command": "bob",
      "args": ["--serve"]
    }
  }
}
```

## What Bob Provides

Bob gives Claude access to:

- **Workflows** - Multi-step orchestrated workflows (brainstorm, code-review, performance, explore)
- **Tasks** - Git-backed task management with dependencies
- **State** - Persistent JSON state files shared across all Claude sessions
- **Guidance** - Step-by-step prompts for each workflow phase

## Available MCP Tools

### Workflow Management
- `bob.workflow_list` - List all available workflows
- `bob.workflow_get` - Get workflow definition by keyword
- `bob.workflow_create` - Start new workflow instance
- `bob.workflow_progress` - Advance workflow to next step
- `bob.workflow_list_running` - List active workflow instances

### Task Management
- `bob.task_create` - Create new task in `.bob/issues/`
- `bob.task_get` - Get task by ID
- `bob.task_list` - List all tasks with optional filters
- `bob.task_update` - Update task properties

## Workflows Available

1. **brainstorm** - Full development workflow with planning
2. **code-review** - Review, fix, and iterate until clean
3. **performance** - Benchmark, analyze, and optimize
4. **explore** - Read-only codebase exploration

See AGENTS.md for detailed workflow descriptions.

## Storage

Bob stores all state in `~/.bob/~/.bob/state/`:
- All Claude sessions share this state
- Workflows and tasks persist across sessions
- Updates from any Claude session appear everywhere

## Custom Workflows

You can add custom workflows to this repo in `.bob/workflows/*.json`.
Bob will automatically discover and make them available.

See AGENTS.md for custom workflow format.

## Troubleshooting

### Bob not appearing
1. Verify Bob path is correct in MCP configuration
2. Check Bob builds: `cd ~/source/bob && make build`
3. Restart your MCP client

### MCP server errors
1. Test Bob directly: `cd ~/source/bob/cmd/bob && ./bob --serve`
2. Check MCP client logs
3. Verify Go dependencies: `cd ~/source/bob/cmd/bob && go mod download`

### State issues
1. Database location: `~/.bob/~/.bob/state/`
2. Check permissions: `ls -la ~/.bob/state/`
3. Reset state: `rm ~/.bob/~/.bob/state/` (will recreate)

## Building Bob

```bash
cd ~/source/bob
make build
```

Binary will be at: `cmd/bob/bob`

---

*🏴‍☠️ Belayin' Pin Bob - Captain of Your Agents*
