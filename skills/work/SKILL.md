---
name: bob:work
description: Full development workflow orchestrator - INIT → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMMIT → MONITOR
user-invocable: true
category: workflow
---

# Work Workflow Orchestrator

You are orchestrating a **full development workflow**. You coordinate specialized subagents via the Task tool to guide through complete feature development from idea to merged PR.


## Workflow Diagram

```
INIT → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMMIT → MONITOR → COMPLETE
          ↑                                      ↓               ↓
          └──────────────────────────────────────┴───────────────┘
                        (loop back on issues)
```

## Flow Control Rules

**Loop-back paths:**
- **REVIEW → BRAINSTORM**: Issues found during review require re-brainstorming
- **MONITOR → BRAINSTORM**: CI failures or PR feedback require re-brainstorming
- **TEST → EXECUTE**: Test failures require code fixes

**Never skip REVIEW** - Always review before commit, even if tests pass.

**NEVER commit or push before the COMMIT phase.** No `git add`, `git commit`, `git push`, or `gh pr create` until you reach Phase 7: COMMIT. Subagents must not commit either.

---

## Execution Rules

**CRITICAL: All subagents MUST run in background**

- ✅ **ALWAYS use `run_in_background: true`** for ALL Task calls
- ✅ **After spawning agents, STOP** - do not poll or check status
- ✅ **Wait for agent completion notification** - you'll be notified automatically
- ❌ **Never use foreground execution** - it blocks the workflow

**Example:**
```
Task(subagent_type: "any-agent",
     description: "Brief description",
     run_in_background: true,  // ← REQUIRED
     prompt: "Detailed instructions...")
```

**Why?** Background execution allows the workflow to continue and enables true parallelism when spawning multiple agents.

---

## Autonomous Progression Rules

**CRITICAL: Progress automatically between phases without user confirmation**

- ✅ **Automatically continue** from EXECUTE → TEST → REVIEW when agents complete successfully
- ✅ **Only prompt user** when there's a problem, decision needed, or at critical gates
- ❌ **Never ask "Should I continue to next phase?"** when the path is clear
- ❌ **Never ask "Do you want me to proceed?"** for standard workflow progression
- ❌ **NEVER skip REVIEW to go directly to COMMIT** — REVIEW must complete before COMMIT, no exceptions
- ❌ **NEVER proceed to COMMIT unless `.bob/state/review.md` exists** — this file is proof REVIEW ran

**When to ask user:**
- ❌ NOT between EXECUTE → TEST (automatic)
- ❌ NOT between TEST → REVIEW (automatic if tests pass)
- ❌ NOT between REVIEW → COMMIT (automatic if no issues)
- ✅ YES if agent fails or returns errors
- ✅ YES at REVIEW → BRAINSTORM/EXECUTE decision (show findings, explain routing)
- ✅ YES at COMPLETE phase (confirm merge)
- ✅ YES if loop-back occurs (explain why and what will be fixed)

**Standard flow without prompts:**
```
EXECUTE (agent completes) → TEST (tests pass) → REVIEW (all reviewers complete) → [decision]
```

**Flow with problems (prompts required):**
```
EXECUTE (agent fails) → [PROMPT: explain error, ask if should retry/modify]
TEST (tests fail) → [PROMPT: show failures, confirm loop to EXECUTE]
REVIEW (issues found) → [PROMPT: show findings, explain routing to BRAINSTORM/EXECUTE]
```

--=

## .bob/planning/ Context Integration

If a `.bob/planning/` directory exists (created by `/bob:project`), use it as persistent project context throughout the workflow:

**Check at INIT:**
```bash
ls .bob/planning/ 2>/dev/null
```

**If `.bob/planning/` exists, load context:**
- `.bob/planning/PROJECT.md` — Project vision, scope, technical decisions
- `.bob/planning/REQUIREMENTS.md` — Traceable requirements with REQ-IDs
- `.bob/planning/CODEBASE.md` — Existing code analysis (if brownfield)

