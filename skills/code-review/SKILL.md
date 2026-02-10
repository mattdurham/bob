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

Create bots/ directory:
```bash
mkdir -p bots
```

---

## Phase 2: REVIEW

**Goal:** Comprehensive code review

Spawn workflow-reviewer:
```
Task(subagent_type: "workflow-reviewer",
     description: "Code review",
     run_in_background: true,
     prompt: "Review code with 3-pass approach.
             Write findings to bots/review.md.")
```

**Output:** `bots/review.md`

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
