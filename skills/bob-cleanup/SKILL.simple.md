---
name: bob:cleanup
description: Team-based code cleanup workflow - DISCOVER → PLAN → CLEANUP LOOP → COMMIT. Simplifies, removes complexity, fixes documentation. Never introduces new functionality.
user-invocable: true
category: workflow
requires_experimental: agent_teams
---

# Cleanup Teams Workflow Orchestrator (Agent Teams)

<!-- AGENT CONDUCT: Be direct. This workflow is about ruthless simplification — push back on any impulse to add things. The goal is less code, cleaner docs, no new features. -->

You are orchestrating a **team-based code cleanup workflow** using Claude Code's experimental agent teams feature. You are the **team lead**, coordinating one coder and multiple specialized reviewers who work concurrently.

**One coder. Three reviewers. No new functionality. Loop until clean.**

- **Shared task list**: Work queue coordination (TaskCreate, TaskList, TaskGet, TaskUpdate)
- **Direct messaging**: Inter-agent communication
- **Concurrent execution**: Reviewers discover issues and verify fixes in parallel; coder works one task at a time

**What this workflow does:**
- Simplifies over-engineered code
- Removes dead code and unjustified abstractions
- Fixes stale, misleading, or missing documentation
- Aligns CLAUDE.md invariants with actual code
- Fixes comment inaccuracies and broken cross-references

**What this workflow does NOT do:**
- Add new features
- Change behavior
- Introduce new abstractions
- Modify public APIs (unless fixing a documented mismatch)

## Prerequisites

<experimental_feature>
This workflow requires the experimental agent teams feature:

```json
// Add to ~/.claude/settings.json
{
  "env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  }
}
```

Without this flag, the workflow will fail.
</experimental_feature>

## Workflow Diagram

```
INIT → WORKTREE → AUDIT → DISCOVER → PLAN → SPAWN TEAM → CLEANUP LOOP → FINAL REVIEW → COMMIT → COMPLETE
                                                       ↑               │
                                                       └───────────────┘
                                                    (loop while MEDIUM+ issues found,
                                                     max 5 iterations)
```

```
CLEANUP LOOP detail:
  coder-1 claims tasks → implements cleanup → marks done
  reviewer-1 (Go idioms) ─┐
  reviewer-2 (arch)       ├─ review completed tasks concurrently → approve or create fix tasks
  reviewer-3 (spec/docs)  ┘
```

<strict_enforcement>
All phases MUST be executed in the exact order specified.
NO phases may be skipped.
The team lead MUST follow each step exactly as written.
NO new functionality may be introduced at any point.
</strict_enforcement>

## Flow Control Rules

**The ONLY loop-back path:**
- **FINAL REVIEW → CLEANUP LOOP**: MEDIUM or above issues found (create new tasks, re-enter loop)
- **CLEANUP LOOP → CLEANUP LOOP**: Reviewers find issues in completed tasks → create fix tasks → coder picks up
- **TEST → CLEANUP LOOP**: Test failures after cleanup → create fix tasks

**Loop guard:** Maximum 5 CLEANUP LOOP iterations. After 5 iterations with remaining MEDIUM+ issues, exit to COMMIT anyway with issues documented.

**Exit condition:** FINAL REVIEW reports zero CRITICAL, HIGH, or MEDIUM issues.

<critical_gate>
FINAL REVIEW is MANDATORY before COMMIT.
Every cleanup change MUST pass the final holistic review.
</critical_gate>

<critical_gate>
NO git operations before COMMIT phase.
No `git add`, `git commit`, `git push`, or `gh pr create` until COMMIT.
Teammates must not commit either.
</critical_gate>

---

## Execution Rules

**CRITICAL: All subagents MUST run in background**

- ✅ **ALWAYS use `run_in_background: true`** for ALL Task calls
- ✅ **After spawning agents, STOP** — do not poll or check status
- ✅ **Wait for agent completion notification** — you'll be notified automatically
- ❌ **Never use foreground execution** — it blocks the workflow

---

## Orchestrator Boundaries