**How to use it:**
- **INIT:** Read PROJECT.md to understand the project
- **BRAINSTORM:** Pass PROJECT.md + REQUIREMENTS.md context to the brainstorm agent so it understands the bigger picture
- **PLAN:** Tell the planner which REQ-IDs from REQUIREMENTS.md this work satisfies
- **REVIEW:** Reviewers can check implementation against acceptance criteria in REQUIREMENTS.md
- **COMMIT:** Reference REQ-IDs in commit messages and PR descriptions

If `.bob/planning/` does NOT exist, proceed normally — it's optional context.

---

## Phase 1: INIT

**Goal:** Initialize and understand requirements

**Actions:**
1. Greet user and understand what they want to build
2. Check for `.bob/planning/` directory — if it exists, read PROJECT.md and REQUIREMENTS.md for context
3. Move to BRAINSTORM phase

---

## Phase 2: BRAINSTORM

**Goal:** Gather information and explore approaches

**Actions:**

**Step 1: Use brainstorming skill for ideation**
```
Invoke: /brainstorming
Topic: [The feature/task to implement]
```

The brainstorming skill will help:
- Generate ideas and approaches
- Consider multiple perspectives
- Identify potential issues early
- Think through edge cases

**⚠️ Step 2: CREATE ISOLATED WORKTREE (REQUIRED BEFORE ANY FILE WRITES) ⚠️**

**CRITICAL: You MUST create a git worktree BEFORE writing any design documents or doing research.**
**This ensures all work is isolated and doesn't interfere with the main branch.**

Create a git worktree for isolated development:

```bash
# Get repo name and create feature name
REPO_NAME=$(basename $(git rev-parse --show-toplevel))
FEATURE_NAME="<descriptive-feature-name>"  # e.g., "add-auth", "fix-parser-bug"

# Create worktree directory structure
WORKTREE_DIR="../${REPO_NAME}-worktrees/${FEATURE_NAME}"
mkdir -p "../${REPO_NAME}-worktrees"

# Create new branch and worktree
git worktree add "$WORKTREE_DIR" -b "$FEATURE_NAME"

# Change to worktree directory
cd "$WORKTREE_DIR"

# Create bots directory in worktree
mkdir -p .bob/state .bob/planning

# Verify we're in the worktree
pwd
git branch --show-current
```

**After worktree creation:**
- ✅ All subsequent file operations happen in the worktree
- ✅ Main branch remains clean
- ✅ Work is isolated and can be easily discarded if needed

**Step 3: Research existing patterns**

Now that you're in the worktree, spawn Explore agent for codebase research:
```
Task(subagent_type: "Explore",
     description: "Research similar implementations",
     run_in_background: true,
     prompt: "Search codebase for patterns related to [task].
             Find existing implementations, identify patterns to follow.
             [If .bob/planning/ exists: Read .bob/planning/PROJECT.md for project context
              and .bob/planning/REQUIREMENTS.md for the requirements being implemented.]
             Document findings.")
```

**Step 4: Document findings in the worktree**

Write consolidated findings to `.bob/state/brainstorm.md` (in the worktree):
- Requirements and constraints (reference REQ-IDs from `.bob/planning/REQUIREMENTS.md` if available)
- Existing patterns discovered
- Approaches considered (2-3 options with pros/cons)
- Recommended approach with rationale
- Risks and open questions

**Output:**
- Isolated worktree in `../<repo>-worktrees/<feature>/`
- `.bob/state/brainstorm.md` (in worktree)

---

## Phase 3: PLAN

**Goal:** Create detailed implementation plan

**Actions:**

Use the writing-plans skill to spawn a planner subagent:
```
Invoke: /writing-plans
```

The skill will:
1. Spawn workflow-planner subagent in background
2. Subagent reads design from `.bob/state/design.md` (or `.bob/state/brainstorm.md`)
3. If `.bob/planning/` exists, subagent also reads `.bob/planning/REQUIREMENTS.md` for acceptance criteria and `.bob/planning/PROJECT.md` for project context
4. Subagent creates concrete, bite-sized implementation plan
5. Subagent writes plan to `.bob/state/plan.md`

**Input:** `.bob/state/design.md` or `.bob/state/brainstorm.md` (+ `.bob/planning/REQUIREMENTS.md` if available)
**Output:** `.bob/state/plan.md`

