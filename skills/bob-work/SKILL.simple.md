---
name: bob-work
description: Single-agent development workflow — INIT → PLAN → EXECUTE → TEST → REVIEW → COMPLETE.
user-invocable: true
category: workflow
---

# Bob Work — Single Agent

Run the development workflow yourself in the current workspace. This variant is
strictly single-agent: never call `Task`, `Agent`, `subagent`, create teammates,
or use agent teams. Do not delegate exploration, planning, implementation,
testing, review, commits, or monitoring.

1. Inspect the repository, current branch, working tree, relevant guidance, and
   applicable specs. Preserve unrelated user changes.
2. Form a concise implementation plan and record it in `.bob/state/plan.md`.
3. Implement the requested change directly, keeping the scope minimal.
4. Run the most relevant tests, linters, formatters, or verification commands.
5. Review the diff yourself for correctness, regressions, spec drift, and missing
   tests. Fix issues found and rerun verification.
6. Report changed files, verification results, remaining risks, and the final
   routing recommendation. Commit only when the user requested a commit. If you
   push and open or update a PR, and `BOB_CONFIRM_BEFORE_PUSH` is exactly `1`
   (check with `echo "confirm-before-push: ${BOB_CONFIRM_BEFORE_PUSH:-unset}"`),
   commit locally first, then show the user every commit the push will publish
   (`git log --oneline HEAD --not --remotes=origin`) and the proposed PR title
   and body verbatim, and ask `Push this branch and create or update its PR
   with the title and body shown above? [push / stop]` before running `git
   push` or `gh pr create`/`gh pr edit`. Only proceed on an exact `push` reply.

Use `.bob/state/brainstorm.md`, `.bob/state/plan.md`, and
`.bob/state/test-results.md` for direct artifacts when useful. If blocked by a
missing decision or external state, explain the exact blocker instead of spawning
another agent.

**Conflicting inputs.** You have no coder to escalate to, so a plan/code conflict
lands on you. Do not guess and do not stall: bisect what the conflict actually
affects, implement everything it does not, and adopt already-landed code as the
tiebreaker (compiling code is the settled shape; a plan or decision record is recorded
intent). Then surface it to the user in your final message as a `DECISION NEEDED` entry
— question, both readings, affected symbols, the tiebreaker you applied, and the impact
if it is wrong — and record the ruling in the plan or decision log so the next plan does
not repeat the conflict. A surfaced conflict is a successful outcome, not a failure.
