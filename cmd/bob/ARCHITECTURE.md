# BPB Architecture - SQLite-Based Shared State

## Overview

Separate executables with shared SQLite database for true multi-session state sharing.

## Architecture

```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│  Claude 1   │      │  Claude 2   │      │  Claude 3   │
│  bob --serve│      │  bob --serve│      │  bob --serve│
└──────┬──────┘      └──────┬──────┘      └──────┬──────┘
       │ writes           │ writes           │ writes
       └────────────────────┼────────────────────┘
                            ▼
                 ~/.bob/state/db.sql
                  (Shared SQLite DB)
                            ▲
                            │ reads
                     ┌──────┴──────┐
                     │   bob-web   │
                     │  HTTP:9090  │
                     └─────────────┘
```

## Components

### 1. `bob` - MCP Server (Multiple Instances)
- **Purpose**: MCP server for Claude integration
- **Mode**: `bob --serve` (stdio)
- **Database**: Writes workflow and task state
- **Port**: None (stdio communication)
- **Instances**: One per Claude session

### 2. `bob-web` - Web UI (Single Instance)
- **Purpose**: HTML dashboard for monitoring
- **Mode**: `bob-web` (HTTP server)
- **Database**: Reads workflow and task state
- **Port**: 9090 (configurable with `--port`)
- **Instances**: One shared instance

### 3. `~/.bob/state/db.sql` - Shared Database
- **Format**: SQLite3
- **Location**: User's home directory
- **Access**: Read/write by all bob instances, read by bob-web
- **Schema**:
  - `workflows` - Active workflow sessions
  - `workflow_progress` - Progress history per workflow
  - `tasks` - All tasks across repositories
  - `task_comments` - Comments on tasks

## Benefits

✅ **True Shared State** - All Claude sessions see the same data
✅ **No Port Conflicts** - MCP uses stdio, web UI is separate
✅ **Persistent** - Database survives restarts
✅ **Scalable** - SQLite handles concurrent reads/writes
✅ **Simple** - Standard SQL, no complex setup

## Database Schema

### workflows
```sql
CREATE TABLE workflows (
    id TEXT PRIMARY KEY,              -- Workflow instance ID
    workflow TEXT NOT NULL,            -- Type (brainstorm, code-review, etc)
    current_step TEXT NOT NULL,        -- Current step in workflow
    task_description TEXT,             -- User's task description
    loop_count INTEGER DEFAULT 0,      -- Number of loops
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### workflow_progress
```sql
CREATE TABLE workflow_progress (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    workflow_id TEXT NOT NULL,         -- Links to workflows(id)
    step TEXT NOT NULL,                -- Step name
    metadata TEXT,                     -- JSON metadata
    timestamp TIMESTAMP
);
```

### tasks
```sql
CREATE TABLE tasks (
    id TEXT PRIMARY KEY,               -- Task ID (task-001, etc)
    repo_path TEXT NOT NULL,           -- Repository path
    title TEXT NOT NULL,
    description TEXT,
    type TEXT,                         -- feature, bug, chore, etc
    priority TEXT,                     -- high, medium, low
    state TEXT,                        -- pending, in_progress, completed, blocked
    assignee TEXT,
    tags TEXT,                         -- JSON array
    blocks TEXT,                       -- JSON array of task IDs
    blocked_by TEXT,                   -- JSON array of task IDs
    metadata TEXT,                     -- JSON object
    workflow_state TEXT,               -- JSON object
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    completed_at TIMESTAMP
);
```

### task_comments
```sql
CREATE TABLE task_comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL,             -- Links to tasks(id)
    author TEXT,
    text TEXT NOT NULL,
    timestamp TIMESTAMP
);
```

## Usage

### Start MCP Server (for Claude)
```bash
# Automatically started by Claude via ~/.mcp.json
bob --serve

# Writes to: ~/.bob/state/db.sql
```

### Start Web UI
```bash
# In a separate terminal
bob-web

# Access at: http://localhost:9090
# Reads from: ~/.bob/state/db.sql
```

### Claude MCP Config (~/.mcp.json)
```json
{
  "mcpServers": {
    "bob": {
      "command": "/path/to/bob",
      "args": ["--serve"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      },
      "autoStart": false
    }
  }
}
```

## Implementation Status

### ✅ Completed
- Database schema design
- Database operations (Save/Get workflows and tasks)
- Separate bob-web executable
- Web UI templates (HTML/CSS/JS)

### 🚧 In Progress
- StateManager database integration
- TaskManager database integration
- MCP server database writes

### 📋 TODO
- Complete database write integration in StateManager
- Complete database write integration in TaskManager
- Test multi-session workflows
- Migration from JSON files to SQLite (optional)

## Migration Path

### Phase 1: Dual Write (Current + SQLite)
- Keep existing JSON file writes
- Add SQLite writes
- Verify data consistency

### Phase 2: SQLite Primary
- Read from SQLite
- Keep JSON as backup
- Test extensively

### Phase 3: SQLite Only
- Remove JSON file operations
- SQLite as single source of truth

## Development

### Build Both Executables
```bash
# Build MCP server
cd cmd/bob
go build -o bob

# Build web UI
cd cmd/bob-web
go build -o bob-web
```

### Test Database
```bash
# Check database location
ls -la ~/.bob/state/db.sql

# Query database directly
sqlite3 ~/.bob/state/db.sql "SELECT * FROM workflows;"
sqlite3 ~/.bob/state/db.sql "SELECT * FROM tasks;"
```

## Advantages Over Previous Design

| Aspect | Old (JSON Files) | New (SQLite) |
|--------|------------------|--------------|
| **State Sharing** | File-based, eventual consistency | Database, immediate consistency |
| **Queries** | Scan all files | SQL queries |
| **Concurrency** | File locking issues | SQLite handles it |
| **Web UI** | Embedded in MCP server | Separate process |
| **Port Conflicts** | Yes (multiple web servers) | No (one web server) |
| **Scalability** | Poor (O(n) scans) | Good (indexed queries) |

## Future Enhancements

- [ ] Add database backups
- [ ] Add database migrations system
- [ ] Add real-time updates via WebSockets
- [ ] Add query filters and search
- [ ] Add historical analytics
- [ ] Add export to JSON/CSV
