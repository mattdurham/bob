# Shipmate Span Attributes

All spans emitted by shipmate carry these attributes. Fields are omitted when empty.

## Identity

| Attribute | Source | Description |
|-----------|--------|-------------|
| `session.id` | All hooks | Claude Code session UUID — links all spans for a session |
| `agent_id` | Subagent events | Unique ID of the subagent instance |
| `agent_type` | Subagent events | Agent type, e.g. `team-coder`, `workflow-planner` |
| `task.teammate` | TaskCreated / TaskCompleted | Name of the teammate that owns the task |
| `task.team` | TaskCreated / TaskCompleted | Team name |

## Session Context

| Attribute | Source | Description |
|-----------|--------|-------------|
| `hook.event` | All hooks | Event name: `PostToolUse`, `SubagentStart`, `TaskCompleted`, etc. |
| `cwd` | All hooks | Working directory when the hook fired |
| `permission_mode` | All hooks | Claude Code permission mode: `default`, `dontAsk`, `auto`, etc. |
| `transcript_path` | All hooks | Path to the session transcript file on disk |
| `session.source` | SessionStart | How the session started: `startup`, `resume`, `clear`, `compact` |
| `session.model` | SessionStart | Model name used for this session |

## User Prompt

| Attribute | Source | Description |
|-----------|--------|-------------|
| `prompt.text` | UserPromptSubmit | The actual text the user typed (capped at 512 chars) |

## Tool Execution

| Attribute | Source | Description |
|-----------|--------|-------------|
| `tool_use_id` | PreToolUse / PostToolUse | Unique ID per tool call — correlates pre/post pairs |
| `tool.success` | PostToolUse | `"true"` or `"false"` — whether the tool succeeded |
| `tool.error` | PostToolUse | Error message when tool fails |
| `tool.file` | Edit / Write / Read | File path the tool operated on |
| `tool.command` | Bash | Shell command that was executed |
| `tool.description` | Bash | Human-readable description of the command |
| `tool.run_in_background` | Bash | `"true"` if the command ran in background |
| `tool.old_string` | Edit | Text that was replaced (capped at 256 chars) |
| `tool.new_string` | Edit | Replacement text (capped at 256 chars) |
| `tool.content` | Write | File content written (capped at 256 chars) |
| `tool.pattern` | Grep | Search pattern |
| `tool.path` | Grep | Directory searched |
| `tool.glob` | Grep | File glob filter |
| `tool.query` | WebSearch | Search query |
| `tool.url` | WebFetch | URL fetched |
| `tool.response_file` | PostToolUse (Write) | File path from Write tool response |

## Subagent Lifecycle

| Attribute | Source | Description |
|-----------|--------|-------------|
| `agent.last_message` | SubagentStop | Final assistant message from the subagent (capped at 512 chars) |
| `agent.transcript_path` | SubagentStop | Path to the subagent's transcript file |

## Task Lifecycle

| Attribute | Source | Description |
|-----------|--------|-------------|
| `task.id` | TaskCreated / TaskCompleted | Task identifier |
| `task.subject` | TaskCreated / TaskCompleted | Short task title |
| `task.description` | TaskCreated / TaskCompleted | Full task description (capped at 256 chars) |

## Memory Annotations

Set when Claude calls `shipmate memory "text"` via the Bash tool.

| Attribute | Source | Description |
|-----------|--------|-------------|
| `memory.text` | memory command | Free-text annotation Claude added to the trace |
| `memory.source` | memory command | Always `"memory"` — identifies annotation spans |

## Quality Scoring

Set by the scorer on session stop after calling `claude -p` for LLM-as-judge evaluation.

| Attribute | Source | Description |
|-----------|--------|-------------|
| `score` | Scorer | Quality score 0.0–1.0 |
| `score.comments` | Scorer | LLM commentary on quality |
| `score.span_ref` | Scorer | Span ID of the original span being scored |

## Service Identity

Set on every span by the OTEL SDK resource.

| Attribute | Source | Description |
|-----------|--------|-------------|
| `service.name` | OTEL resource | Always `"shipmate"` |