Plan includes:
- Exact file paths
- Complete code snippets
- Step-by-step actions (2-5 min each)
- TDD approach (test first!)
- Verification steps

**If looping from REVIEW:** Update plan to address review findings

---

## Phase 4: EXECUTE

**Goal:** Implement the planned changes.

**CRITICAL: You are the orchestrator. You NEVER write code, edit files, or fix issues yourself. You ALWAYS spawn workflow-coder to do the work.**

**Actions:**

Spawn workflow-coder agent:
```
Task(subagent_type: "workflow-coder",
     description: "Implement feature",
     run_in_background: true,
     prompt: "Follow plan in .bob/state/plan.md.
             Use TDD: write tests first, verify they fail, then implement.
             Keep functions small (complexity < 40).
             Follow existing code patterns.
             Working directory: [worktree-path]")
```

**Input:** `.bob/state/plan.md`
**Output:** Code implementation

**After completion:**
- ✅ If agent succeeds → **Automatically proceed to TEST** (no prompt)
- ❌ If agent fails → Prompt user with error details

**If looping from TEST:** Spawn workflow-coder again with test failure details:
```
Task(subagent_type: "workflow-coder",
     description: "Fix test failures",
     run_in_background: true,
     prompt: "Tests failed. Read .bob/state/test-results.md for failure details.
             Fix the failing tests. Do not rewrite working code.
             Working directory: [worktree-path]")
```

**If looping from REVIEW (MEDIUM/LOW issues):** Spawn workflow-coder again with review findings:
```
Task(subagent_type: "workflow-coder",
     description: "Fix review issues",
     run_in_background: true,
     prompt: "Code review found issues. Read .bob/state/review.md for details.
             Fix only the MEDIUM and LOW severity issues listed.
             Do not rewrite working code — make targeted fixes only.
             Working directory: [worktree-path]")
```

---

## Phase 5: TEST

**Goal:** Run all tests and quality checks

**Actions:**

