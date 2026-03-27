# Belayin' Pin Bob - Agent Guidance

This repository uses **Belayin' Pin Bob** for workflow orchestration through Claude skills and subagents. Skills invoke specialized subagents, pass state through `.bob/` artifacts, and enforce quality gates automatically.

## What Bob Provides

Bob installs workflow skills and specialized subagents to `~/.claude/`:

1. **Workflow Skills** - User-invocable workflows (`/bob:work`, `/bob:work-teams`, `/bob:explore`, `/bob:design`)
2. **Subagent Orchestration** - Specialized agents for each workflow phase
3. **State Management** - Persistent workflow artifacts in `.bob/state/` directory
4. **Git Worktrees** - Isolated development environments

## Available Workflows

### /bob:work — Simple Direct Workflow
```
INIT → WORKTREE → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMMIT → COMPLETE
```

You do all the work yourself. No subagents, no orchestration. Linear flow, local commit only.

### /bob:work-teams — Concurrent Agent Team Workflow
```
INIT → WORKTREE → BRAINSTORM → PLAN → SPAWN TEAM → EXECUTE <-> REVIEW → COMMIT → MONITOR → COMPLETE
```

Multiple coder and reviewer teammates work in parallel. Requires `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`.

### /bob:explore — Read-Only Exploration
```
INIT → DISCOVER → ANALYZE → DOCUMENT → COMPLETE
```

Spec-aware codebase exploration. No code changes.

### /bob:audit — Spec Audit
```
INIT → DISCOVER → AUDIT → REPORT → COMPLETE
```

Verify code satisfies stated invariants in spec-driven modules. Read-only — reports drift but doesn't fix it.

### /bob:design — Spec Scaffolding
```
INIT → GATHER → [ANALYZE] → SCAFFOLD → COMPLETE
```

Create or apply spec-driven module structure.

## Loop-Back Rules

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
| workflow-coder | EXECUTE | Code implementation coordination |
| workflow-implementer | EXECUTE | Actual code writing (used by workflow-coder and design) |
| workflow-tester | TEST | Test execution and quality checks |
| review-consolidator | REVIEW | Multi-domain code review |
| commit-agent | COMMIT | Git operations and PR creation |
| monitor-agent | MONITOR | CI/CD and PR monitoring |
| team-coder | EXECUTE | Concurrent coder teammate |
| team-reviewer | REVIEW | Concurrent reviewer teammate |
| Explore | DISCOVER | Codebase exploration |

## Workflow Artifacts

All workflows store artifacts in `.bob/state/` within the worktree:

```
.bob/
  state/
    brainstorm.md       # Research and approach decisions
    plan.md             # Detailed implementation plan
    test-results.md     # Test execution results
    review.md           # Consolidated code review findings
```

These files persist across Claude sessions and serve as context for subsequent phases.

## Spec-Driven Development

Bob treats spec documents as the **source of truth** for module behavior:

- **Planner** reads SPECS.md/CLAUDE.md invariants and derives verification tests from them
- **Coder** receives the actual invariants as hard constraints in the implementation prompt
- **Reviewer** verifies code satisfies stated invariants (not just that docs were updated)

See CLAUDE.md for full spec-driven module documentation.

## Installation

```bash
make install                # Everything (skills + agents + LSP)
make install-skills         # Skills only
make install-agents         # Subagents only
make enable-agent-teams     # Enable /bob:work-teams
make hooks                  # Optional: pre-commit quality checks
```

## Starting a Workflow

```
/bob:work-teams "Add rate limiting to API"
```

The skill creates a worktree, spawns teammate agents that work concurrently, and drives autonomously from INIT through COMMIT — only prompting at the final merge.

---

*Bob - Captain of Your Agents*
