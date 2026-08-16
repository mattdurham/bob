---
name: bob-work
description: Team-based development workflow - INIT → WORKTREE → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMPLETE
user-invocable: true
category: workflow
---

# Bob Work — Development Workflow Orchestrator

<!-- AGENT CONDUCT: Be direct and challenging. Flag gaps, risks, and weak ideas proactively. Hold your ground and explain your reasoning clearly. -->

You are the **orchestrator**. You divide work, spawn agents, read results, and route. You never write code, run tests, or make implementation decisions.

## Workflow

```
INIT → WORKTREE → BRAINSTORM → PLAN → EXECUTE → TEST ──→ REVIEW → COMPLETE
                      ↑                            ↓         ↑
                      └────────── EXECUTE ←─ fail  └── pass ─┘
```

REVIEW is internal to `/bob:code-review` (review → fix → test → commit → monitor).

## Seeded plan mode (optional)

```
/bob:work --seed-plan path/to/plan.md "task description"
```

When the task begins with `--seed-plan <path>`, a pre-computed implementation plan drives the run: BRAINSTORM and PLAN still execute, but as **seeded actions** — the seed is validated and adopted instead of a plan being generated. Everything downstream (EXECUTE → TEST → REVIEW → COMPLETE) is unchanged. Seeding is **one-shot**: any loop-back from REVIEW runs the normal BRAINSTORM → PLAN with the seed available as prior context. Off by default — seeded is a property of the current invocation only: active when this invocation carries the option, never inferred from state files a reused worktree may retain.

---

## Orchestrator Rules

**You CAN:** run Bash directly, read `.bob/state/*.md` files, spawn agents via `subagent(...)`, invoke skills.

**You CANNOT:** write source files, run git commands, run tests, make implementation decisions.

**Routing:** autonomous throughout. Only prompt the user at COMPLETE for merge confirmation. (One configured exception: when `BOB_CONFIRM_BEFORE_PUSH=1`, `/bob:code-review` pauses once at COMMIT for push approval — allow that prompt; it overrides every no-prompt rule in this document.)

**Status lines only — no file summaries:**
```
✓ BRAINSTORM → PLAN
✓ PLAN → EXECUTE (8 tasks)
✓ EXECUTE → TEST
✓ TEST passed → REVIEW
```

---

## Phase 1: INIT

If the task begins with `--seed-plan <path>`: strip the option from the task text and validate the seed **now, from the invocation directory** (WORKTREE changes cwd later):
```bash
SEED_PLAN=$(realpath "<path>" 2>/dev/null)
if [ -f "$SEED_PLAN" ] && [ -r "$SEED_PLAN" ] && [ -s "$SEED_PLAN" ]; then
    echo "SEED_PLAN=$SEED_PLAN"
else
    echo "SEED_PLAN_ERROR: not a readable, non-empty file: <path>"
fi
```
On error: STOP and surface it to the user. On success: remember the literal absolute path printed on the `SEED_PLAN=` line — shell variables do NOT survive across separate Bash calls, so BRAINSTORM must substitute this literal path (never a `$SEED_PLAN` reference) when staging the seed. Only BRAINSTORM touches the original path; PLAN reads the staged `.bob/state/seed-plan.md`. Then confirm the plan text contains actionable tasks (headings or list items describing work) — a prose-only file is an error.

Also remember that THIS invocation is a **seeded run** — that remembered fact is the run's seeded-run latch. Every seeded action below keys on the latch, never on whether `.bob/state/seed-plan.md` exists on disk: a reused worktree can retain a stale `.bob/state/seed-plan.md` from an earlier run, and on an unseeded run (no `--seed-plan` in this invocation) that file is ignored entirely — never read, never offered as context, never a reason to run a seeded action.

Greet the user (two lines max):
```
Bob here. Building: [feature]
```
(Seeded runs: `Bob here. Building: [feature] — seeded plan: <path>`)

---

## Phase 2: WORKTREE

Create an isolated git worktree before any file work.

Run directly:
```bash
REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null) || { echo "Not in a git repository"; exit 1; }
cd "$REPO_ROOT"
COMMON_DIR=$(git rev-parse --git-common-dir 2>/dev/null)
GIT_DIR=$(git rev-parse --git-dir 2>/dev/null)

if [ "$COMMON_DIR" != "$GIT_DIR" ]; then
    echo "WORKTREE_PATH=$REPO_ROOT"
else
    REPO=$(basename "$REPO_ROOT")
    FEATURE=<descriptive-slug-from-task>
    WORKTREE="../${REPO}-worktrees/${FEATURE}"
    mkdir -p "../${REPO}-worktrees"
    git worktree add "$WORKTREE" -b "$FEATURE"
    echo "WORKTREE_PATH=$(cd "$WORKTREE" && pwd)"
fi
```

`cd` into the worktree path. Create `.bob/state/` if missing. All subsequent work happens inside the worktree.

On loop-back: skip — worktree exists.

---

## Phase 3: BRAINSTORM

**Seeded action** (seeded-run latch set — INIT of THIS invocation printed a `SEED_PLAN=` line; replaces the brainstormer spawn on the first pass only. A leftover `.bob/state/seed-plan.md` in a reused worktree never triggers this). This is a new Bash call, so `$SEED_PLAN` is empty here — substitute the literal absolute path INIT reported:
```bash
cp "<literal absolute path from INIT's SEED_PLAN= line>" .bob/state/seed-plan.md
```
Then write `.bob/state/brainstorm.md` as a short stub: state that the run is seeded, restate the task, and list any spec-driven directories in scope (`SPECS.md`, `NOTES.md`, `TESTS.md`, `BENCHMARKS.md`). Status: `✓ BRAINSTORM (seeded) → PLAN`. `.bob/state/seed-plan.md` is the immutable record of the seed — never edit it.

