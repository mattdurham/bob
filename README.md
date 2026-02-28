# Belayin' Pin Bob

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

Workflow orchestration for Claude Code through skills and subagents.

## What is Bob?

Bob coordinates AI agent workflows for feature development. No MCP servers, no daemons — skills invoke specialized subagents, pass state through `.bob/` artifacts, and enforce quality gates automatically.

## Spec-Driven Development

Bob treats **SPECS.md as the source of truth** for module behavior. Every workflow is spec-aware:

- **`/bob:design`** creates or applies spec-driven module structure. Call this first when starting a new module, or before major design changes. It scaffolds SPECS.md, NOTES.md, TESTS.md, BENCHMARKS.md, and adds the NOTE invariant to `.go` files.

- **`/bob:work`**, **`/bob:work-agents`**, and **`/bob:work-teams`** all read existing specs before making changes. If a request contradicts a contract or invariant in SPECS.md, the workflow will question it — specs can be changed, but only deliberately. Code changes to spec-driven modules must be reflected in the corresponding spec docs.

- **`/bob:explore`** prioritizes spec docs when analyzing a codebase. For spec-driven modules, it reads SPECS.md and NOTES.md first to understand contracts and design decisions before diving into implementation code.

A spec-driven module is any directory containing SPECS.md, NOTES.md, TESTS.md, BENCHMARKS.md, or `.go` files with this comment:

```go
// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.
```

## Quick Start

```bash
git clone https://github.com/mattdurham/bob.git
cd bob
make install
```

This installs workflow skills to `~/.claude/skills/` and subagents to `~/.claude/agents/`. Restart Claude Code after installation.

## Workflows

### `/bob:design` — Spec Scaffolding

```
INIT → GATHER → [ANALYZE] → SCAFFOLD → COMPLETE
```

Two modes:
- **New module** — describe what you're building, get SPECS.md, NOTES.md, TESTS.md, BENCHMARKS.md, and stub `.go` files
- **Apply to existing** — point at a directory, get spec docs generated from the existing implementation

### `/bob:work` — Simple Direct Workflow

```
INIT → WORKTREE → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMMIT → COMPLETE
```

You do all the work yourself. No subagents, no orchestration. Linear flow, local commit only.

### `/bob:work-agents` — Sequential Subagent Workflow

```
INIT → WORKTREE → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMMIT → MONITOR → COMPLETE
                      ↑                                    ↓               ↓
                      └────────────────────────────────────┴───────────────┘
```

Full orchestration with specialized subagents for each phase. You coordinate; agents do the work. Autonomous progression — only prompts at the final merge.

### `/bob:work-teams` — Concurrent Agent Team Workflow

```
INIT → WORKTREE → BRAINSTORM → PLAN → SPAWN TEAM → EXECUTE ↔ REVIEW → COMMIT → MONITOR → COMPLETE
```

Multiple coder and reviewer teammates work in parallel through a shared task list. Requires `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`.

### `/bob:explore` — Read-Only Exploration

```
INIT → DISCOVER → ANALYZE → DOCUMENT → COMPLETE
```

Spec-aware codebase exploration. No code changes.

## Loop-Back Rules

All work workflows enforce these routing rules:

| Trigger | Route to | Reason |
|---------|----------|--------|
| CRITICAL/HIGH review issues | BRAINSTORM | Re-think the approach |
| MEDIUM/LOW review issues | EXECUTE | Targeted fixes |
| Test failures | EXECUTE | Fix the code |
| CI failures or PR feedback | BRAINSTORM | Always re-brainstorm |

REVIEW is mandatory — it cannot be skipped even if tests pass.

## Subagents

| Agent | Phase | Purpose |
|-------|-------|---------|
| workflow-brainstormer | BRAINSTORM | Research and creative ideation |
| workflow-planner | PLAN | Implementation planning |
| workflow-coder | EXECUTE | Code implementation (TDD) |
| workflow-implementer | EXECUTE | Used by workflow-coder and design |
| workflow-tester | TEST | Test execution and quality checks |
| review-consolidator | REVIEW | Multi-domain code review |
| commit-agent | COMMIT | Git operations and PR creation |
| monitor-agent | MONITOR | CI/CD and PR monitoring |
| team-coder | EXECUTE | Concurrent coder teammate |
| team-reviewer | REVIEW | Concurrent reviewer teammate |
| Explore | DISCOVER | Codebase exploration |

## Git Worktrees

All work workflows create isolated git worktrees before any file operations:

```
repo/
repo-worktrees/
  ├── add-auth/          # Feature worktree
  │   ├── .bob/state/    # Workflow artifacts
  │   └── ...
  └── fix-parser/
      ├── .bob/state/
      └── ...
```

## Installation

```bash
make install                # Everything (skills + agents + LSP)
make install-skills         # Skills only
make install-agents         # Subagents only
make install-mcp            # Filesystem MCP server
make enable-agent-teams     # Enable /bob:work-teams
make hooks                  # Optional: pre-commit quality checks
```

## Requirements

- Claude Code CLI
- Git

Optional: Go, golangci-lint, gocyclo (for Go-specific features)

---

*Bob - Captain of Your Agents*