Spawn workflow-tester agent:
```
Task(subagent_type: "workflow-tester",
     description: "Run all tests and checks",
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

**Input:** Code to test
**Output:** `.bob/state/test-results.md`

Checks:
- All tests pass
- No race conditions
- Good coverage (>80%)
- Code formatted
- Linter clean
- Complexity < 40

**After completion:**
- ✅ If tests pass → **Automatically proceed to REVIEW** (no prompt)
- ❌ If tests fail → Prompt user with failures, confirm loop to EXECUTE

**If tests fail:** Loop to EXECUTE to fix issues

---

## Phase 6: REVIEW (Parallel Multi-Agent Review)

**Goal:** Comprehensive code review by 9 specialized agents in parallel

**Actions:**

Spawn 9 reviewer agents in parallel (single message, 9 Task calls):

```
Task(subagent_type: "workflow-reviewer",
     description: "Code quality review",
     run_in_background: true,
     prompt: "Perform 3-pass code review focusing on code logic, bugs, and best practices.
             Pass 1: Cross-file consistency
             Pass 2: Code quality and logic errors
             Pass 3: Best practices compliance
             Write findings to .bob/state/review-code.md with severity levels.")

Task(subagent_type: "security-reviewer",
     description: "Security vulnerability review",
     run_in_background: true,
     prompt: "Scan code for security vulnerabilities:
             - OWASP Top 10 (injection, XSS, CSRF, etc.)
             - Secret detection (API keys, passwords)
             - Authentication/authorization issues
             - Input validation gaps
             Write findings to .bob/state/review-security.md with severity levels.")

Task(subagent_type: "performance-analyzer",
     description: "Performance bottleneck review",
     run_in_background: true,
     prompt: "Analyze code for performance issues:
             - Algorithmic complexity (O(n²) opportunities)
             - Memory leaks and inefficient allocations
             - N+1 patterns and missing caching
             - Expensive operations in loops
             Write findings to .bob/state/review-performance.md with severity levels.")

Task(subagent_type: "docs-reviewer",
     description: "Documentation accuracy review",
     run_in_background: true,
     prompt: "Review documentation for accuracy and completeness:
             - README accuracy (features match implementation)
             - Example validity (code examples work)
             - API documentation alignment
             - Comment correctness
             Write findings to .bob/state/review-docs.md with severity levels.")

Task(subagent_type: "architect-reviewer",
     description: "Architecture and design review",
     run_in_background: true,
     prompt: "Evaluate system architecture and design decisions:
             - Design patterns appropriateness
             - Scalability assessment
             - Technology choices justification
             - Integration patterns validation
             - Technical debt analysis
             Write findings to .bob/state/review-architecture.md with severity levels.")

Task(subagent_type: "code-reviewer",
     description: "Comprehensive code quality review",
     run_in_background: true,
     prompt: "Conduct deep code review across all aspects:
             - Logic correctness and error handling
             - Code organization and readability
             - Security best practices
             - Performance optimization opportunities
             - Maintainability and test coverage
             Write findings to .bob/state/review-code-quality.md with severity levels.")

Task(subagent_type: "golang-pro",
     description: "Go-specific code review",
     run_in_background: true,
     prompt: "Review Go code for idiomatic patterns and best practices:
             - Idiomatic Go patterns (effective Go guidelines)
             - Concurrency patterns (goroutines, channels, context)
             - Error handling excellence
             - Performance and race condition analysis
             - Go-specific security concerns
             Write findings to .bob/state/review-go.md with severity levels.")

Task(subagent_type: "debugger",
     description: "Bug diagnosis and debugging review",
     run_in_background: true,
     prompt: "Perform systematic debugging analysis on the code:
             - Potential null pointer dereferences and panic conditions
             - Race conditions and concurrency bugs
             - Off-by-one errors and boundary conditions
             - Resource leaks (connections, file handles, memory)
             - Logic errors in control flow and state management
             - Error propagation and handling gaps
             Write findings to .bob/state/review-debug.md with severity levels.")

Task(subagent_type: "error-detective",
     description: "Error pattern analysis review",
     run_in_background: true,
     prompt: "Analyze code for error handling patterns and potential failure modes:
             - Error handling consistency across the codebase
             - Missing error checks and silent failures
             - Error message clarity and actionability
             - Retry logic and failure recovery patterns
             - Timeout and deadline handling
             - Circuit breaker and fallback patterns
             Write findings to .bob/state/review-errors.md with severity levels.")
```

**Wait for ALL 9 agents to complete.** If any agent fails, abort and report error.

**Input:** Code changes, `.bob/state/plan.md`
**Output:**
- `.bob/state/review-code.md` (code quality findings)
- `.bob/state/review-security.md` (security findings)
- `.bob/state/review-performance.md` (performance findings)
- `.bob/state/review-docs.md` (documentation findings)
- `.bob/state/review-architecture.md` (architecture findings)
- `.bob/state/review-code-quality.md` (comprehensive code quality findings)
- `.bob/state/review-go.md` (Go-specific findings)
- `.bob/state/review-debug.md` (debugging and bug diagnosis findings)
- `.bob/state/review-errors.md` (error handling pattern findings)
- `.bob/state/review.md` (consolidated report - created in next step)

**Step 2: Consolidate Findings**

After all 9 agents complete successfully:

1. **Read all 9 review files:**
   ```
   Read(file_path: "/path/to/worktree/.bob/state/review-code.md")
   Read(file_path: "/path/to/worktree/.bob/state/review-security.md")
   Read(file_path: "/path/to/worktree/.bob/state/review-performance.md")
   Read(file_path: "/path/to/worktree/.bob/state/review-docs.md")
   Read(file_path: "/path/to/worktree/.bob/state/review-architecture.md")
   Read(file_path: "/path/to/worktree/.bob/state/review-code-quality.md")
   Read(file_path: "/path/to/worktree/.bob/state/review-go.md")
   Read(file_path: "/path/to/worktree/.bob/state/review-debug.md")
   Read(file_path: "/path/to/worktree/.bob/state/review-errors.md")
   ```

2. **Parse and merge findings:**
   - Extract all issues from each file
   - Sort by severity (CRITICAL, HIGH, MEDIUM, LOW)
   - Deduplicate similar issues:
     - Same file:line → Merge descriptions
     - Keep highest severity
     - Note which agents found it

3. **Generate consolidated report:**
   Write to `.bob/state/review.md`:
   ```markdown
   # Consolidated Code Review Report

   ## Critical Issues (Must Fix Before Commit)

   ### Issue 1: SQL Injection in Login Handler
   **Severity:** CRITICAL
   **Category:** security
   **Found by:** security-reviewer, workflow-reviewer
   **Files:** auth/login.go:45
   **Description:** User input directly concatenated into SQL query...
   **Impact:** Database compromise possible
   **Fix:** Use parameterized queries

   ## High Priority Issues

   ### Issue 2: O(n²) Algorithm
   **Severity:** HIGH
   **Category:** performance
   **Found by:** workflow-performance-analyzer
   ...

   ## Medium Priority Issues

   [List MEDIUM severity issues]

   ## Low Priority Issues

   [List LOW severity issues]

   ## Summary

   **Total Issues:** 15
   - CRITICAL: 2 (security: 2, code: 0, performance: 0, docs: 0, architecture: 0, code-quality: 0, go: 0, debug: 0, errors: 0)
   - HIGH: 4 (security: 1, code: 1, performance: 1, docs: 0, architecture: 1, code-quality: 0, go: 0, debug: 0, errors: 0)
   - MEDIUM: 6 (security: 0, code: 2, performance: 1, docs: 2, architecture: 0, code-quality: 1, go: 0, debug: 0, errors: 0)
   - LOW: 3 (security: 0, code: 0, performance: 0, docs: 2, architecture: 0, code-quality: 0, go: 1, debug: 0, errors: 0)

   **Agents Executed:**
   - Code Quality Review: ✓ (workflow-reviewer)
   - Security Review: ✓ (security-reviewer)
   - Performance Review: ✓ (performance-analyzer)
   - Documentation Review: ✓ (docs-reviewer)
   - Architecture Review: ✓ (architect-reviewer)
   - Code Quality Deep Review: ✓ (code-reviewer)
   - Go-Specific Review: ✓ (golang-pro)
   - Debugging Review: ✓ (debugger)
   - Error Pattern Review: ✓ (error-detective)

   **Recommendation:** BRAINSTORM (2 CRITICAL issues require architectural review)
   ```

**Step 3: Severity-Based Routing**

Based on the consolidated findings:

```
if (any CRITICAL or HIGH issues):
  → Loop to BRAINSTORM
  Reason: Major issues require re-thinking approach

elif (any MEDIUM or LOW issues):
  → Loop to EXECUTE
  Reason: Quick fixes needed, no architectural changes

else:
  → Continue to COMMIT
  Reason: Clean code, ready to commit
```

**Routing Logic:**

1. **CRITICAL or HIGH issues → BRAINSTORM**
   - Security vulnerabilities
   - Major bugs or logic errors
   - Severe performance problems
   - Breaking API changes
   - These require re-thinking the approach

2. **MEDIUM or LOW issues → EXECUTE**
   - Minor bugs
   - Missing validation
   - Documentation issues
   - Style improvements
   - These are quick fixes

3. **No issues → COMMIT**
   - All checks passed
   - Code is ready

**After routing decision:**
- ✅ **Always show user the findings summary** (even if no issues)
- ✅ **Explain the routing decision** (why BRAINSTORM/EXECUTE/COMMIT)
- ✅ If no issues found → **Automatically proceed to COMMIT** (no confirmation needed)
- ✅ If issues found → **Explain what will be fixed** and automatically loop (no "should I continue?" prompt)

**Error Handling:**

- **Any agent fails:** Abort review, report to user, retry once
- **Empty results:** Valid (agent found no issues)
- **Consolidation fails:** Show individual files to user, ask for manual review

**Note:** This replaces the simple single-agent review with parallel multi-agent review (9 specialized agents) while maintaining backward compatibility (still produces `.bob/state/review.md`).


---

## Phase 7: COMMIT

**Goal:** Commit changes and create a PR

**This is the FIRST phase where git operations are allowed.**

**PREREQUISITE:** `.bob/state/review.md` MUST exist. If it does not, STOP and go back to REVIEW. Never commit unreviewed code.

**Actions:**

1. Verify review was completed:
   ```bash
   # HARD GATE — do not proceed if this file is missing
   test -f .bob/state/review.md || { echo "REVIEW not completed"; exit 1; }
   ```

2. Review changes:
   ```bash
   git status
   git diff
   ```

3. Show the user a summary of all changes and review findings.

4. Create PR (default — no need to ask):
   ```bash
   git add [relevant-files]
   git commit -m "..."
   git push -u origin $(git branch --show-current)
   gh pr create --title "Title" --body "Description"
   ```

---

## Phase 8: MONITOR

**Goal:** Monitor CI/PR checks and handle feedback

**Actions:**

1. Check CI status:
   ```bash
   gh pr checks --json name,status,conclusion
   ```

2. Monitor for:
   - CI check failures
   - PR review comments
   - Requested changes
   - Conversation threads

3. **If ANY issues/failures:**
   - **LOOP TO BRAINSTORM** (not REVIEW, not EXECUTE)
   - Review the issues
   - Re-brainstorm to address them
   - Then proceed through EXECUTE → TEST → REVIEW again

4. **If all checks pass and approved:**
   - Move to COMPLETE

**Critical:** MONITOR always loops to BRAINSTORM when issues found, ensuring proper re-brainstorming before fixes.

---

## Phase 9: COMPLETE

**Goal:** Workflow complete

**Actions:**

1. Confirm with user: "All checks passed. Ready to merge?"
2. If approved, merge PR:
   ```bash
   gh pr merge --squash
   ```
3. Celebrate! 🎉

---

## State Management (No Bob MCP)

Workflow state is maintained through:
- **.bob/state/*.md files** - Persistent artifacts between phases
- **Git branch** - Feature branch tracks work
- **Git worktree** - Isolated development environment

**Key files:**
- `.bob/state/brainstorm.md` - Research and approach
- `.bob/state/plan.md` - Implementation plan
- `.bob/state/test-results.md` - Test execution results
- `.bob/state/review.md` - Code review findings

---

## Subagent Chain

Each phase spawns specialized agents with clear inputs/outputs:

```
BRAINSTORM:
  Explore → .bob/state/brainstorm.md

PLAN:
  workflow-planner(.bob/state/brainstorm.md) → .bob/state/plan.md

EXECUTE:
  workflow-coder(.bob/state/plan.md) → code changes

TEST:
  workflow-tester(code) → .bob/state/test-results.md

REVIEW (9 agents in parallel):
  workflow-reviewer(code, .bob/state/plan.md) → .bob/state/review-code.md
  security-reviewer(code) → .bob/state/review-security.md
  performance-analyzer(code) → .bob/state/review-performance.md
  docs-reviewer(code, docs) → .bob/state/review-docs.md
  architect-reviewer(code, design) → .bob/state/review-architecture.md
  code-reviewer(code) → .bob/state/review-code-quality.md
  golang-pro(*.go files) → .bob/state/review-go.md
  debugger(code) → .bob/state/review-debug.md
  error-detective(code) → .bob/state/review-errors.md
  → Consolidate all → .bob/state/review.md
```

---

## Best Practices

**Orchestration:**
- Let subagents do the work
- Pass context via .bob/state/*.md files
- Clear input/output for each phase
- Chain agents together systematically

**Flow Control:**
- Enforce loop-back rules strictly
- MONITOR → PLAN (not REVIEW or EXECUTE)
- Never skip REVIEW phase
- Always validate test passage

**Quality:**
- TDD throughout (tests first)
- Comprehensive code review
- Fix issues properly (replan if needed)
- Maintain code quality standards

---

## Summary

**Remember:**
- You are the **orchestrator**, not the implementer
- Spawn **specialized subagents** for each phase
- Use **.bob/state/*.md files** for state
- Follow **flow control rules** strictly
- **MONITOR → PLAN** when issues found

**Goal:** Guide complete, high-quality feature development from idea to merged PR.

Good luck! 🏴‍☠️
