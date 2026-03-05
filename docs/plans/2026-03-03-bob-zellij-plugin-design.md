# Bob Zellij Plugin — Design

*Date: 2026-03-03*

## Overview

`bob` is a native Rust/WASM Zellij plugin that provides a dashboard UI for managing multiple
concurrent Claude Code agentic sessions. It is launched from the root of a git repository and
manages git worktrees, Claude processes, and agent status in a single Zellij session.

---

## Architecture

Three components:

### 1. `bob` Launcher (shell script)

A shell script installed to `~/.local/bin/bob` that:
- Validates it is run from a git repository root
- Passes the repo path to Zellij as a startup environment variable (`BOB_REPO_ROOT`)
- Launches Zellij with the `bob` layout: `zellij --layout ~/.config/zellij/layouts/bob.kdl`

### 2. `bob-plugin` (Rust/WASM Zellij plugin)

The plugin owns the left sidebar pane. It:
- Renders the rich agent list (see UI section)
- Handles keyboard and mouse input for navigation and actions
- Manages an in-memory agent registry (list of active agents + metadata)
- Polls `bob-status.json` files every 500ms to update agent status
- Spawns new agents via a floating modal
- Controls Zellij tab switching (hidden tab bar — sidebar is the nav)

### 3. Status Feed (statusline script + approval hook)

**Primary feed — statusline side-effect:**
`~/.claude/statusline-command.sh` is modified to write a side-effect file alongside its normal
output:

```
~/.claude/projects/<session-hash>/bob-status.json
{
  "cwd": "/path/to/worktree",
  "context_remaining": 72,
  "model": "claude-sonnet-4-6",
  "updated_at": 1709500000
}
```

Written on every statusline refresh (frequent). Gives the plugin: working directory, model,
context %, liveness.

**Secondary feed — approval hook:**
A minimal hook at `.claude/hooks/stop.sh` (or global `~/.claude/hooks/stop.sh`) writes approval
state when Claude is waiting for user input. Cleared when the user responds.

```
~/.claude/projects/<session-hash>/bob-approval.json
{
  "pending": true,
  "tool": "Bash",
  "preview": "git push origin add-auth"
}
```

Bob installs this hook automatically when spawning a new agent.

---

## Layout

```
┌──────────────┬────────────────────────────────────┐
│ bob sidebar  │  [active agent's Zellij tab]        │
│  (plugin)    │                                     │
│              │  $ claude --model sonnet            │
│ > add-auth   │  ...interactive session...          │
│   sonnet     │                                     │
│   editing..  │                                     │
│  ─────────   │                                     │
│   fix-bug    │                                     │
│   opus       │                                     │
│   ⚠ waiting  │                                     │
│   ─────────  │                                     │
│   [+ new]    │                                     │
└──────────────┴────────────────────────────────────┘
```

- Left sidebar: ~24 columns wide, owned by `bob-plugin`
- Right area: Zellij tabs, tab bar hidden, plugin switches tabs programmatically
- Each agent gets its own hidden Zellij tab with a full interactive `claude` terminal

---

## Sidebar — Rich Entry Format

Each agent entry (3 lines + divider):

```
> add-auth                  ← worktree name, > = selected
  sonnet • Writing auth.go  ← model • last_action (truncated)
  ⚠ git push origin add-auth ← pending approval (null = blank line)
  ─────────────────────────  ← divider
```

Status values displayed in `last_action`:
- `running` — tool name + argument preview
- `idle` — "idle"
- `waiting` — shown as ⚠ with tool + preview
- `error` — "error" in red
- `stale` — "no updates" if `updated_at` > 30s ago

---

## Spawn Flow

1. User presses `Ctrl+b n` (or clicks `[+ new]` in sidebar)
2. Plugin renders floating modal overlay:

```
╔══════════════════════════════╗
║  New Agent                   ║
║                              ║
║  Model:  [ sonnet ▼ ]        ║
║                              ║
║  Prompt:                     ║
║  [                         ] ║
║  [                         ] ║
║                              ║
║         [ Spawn ]  [ Cancel ]║
╚══════════════════════════════╝
```

3. User fills prompt, selects model (sonnet / opus / haiku), presses Enter or clicks Spawn
4. Plugin:
   a. Derives `feature-name` from first 4 words of prompt (slugified)
   b. Runs `git worktree add ../repo-worktrees/feature-name -b feature-name` via Zellij `RunCommand`
   c. Writes `.claude/hooks/stop.sh` into the worktree (approval hook)
   d. Opens a new Zellij tab titled `feature-name`
   e. Runs `claude --model <model>` in that tab
   f. Adds agent entry to sidebar registry
   g. Switches sidebar selection to the new agent

---

## Navigation

| Input | Action |
|-------|--------|
| `↑` / `↓` (sidebar focused) | Navigate agent list |
| Mouse click on agent entry | Select agent, switch to its tab |
| `Ctrl+b n` | Open spawn modal |
| `Ctrl+b k` / `Ctrl+b j` | Navigate up/down (vi-style) |
| `Ctrl+b x` | Kill selected agent, remove worktree |
| `Ctrl+b r` | Force re-read all status files |
| `Enter` (sidebar focused) | Focus the agent's terminal pane |
| `Esc` (modal open) | Close modal without spawning |

Mouse events: plugin subscribes to `Mouse` events, maps click `y` coordinate to agent entry index.
Arrow keys active when plugin pane is focused.

---

## Project Structure

This is a new crate inside the `bob` repository:

```
cmd/bob-plugin/          ← Rust crate (Zellij plugin)
  Cargo.toml
  src/
    main.rs              ← ZellijPlugin impl, event loop
    state.rs             ← AgentRegistry, AgentEntry, StatusFile
    ui.rs                ← Render logic (sidebar, modal)
    spawn.rs             ← Worktree creation, claude launch
    poll.rs              ← Status file polling (500ms timer)

scripts/
  bob                    ← launcher shell script
  statusline-command.sh  ← modified to write bob-status.json side-effect

config/
  bob.kdl               ← Zellij layout definition
```

The Rust crate compiles to WASM (`wasm32-wasi` target). The `bob` launcher script is installed
to `~/.local/bin/bob` via `make install-bob-plugin`.

---

## State File Locations

Given a Claude session running in worktree `/path/to/repo-worktrees/add-auth`:

- Claude session hash: derived from `cwd` → `~/.claude/projects/<hash>/`
- Status file: `~/.claude/projects/<hash>/bob-status.json`
- Approval file: `~/.claude/projects/<hash>/bob-approval.json`

Plugin maps worktree path → session hash by scanning `~/.claude/projects/*/bob-status.json`
and matching `cwd` field.

---

## Out of Scope (v1)

- Multiple Zellij sessions (one `bob` session per repo)
- Persistent agent history across bob restarts
- Agent-to-agent communication
- Custom per-agent keybindings