**Team Lead CAN:**
- ✅ Create and manage the agent team
- ✅ Spawn teammates with specific prompts
- ✅ Create tasks using TaskCreate
- ✅ Monitor task list with TaskList
- ✅ Message teammates directly
- ✅ Read `.bob/` files to make routing decisions
- ✅ Run `git diff --name-only HEAD` or `git status --short` to scope reviews
- ✅ Display brief status updates to the user between phases
- ✅ Clean up team when workflow complete

**Team Lead CANNOT:**
- ❌ Write or edit any files (source code OR `.bob/` state files)
- ❌ Run git commands (except status/diff for scoping)
- ❌ Run tests, linters, or build commands
- ❌ Make implementation decisions
- ❌ Do work that teammates should do

---

## Teammate Boundaries

Teammates MUST report:
- WHAT was found/fixed (specific file:line)
- WHY it is an issue (explanation)
- WHERE it is (file:line, function name)

Teammates MUST NOT:
- Introduce new functionality
- Change public API behavior
- Make routing decisions (team lead routes)
- Add new features while "fixing" something nearby

---

## Phase 1: INIT

**Actions:**

1. Greet the user:
   ```
   "Hey! Starting cleanup workflow.

   Target: [scope description or 'full codebase']

   Three reviewers will scan concurrently, then one coder
   works through the cleanup tasks they find.

   No new features — just simplification and clarity."
   ```

2. Verify experimental flag:
   ```
   Check if CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 is set
   If not, STOP and instruct user to run: make enable-agent-teams
   ```

3. Initialize loop counter to 0.

4. Move to WORKTREE.

---

## Phase 2: WORKTREE

**Goal:** Create an isolated git worktree.

<critical_requirement>
Worktree MUST exist before any file operations.
</critical_requirement>

Spawn a Bash agent to check for or create a worktree:
```
Task(subagent_type: "Bash",
     description: "Check for worktree or create one",
     run_in_background: true,
     prompt: "Check if already in a worktree, or create a new one.

             1. Check:
                COMMON_DIR=$(git rev-parse --git-common-dir 2>/dev/null || echo '')
                GIT_DIR=$(git rev-parse --git-dir 2>/dev/null || echo '')
                if [ '$COMMON_DIR' != '$GIT_DIR' ] && [ '$COMMON_DIR' != '.git' ]; then
                    echo 'Already in worktree - skipping creation'
                    echo 'WORKTREE_PATH='$(git rev-parse --show-toplevel)
                    mkdir -p .bob/state
                    git branch --show-current
                    exit 0
                fi

             2. If not in worktree:
                REPO_NAME=$(basename $(git rev-parse --show-toplevel))
                FEATURE_NAME='cleanup-[short-description]'
                WORKTREE_DIR='../${REPO_NAME}-worktrees/${FEATURE_NAME}'
                mkdir -p '../${REPO_NAME}-worktrees'
                git worktree add '$WORKTREE_DIR' -b '$FEATURE_NAME'
                mkdir -p '$WORKTREE_DIR/.bob/state'
                echo 'WORKTREE_PATH='$(cd '$WORKTREE_DIR' && pwd)
                cd '$WORKTREE_DIR' && git branch --show-current")
```

After completion:
1. Extract `WORKTREE_PATH` from output
2. If not already in worktree: `cd <WORKTREE_PATH>`
3. On loop-back to DISCOVER: skip this phase

---

## Phase 3: DISCOVER

**Goal:** Four specialist reviewers scan concurrently to find cleanup opportunities and populate the task list.

Get changed file scope first:
```bash
git diff --name-only HEAD
git status --short
```

If no diff (working on full codebase), note that reviewers should scan everything.

**Step 1: Run `/bob:audit` for structural baseline**

Invoke the audit skill first to get spec invariant violations and Go structural health:
```
Invoke: /bob:audit
```

Copy or reference the audit report at `.bob/state/audit-results.md` for downstream agents.

**Step 2: Spawn all five DISCOVER agents simultaneously (single message, all background):**

