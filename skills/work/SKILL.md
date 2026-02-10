---
name: work
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

## Phase 1: INIT

**Goal:** Initialize and understand requirements

**Actions:**
1. Greet user and understand what they want to build
2. Create `bots/` directory for workflow artifacts:
   ```bash
   mkdir -p bots
   ```
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

**Step 2: Research existing patterns**

Spawn Explore agent for codebase research:
```
Task(subagent_type: "Explore",
     description: "Research similar implementations",
     run_in_background: true,
     prompt: "Search codebase for patterns related to [task].
             Find existing implementations, identify patterns to follow.
             Document findings.")
```

**Step 3: Document findings**

Write consolidated findings to `bots/brainstorm.md`:
- Requirements and constraints
- Existing patterns discovered
- Approaches considered (2-3 options with pros/cons)
- Recommended approach with rationale
- Risks and open questions

**Output:** `bots/brainstorm.md`

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
2. Subagent reads design from `bots/design.md` (or `bots/brainstorm.md`)
3. Subagent creates concrete, bite-sized implementation plan
4. Subagent writes plan to `bots/plan.md`

**Input:** `bots/design.md` or `bots/brainstorm.md`
**Output:** `bots/plan.md`

Plan includes:
- Exact file paths
- Complete code snippets
- Step-by-step actions (2-5 min each)
- TDD approach (test first!)
- Verification steps

**If looping from REVIEW:** Update plan to address review findings

---

## Phase 4: EXECUTE

**Goal:** Implement the planned changes

**Actions:**

Spawn workflow-coder agent(s):
```
Task(subagent_type: "workflow-coder",
     description: "Implement feature",
     run_in_background: true,
     prompt: "Follow plan in bots/plan.md.
             Use TDD: write tests first, verify they fail, then implement.
             Keep functions small (complexity < 40).
             Follow existing code patterns.")
```

**Input:** `bots/plan.md`
**Output:** Code implementation

For parallel work, spawn multiple coder agents for independent files.

**If looping from TEST/REVIEW:** Fix specific issues identified

---

## Phase 5: TEST

**Goal:** Run all tests and quality checks

**Actions:**

Spawn workflow-tester agent:
```
Task(subagent_type: "workflow-tester",
     description: "Run all tests and checks",
     run_in_background: true,
     prompt: "Run complete test suite and quality checks.
             Report results in bots/test-results.md.")
```

**Input:** Code to test
**Output:** `bots/test-results.md`

Checks:
- All tests pass
- No race conditions
- Good coverage (>80%)
- Code formatted
- Linter clean
- Complexity < 40

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
             Write findings to bots/review-code.md with severity levels.")

