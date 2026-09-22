---
name: bob-work
description: Goose-native development workflow - INIT → WORKTREE → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMPLETE
---

# Bob Work for Goose

You are the parent orchestrator. Keep orchestration in this session because a
Goose delegate cannot recursively delegate. Use the Summon extension and run in
autonomous permission mode.

For each delegated job, record the returned task id. Independent jobs may use:

```text
job = delegate(source: "workflow-coder", instructions: "One bounded assignment with exact worktree, files, and artifact path", async: true)
result = load(source: "<job task id>")
```

Every `load` must succeed and every promised artifact must exist before routing.
Missing output is a phase failure, never a pass.

## Workflow

`INIT → WORKTREE → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMPLETE`

1. **INIT:** Parse the request. Support `--seed-plan <path>` exactly as the
   standard Bob workflow does: validate a readable non-empty actionable plan
   and remember that this invocation is seeded.
2. **WORKTREE:** Reuse the current worktree when already in one; otherwise create
   a feature worktree. Create `.bob/state/` there. All delegate instructions
   must name its absolute path.
3. **BRAINSTORM:** For a normal run, delegate to `workflow-brainstormer` and
   require `.bob/state/brainstorm.md`. For the first seeded pass, copy the seed
   to `.bob/state/seed-plan.md` and write the short brainstorm stub locally.
4. **PLAN:** Delegate to `workflow-planner` and require `.bob/state/plan.md`.
   On the first seeded pass adopt the staged seed instead. Split the plan into
   assignments with disjoint file ownership.
5. **EXECUTE:** Start one async `workflow-coder` delegate per disjoint assignment.
   Record ownership and task ids in `.bob/state/team-status.md`, then call
   `load(source: "<task id>")` for every job. Serialize overlapping writes.
   Require a status artifact from each coder.
6. **TEST:** Delegate to `tester`; load the result and require
   `.bob/state/test-results.md`. Failures route to EXECUTE, maximum three loops.
7. **REVIEW:** Run `workflow-task-reviewer` and `workflow-code-quality` as async,
   read-only delegates, load both, then delegate to `review-consolidator` and
   require `.bob/state/review.md`. REVIEW is mandatory.
8. **ROUTE:** CRITICAL or HIGH findings, CI failure, PR feedback, or
   `NEEDS_BRAINSTORM` route to BRAINSTORM. MEDIUM/LOW findings route to EXECUTE.
   Preserve any `BOB_CONFIRM_BEFORE_PUSH=1` publication prompt.
9. **COMPLETE:** Only after tests and review pass, report completion and ask for
   the existing final merge decision.

On a review loop, seeded mode is no longer special: use the normal BRAINSTORM
and PLAN phases. Preserve `.bob/state/brainstorm.md`, `.bob/state/plan.md`,
`.bob/state/test-results.md`, `.bob/state/review.md`, and all spec-driven module
updates.