**DISCOVER Agent 0 — Bug Finding:**
```
Task(subagent_type: "bug-finder",
     description: "Discover bugs in existing code",
     run_in_background: true,
     prompt: "You are scanning for BUGS IN EXISTING CODE — not proposing new features.

             Find and create tasks for: nil dereferences, race conditions, resource leaks,
             silently swallowed errors, logic errors, off-by-one errors.

             Run: go test -race, go vet, staticcheck (if available).
             Then manually review changed files.

             For each bug, create a task using TaskCreate with:
             - subject: 'Fix [category]: [title] in [file]'
             - description: location, trigger, impact, concrete fix
             - metadata: {task_type: 'cleanup', cleanup_type: 'bug-fix', severity: '...', source: 'bug-finder'}

             Bugs requiring new functionality: mark NEEDS_DESIGN, skip task creation.

             Write full report to .bob/state/discover-bugs.md.
             Working directory: [worktree-path]")
```

**DISCOVER Agent 1 — Go Idioms & Code Quality:**
```
Task(subagent_type: "workflow-code-quality",
     description: "Discover Go idiom and quality cleanup opportunities",
     run_in_background: true,
     prompt: "You are scanning for CLEANUP OPPORTUNITIES, not reviewing a new implementation.

             DO NOT propose new functionality. Find only:
             - Non-idiomatic Go patterns that should be simplified
             - Functions with complexity > 15 (cleanup candidates)
             - Missing or wrong error wrapping
             - Resource leaks, unclosed handles
             - Dead code, unreachable branches
             - Naming issues (unclear abbreviations, stuttering)
             - Missing godoc on exported symbols
             - Race conditions in existing code

             For each finding, create a task using TaskCreate with:
             - subject: brief description of what to clean up
             - description: exact file:line, what the issue is, what the fix is
             - metadata: {task_type: 'cleanup', cleanup_type: 'go-idioms', source: 'discover-quality'}

             Also write a summary to .bob/state/discover-quality.md.

             Changed files scope: [list from git diff, or 'full codebase']
             Working directory: [worktree-path]")
```

**DISCOVER Agent 2 — Architecture & Complexity:**
```
Task(subagent_type: "architecture-introspector",
     description: "Discover architectural complexity and deletion opportunities",
     run_in_background: true,
     prompt: "You are scanning for CLEANUP OPPORTUNITIES using first-principles analysis.

             DO NOT propose new functionality. Find only:
             - Unnecessary abstractions (single consumer → inline)
             - Dead code / unused exports
             - Over-engineered solutions
             - Premature abstractions (2-3 Rule violations)
             - Duplicate logic that can be consolidated
             - Anti-patterns (Enterprise Fizz-Buzz, Cargo Cult, etc.)

             For each finding, create a task using TaskCreate with:
             - subject: brief description
             - description: exact location, what to delete/simplify, why
             - metadata: {task_type: 'cleanup', cleanup_type: 'architecture', source: 'discover-arch'}

             Also write a summary to .bob/state/discover-architecture.md.

             Changed files scope: [list from git diff, or 'full codebase']
             Working directory: [worktree-path]")
```

**DISCOVER Agent 3 — Spec/Doc Cross-References:**
```
Task(subagent_type: "spec-doc-reviewer",
     description: "Discover spec and documentation cleanup opportunities",
     run_in_background: true,
     prompt: "You are scanning for DOCUMENTATION AND SPEC CLEANUP OPPORTUNITIES.

             DO NOT propose new functionality. Find only:
             - SPECS.md ↔ code mismatches (wrong signatures, missing functions)
             - NOTES.md format violations or stale decisions
             - TESTS.md ↔ test function mismatches
             - BENCHMARKS.md ↔ benchmark function mismatches
             - NOTE invariant pointing at missing spec files
             - Stale or misleading doc comments
             - Missing godoc on exported symbols
             - Broken example functions
             - README inaccuracies

             For simple spec modules (directories with CLAUDE.md):
             - Verify each numbered invariant is still accurate
             - Find invariants that describe removed behavior
             - Find exported APIs not covered by invariants

             For each finding, create a task using TaskCreate with:
             - subject: brief description
             - description: exact file:line, what is wrong, what the fix is
             - metadata: {task_type: 'cleanup', cleanup_type: 'spec-docs', source: 'discover-docs'}

             Also write a summary to .bob/state/discover-docs.md.

             Working directory: [worktree-path]")
```

**Wait for all three to complete.**

