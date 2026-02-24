---
name: bob:code-review
description: Code review workflow orchestrator - REVIEW → FIX → TEST → COMMIT → MONITOR
user-invocable: true
category: workflow
---

# Code Review Workflow Orchestrator

You orchestrate a **code review workflow** for reviewing, fixing issues, and verifying quality.

## Workflow Diagram

```
INIT → REVIEW → FIX → TEST → COMMIT → MONITOR → COMPLETE
          ↑      ↓                    ↓
          └──────┴────────────────────┘
         (loop on issues)
```

## Flow Control Rules

- **REVIEW → FIX**: Issues found
- **TEST → REVIEW**: Re-verify after fixes
- **MONITOR → REVIEW**: CI failures or PR feedback (review and fix again)

**NEVER commit or push before the COMMIT phase.** No `git add`, `git commit`, `git push`, or `gh pr create` until you reach Phase 5: COMMIT. Subagents must not commit either.

---

## Execution Rules

**CRITICAL: All subagents MUST run in background**

- ✅ **ALWAYS use `run_in_background: true`** for ALL Task calls
- ✅ **After spawning agents, STOP** - do not poll or check status
- ✅ **Wait for agent completion notification** - you'll be notified automatically
- ❌ **Never use foreground execution** - it blocks the workflow

---

## Phase 1: INIT

Check if already in worktree, or create .bob directory and isolated worktree:
```bash
# Check if we're already in a worktree
COMMON_DIR=$(git rev-parse --git-common-dir 2>/dev/null || echo "")
GIT_DIR=$(git rev-parse --git-dir 2>/dev/null || echo "")

if [ "$COMMON_DIR" != "$GIT_DIR" ] && [ "$COMMON_DIR" != ".git" ]; then
    echo "Already in worktree - skipping creation"
    mkdir -p .bob/state .bob/planning
else
    mkdir -p .bob/state .bob/planning

    # Create worktree for isolated code review
    REPO_NAME=$(basename $(git rev-parse --show-toplevel))
    FEATURE_NAME="<review-name>"  # e.g., "review-security-fix", "review-pr-123"

    # Create worktree directory structure
    WORKTREE_DIR="../${REPO_NAME}-worktrees/${FEATURE_NAME}"
    mkdir -p "../${REPO_NAME}-worktrees"

    # Create new branch and worktree
    git worktree add "$WORKTREE_DIR" -b "$FEATURE_NAME"

    # Change to worktree directory
    cd "$WORKTREE_DIR"
fi
```

---

## Phase 2: REVIEW

**Goal:** Comprehensive code review across all quality domains

Spawn a single review-consolidator agent:
```
Task(subagent_type: "review-consolidator",
     description: "Comprehensive code review",
     run_in_background: true,
     prompt: "Perform a thorough multi-domain code review covering: security, bug diagnosis,
             error handling, code quality, performance, Go idioms, architecture, and documentation.
             Write consolidated report to .bob/state/review.md with severity levels and routing recommendation.")
```

**Output:** `.bob/state/review.md` (consolidated report)

**Decision:**
- No issues → COMMIT
- Issues found → FIX

---

## Phase 3: FIX

**Goal:** Fix identified issues

Spawn workflow-coder:
```
Task(subagent_type: "workflow-coder",
     description: "Fix review issues",
     run_in_background: true,
     prompt: "Fix issues in .bob/state/review.md.
             Keep changes minimal and focused.")
```

**Input:** `.bob/state/review.md`
**Output:** Code fixes

---

## Phase 4: TEST

**Goal:** Verify fixes

Spawn workflow-tester:
```
Task(subagent_type: "workflow-tester",
     description: "Run tests",
     run_in_background: true,
     prompt: "Run the complete test suite and all quality checks:
             1. go test ./... (all tests must pass)
             2. go test -race ./... (no race conditions)
             3. go test -cover ./... (report coverage)
             4. go fmt ./... (code must be formatted)
             5. golangci-lint run (no lint issues)
             6. gocyclo -over 40 . (no complex functions)
             Report all results in .bob/state/test-results.md.
             Working directory: [worktree-path]")
```

**Output:** `.bob/state/test-results.md`

**Decision:**
- Tests pass → REVIEW (verify fixes are correct!)
- Tests fail → FIX

---

## Phase 5: COMMIT

**This is the FIRST phase where git operations are allowed.**

**PREREQUISITE:** `.bob/state/review.md` MUST exist. If it does not, STOP and go back to REVIEW. Never commit unreviewed code.

1. Verify review was completed:
   ```bash
   test -f .bob/state/review.md || { echo "REVIEW not completed"; exit 1; }
   ```

2. Show the user a summary of all changes and review findings.

3. Create PR (default — no need to ask):
   ```bash
   git add [relevant-files]
   git commit -m "$(cat <<'EOF'
   fix: address code review findings

   Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
   EOF
   )"
   git push -u origin $(git branch --show-current)
   gh pr create --title "Code review fixes" --body "Description"
   ```

---

## Phase 6: MONITOR

Check CI (only if a PR was created):
```bash
gh pr checks
```

**If failures:** Loop to REVIEW to analyze and replan

---

## Phase 7: COMPLETE

Review complete!
