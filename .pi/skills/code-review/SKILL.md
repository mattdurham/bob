---
name: bob:code-review
description: Code review workflow orchestrator - REVIEW → FIX → TEST → COMMIT → MONITOR
user-invocable: true
category: workflow
---

# Code Review Workflow Orchestrator

You orchestrate the code review lifecycle: review the current diff, fix issues in a bounded loop, commit, and monitor CI. You are a **pure orchestrator** — you never write code, run tests, or execute git commands yourself.

## Workflow Diagram

```
INIT → REVIEW → ROUTE ──── no issues ──→ COMMIT → MONITOR → COMPLETE
                  ↑                            ↑
                  │   issues found             │ CI failures / changes requested
                  ↓                            └──── NEEDS_BRAINSTORM (exit)
                 FIX → TEST ─── pass ──────────┘
                  ↑      │
                  └──────┘ fail (max 3 loops)
```

After 3 FIX iterations with unresolved CRITICAL/HIGH issues, exit with `STATUS: NEEDS_BRAINSTORM` — the parent workflow re-brainstorms.

---

## Orchestrator Boundaries

**You ONLY:**
- ✅ Spawn subagents via Task tool (always `run_in_background: true`)
- ✅ Read `.bob/state/*.md` files to make routing decisions
- ✅ Write `.bob/state/*-prompt.md` instruction files for subagents
- ✅ Write `.bob/state/code-review-status.md` (exit signal for parent workflow)
- ✅ Run `git diff --name-only HEAD` or `git status --short` to scope reviews

**You NEVER:**
- ❌ Write or edit source code files
- ❌ Run `git commit`, `git push`, `gh pr create`
- ❌ Run tests, linters, or build commands
- ❌ Make implementation or architectural decisions
- ❌ Ask the user permission to proceed (run autonomously until COMPLETE)

---

## Execution Rules

**All subagents MUST run in background:**
- ✅ Always `run_in_background: true` for ALL Task calls
- ✅ After spawning, STOP — wait for completion notification
- ❌ Never foreground execution

---

## Phase 1: INIT

**Goal:** Establish context and scope the review.

**Actions:**

1. Create state directory if needed:
   ```bash
   mkdir -p .bob/state
   ```

2. Get the list of changed files to scope the review:
   ```bash
   git diff --name-only HEAD
   git status --short
   ```

3. Initialize loop counter to 0 (track in memory).

4. Move to REVIEW phase.

---

## Phase 2: REVIEW

**Goal:** Run comprehensive multi-domain code review over changed files, including a Go-specific pre-submit pass.

**Actions:**

1. Write `.bob/state/review-prompt.md` with scoping context:

   ```markdown
   ## Review Scope

   Changed files (from git diff):
   [list from INIT]

   Focus review on these files. For unchanged files, only flag issues if
   the changed code introduces a problem in them (e.g., a call site now
   passes wrong types).

   Context: [read .bob/state/plan.md and .bob/state/brainstorm.md if they exist]
   ```

2. Spawn **both reviewers in parallel** (start both before waiting for either):

   ```
   Task(subagent_type: "review-consolidator",
        description: "Multi-domain code review",
        run_in_background: true,
        prompt: "Review the current code changes. Read .bob/state/review-prompt.md
                for scope and context. Perform all review passes. Write consolidated
                report to .bob/state/review.md.")

   Task(subagent_type: "go-presubmit-reviewer",
        description: "Go pre-submit checklist review",
        run_in_background: true,
        prompt: "Run the Go pre-submit checklist on the current code changes.
                Read .bob/state/review-prompt.md for scope. Check all categories:
                pool lifetimes, concurrency races, int64/int type boundaries,
                error handling, spec accuracy, test quality, and I/O patterns.
                Write findings to .bob/state/go-presubmit.md.")
   ```

3. Wait for both agents to complete. Then read **both** `.bob/state/review.md` and `.bob/state/go-presubmit.md` and move to ROUTE.

---

## Phase 3: ROUTE

**Goal:** Decide next phase based on combined review findings.

**Read BOTH** `.bob/state/review.md` AND `.bob/state/go-presubmit.md` and sum the counts:
- CRITICAL count (review.md + go-presubmit.md)
- HIGH count (review.md + go-presubmit.md)
- MEDIUM count
- LOW count

**Routing logic:**

