# Belayin' Pin Bob - Agent Guidance

This repository uses **Belayin' Pin Bob** for workflow orchestration through Claude skills and subagents. Skills invoke specialized subagents, pass state through `.bob/` artifacts, and enforce quality gates automatically.

## What Bob Provides

Bob installs workflow skills and specialized subagents to `~/.claude/`:

1. **Workflow Skills** - User-invocable workflows (`/bob:work`, `/bob:matt-research`)
2. **Subagent Orchestration** - Specialized agents for each workflow phase
3. **State Management** - Persistent workflow artifacts in `.bob/state/` directory
4. **Git Worktrees** - Isolated development environments

## Available Workflows

### /bob:work — Development Workflow
```
Plan intake → WORKTREE → supplied plan or bob:matt-grill-me → coding worklist → EXECUTE → TEST → REVIEW → PR → merge decision
```

Ask whether the user has an existing plan. Use the supplied plan, or run the
imported `bob:matt-grill-me` interview and wait for confirmed shared understanding before
planning. Both paths produce a coding worklist; choose direct execution or an
appropriate number of coders based on dependencies and file ownership. Completed assignments may be reviewed while independent coding continues;
verification and final review wait for all writers and reviewers to finish.

### /bob:matt-research — Research and Codebase Exploration

Investigate a question against primary sources, trace relevant codebase behavior,
check assumptions, and save one grounded Markdown report. Delegation is optional.
No source changes.

### /bob:audit — Spec Audit
```
INIT → DISCOVER → AUDIT → REPORT → COMPLETE
```

Check documented guarantees against implementation and tests. Report each as
satisfied, violated, or unverified, with evidence and suggested remedies in
`.bob/state/audit-report.md`. Code and specs remain unchanged.

## Feedback and Verification

The calling workflow owns fixes and plan decisions. Reconsider agreed assumptions
when findings invalidate the approach; make bounded fixes for other issues. Tests,
CI, and PR feedback requiring changes return to implementation, followed by
relevant verification and final review. Failed or incomplete gates block completion.

Final review is mandatory, even when tests pass.

## Subagents

| Agent | Phase | Purpose |
|-------|-------|---------|
| workflow-brainstormer | BRAINSTORM | Research and creative ideation |
| workflow-planner | PLAN | Implementation planning |
| workflow-coder | EXECUTE | Code implementation coordination |
| workflow-implementer | EXECUTE | Actual code writing (used by workflow-coder and specs) |
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
make enable-agent-teams     # Optional: enable native Claude teams
make hooks                  # Optional: pre-commit quality checks
```

## Starting a Workflow

```
/bob:work "Add rate limiting to API"
```

The skill asks about an existing plan and, when needed, runs `bob:matt-grill-me`. After
planning it creates coding assignments as appropriate and drives autonomously
through verification, review, and publication, with the final merge decision
and the configured push-approval prompt when `BOB_CONFIRM_BEFORE_PUSH=1`.

---

*Bob - Captain of Your Agents*
