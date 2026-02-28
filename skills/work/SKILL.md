---
name: bob:work
description: Simple direct workflow - no agents, no ceremony. INIT → WORKTREE → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMPLETE
user-invocable: true
category: workflow
---

# Simple Work Workflow

<!-- AGENT CONDUCT: Be direct and challenging. Flag gaps, risks, and weak ideas proactively. -->

You are executing a **direct development workflow**. No subagents, no orchestration layers — you do all the work yourself, linearly, from start to finish.

## Workflow Diagram

```
INIT → WORKTREE → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMPLETE
```

No loop-backs in the outer workflow. The REVIEW phase invokes `/bob:code-review`, which handles its own FIX loop, commit, and CI monitoring.

---

## Phase 1: INIT

**Goal:** Brief greeting and acknowledgment.

```
"Hey! Bob here. Let's build this directly — no agents, no ceremony.

Building: [feature description]

Starting up..."
```

Proceed immediately to WORKTREE.

---

## Phase 2: WORKTREE

**Goal:** Create an isolated git worktree for development.

**Actions:**

1. Check if already in a worktree:
   ```bash
   COMMON_DIR=$(git rev-parse --git-common-dir 2>/dev/null || echo "")
   GIT_DIR=$(git rev-parse --git-dir 2>/dev/null || echo "")
   ```
   If `$COMMON_DIR != $GIT_DIR` and `$COMMON_DIR != ".git"`, you're already in a worktree — skip creation.

2. If not in a worktree, create one:
   ```bash
   REPO_NAME=$(basename $(git rev-parse --show-toplevel))
   FEATURE_NAME="<descriptive-feature-name>"  # derive from task
   WORKTREE_DIR="../${REPO_NAME}-worktrees/${FEATURE_NAME}"
   mkdir -p "../${REPO_NAME}-worktrees"
   git worktree add "$WORKTREE_DIR" -b "$FEATURE_NAME"
   ```

3. Create `.bob/state/` directory and `cd` into the worktree:
   ```bash
   mkdir -p "$WORKTREE_DIR/.bob/state"
   cd "$WORKTREE_DIR"
   ```

4. Confirm the worktree path and branch name.

**From this point forward, ALL work happens in the worktree.**

---

## Phase 3: BRAINSTORM

**Goal:** Interactive ideation and design exploration.

**Actions:**

Invoke the brainstorming skill:
```
/bob:internal:brainstorming
Topic: [The feature/task to implement]
```

The brainstorming skill will:
- Ask clarifying questions interactively
- Explore approaches and trade-offs
- Detect spec-driven modules in scope (directories with SPECS.md, NOTES.md, TESTS.md, BENCHMARKS.md, or .go files with the NOTE invariant)
- Write the design document to `.bob/state/design.md`

Wait for brainstorming to complete before proceeding.

---

## Phase 4: PLAN

**Goal:** Write a concrete implementation plan directly.

**Actions:**

1. Read `.bob/state/design.md` (or `.bob/state/brainstorm.md` if design.md doesn't exist).

2. Scan target directories for spec-driven modules:
   - Look for SPECS.md, NOTES.md, TESTS.md, BENCHMARKS.md
   - Look for `.go` files with: `// NOTE: Any changes to this file must be reflected in the corresponding specs.md or NOTES.md.`
   - Note which modules require doc updates alongside code changes.

3. Write `.bob/state/plan.md` directly with:
   - Numbered tasks, each with exact file paths
   - TDD format: write test first, then implement
   - Spec-driven doc updates inline with their corresponding code tasks (not as a separate step)
   - Each task should be 2-5 minutes of work
   - Verification steps per task

**Format:**
```markdown
# Implementation Plan

## Task 1: [description]
- **Files:** path/to/file.go, path/to/file_test.go
- **Test first:** Write test for [behavior]
- **Implement:** [what to change]
- **Spec docs:** Update SPECS.md section X (if spec-driven module)
- **Verify:** `go test ./path/to/...`

## Task 2: ...
```

**Only artifact:** `.bob/state/plan.md`

---

## Phase 5: EXECUTE

**Goal:** Implement the plan directly using TDD.

**Actions:**

For each task in `.bob/state/plan.md`:

1. **Write the test first** — create or update the `_test.go` file
2. **Run the test** — verify it fails (confirms the test is meaningful)
3. **Implement the code** — make the test pass
4. **Run the test again** — verify it passes
5. **Update spec docs** — if the target directory is a spec-driven module:
   - Update SPECS.md for API/contract changes
   - Add dated entry to NOTES.md for design decisions
   - Update TESTS.md with scenario/setup/assertions for new tests
   - Update BENCHMARKS.md for new benchmarks
   - Add NOTE invariant comment to new .go files
   - Never delete NOTES.md entries — add Addendum notes if reversing a decision

Work through tasks sequentially. Fix issues as you encounter them — no looping back to a separate phase.

---

## Phase 6: TEST

**Goal:** Run full test suite and fix any failures.

**Actions:**

1. Run `make ci` (or equivalent test commands if make target unavailable):
   ```bash
   make ci
   ```

2. If any failures:
   - Read the output carefully
   - Fix the failing code directly
   - Re-run `make ci`
   - Repeat until clean

3. If `make ci` is not available, run individually:
   ```bash
   go test ./...
   go test -race ./...
   go fmt ./...
   golangci-lint run   # if available
   ```

Do not proceed until tests pass. Fix issues inline — no separate EXECUTE loop.

---

## Phase 7: REVIEW

**Goal:** Comprehensive code review, fix, commit, and CI monitoring.

**Actions:**

Invoke the code-review skill:
```
Invoke: /bob:code-review
```

The code-review skill handles:
1. Multi-domain review (security, bugs, errors, quality, performance, Go idioms, architecture, docs)
2. Spec-driven compliance check (SPECS.md, NOTES.md, TESTS.md, BENCHMARKS.md)
3. FIX loop — fixes issues, re-runs tests until clean
4. Creates commit and pushes PR
5. Monitors CI

After code-review completes, proceed to COMPLETE.

---

## Phase 8: COMPLETE

**Goal:** Summary.

**Actions:**

Display a summary:

```
Done!

Branch: <branch-name>
Worktree: <worktree-path>

Changes:
  - [brief list of what was implemented]

Next steps:
  - Clean up worktree: git worktree remove <worktree-path>

-- Bob
```

---

## Rules

- **No agents (except via skills).** You do all work directly — reading, writing, testing. Skills handle their own subagents internally.
- **No outer loop-backs.** Linear flow only. Fix issues inline in EXECUTE. The REVIEW phase's `/bob:code-review` skill handles its own FIX loop.
- **One artifact.** Only `.bob/state/plan.md` is written as a workflow artifact.
- **Spec-driven compliance.** Detect and enforce doc updates in PLAN and EXECUTE.
- **TDD.** Write tests first in EXECUTE phase.
- **Skill invocations.** `/bob:internal:brainstorming` in BRAINSTORM; `/bob:code-review` in REVIEW.