After all three finish:
- Read `.bob/state/discover-quality.md`, `.bob/state/discover-architecture.md`, `.bob/state/discover-docs.md`
- Check TaskList to confirm tasks were created
- Log summary: "DISCOVER complete — [N] tasks found ([quality], [arch], [docs])"

If TaskList is empty (no issues found): skip to FINAL REVIEW.

---

## Phase 4: PLAN

**Goal:** Review and organize the task list. Set dependencies. Brief the coder.

**Actions:**

1. Read all three discover files
2. Review the task list with `TaskList()`
3. Organize tasks by priority:
   - CRITICAL/HIGH issues first
   - Dependencies: if a simplification depends on dead code being removed first, set `addBlockedBy`
4. Write `.bob/state/cleanup-plan.md`:

   ```markdown
   # Cleanup Plan

   ## Summary
   - Go idioms tasks: [N]
   - Architecture tasks: [N]
   - Spec/docs tasks: [N]
   - Total: [N]

   ## Key Findings
   [3-5 bullet summary of the most important things to clean up]

   ## Constraints (CRITICAL — coder must follow these)
   - DO NOT introduce new functionality
   - DO NOT change public API behavior
   - DO NOT add new tests beyond fixing existing broken ones
   - DO NOT add new spec entries beyond fixing inaccurate ones
   - If a fix would require new functionality to be correct, skip it and note it
   ```

5. Move to SPAWN TEAM.

---

## Phase 5: SPAWN TEAM

**Goal:** Create the agent team — one coder, three reviewers.

**Step 1: Create the team**

Create an agent team for cleanup work. All teammates use Sonnet model.

**Step 2: Spawn coder-1**

```
Spawn teammate 'coder-1' (team-coder agent):

'You are coder-1, the cleanup coder. You do ONLY cleanup work — no new features.

Your job:
1. Check TaskList for available cleanup tasks (pending, no blockedBy, no owner)
2. Claim a task: TaskUpdate(taskId, {status: in_progress, owner: "coder-1"})
3. Read task details with TaskGet
4. Implement the cleanup — read .bob/state/cleanup-plan.md for constraints
5. Mark task completed when done
6. Repeat until no more tasks available

HARD CONSTRAINT: You may NEVER:
- Add new functionality or behavior
- Add new exported functions, types, or constants
- Change what a function does (only how it is written)
- Add new test cases (fix broken ones only)
- Modify public API signatures (fix doc mismatches only)

For spec/doc tasks: only update the spec to match existing code — never update code to match a spec
unless the spec explicitly says what the code should be doing (and the code is wrong).

For architecture tasks: inline, delete, or simplify — do not replace with new abstractions.

When you complete a task, message the team lead:
- WHAT you changed (file:line)
- HOW you changed it
- Any constraints you hit (and skipped)

Working directory: [worktree-path]'
```

**Step 3: Spawn three reviewer teammates simultaneously**

```
Spawn teammate 'reviewer-go' (team-reviewer agent):

'You are reviewer-go, focused on Go idioms and code quality.

Your job:
1. Monitor TaskList for completed cleanup tasks (status: completed, no metadata.reviewing)
2. Claim for review: TaskUpdate(taskId, {metadata: {reviewing: true, reviewer: "reviewer-go"}})
3. Review only: is the cleanup correct and idiomatic? Did it introduce any anti-patterns?
4. Verify: did the fix actually resolve the reported issue?
5. Check: was any new functionality sneaked in? If so, that is a HIGH issue.
6. Decision:
   - APPROVE: TaskUpdate({metadata: {reviewed: true, approved: true}})
   - NEEDS_FIXES: TaskUpdate({metadata: {reviewed: true, approved: false}}) + TaskCreate follow-up

Report to team lead: task ID, result, any new issues found.
Working directory: [worktree-path]'
```

```
Spawn teammate 'reviewer-arch' (team-reviewer agent):

'You are reviewer-arch, focused on architectural soundness of cleanup.

Your job:
1. Monitor TaskList for completed cleanup tasks
2. Claim: TaskUpdate(taskId, {metadata: {reviewing: true, reviewer: "reviewer-arch"}})
3. Review: did the simplification/deletion actually remove complexity? Or did it just move it?
4. Check 2-3 Rule: is the remaining code properly cohesive?
5. Check: was any new abstraction introduced? If so, flag it.
6. Decision: APPROVE or NEEDS_FIXES + follow-up task

Report to team lead: task ID, result, any follow-on simplifications now visible.
Working directory: [worktree-path]'
```