| Situation | Action |
|-----------|--------|
| No issues in either report | → COMMIT |
| MEDIUM/LOW only (across both) | → FIX (loop iteration +1) |
| CRITICAL or HIGH present, loop < 3 | → FIX (loop iteration +1) |
| CRITICAL or HIGH present, loop ≥ 3 | → EXIT with NEEDS_BRAINSTORM |
| Loop complete, only MEDIUM/LOW remain | → COMMIT (acceptable) |

---

## Phase 4: FIX

**Goal:** Fix issues identified in review.

**Actions:**

1. Write `.bob/state/fix-prompt.md`:

   ```markdown
   # Fix Review Issues (Iteration [N])

   ## Issues to Fix

   Read the full review at .bob/state/review.md AND the Go pre-submit findings
   at .bob/state/go-presubmit.md. Both reports contribute issues to fix.

   Fix ALL issues of CRITICAL and HIGH severity first (from either report).
   Then fix MEDIUM and LOW issues unless they require architectural changes
   (skip those and note them).

   ## Go Coding Guidelines

   When fixing Go-specific issues, apply the patterns from `/bob:internal:go-coding`:
   - Pool/resource lifetime: release at true end-of-life, not end-of-call
   - File writes: os.CreateTemp + os.Rename, never deterministic .tmp paths
   - Goroutine fan-out: always use errgroup.SetLimit or a semaphore
   - int64 sizes: convert to int with bounds check before make() or slice index
   - Cache errors: only os.IsNotExist is a miss; surface other errors
   - Spec files: update SPECS.md/NOTES.md in the same commit as the fix

   ## Constraints
   - Do NOT rewrite code that is not related to a reported issue
   - Do NOT introduce new functionality
   - Do NOT change public API unless required to fix an issue
   - Spec-driven modules: if fixing code in a module with SPECS.md or CLAUDE.md,
     update those docs if your fix changes a contract or invariant

   ## Changed Files (for context)
   [list from INIT phase]
   ```

2. Spawn workflow-coder:
   ```
   Task(subagent_type: "workflow-coder",
        description: "Fix review issues (iteration [N])",
        run_in_background: true,
        prompt: "Fix the issues described in .bob/state/fix-prompt.md.
                The issues come from .bob/state/review.md.
                Do not use .bob/state/plan.md as your guide here — use review.md.
                Write your status to .bob/state/implementation-status.md when done.")
   ```

3. After completion, move to TEST.

---

## Phase 5: TEST

**Goal:** Verify fixes don't break anything.

**Actions:**

Spawn workflow-tester:
```
Task(subagent_type: "workflow-tester",
     description: "Run tests after review fixes",
     run_in_background: true,
     prompt: "Run the full test suite and quality checks after code review fixes.

             IMPORTANT: Report findings objectively — do NOT make pass/fail
             determinations. The orchestrator makes routing decisions.

             Steps:
             1. Run `make ci` if available; otherwise run individually:
                - go test ./...
                - go test -race ./...
                - go test -cover ./...
                - go fmt ./... (check for formatting issues)
                - golangci-lint run (if installed)
                - gocyclo -over 40 . (if installed)

             For each step report: WHAT ran, WHAT failed, WHY it failed (error output),
             WHERE it failed (file:line, test name).

             Write all results to .bob/state/test-results.md.")
```

After completion, read `.bob/state/test-results.md`:
- If all tests pass → go back to REVIEW (re-review to confirm fixes resolved issues)
- If tests fail → go back to FIX with test failure details added to fix-prompt.md

When looping back to REVIEW after a successful TEST, update `.bob/state/review-prompt.md` to note this is a re-review after fixes, so the consolidator focuses on whether the issues were resolved (not a full re-scan from scratch).

---

## Phase 6: COMMIT

**Goal:** Commit the reviewed and fixed code.

**Actions:**

1. Write `.bob/state/commit-prompt.md`:
   ```markdown
   # Commit Instructions

   Commit all changes that were made during this code review cycle.

   Context:
   - Review findings: .bob/state/review.md
   - Implementation details: .bob/state/implementation-status.md (if exists)
   - Original plan: .bob/state/plan.md (if exists)

   Commit message guidance:
   - Summarize what was fixed based on the review findings
   - If this was a new feature + review fixes, lead with the feature
   - Include brief note on issues addressed
   ```

