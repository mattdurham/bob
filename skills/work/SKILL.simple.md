---
name: bob:work
description: Simple direct workflow - no agents, no ceremony. INIT → WORKTREE → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMMIT → COMPLETE
user-invocable: true
category: workflow
---

# Simple Work Workflow

<!-- AGENT CONDUCT: Be direct and challenging. Flag gaps, risks, and weak ideas proactively. -->

You are executing a **direct development workflow**. No subagents, no orchestration layers — you do all the work yourself, linearly, from start to finish.

## Workflow Diagram

```
INIT → WORKTREE → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMMIT → COMPLETE
```

No loop-backs. No severity routing. Fix issues inline as you encounter them.

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
- Detect CLAUDE.md modules in scope (directories containing CLAUDE.md)
- Write the design document to `.bob/state/design.md`

Wait for brainstorming to complete before proceeding.

---

## Phase 4: PLAN

**Goal:** Write a concrete implementation plan directly.

**Actions:**

1. Read `.bob/state/design.md` (or `.bob/state/brainstorm.md` if design.md doesn't exist).

2. Scan target directories for CLAUDE.md modules:
   - Look for `CLAUDE.md` in each directory you plan to modify
   - Note which modules require CLAUDE.md updates alongside code changes.

3. Write `.bob/state/plan.md` directly with:
   - Numbered tasks, each with exact file paths
   - TDD format: write test first, then implement
   - CLAUDE.md updates inline with their corresponding code tasks (not as a separate step)
   - Each task should be 2-5 minutes of work
   - Verification steps per task

**Format:**
```markdown
# Implementation Plan

## Task 1: [description]
- **Files:** path/to/file.go, path/to/file_test.go
- **Test first:** Write test for [behavior]
- **Implement:** [what to change]
- **CLAUDE.md:** Update invariant N if affected (if CLAUDE.md module)
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
5. **Update CLAUDE.md** — if the target directory contains CLAUDE.md and your changes
   affect any numbered invariant, update CLAUDE.md to reflect the current truth.
   Keep it tidy: only numbered invariants, axioms, assumptions, and non-obvious constraints.
   Never add trivial, ephemeral, or obviously code-derivable content.

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

**Goal:** Lightweight self-review via git diff.

**Actions:**

1. Review all changes:
   ```bash
   git diff
   git diff --stat
   ```

2. Check for common issues:
   - Security: SQL injection, command injection, XSS, hardcoded secrets
   - Bugs: nil pointer dereferences, off-by-one errors, unchecked errors
   - Quality: functions too long, unclear naming, missing error handling
   - Performance: N+1 queries, unnecessary allocations, missing caching

3. **CLAUDE.md compliance check:** For each changed directory that contains CLAUDE.md,
   verify that CLAUDE.md accurately reflects any affected numbered invariants.

4. Fix any issues found directly — edit the files, re-run tests if needed.

No `.bob/state/review.md` artifact. No severity routing. Just fix what you find and move on.

---

## Phase 8: COMMIT

**Goal:** Create a local commit. No push, no PR.

**Actions:**

1. Stage relevant files (never `git add -A`):
   ```bash
   git add <specific files>
   ```

2. Create commit with descriptive message:
   ```bash
   git commit -m "feat: [description of changes]"
   ```

**Do NOT:**
- `git push`
- `gh pr create`
- Push to any remote

The commit stays local on the worktree branch.

---

## Phase 9: COMPLETE

**Goal:** Summary and next-step hints.

**Actions:**

Display a summary:

```
Done!

Branch: <branch-name>
Commit: <short-hash> <commit-message>
Worktree: <worktree-path>

Changes:
  - [brief list of what was implemented]

Next steps:
  - Review the changes: cd <worktree-path> && git log --oneline -5
  - Push when ready: git push -u origin <branch-name>
  - Create PR: gh pr create --title "..." --body "..."
  - Clean up worktree: git worktree remove <worktree-path>

-- Bob
```

---

## Rules

- **No agents.** You do all work directly — reading, writing, testing, reviewing.
- **No loop-backs.** Linear flow only. Fix issues inline when you find them.
- **One artifact.** Only `.bob/state/plan.md` is written as a workflow artifact.
- **Local commit only.** No push, no PR, no MONITOR phase.
- **CLAUDE.md compliance.** Detect and enforce invariant updates in PLAN, EXECUTE, and REVIEW.
- **TDD.** Write tests first in EXECUTE phase.
- **Only invocation.** The only skill invoked is `/bob:internal:brainstorming` in the BRAINSTORM phase.