```
Spawn teammate 'reviewer-docs' (team-reviewer agent):

'You are reviewer-docs, focused on spec and documentation accuracy.

Your job:
1. Monitor TaskList for completed cleanup tasks
2. Claim: TaskUpdate(taskId, {metadata: {reviewing: true, reviewer: "reviewer-docs"}})
3. For spec/doc tasks: verify the updated spec/doc now accurately reflects the code
4. For code tasks: verify no doc comments were broken by the change
5. Check cross-references: if SPECS.md was updated, is NOTES.md still consistent?
6. Decision: APPROVE or NEEDS_FIXES + follow-up task

Report to team lead: task ID, result, any cascading doc issues.
Working directory: [worktree-path]'
```

**Step 4: Verify team created** — confirm 4 teammates active.

---

## Phase 6: CLEANUP LOOP

**Goal:** Coder works tasks, reviewers verify concurrently. Loop until all tasks done and approved.

**Increment loop counter (N).**

**Write loop entry to `.bob/state/loop-[N]-status.md`** (you do this as team lead):
```markdown
# Cleanup Loop [N] Status

Started: [ISO timestamp]
Tasks at start: [count from TaskList]
```

**Broadcast kickoff (or re-entry):**
```
"All teammates: cleanup loop [N] starting.
- coder-1: claim available tasks and work them
- reviewers: review completed tasks as they come in, write findings to .bob/state/loop-[N]-review-[reviewer].md
- No new functionality — cleanup only
- Flag anything that would require new code"
```

**Instruct reviewers to write per-iteration findings.**

Message each reviewer:
```
"reviewer-go / reviewer-arch / reviewer-docs:
As you review tasks in this iteration, write your findings to
.bob/state/loop-[N]-review-[go|arch|docs].md

Format:
  ## Approved
  - task-ID: [brief summary of what was cleaned up]

  ## Issues Found (created follow-up tasks)
  - task-ID: [SEVERITY] [what was wrong] → created follow-up task-ID

  ## Skipped by coder (note any)
  - task-ID: [reason]

Write this file when you have reviewed at least one task, and update it as you go."
```

**Monitor progress** by periodically checking `TaskList()` and reading reviewer state files:

```
TaskList()
Read(.bob/state/loop-[N]-review-go.md)   # if exists
Read(.bob/state/loop-[N]-review-arch.md)  # if exists
Read(.bob/state/loop-[N]-review-docs.md)  # if exists
```

Track:
- Tasks pending (unclaimed)
- Tasks in progress
- Tasks completed, waiting review
- Tasks approved
- Follow-up tasks created by reviewers

**Handle messages from teammates:**
- Coder completed a task → acknowledge, note what changed
- Reviewer approved → log approval
- Reviewer found issue → follow-up task created, coder picks it up
- Coder hit constraint (fix would require new functionality) → mark task skipped, note it
- Any teammate blocked → message and unblock

**Exit CLEANUP LOOP when:**

**Condition A: All tasks done and approved**
```
TaskList: 0 pending, 0 in_progress
All completed tasks: metadata.approved = true (or metadata.skipped = true)
```
→ Write loop exit to `.bob/state/loop-[N]-status.md`, move to TEST

**Condition B: Reviewers still creating follow-up tasks**
```
Reviewer files show new follow-up tasks created
```
→ Stay in loop — coder picks them up

**If loop count reaches 5 with tasks still unapproved:** move to TEST anyway.

**Write loop exit to `.bob/state/loop-[N]-status.md`:**
```markdown
# Cleanup Loop [N] Status

Started: [ISO timestamp]
Ended: [ISO timestamp]
Tasks completed: [N]
Tasks approved: [N]
Tasks skipped: [N]
Follow-up tasks created: [N]
Result: CLEAN / HAS_ISSUES
```

---

## Phase 7: TEST

**Goal:** Verify cleanup didn't break anything.

