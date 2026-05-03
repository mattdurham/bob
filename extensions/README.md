# Bob Pi Extensions

Extensions installed to `~/.pi/agent/extensions/` via `make install-pi`.

---

## bob-agents

**File:** `bob-agents/index.ts`  
**Purpose:** Core multi-agent orchestration for Bob workflows in pi.

The pi SDK's built-in subagent support runs each agent as a separate process. Bob requires in-process agents so they can share memory — specifically the message bus and task board needed for reviewer↔coder communication.

### What it provides

**Subagent spawning** (`subagent` tool)  
Spawn agents in-process using `agents/*/SKILL.md` definitions. Supports single, parallel, and chain modes. All agents run in background by default so the orchestrator stays responsive.

**Teams** (`TeamCreate`, `TeamStatus`, `TeamDisband`)  
Named groups of agents with isolated state. Each team has its own message bus, task board, and agent registry. A subagent can create a sub-team and become its team lead, enabling nested hierarchies. The root team (`__root__`) represents the main pi session.

**Message bus** (`mailbox_send`, `mailbox_receive`, `mailbox_broadcast`, `mailbox_send_as`, `mailbox_read`)  
In-process pub/sub so agents can talk to each other and to the orchestrator without going through files. Each team has an isolated bus — the team lead is "orchestrator" within that bus.

**Task board** (`TaskCreate`, `TaskList`, `TaskGet`, `TaskUpdate`)  
Shared per-team work queue. Agents claim tasks, mark them done, and the orchestrator routes based on status. Replaces Claude Code's native Task tool so Bob's existing agent prompts work unchanged.

**Agent observability** (`agent_status`, `agent_output`, `agent_wait`)

- `agent_status` — live status of all agents with last 3 lines of stdout
- `agent_output` — full 8KB stdout buffer + lifecycle log ring for a specific agent
- `agent_wait` — poll until named agents finish (without blocking the orchestrator)

**Context injection**  
On each agent spawn, injects: agent identity, team name, team lead, and the contents of `~/.pi/agent/AGENTS.md` / `CLAUDE.md` so user preferences (conciseness, tooling guidelines) apply to subagents too.

**AGENTS.md / CLAUDE.md loading**  
The `before_agent_start` hook loads context files from `~/.pi/agent/` and the project root into the main session's system prompt on each turn.

### Key design decisions

- **In-process over subprocess**: pi-subagents spawns real `pi` processes; bob-agents uses `createAgentSession()` so agents share a single Node.js process and memory. Required for the message bus.
- **`extensionFactories` for custom tools**: Custom tools (mailbox, task board) are injected via `extensionFactories` in `DefaultResourceLoader`, not via the `customTools` param. This ensures they're activated correctly — passing `customTools` directly caused a tool activation bug due to how `initialActiveToolNames` is computed.
- **No `tools: string[]`**: Passing string tool names to `createAgentSession` causes all names to map to `undefined` (strings have no `.name`), leaving only default tools active.
- **`new ModelRegistry(authStorage)` not `ModelRegistry.create()`**: `create()` doesn't exist; using it caused the extension to fail silently on every child session spawn.
- **`DefaultResourceLoader` with `noExtensions: true`**: Prevents the parent bob-agents extension from loading inside child sessions, which would fire `session_shutdown` and reset shared singletons (bus, registry, task board).
- **No singleton resets in `session_shutdown`**: Bus, registry, and task board are process-lifetime objects. Resetting them when a child session ends would wipe the parent orchestrator's state.

---

## otel

**File:** `otel.ts`  
**Purpose:** Export pi session traces to any OTLP-compatible backend (Grafana Cloud, Jaeger, etc.).

Enabled only when `OTEL_ENDPOINT`, `OTEL_USER`, and `OTEL_TOKEN` env vars are set. Silent otherwise.

### What it records

- **Session span** (root): cwd, last-used model
- **Agent-loop spans** (child): one per user prompt — model, token counts (input/output/cache), cost, elapsed time
- **Events on each span**: `user.message`, `assistant.message` (truncated to 2000 chars), `tool.call` (name + safe params)

### What it does NOT record

- File contents (read/write/edit tool calls skipped entirely)
- Bash command strings
- File paths

### Service name

`bob` — configure via `OTEL_ENDPOINT`/`OTEL_USER`/`OTEL_TOKEN`.

---

## bash-compact

**File:** `bash-compact.ts`  
**Purpose:** Reduce bash tool noise — show only the command and a ✓/✗ indicator.

The default pi bash renderer shows the full command output inline. For successful commands this is usually noise. This extension overrides the bash tool renderer:

- **Success (exit 0)**: shows `✓` only — no output
- **Error (non-zero exit)**: shows `✗ exit N` + full output
- **Expanded (Ctrl+E)**: always shows full output

The agent still receives the full output in its context — only the UI display changes.

Uses `createBashTool(cwd)` and delegates `execute()` to the original, so behavior is identical.

---

## quiet-thoughts

**File:** `quiet-thoughts.ts`  
**Purpose:** Suppress assistant preamble text ("Let me look at X...") on turns that contain tool calls.

When a model writes reasoning text before calling tools, it's visible in the pi UI but adds no value to the user. This extension intercepts `message_end` events and strips text content from assistant messages that also contain tool calls.

- **Turn with tool calls**: text stripped, tool calls preserved
- **Turn with no tool calls** (final answer): shown in full, unmodified

The model's tool calls remain in context so it can reference results in subsequent turns. Only the visible preamble is removed.

> **Note**: An alternative is `/thinking high` (extended thinking mode), which moves all reasoning to hidden thinking blocks. That works at the model level; this extension works at the display level and applies regardless of thinking mode.

---

## Installation

All extensions are installed by `make install-pi`:

```bash
make install-pi
```

To add a new extension: create the `.ts` file in `extensions/`, add a copy step to the `install-pi` target in the Makefile, and document it here.