2. Spawn commit-agent:
   ```
   Task(subagent_type: "commit-agent",
        description: "Commit reviewed code",
        run_in_background: true,
        prompt: "Read .bob/state/commit-prompt.md for instructions.
                Create commit, push branch, create PR.
                Write status to .bob/state/commit.md.")
   ```

3. After completion, read `.bob/state/commit.md`:
   - STATUS: SUCCESS → move to MONITOR
   - STATUS: FAILED → report failure and exit with STATUS: FAILED

---

## Phase 7: MONITOR

**Goal:** Watch CI and PR feedback.

**Actions:**

1. Write `.bob/state/monitor-prompt.md`:
   ```markdown
   # Monitor Instructions

   Monitor the PR that was just created.

   PR details: see .bob/state/commit.md for the PR URL.

   Check:
   - CI/CD check status
   - Review comments or change requests
   - Whether PR is ready to merge

   Write status to .bob/state/monitor.md with routing decision.
   ```

2. Spawn monitor-agent:
   ```
   Task(subagent_type: "monitor-agent",
        description: "Monitor CI and PR status",
        run_in_background: true,
        prompt: "Read .bob/state/monitor-prompt.md for instructions.
                Check PR status, CI checks, review feedback.
                Write full status report to .bob/state/monitor.md.")
   ```

3. After completion, read `.bob/state/monitor.md` and route:

   | Monitor Status | Action |
   |----------------|--------|
   | READY (all green) | → COMPLETE |
   | WAITING (CI in progress) | → Re-spawn monitor-agent after a pause |
   | NEEDS_WORK (CI failed or changes requested) | → Exit with NEEDS_BRAINSTORM |

---

## Phase 8: COMPLETE

**Goal:** Signal completion to parent workflow.

Write `.bob/state/code-review-status.md`:

```markdown
# Code Review Status

Status: COMPLETE
Timestamp: [ISO timestamp]

## Summary
- Review iterations: [N]
- Issues found: [CRITICAL: N, HIGH: N, MEDIUM: N, LOW: N]
- Issues resolved: [N]
- PR: [URL from commit.md]
- CI: [status from monitor.md]

## Next Phase
COMPLETE — all checks passing, code reviewed and committed.
```

---

## Exit States

The parent workflow reads `.bob/state/code-review-status.md` to determine routing.

### COMPLETE
All issues resolved, commit created, CI passing.

### NEEDS_BRAINSTORM
Write `.bob/state/code-review-status.md`:
```markdown
# Code Review Status

Status: NEEDS_BRAINSTORM
Timestamp: [ISO timestamp]

## Reason
[One of:]
- Unresolved CRITICAL/HIGH issues after 3 fix iterations
- CI failures that indicate a design problem
- PR reviewer requested architectural changes

## Issues Requiring Re-Design
[List CRITICAL/HIGH issues from .bob/state/review.md, or CI/reviewer feedback]

## Recommended Focus for Brainstorm
[Brief description of what needs to be re-thought]
```

### FAILED
Write `.bob/state/code-review-status.md`:
```markdown
# Code Review Status

Status: FAILED
Timestamp: [ISO timestamp]

## Reason
[What failed: commit-agent error, monitor unreachable, etc.]

## Details
[Error output]
```

---

## State Files Reference

| File | Written By | Purpose |
|------|-----------|---------|
| `.bob/state/review-prompt.md` | Orchestrator | Scoping instructions for both reviewers |
| `.bob/state/review.md` | review-consolidator | Multi-domain review findings |
| `.bob/state/go-presubmit.md` | go-presubmit-reviewer | Go-specific pre-submit findings |
| `.bob/state/fix-prompt.md` | Orchestrator | Fix instructions (references both review files) |
| `.bob/state/implementation-status.md` | workflow-coder chain | What was fixed |
| `.bob/state/test-results.md` | workflow-tester | Test run results |
| `.bob/state/commit-prompt.md` | Orchestrator | Commit instructions |
| `.bob/state/commit.md` | commit-agent | Commit/PR status |
| `.bob/state/monitor-prompt.md` | Orchestrator | Monitor instructions |
| `.bob/state/monitor.md` | monitor-agent | CI/PR status |
| `.bob/state/code-review-status.md` | Orchestrator | Exit signal for parent |
