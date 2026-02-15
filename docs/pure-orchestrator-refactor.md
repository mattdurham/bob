# Pure Orchestrator Refactor - Progress Report

**Date:** 2026-02-11
**Status:** BRAINSTORM ✅ | REVIEW ✅ | Others 🚧

---

## Overview

Refactoring the work workflow orchestrator to be **pure** - it only coordinates via `.bob/` folder files and never does actual work itself.

### Pure Orchestrator Principle

**Orchestrator CAN:**
- ✅ Read/write files in `.bob/` folder only
- ✅ Spawn Task tool subagents
- ✅ Route between phases based on `.bob/` files

**Orchestrator CANNOT:**
- ❌ Touch source code files
- ❌ Run git commands (except worktree setup)
- ❌ Run tests
- ❌ Do research/exploration
- ❌ Make implementation decisions
- ❌ Consolidate data
- ❌ Make routing decisions

---

## Completed Phases

### ✅ Phase 2: BRAINSTORM (COMPLETE)

**New Agent:** `workflow-brainstormer`
**Location:** `~/.claude/agents/workflow-brainstormer/SKILL.md`

**Flow:**
```
1. Orchestrator writes task → .bob/brainstorm-prompt.md
2. Orchestrator spawns workflow-brainstormer agent
3. Agent reads prompt, does research, iterates
4. Agent writes findings → .bob/brainstorm.md (with timestamps)
5. Agent signals completion: "BRAINSTORM COMPLETE"
6. Orchestrator reads .bob/brainstorm.md
7. Orchestrator routes to PLAN
```

**Key Features:**
- Autonomous (no user interaction)
- Uses Explore agent for codebase research
- Appends findings with ISO timestamps
- Considers 2-3 approaches with pros/cons
- Makes recommendation with rationale
- Self-determines completion

**Output Format:**
```markdown
# Brainstorm

## 2026-02-11 14:30:15 - Task Received
[Task description]

## 2026-02-11 14:31:42 - Research Findings
[Patterns found, dependencies, etc.]

## 2026-02-11 14:33:28 - Approaches Considered
[Approach 1, 2, 3 with pros/cons]

## 2026-02-11 14:35:51 - Recommendation
[Chosen approach with rationale]

## 2026-02-11 14:36:15 - BRAINSTORM COMPLETE
Ready for PLAN phase
```

---

### ✅ Phase 6: REVIEW (COMPLETE)

**New Agents:**
1. `review-consolidator` - `~/.claude/agents/review-consolidator/SKILL.md`
2. `review-router` - `~/.claude/agents/review-router/SKILL.md`

**Flow:**
```
1. Orchestrator spawns 9 reviewer agents in parallel
   → .bob/review-code.md
   → .bob/review-security.md
   → .bob/review-performance.md
   → .bob/review-docs.md
   → .bob/review-architecture.md
   → .bob/review-code-quality.md
   → .bob/review-go.md
   → .bob/review-debug.md
   → .bob/review-errors.md

2. Orchestrator spawns review-consolidator agent
   - Reads all 9 review files
   - Parses and deduplicates findings
   - Sorts by severity
   → .bob/review.md

3. Orchestrator spawns review-router agent
   - Reads .bob/review.md
   - Analyzes severity and scope
   - Makes routing decision
   → .bob/routing.md

4. Orchestrator reads .bob/routing.md
5. Orchestrator routes to next phase (BRAINSTORM/EXECUTE/COMMIT)
```

**review-consolidator Features:**
- Reads 9 review files in parallel
- Flexible parsing (handles different formats)
- Deduplicates by file:line
- Sorts by severity (CRITICAL → HIGH → MEDIUM → LOW)
- Tracks which agents found each issue
- Generates comprehensive consolidated report
- Handles missing files gracefully

**review-router Features:**
- Analyzes severity distribution
- Considers scope and complexity
- Applies routing rules:
  - CRITICAL/HIGH → BRAINSTORM
  - MEDIUM/LOW → EXECUTE
  - None → COMMIT
- Provides detailed rationale
- Conservative approach (err on side of BRAINSTORM)
- Clear action recommendations

**Output Format (routing.md):**
```markdown
# Routing Decision

Decision: BRAINSTORM

## Analysis
- CRITICAL: 2
- HIGH: 3
- MEDIUM: 4
- LOW: 1

## Decision Rationale
Route to: BRAINSTORM

Primary Reasons:
1. 2 CRITICAL security vulnerabilities require architectural review
2. Issues span 8 files indicating systemic problems
3. Root causes need addressing; quick fixes insufficient

## Recommended Actions
Focus on:
- SQL injection in login handler
- XSS vulnerability in user input
- O(n²) algorithm in dashboard

## For Orchestrator
ROUTING: BRAINSTORM
CONTEXT: 2 CRITICAL + 3 HIGH issues require architectural review
ACTION: Update .bob/brainstorm-prompt.md and spawn workflow-brainstormer
```

---

## Remaining Phases

### 🚧 Phase 3: PLAN (NEEDS UPDATE)

**Current State:** Mostly pure, uses writing-plans skill
**Status:** May need minor updates for consistency

**Potential Changes:**
- Orchestrator writes `.bob/plan-prompt.md`
- Spawn workflow-planner agent
- Agent writes `.bob/plan.md`
- Orchestrator reads plan and continues

**Priority:** Low (already quite clean)

---

### 🚧 Phase 4: EXECUTE (NEEDS UPDATE)

**Current State:** Spawns workflow-coder agent
**Status:** May need prompt file approach