Task(subagent_type: "security-reviewer",
     description: "Security vulnerability review",
     run_in_background: true,
     prompt: "Scan code for security vulnerabilities:
             - OWASP Top 10 (injection, XSS, CSRF, etc.)
             - Secret detection (API keys, passwords)
             - Authentication/authorization issues
             - Input validation gaps
             Write findings to bots/review-security.md with severity levels.")

Task(subagent_type: "performance-analyzer",
     description: "Performance bottleneck review",
     run_in_background: true,
     prompt: "Analyze code for performance issues:
             - Algorithmic complexity (O(n²) opportunities)
             - Memory leaks and inefficient allocations
             - N+1 patterns and missing caching
             - Expensive operations in loops
             Write findings to bots/review-performance.md with severity levels.")

Task(subagent_type: "docs-reviewer",
     description: "Documentation accuracy review",
     run_in_background: true,
     prompt: "Review documentation for accuracy and completeness:
             - README accuracy (features match implementation)
             - Example validity (code examples work)
             - API documentation alignment
             - Comment correctness
             Write findings to bots/review-docs.md with severity levels.")

Task(subagent_type: "architect-reviewer",
     description: "Architecture and design review",
     run_in_background: true,
     prompt: "Evaluate system architecture and design decisions:
             - Design patterns appropriateness
             - Scalability assessment
             - Technology choices justification
             - Integration patterns validation
             - Technical debt analysis
             Write findings to bots/review-architecture.md with severity levels.")

Task(subagent_type: "code-reviewer",
     description: "Comprehensive code quality review",
     run_in_background: true,
     prompt: "Conduct deep code review across all aspects:
             - Logic correctness and error handling
             - Code organization and readability
             - Security best practices
             - Performance optimization opportunities
             - Maintainability and test coverage
             Write findings to bots/review-code-quality.md with severity levels.")

Task(subagent_type: "golang-pro",
     description: "Go-specific code review",
     run_in_background: true,
     prompt: "Review Go code for idiomatic patterns and best practices:
             - Idiomatic Go patterns (effective Go guidelines)
             - Concurrency patterns (goroutines, channels, context)
             - Error handling excellence
             - Performance and race condition analysis
             - Go-specific security concerns
             Write findings to bots/review-go.md with severity levels.")

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
             Write findings to bots/review-debug.md with severity levels.")

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
             Write findings to bots/review-errors.md with severity levels.")
```

**Wait for ALL 9 agents to complete.** If any agent fails, abort and report error.

**Input:** Code changes, `bots/plan.md`
**Output:**
- `bots/review-code.md` (code quality findings)
- `bots/review-security.md` (security findings)
- `bots/review-performance.md` (performance findings)
- `bots/review-docs.md` (documentation findings)
- `bots/review-architecture.md` (architecture findings)
- `bots/review-code-quality.md` (comprehensive code quality findings)
- `bots/review-go.md` (Go-specific findings)
- `bots/review-debug.md` (debugging and bug diagnosis findings)
- `bots/review-errors.md` (error handling pattern findings)
- `bots/review.md` (consolidated report - created in next step)

**Step 2: Consolidate Findings**

After all 9 agents complete successfully:

1. **Read all 9 review files:**
   ```
   Read(file_path: "/path/to/worktree/bots/review-code.md")
   Read(file_path: "/path/to/worktree/bots/review-security.md")
   Read(file_path: "/path/to/worktree/bots/review-performance.md")
   Read(file_path: "/path/to/worktree/bots/review-docs.md")
   Read(file_path: "/path/to/worktree/bots/review-architecture.md")
   Read(file_path: "/path/to/worktree/bots/review-code-quality.md")
   Read(file_path: "/path/to/worktree/bots/review-go.md")
   Read(file_path: "/path/to/worktree/bots/review-debug.md")
   Read(file_path: "/path/to/worktree/bots/review-errors.md")
   ```

2. **Parse and merge findings:**
   - Extract all issues from each file
   - Sort by severity (CRITICAL, HIGH, MEDIUM, LOW)
   - Deduplicate similar issues:
     - Same file:line → Merge descriptions
     - Keep highest severity
     - Note which agents found it

3. **Generate consolidated report:**
   Write to `bots/review.md`:
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

**Error Handling:**

- **Any agent fails:** Abort review, report to user, retry once
- **Empty results:** Valid (agent found no issues)
- **Consolidation fails:** Show individual files to user, ask for manual review

**Note:** This replaces the simple single-agent review with parallel multi-agent review (9 specialized agents) while maintaining backward compatibility (still produces `bots/review.md`).


---

## Phase 7: COMMIT

**Goal:** Commit changes with clear message

**Actions:**

1. Review changes:
   ```bash
   git status
   git diff
   ```

2. Create commit:
   ```bash
   git add [relevant-files]
   
   git commit -m "$(cat <<'EOF'
   type: brief description
   
   Detailed explanation of changes and why.
   
   Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
   EOF
   )"
   ```

3. Push and create PR:
   ```bash
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
- **bots/*.md files** - Persistent artifacts between phases
- **Git branch** - Feature branch tracks work
- **Git worktree** - Isolated development environment

**Key files:**
- `bots/brainstorm.md` - Research and approach
- `bots/plan.md` - Implementation plan
- `bots/test-results.md` - Test execution results
- `bots/review.md` - Code review findings

---

## Subagent Chain

Each phase spawns specialized agents with clear inputs/outputs:

```
BRAINSTORM:
  Explore → bots/brainstorm.md

PLAN:
  workflow-planner(bots/brainstorm.md) → bots/plan.md

EXECUTE:
  workflow-coder(bots/plan.md) → code changes

TEST:
  workflow-tester(code) → bots/test-results.md

REVIEW (9 agents in parallel):
  workflow-reviewer(code, bots/plan.md) → bots/review-code.md
  security-reviewer(code) → bots/review-security.md
  performance-analyzer(code) → bots/review-performance.md
  docs-reviewer(code, docs) → bots/review-docs.md
  architect-reviewer(code, design) → bots/review-architecture.md
  code-reviewer(code) → bots/review-code-quality.md
  golang-pro(*.go files) → bots/review-go.md
  debugger(code) → bots/review-debug.md
  error-detective(code) → bots/review-errors.md
  → Consolidate all → bots/review.md
```

---

## Best Practices

**Orchestration:**
- Let subagents do the work
- Pass context via .md files
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
- Use **bots/*.md files** for state
- Follow **flow control rules** strictly
- **MONITOR → PLAN** when issues found

**Goal:** Guide complete, high-quality feature development from idea to merged PR.

Good luck! 🏴‍☠️