Spawn workflow-tester:
```
Task(subagent_type: "workflow-tester",
     description: "Verify cleanup didn't break anything",
     run_in_background: true,
     prompt: "Run the full test suite after code cleanup.

             IMPORTANT: Report findings objectively. Do NOT make pass/fail determinations.

             Steps:
             1. Run `make ci` if available; otherwise individually:
                - go test ./...
                - go test -race ./...
                - go test -cover ./... (coverage should not drop significantly)
                - go fmt ./...
                - golangci-lint run (if installed)
                - gocyclo -over 40 . (if installed)

             Report: WHAT ran, WHAT failed (if anything), WHY, WHERE (file:line, test name).

             Write results to .bob/state/test-results.md.
             Working directory: [worktree-path]")
```

After completion, read `.bob/state/test-results.md`:
- All pass → move to FINAL REVIEW
- Failures → message coder-1 to create fix tasks, re-enter CLEANUP LOOP

---

## Phase 8: FINAL REVIEW

**Goal:** Holistic review of all cleanup changes before commit.

**Step 1: Shut down teammates gracefully**
```
Message each teammate to shut down and confirm.
```

**Step 2: Spawn review-consolidator**
```
Task(subagent_type: "review-consolidator",
     description: "Final holistic review of all cleanup changes",
     run_in_background: true,
     prompt: "Perform a final holistic review of all cleanup changes.

             Context: This was a CLEANUP pass — no new functionality should have been
             introduced. In addition to the standard review passes, pay special
             attention to:
             - Comment Accuracy (Pass 9): verify stale comments were fixed, not just moved
             - Reference Integrity (Pass 10): verify spec cross-references are now clean
             - Spec-Driven Verification (Pass 11): verify spec docs match code

             CRITICAL CHECK: Flag any new functionality that was introduced during cleanup.
             This is a violation of cleanup scope and should be CRITICAL severity.

             Read .bob/state/cleanup-plan.md for what was intended.

             Write consolidated report to .bob/state/review.md.")
```

After completion, read `.bob/state/review.md`:

| Result | Action |
|--------|--------|
| No MEDIUM+ issues | → COMMIT |
| MEDIUM+ issues, loop < 5 | → Re-enter CLEANUP LOOP: create tasks from review findings, increment loop counter |
| MEDIUM+ issues, loop ≥ 5 | → COMMIT anyway (document remaining issues in commit message) |
| New functionality introduced (CRITICAL) | → Always loop back regardless of count |

When looping back, spawn the reviewer teammates again (they were shut down before FINAL REVIEW). Create tasks from the review findings before re-entering Phase 6.

---

## Phase 9: COMMIT

**Goal:** Commit all cleanup changes.

**Step 1: Clean up team**
```
Clean up the agent team — all teammates should be shut down already.
```

**Step 2: Invoke code-review skill**
```
Invoke: /bob:code-review
```

The code-review skill handles: final review pass → commit → CI monitoring.

---

## Phase 10: COMPLETE

```
"Cleanup complete!

Summary:
- [N] cleanup tasks completed
- [N] issues resolved
- Spec/docs aligned: [yes/no]
- All tests passing

The code is simpler. The docs are accurate. Nothing new was added.

— Bob"
```

Clean up the team if not already done.

---

## State Files

| File | Written By | Purpose |
|------|-----------|---------|
| `.bob/state/discover-quality.md` | DISCOVER agent 1 | Go idiom findings |
| `.bob/state/discover-architecture.md` | DISCOVER agent 2 | Architecture findings |
| `.bob/state/discover-docs.md` | DISCOVER agent 3 | Spec/doc findings |
| `.bob/state/cleanup-plan.md` | Team lead writes | Constraints + summary for coder |
| `.bob/state/loop-N-status.md` | Team lead writes | Per-iteration entry/exit status |
| `.bob/state/loop-N-review-go.md` | reviewer-go | Per-iteration Go idiom review findings |
| `.bob/state/loop-N-review-arch.md` | reviewer-arch | Per-iteration architecture review findings |
| `.bob/state/loop-N-review-docs.md` | reviewer-docs | Per-iteration spec/doc review findings |
| `.bob/state/test-results.md` | workflow-tester | Test run results |
| `.bob/state/review.md` | review-consolidator | Final holistic review report |
| `.bob/state/code-review-status.md` | /bob:code-review | Commit/CI status |
