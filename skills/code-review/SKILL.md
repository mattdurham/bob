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

**Goal:** Comprehensive code review by 9 specialized agents in parallel

Spawn 9 reviewer agents in parallel (single message, 9 Task calls):
```
Task(subagent_type: "workflow-reviewer",
     description: "Code quality review",
     run_in_background: true,
     prompt: "Perform 3-pass code review focusing on code logic, bugs, and best practices.
             Write findings to .bob/state/review-code.md with severity levels.")

Task(subagent_type: "security-reviewer",
     description: "Security vulnerability review",
     run_in_background: true,
     prompt: "Scan code for security vulnerabilities (OWASP Top 10, secrets, input validation).
             Write findings to .bob/state/review-security.md with severity levels.")

Task(subagent_type: "performance-analyzer",
     description: "Performance bottleneck review",
     run_in_background: true,
     prompt: "Analyze code for performance issues (complexity, memory, N+1 patterns).
             Write findings to .bob/state/review-performance.md with severity levels.")

Task(subagent_type: "docs-reviewer",
     description: "Documentation accuracy review",
     run_in_background: true,
     prompt: "Review documentation for accuracy and completeness.
             Write findings to .bob/state/review-docs.md with severity levels.")

Task(subagent_type: "architect-reviewer",
     description: "Architecture and design review",
     run_in_background: true,
     prompt: "Evaluate system architecture and design decisions.
             Write findings to .bob/state/review-architecture.md with severity levels.")

Task(subagent_type: "code-reviewer",
     description: "Comprehensive code quality review",
     run_in_background: true,
     prompt: "Conduct deep code review (logic, security, performance, maintainability).
             Write findings to .bob/state/review-code-quality.md with severity levels.")

Task(subagent_type: "golang-pro",
     description: "Go-specific code review",
     run_in_background: true,
     prompt: "Review Go code for idiomatic patterns and best practices.
             Write findings to .bob/state/review-go.md with severity levels.")

Task(subagent_type: "debugger",
     description: "Bug diagnosis and debugging review",
     run_in_background: true,
     prompt: "Perform systematic debugging analysis on the code:
             - Potential null pointer dereferences and panic conditions
             - Race conditions and concurrency bugs
             - Resource leaks and error handling gaps
             Write findings to .bob/state/review-debug.md with severity levels.")

Task(subagent_type: "error-detective",
     description: "Error pattern analysis review",
     run_in_background: true,
     prompt: "Analyze code for error handling patterns and potential failure modes:
             - Error handling consistency and missing error checks
             - Retry logic and failure recovery patterns
             - Timeout and circuit breaker patterns
             Write findings to .bob/state/review-errors.md with severity levels.")
```

After all 9 agents complete, consolidate findings into `.bob/state/review.md`.

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
