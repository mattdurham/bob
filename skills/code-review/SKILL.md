---
name: code-review
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

---

## Execution Rules

**CRITICAL: All subagents MUST run in background**

- ✅ **ALWAYS use `run_in_background: true`** for ALL Task calls
- ✅ **After spawning agents, STOP** - do not poll or check status
- ✅ **Wait for agent completion notification** - you'll be notified automatically
- ❌ **Never use foreground execution** - it blocks the workflow

---

## Phase 1: INIT

Create bots/ directory and isolated worktree:
```bash
mkdir -p bots

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
             Write findings to bots/review-code.md with severity levels.")

Task(subagent_type: "security-reviewer",
     description: "Security vulnerability review",
     run_in_background: true,
     prompt: "Scan code for security vulnerabilities (OWASP Top 10, secrets, input validation).
             Write findings to bots/review-security.md with severity levels.")

Task(subagent_type: "performance-analyzer",
     description: "Performance bottleneck review",
     run_in_background: true,
     prompt: "Analyze code for performance issues (complexity, memory, N+1 patterns).
             Write findings to bots/review-performance.md with severity levels.")

Task(subagent_type: "docs-reviewer",
     description: "Documentation accuracy review",
     run_in_background: true,
     prompt: "Review documentation for accuracy and completeness.
             Write findings to bots/review-docs.md with severity levels.")

Task(subagent_type: "architect-reviewer",
     description: "Architecture and design review",
     run_in_background: true,
     prompt: "Evaluate system architecture and design decisions.
             Write findings to bots/review-architecture.md with severity levels.")

Task(subagent_type: "code-reviewer",
     description: "Comprehensive code quality review",
     run_in_background: true,
     prompt: "Conduct deep code review (logic, security, performance, maintainability).
             Write findings to bots/review-code-quality.md with severity levels.")

Task(subagent_type: "golang-pro",
     description: "Go-specific code review",
     run_in_background: true,
     prompt: "Review Go code for idiomatic patterns and best practices.
             Write findings to bots/review-go.md with severity levels.")

Task(subagent_type: "debugger",
     description: "Bug diagnosis and debugging review",
     run_in_background: true,
     prompt: "Perform systematic debugging analysis on the code:
             - Potential null pointer dereferences and panic conditions
             - Race conditions and concurrency bugs
             - Resource leaks and error handling gaps
             Write findings to bots/review-debug.md with severity levels.")

Task(subagent_type: "error-detective",
     description: "Error pattern analysis review",
     run_in_background: true,
     prompt: "Analyze code for error handling patterns and potential failure modes:
             - Error handling consistency and missing error checks
             - Retry logic and failure recovery patterns
             - Timeout and circuit breaker patterns
             Write findings to bots/review-errors.md with severity levels.")
```

After all 9 agents complete, consolidate findings into `bots/review.md`.

**Output:** `bots/review.md` (consolidated report)

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
     prompt: "Fix issues in bots/review.md.
             Keep changes minimal and focused.")
```

**Input:** `bots/review.md`
**Output:** Code fixes

---

## Phase 4: TEST

**Goal:** Verify fixes

Spawn workflow-tester:
```
Task(subagent_type: "workflow-tester",
     description: "Run tests",
     run_in_background: true,
     prompt: "Run full test suite.
             Write results to bots/test-results.md.")
```

**Output:** `bots/test-results.md`

**Decision:**
- Tests pass → REVIEW (verify fixes are correct!)
- Tests fail → FIX

---

## Phase 5: COMMIT

Create commit:
```bash
git add [files]
git commit -m "fix: address code review findings

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
git push
gh pr create --title "Code review fixes"
```

---

## Phase 6: MONITOR

Check CI:
```bash
gh pr checks
```

**If failures:** Loop to REVIEW to analyze and replan

---

## Phase 7: COMPLETE

Review complete! Ready to merge.
