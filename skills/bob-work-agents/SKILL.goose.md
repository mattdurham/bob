---
name: bob-work-agents
description: Goose-native agent development workflow - INIT → WORKTREE → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMPLETE
---

# Bob Work Agents for Goose

The parent session owns this entire workflow. Delegates are bounded workers and
must not orchestrate other delegates.

Use `delegate(source: "<custom-agent>", instructions: "<bounded task>", async: true)`
for independent work, retain each task id, and collect it with
`load(source: "<task id>")`. Validate the requested `.bob/state/` artifact after
every load; a timeout, failed load, or missing artifact fails the phase.

## Required phases

1. INIT and create/reuse an isolated WORKTREE.
2. BRAINSTORM through `workflow-brainstormer`, producing
   `.bob/state/brainstorm.md`.
3. PLAN through `workflow-planner`, producing `.bob/state/plan.md`.
4. Partition the plan by non-overlapping file ownership. Delegate each slice
   directly to `workflow-implementer` asynchronously, then load all results.
5. After the write barrier, delegate `workflow-task-reviewer` and
   `workflow-code-quality` asynchronously. Load both and send bounded fixes to
   `workflow-implementer` when needed.
6. Delegate `tester`; require `.bob/state/test-results.md`. Test failures route
   back to implementation, with at most three attempts.
7. Delegate `review-consolidator`; require `.bob/state/review.md`. CRITICAL/HIGH
   routes to BRAINSTORM; MEDIUM/LOW routes to implementation. REVIEW cannot be
   skipped.
8. Delegate publication to `commit-agent` only after all gates pass, then
   delegate CI/feedback monitoring to `monitor-agent`. `NEEDS_BRAINSTORM` routes
   to BRAINSTORM. Preserve push confirmation configuration.
9. COMPLETE only when the review, tests, CI, and feedback gates pass.

Read spec documents before code and update them with implementation changes.
Never allow concurrently delegated writers to touch the same path.
