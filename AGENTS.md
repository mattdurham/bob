# 🏴‍☠️ Belayin' Pin Bob - Agent Guidance

This repository uses **Belayin' Pin Bob** (bob) for workflow orchestration and task management.

Bob is your ship's captain - he keeps your AI agent workflows organized, coordinated, and running smoothly through the Model Context Protocol (MCP).

## What Bob Provides

Bob is an MCP server that gives Claude access to:

1. **Workflow Orchestration** - Multi-step workflows with loop-back rules
2. **Task Management** - Git-backed task tracking with dependencies
3. **State Persistence** - Shared JSON state across all sessions
4. **Workflow Guidance** - Step-by-step prompts for each workflow phase

## Available Workflows

### work
Full development workflow with planning and iteration:
```
INIT → WORKTREE → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMMIT → MONITOR → COMPLETE
```

**Loop rules:**
- `REVIEW → PLAN` (issues found require replanning)
- `MONITOR → REVIEW` (always review before fixing)
- `TEST → EXECUTE` (test failures require fixes)

### code-review
Review, fix, and iterate until clean:
```
INIT → REVIEW → FIX → TEST → COMMIT → MONITOR → COMPLETE
```

**Loop rules:**
- `REVIEW → FIX` (issues found)
- `TEST → REVIEW` (re-verify after fixes)
- `MONITOR → REVIEW` (CI failures or feedback)

### performance
Benchmark, analyze, optimize, and verify:
```
INIT → BENCHMARK → ANALYZE → OPTIMIZE → VERIFY → COMMIT → MONITOR → COMPLETE
```

**Loop rules:**
- `VERIFY → ANALYZE` (targets not met)
- `MONITOR → ANALYZE` (CI failures)

### explore
Read-only codebase exploration:
```
DISCOVER → ANALYZE → DOCUMENT → COMPLETE
```

No loops, no file changes, no worktree needed.

## Using Bob

### Start a Workflow

Bob will guide you through workflows using these MCP tools:

```typescript
// List available workflows
bob.workflow_list()

// Get workflow definition
bob.workflow_get({ keyword: "work" })

// Create workflow instance
bob.workflow_create({
  workflowKeyword: "work",
  repoPath: "/path/to/repo",
  taskDescription: "Add user authentication"
})

// Progress through steps
bob.workflow_progress({
  instanceId: "...",
  toStep: "PLAN"
})
```

### Task Management

```typescript
// Create task
bob.task_create({
  repoPath: "/path/to/repo",
  title: "Fix authentication bug",
  description: "...",
  priority: "high"
})

// List tasks
bob.task_list({
  repoPath: "/path/to/repo",
  state: "pending"
})

// Update task
bob.task_update({
  repoPath: "/path/to/repo",
  taskId: "...",
  updates: { state: "in_progress" }
})
```

## Bob Storage

Bob stores state in `~/.bob/`:
- `~/.bob/state/` - JSON state with all workflows and tasks
- All Claude sessions share this state
- Updates from any session appear everywhere

## Custom Workflows

Create custom workflows in `.bob/workflows/*.json` in your repo:

```json
{
  "keyword": "my-workflow",
  "name": "My Custom Workflow",
  "description": "STEP1 → STEP2 → STEP3",
  "steps": [
    {
      "name": "STEP1",
      "description": "First step",
      "requirements": ["git_repo"]
    }
  ],
  "loopRules": [
    {
      "fromStep": "STEP2",
      "toStep": "STEP1",
      "condition": "retry_needed",
      "description": "Retry if needed"
    }
  ]
}
```

Bob will automatically discover custom workflows in `.bob/workflows/`.

## Planning Documents

All planning documents, brainstorming notes, and workflow artifacts should be stored in the `bots/` folder at the root of your repository. This folder is ignored by git and provides a clean workspace for agent-generated planning materials.

```
your-repo/
├── bots/              # All planning docs go here (git ignored)
│   ├── plans/
│   ├── notes/
│   └── research/
├── .bob/              # Bob configuration and custom workflows
└── src/               # Your source code
```

## Workflow Principles

1. **Workflows are loops** - Most work needs iteration
2. **Review before fix** - MONITOR → REVIEW → FIX (not MONITOR → FIX)
3. **State persists** - Resume workflows across sessions
4. **Git-based tasks** - Tasks stored in git on `bob` branch
5. **Guidance-driven** - Bob provides step-by-step prompts

## Task Files

Tasks are stored as JSON in `.bob/issues/<id>.json` on the `bob` git branch:

```json
{
  "id": "task-001",
  "title": "Add authentication",
  "description": "Implement JWT authentication",
  "type": "feature",
  "priority": "high",
  "state": "in_progress",
  "assignee": "claude",
  "blocks": [],
  "blockedBy": [],
  "tags": ["auth", "security"],
  "metadata": {},
  "createdAt": "2026-02-09T12:00:00Z",
  "updatedAt": "2026-02-09T12:30:00Z"
}
```

## MCP Configuration

Bob is configured in CLAUDE.md - see that file for MCP server setup.

---

*🏴‍☠️ Fair winds and following seas! - Captain Bob*
