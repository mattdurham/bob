# Claude Configuration for Bob

This repository uses **Belayin' Pin Bob** for workflow orchestration via MCP (Model Context Protocol).

## MCP Server Configuration

Bob provides two MCP servers that work together:
1. **Bob** - Workflow orchestration, task management, and workflow guidance
2. **Filesystem** - Secure filesystem operations (read/write files, search, etc.)

### Complete Configuration

After running `make install-mcp`, your MCP configuration will include both servers:

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

Both servers run independently and can be used simultaneously in your Claude sessions.

## What Bob Provides

Bob gives Claude access to:

- **Workflows** - Multi-step orchestrated workflows (work, code-review, performance, explore)
- **Tasks** - Git-backed task management with dependencies
- **State** - Persistent JSON state files shared across all Claude sessions
- **Guidance** - Step-by-step prompts for each workflow phase
- **Filesystem** - Secure file operations (read, write, search) in allowed directories

## Platform Compatibility

Bob works with both **Claude** and **Codex**. The `make install-mcp` command automatically registers Bob with both platforms (if their CLIs are available).

- **For Claude users**: This document contains Claude-specific configuration and usage
- **For Codex users**: See [CODEX.md](CODEX.md) for Codex-specific documentation
- **Shared state**: Workflows and tasks are shared across both platforms, so you can start work in Claude and continue in Codex, or vice versa

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

### Filesystem Operations
- `filesystem.read_file` - Read file contents
- `filesystem.write_file` - Write or create files
- `filesystem.list_directory` - List directory contents
- `filesystem.create_directory` - Create new directory
- `filesystem.search_files` - Search files by name pattern
- `filesystem.search_within_files` - Search file contents
- `filesystem.get_file_info` - Get file metadata
- `filesystem.copy_file` - Copy files
- `filesystem.move_file` - Move/rename files
- `filesystem.delete_file` - Delete files
- `filesystem.tree` - Get directory tree structure
- `filesystem.read_multiple_files` - Read multiple files at once

**Allowed Directories**: `$HOME/source`, `/tmp`

**Security**: Filesystem server only allows access to explicitly allowed directories. Directory traversal attempts are blocked.

## Workflows Available

1. **work** - Full development workflow with planning
2. **code-review** - Review, fix, and iterate until clean
3. **performance** - Benchmark, analyze, and optimize
4. **explore** - Read-only codebase exploration

See AGENTS.md for detailed workflow descriptions.

## Storage

Bob stores all state in `~/.bob/state/`:
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
1. Database location: `~/.bob/state/`
2. Check permissions: `ls -la ~/.bob/state/`
3. Reset state: `rm -rf ~/.bob/state/` (will recreate)

## Building Bob

```bash
cd ~/source/bob
make build
```

Binary will be at: `cmd/bob/bob`

---

*🏴‍☠️ Belayin' Pin Bob - Captain of Your Agents*
