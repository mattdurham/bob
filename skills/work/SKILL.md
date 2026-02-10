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
          ↑                                              ↓
          └──────────────────────────────────────────────┘
                        (loop on issues)
```

## Flow Control Rules

**Loop-back paths:**
- **REVIEW → PLAN**: Major issues requiring replanning
- **REVIEW → EXECUTE**: Minor implementation fixes
- **TEST → EXECUTE**: Test failures
- **MONITOR → PLAN**: CI failures or PR feedback (always replan when issues found)

**Never skip REVIEW** - Always review before commit, even if tests pass.

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

Spawn workflow-planner agent:
```
Task(subagent_type: "workflow-planner",
     description: "Create implementation plan",
     prompt: "Read findings in bots/brainstorm.md.
             Create detailed implementation plan following TDD.
             Write plan to bots/plan.md.")
```

**Input:** `bots/brainstorm.md`
**Output:** `bots/plan.md`

Plan should include:
- Files to create/modify
- Implementation steps (tests first!)
- Edge cases to handle
- Test strategy
- Risks and mitigations

**If looping from REVIEW:** Update plan to address review findings

---

## Phase 4: EXECUTE

**Goal:** Implement the planned changes

**Actions:**

Spawn workflow-coder agent(s):
```
Task(subagent_type: "workflow-coder",
     description: "Implement feature",
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

## Phase 6: REVIEW

**Goal:** Comprehensive code review

**Actions:**

Spawn workflow-reviewer agent:
```
Task(subagent_type: "workflow-reviewer",
     description: "Code review",
     prompt: "Perform 3-pass code review:
             Pass 1: Cross-file consistency
             Pass 2: Code quality and security
             Pass 3: Documentation accuracy
             Write findings to bots/review.md.")
```

**Input:** Code changes, `bots/plan.md`
**Output:** `bots/review.md`

Review covers:
- Semantic correctness
- Security issues
- Code quality
- Best practices
- Documentation alignment

**Decision based on bots/review.md:**
- No issues → COMMIT
- Minor issues → EXECUTE (fix and return to REVIEW)
- Major issues → PLAN (need to replan)

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
   - **LOOP TO PLAN** (not REVIEW, not EXECUTE)
   - Review the issues
   - Update plan to address them
   - Then proceed through EXECUTE → TEST → REVIEW again

4. **If all checks pass and approved:**
   - Move to COMPLETE

**Critical:** MONITOR always loops to PLAN when issues found, ensuring proper replanning before fixes.

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
  
REVIEW:
  workflow-reviewer(code, bots/plan.md) → bots/review.md
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