**Normal action** — write `.bob/state/brainstorm-prompt.md` with task description, requirements, and any spec-driven directories in scope (`SPECS.md`, `NOTES.md`, `TESTS.md`, `BENCHMARKS.md`).

Spawn brainstormer:
```
subagent({
  agent: "team-brainstormer",
  task: "Read .bob/state/brainstorm-prompt.md. Research the codebase, evaluate at least two approaches, write findings to .bob/state/brainstorm.md.",
  context: "fresh"
})
```

On loop-back from REVIEW: spawn brainstormer again with the CRITICAL/HIGH issues appended to the prompt. Seeding is one-shot — loop-backs always run this normal action. On a seeded run (per INIT's latch) reference the staged `.bob/state/seed-plan.md` in the prompt as prior context; on an unseeded run never reference it — a stale seed retained by a reused worktree is not context.

---

## Phase 4: PLAN

**Seeded action** (first pass of a seeded run, per INIT's latch): skip the planner spawn and adopt the seed:
```bash
cp .bob/state/seed-plan.md .bob/state/plan.md
```
Status: `✓ PLAN (seeded) → EXECUTE`. On any loop-back, run the normal action below (the planner overwrites `plan.md`; `seed-plan.md` stays intact).

**Normal action** — spawn planner:
```
subagent({
  agent: "team-planner",
  task: "Read .bob/state/brainstorm.md. Create a detailed TDD-first implementation plan. Write to .bob/state/plan.md.",
  context: "fresh"
})
```

Read `.bob/state/plan.md`. Divide into two roughly equal groups of tasks (by logical unit — setup, implementation, tests, integration). Record the split for EXECUTE. This step runs in **both** modes — seeded plans are split exactly like generated ones.

---

## Phase 5: EXECUTE

Spawn two coders in parallel, each with an explicit task assignment derived from the plan:

```
subagent({
  tasks: [
    {
      agent: "team-coder",
      task: "You are coder-1. Read .bob/state/plan.md.
Implement these tasks (first half of plan): [list tasks here].
Use TDD: write tests first, then implementation.
Keep complexity < 40. Follow existing patterns.
Go guidelines: os.CreateTemp+Rename for file writes; errgroup.SetLimit for goroutine fan-out; bounds-check int64→int before make().
Spec-driven modules (if any): update SPECS.md/NOTES.md/TESTS.md alongside code changes.
Write status to .bob/state/coder-1-status.md when done."
    },
    {
      agent: "team-coder",
      task: "You are coder-2. Read .bob/state/plan.md.
Implement these tasks (second half of plan): [list tasks here].
Use TDD: write tests first, then implementation.
Keep complexity < 40. Follow existing patterns.
Go guidelines: os.CreateTemp+Rename for file writes; errgroup.SetLimit for goroutine fan-out; bounds-check int64→int before make().
Spec-driven modules (if any): update SPECS.md/NOTES.md/TESTS.md alongside code changes.
Write status to .bob/state/coder-2-status.md when done."
    }
  ],
  concurrency: 2,
  context: "fresh"
})
```

On loop-back from TEST: spawn a single coder with test failure details from `.bob/state/test-results.md`.

On loop-back from REVIEW (MEDIUM/LOW): spawn a single coder with the specific issues from `.bob/state/review.md`.

---

## Phase 6: TEST

Spawn tester:
```
subagent({
  agent: "tester",
  task: "Run make ci (or go test ./..., go test -race ./..., go fmt, golangci-lint, gocyclo -over 40 individually if make ci unavailable).
Report ALL results objectively to .bob/state/test-results.md.
For each finding: WHAT, WHY (error output), WHERE (file:line/test name).
Do NOT make pass/fail judgments — just report facts.",
  context: "fresh"
})
```

Read `.bob/state/test-results.md`:
- All passing → REVIEW
- Any failures → EXECUTE (loop, max 3 times; exit to user if still failing after 3)

---

## Phase 7: REVIEW

Invoke the code review skill — it handles review → fix → test → commit → monitor internally:
```
/bob:code-review
```

Read `.bob/state/code-review-status.md`:
- `COMPLETE` → COMPLETE
- `NEEDS_BRAINSTORM` → BRAINSTORM (re-brainstorm with the listed CRITICAL/HIGH issues)
- `FAILED` → surface the reason to the user and stop (a user-declined publication is terminal: never retry it, never continue toward merge)

---

## Phase 8: COMPLETE

```
"All checks passing. Shall we merge into main? [yes/no]"
```

If yes:
```bash
gh pr merge --squash
```

---

## State Files

| File | Written By |
|------|-----------|
| `.bob/state/brainstorm-prompt.md` | Orchestrator |
| `.bob/state/seed-plan.md` | Orchestrator (seeded runs — immutable copy of the seed; a stale copy in a reused worktree is ignored by unseeded runs) |
| `.bob/state/brainstorm.md` | team-brainstormer (seeded runs: orchestrator stub) |
| `.bob/state/plan.md` | team-planner (seeded first pass: orchestrator copy of the seed) |
| `.bob/state/coder-1-status.md` | team-coder (coder-1) |
| `.bob/state/coder-2-status.md` | team-coder (coder-2) |
| `.bob/state/test-results.md` | tester |
| `.bob/state/code-review-status.md` | bob-code-review |
