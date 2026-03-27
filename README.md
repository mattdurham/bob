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

Bob coordinates AI agent workflows for feature development. Skills invoke specialized subagents, pass state through `.bob/` artifacts, and enforce quality gates automatically.

## Spec-Driven Development

Bob treats **SPECS.md as the source of truth** for module behavior. Every workflow is spec-aware:

- **`/bob:work`** reads existing specs before making changes. If a request contradicts a contract or invariant in SPECS.md, the workflow will question it — specs can be changed, but only deliberately. Code changes to spec-driven modules must be reflected in the corresponding spec docs.

- **`/bob:explore`** prioritizes spec docs when analyzing a codebase. For spec-driven modules, it reads SPECS.md and NOTES.md first to understand contracts and design decisions before diving into implementation code. Uses concurrent analysis and adversarial challenge phases for deep, reliable exploration.

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

### `/bob:work` — Concurrent Agent Team Workflow

```
INIT → WORKTREE → BRAINSTORM → PLAN → SPAWN TEAM → EXECUTE ↔ REVIEW → COMMIT → MONITOR → COMPLETE
```

Multiple coder and reviewer teammates work in parallel through a shared task list. Requires `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`.

### `/bob:audit` — Spec Audit

```
INIT → DISCOVER → AUDIT → REPORT → COMPLETE
```

Verify code satisfies stated invariants in spec-driven modules. Read-only — reports drift but doesn't fix it.

### `/bob:explore` — Team-Based Exploration with Adversarial Challenge

```
INIT → DISCOVER → ANALYZE (4 agents) → CHALLENGE (5 agents) → DOCUMENT → COMPLETE
                     ↑                       ↓
                     └───────────────────────┘
                          (any FAIL, max 2 loops)
```

Concurrent specialist agents for codebase exploration. ANALYZE spawns 4 agents (structure, flow, patterns, dependencies). CHALLENGE spawns adversarial agents that stress-test the analysis. Failures loop back to re-analyze. No code changes.

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
make enable-agent-teams     # Enable /bob:work
make hooks                  # Optional: pre-commit quality checks
```

## Requirements

- Claude Code CLI
- Git

Optional: Go, golangci-lint, gocyclo (for Go-specific features)

---

*Bob - Captain of Your Agents*
