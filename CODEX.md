# Codex Configuration for Bob

This repository uses **Belayin' Pin Bob** for workflow orchestration via MCP (Model Context Protocol).

## MCP Server Configuration

### Automatic Installation (Recommended)

The easiest way to register Bob with Codex:

```bash
cd ~/source/bob
make install-mcp
```

This will automatically:
- Build and install Bob to `~/.bob/bob`
- Register Bob with Codex (if Codex CLI is available)
- Register Bob with Claude (if Claude CLI is available)

### Manual Configuration

If the automatic installation doesn't work or you prefer manual setup:

**Option 1: Using Codex CLI**
```bash
codex mcp add bob -- ~/.bob/bob --serve
```

**Option 2: Edit ~/.codex/config.toml**
```toml
[[mcp_servers]]
name = "bob"
command = "/home/username/.bob/bob"  # Replace 'username' with your actual username
args = ["--serve"]
```

Note: Use the full path from `make install-mcp` output, or replace `/home/username` with your home directory path. Some config parsers may not expand `~`.

**Verify Installation:**
```bash
codex mcp list
# Should show "bob" in the list
```

## What Bob Provides

Bob gives Codex access to:

- **Workflows** - Multi-step orchestrated workflows (brainstorm, code-review, performance, explore)
- **Tasks** - Git-backed task management with dependencies
- **State** - Persistent JSON state files shared across all Bob sessions
- **Guidance** - Step-by-step prompts for each workflow phase

## Available MCP Tools

### Workflow Management
- `bob.workflow_list_workflows` - List all available workflows
- `bob.workflow_get_definition` - Get workflow definition by keyword
- `bob.workflow_register` - Start new workflow instance
- `bob.workflow_get_guidance` - Get current step guidance
- `bob.workflow_report_progress` - Advance to next step
- `bob.workflow_get_status` - Get workflow status
- `bob.workflow_record_issues` - Record issues found during workflow
- `bob.workflow_list_agents` - List active agents in workflow
- `bob.workflow_get_session_status` - Get session-specific status

### Task Management
- `bob.task_create` - Create new task in `.bob/issues/`
- `bob.task_get` - Get task by ID
- `bob.task_list` - List all tasks with optional filters
- `bob.task_update` - Update task properties
- `bob.task_add_dependency` - Add dependency between tasks
- `bob.task_add_comment` - Add comment to task
- `bob.task_get_ready` - Get tasks ready to work on
- `bob.task_set_workflow_state` - Set workflow state for task
- `bob.task_get_workflow_state` - Get workflow state for task
- `bob.task_delete_workflow_state_key` - Delete workflow state key

## Workflows Available

1. **brainstorm** - Full development workflow with planning
   - INIT → WORKTREE → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMMIT → MONITOR → COMPLETE
   - Loops back from REVIEW and MONITOR if issues found

2. **code-review** - Review, fix, and iterate until clean
   - INIT → REVIEW → FIX → TEST → COMMIT → MONITOR → COMPLETE
   - Loops between REVIEW, FIX, and TEST phases

3. **performance** - Benchmark, analyze, and optimize
   - INIT → BENCHMARK → ANALYZE → OPTIMIZE → VERIFY → COMMIT → MONITOR → COMPLETE
   - Loops back from ANALYZE and VERIFY if improvements needed

4. **explore** - Read-only codebase exploration
   - DISCOVER → ANALYZE → DOCUMENT → COMPLETE
   - No code modifications, pure exploration

See [AGENTS.md](AGENTS.md) for detailed workflow descriptions.

## Usage Examples

### Starting a Workflow

From Codex, you can start Bob workflows:

```
You: Start a brainstorm workflow for adding user authentication

Codex will use Bob tools:
1. bob.workflow_register(workflow="brainstorm", ...)
2. bob.workflow_get_guidance(...)
3. Follow the workflow steps
```

### Managing Tasks

```
You: Create a task for implementing OAuth2

Codex will use Bob tools:
bob.task_create(
    repoPath="/path/to/repo",
    title="Implement OAuth2 authentication",
    description="Add OAuth2 support with token refresh",
    priority="high"
)
```

### Checking Workflow Status

```
You: What's the status of my current workflow?

Codex will use:
bob.workflow_get_status(worktreePath="/path/to/worktree")
```

## Storage

Bob stores all state in `~/.bob/state/`:
- All Codex sessions (and Claude sessions) share this state
- Workflows and tasks persist across sessions
- Updates from any session appear everywhere

This means you can start a workflow in Codex and continue it in Claude, or vice versa!

## Custom Workflows

You can add custom workflows to your repo in `.bob/workflows/*.json`.
Bob will automatically discover and make them available to Codex.

See [AGENTS.md](AGENTS.md) for custom workflow format.

## Troubleshooting

### Bob not appearing in Codex

1. Verify Bob is in the MCP server list:
   ```bash
   codex mcp list
   ```

2. Check Bob path is correct:
   ```bash
   ls -la ~/.bob/bob
   ```

3. Rebuild and reinstall:
   ```bash
   cd ~/source/bob
   make build
   make install-mcp
   ```

4. Restart Codex or start a new session

### MCP server errors

1. Test Bob directly:
   ```bash
   cd ~/source/bob/cmd/bob
   ./bob --serve
   ```
   (Press Ctrl+C to stop)

2. Check Bob builds successfully:
   ```bash
   cd ~/source/bob
   make build
   ```

3. Verify Go dependencies:
   ```bash
   cd ~/source/bob/cmd/bob
   go mod download
   ```

### State issues

1. Database location: `~/.bob/state/`
2. Check permissions:
   ```bash
   ls -la ~/.bob/state/
   ```
3. Reset state (if needed):
   ```bash
   rm -rf ~/.bob/state/
   ```
   (State will be recreated automatically)

### Workflow safety check errors

Bob prevents workflows from running on main/master branches for safety. If you see:
```
workflows cannot run on main/master branch for safety
```

Create a git worktree instead:
```bash
git worktree add ../repo-worktrees/feature-name -b feature/branch-name
```

Then register the workflow using the worktree path.

## Building Bob

```bash
cd ~/source/bob
make build
```

Binary will be at: `cmd/bob/bob`

## Compatibility

Bob works with both:
- **Codex** (OpenAI's tool)
- **Claude** (Anthropic's tool)

Both can access the same workflows and task state. See [CLAUDE.md](CLAUDE.md) for Claude-specific documentation.

---

*🏴‍☠️ Belayin' Pin Bob - Captain of Your Agents*