**Potential Changes:**
- Orchestrator writes `.bob/execute-prompt.md`
- Include: specific fixes needed, plan to follow
- Spawn workflow-coder agent(s)
- Orchestrator reads completion status

**Priority:** Medium

---

### 🚧 Phase 5: TEST (NEEDS UPDATE)

**Current State:** Spawns workflow-tester agent
**Status:** May need prompt file approach

**Potential Changes:**
- Orchestrator writes `.bob/test-prompt.md`
- Spawn workflow-tester agent
- Agent writes `.bob/test-results.md`
- Orchestrator reads results and routes

**Priority:** Medium

---

### 🚧 Phase 7: COMMIT (NEEDS NEW AGENT)

**Current State:** Orchestrator runs git commands directly
**Status:** VIOLATES pure orchestrator principle

**New Agent Needed:** `commit-agent`
**Location:** `~/.claude/agents/commit-agent/SKILL.md`

**Flow:**
```
1. Orchestrator writes .bob/commit-prompt.md
2. Orchestrator spawns commit-agent
3. Agent runs git status, git diff
4. Agent creates commit with co-author tag
5. Agent pushes and creates PR
6. Agent writes status → .bob/commit.md
7. Orchestrator reads status and continues
```

**Priority:** High (violates pure orchestrator)

---

### 🚧 Phase 8: MONITOR (NEEDS NEW AGENT)

**Current State:** Orchestrator runs gh commands directly
**Status:** VIOLATES pure orchestrator principle

**New Agent Needed:** `monitor-agent`
**Location:** `~/.claude/agents/monitor-agent/SKILL.md`

**Flow:**
```
1. Orchestrator writes .bob/monitor-prompt.md
2. Orchestrator spawns monitor-agent
3. Agent checks CI status via gh
4. Agent checks for PR comments/reviews
5. Agent writes status → .bob/monitor.md
6. Orchestrator reads status and routes
```

**Priority:** High (violates pure orchestrator)

---

## Benefits of Pure Orchestrator

### Separation of Concerns
- Orchestrator: Coordination only
- Agents: Actual work
- Clear responsibilities

### Composability
- Agents can be used independently
- Easy to add new agents
- Easy to modify agent behavior
- No orchestrator changes needed

### Testability
- Each agent independently testable
- File-based contracts make testing simple
- Mock .bob/ files for testing

### Observability
- All state visible in .bob folder
- Easy to debug (read the files)
- Conversation history preserved
- Audit trail included

### Maintainability
- Small, focused agents
- Clear inputs/outputs
- Self-documenting via files
- Easy to understand flow

---

## Next Steps

**Priority Order:**

1. **Test BRAINSTORM refactor** ✅
   - Run real workflow with new workflow-brainstormer
   - Verify conversation history format works
   - Check completion signal

2. **Test REVIEW refactor** ✅
   - Run with 9 reviewers + consolidator + router
   - Verify deduplication works
   - Check routing logic

3. **Build commit-agent** 🚧
   - Handle git operations
   - Create commits and PRs
   - Report status

4. **Build monitor-agent** 🚧
   - Check CI status
   - Monitor PR feedback
   - Report results

5. **Update EXECUTE/TEST** 🚧
   - Add prompt file approach
   - Ensure consistency
   - Polish flow

6. **End-to-end test** 🚧
   - Run complete workflow
   - Verify all phases work
   - Fix any issues

7. **Documentation** 🚧
   - Update CLAUDE.md
   - Add agent documentation
   - Create examples

---

## File Structure

```
~/.claude/
  agents/
    workflow-brainstormer/
      SKILL.md ✅
    review-consolidator/
      SKILL.md ✅
    review-router/
      SKILL.md ✅
    commit-agent/
      SKILL.md 🚧 (TODO)
    monitor-agent/
      SKILL.md 🚧 (TODO)

  skills/
    work/
      SKILL.md ✅ (updated for BRAINSTORM + REVIEW)

worktree/
  .bob/
    brainstorm-prompt.md    (orchestrator writes)
    brainstorm.md           (brainstormer writes)
    plan-prompt.md          (orchestrator writes)
    plan.md                 (planner writes)
    execute-prompt.md       (orchestrator writes)
    test-prompt.md          (orchestrator writes)
    test-results.md         (tester writes)
    review-code.md          (reviewer writes)
    review-security.md      (reviewer writes)
    review-performance.md   (reviewer writes)
    review-docs.md          (reviewer writes)
    review-architecture.md  (reviewer writes)
    review-code-quality.md  (reviewer writes)
    review-go.md            (reviewer writes)
    review-debug.md         (reviewer writes)
    review-errors.md        (reviewer writes)
    review.md               (consolidator writes)
    routing.md              (router writes)
    commit-prompt.md        (orchestrator writes)
    commit.md               (commit-agent writes)
    monitor-prompt.md       (orchestrator writes)
    monitor.md              (monitor-agent writes)
```

---

## Success Criteria

**Pure orchestrator is complete when:**

- ✅ Orchestrator ONLY reads/writes `.bob/` folder
- ✅ All actual work done by subagents
- ✅ Clear file-based contracts between phases
- ✅ Conversation history preserved in files
- ✅ Each agent is self-contained and testable
- ✅ Orchestrator can route based on file contents alone
- ❌ No direct git/test/research in orchestrator (2 phases remaining)
- ❌ All agents follow consistent patterns
- ❌ End-to-end workflow succeeds

**2 of 9 phases complete, 7 remaining**

---

*Progress: BRAINSTORM ✅ | REVIEW ✅ | PLAN 🚧 | EXECUTE 🚧 | TEST 🚧 | COMMIT 🚧 | MONITOR 🚧*
