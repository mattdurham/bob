# BPB Web UI

Human-readable HTML interface for monitoring workflows and tasks.

## Starting the Web UI

**By default, bob starts the web UI on port 9090:**

```bash
./bob
```

Then open http://localhost:9090 in your browser.

**To use a different port:**

```bash
./bob --port 8080
```

**To run as MCP server instead (for Claude integration):**

```bash
./bob --serve
```

## Features

### 1. Dashboard (`/`)
- **Task Statistics**: See total, pending, in-progress, and completed tasks
- **Active Workflows**: View all running workflows with their current step
- **Quick Links**: Navigate to workflows and tasks

### 2. Workflow Diagrams (`/workflow/<name>`)
- **Interactive Mermaid Diagrams**: Visual representation of workflow steps
- **Clickable Steps**: Click any step to see its description
- **Active Step Highlighting**: Currently active steps are highlighted in green
- **Loop Visualization**: Dotted lines show loop-back paths
- **Active Sessions**: See which workflows are currently running

Available workflows:
- http://localhost:9090/workflow/brainstorm
- http://localhost:9090/workflow/code-review
- http://localhost:9090/workflow/performance
- http://localhost:9090/workflow/explore

### 3. Task List (`/tasks`)
- **All Tasks**: View all tasks across all repositories
- **Filtering**: Filter by state (pending, in_progress, blocked, completed)
- **Priority Sorting**: Filter by priority (high, medium, low)
- **Dependency Tracking**: See which tasks are blocked

### 4. Task Details (`/task/<repo>/<id>`)
- **Complete Information**: Title, description, state, priority, type
- **Dependencies**: View blockedBy and blocks relationships
- **Comments**: Read all task comments with timestamps
- **Workflow State**: See stored workflow state/metadata
- **Tags**: View all task tags

## Use Cases

### Monitoring Active Workflows
1. Go to `/` to see all active workflows
2. Click on a workflow name to see its diagram
3. Active steps are highlighted in green
4. Click on steps to see their prompts

### Tracking Task Progress
1. Go to `/tasks` to see all tasks
2. Use filters to find specific tasks
3. Click on a task ID to see full details
4. Check dependencies to understand blockers

### Debugging Workflow Issues
1. Open workflow diagram to see current step
2. Click on the step to read its prompt
3. Check if the workflow is looping (see loop count)
4. Review progress history in active sessions

### Understanding Workflow Flow
1. Open a workflow diagram
2. Follow the arrows from step to step
3. Dotted lines show where loops can occur
4. Click steps to understand what each does

## API Endpoints

For programmatic access:
- `GET /api/workflows` - JSON list of active workflows
- `GET /api/tasks` - JSON list of all tasks

## Tips

- **Refresh**: The UI doesn't auto-refresh - reload the page to see updates
- **Bookmarks**: Bookmark specific workflow diagrams you use frequently
- **Multiple Tabs**: Open workflows and tasks in separate tabs for easy switching
- **Keyboard**: Press `Escape` to close step detail modals

## Example: Monitoring a Code Review

1. Start web UI: `./bob --web localhost:9090`
2. Register a code review workflow via MCP
3. Go to http://localhost:9090/workflow/code-review
4. Watch the green highlight move through steps as the workflow progresses
5. Click on the current step to see what the agent should be doing

## Architecture

- **Templates**: HTML templates in `templates/`
- **Embedded**: Templates are embedded in the binary
- **Mermaid.js**: Diagram rendering via CDN
- **No Database**: Reads directly from state files
- **Read-Only**: Web UI only displays data, doesn't modify it
