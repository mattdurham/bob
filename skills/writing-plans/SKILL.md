---
name: writing-plans
description: Use when you have a spec or requirements for a multi-step task, before touching code
---

# Writing Plans

## Overview

Spawn a subagent to write comprehensive implementation plans assuming the engineer has zero context for our codebase. The subagent documents everything: which files to touch, code, testing, docs. Plans are bite-sized tasks following DRY, YAGNI, TDD principles with frequent commits.

**Announce at start:** "I'm using the writing-plans skill to spawn a planner subagent."

**Process:**
1. Spawn workflow-planner subagent with background execution
2. Subagent reads design from `bots/design.md`
3. Subagent writes plan to `bots/plan.md`
4. Wait for completion notification

## Bite-Sized Task Granularity

**Each step is one action (2-5 minutes):**
- "Write the failing test" - step
- "Run it to make sure it fails" - step
- "Implement the minimal code to make the test pass" - step
- "Run the tests and make sure they pass" - step
- "Commit" - step

## Plan Document Header

**Every plan MUST start with this header:**

```markdown
# [Feature Name] Implementation Plan

**Goal:** [One sentence describing what this builds]

**Architecture:** [2-3 sentences about approach]

**Tech Stack:** [Key technologies/libraries]

---
```

## Task Structure

```markdown
### Task N: [Component Name]

**Files:**
- Create: `exact/path/to/file.py`
- Modify: `exact/path/to/existing.py:123-145`
- Test: `tests/exact/path/to/test.py`

**Step 1: Write the failing test**

```python
def test_specific_behavior():
    result = function(input)
    assert result == expected
```

**Step 2: Run test to verify it fails**

Run: `pytest tests/path/test.py::test_name -v`
Expected: FAIL with "function not defined"

**Step 3: Write minimal implementation**

```python
def function(input):
    return expected
```

**Step 4: Run test to verify it passes**

Run: `pytest tests/path/test.py::test_name -v`
Expected: PASS

**Step 5: Commit**

```bash
git add tests/path/test.py src/path/file.py
git commit -m "feat: add specific feature"
```
```

## Remember
- Exact file paths always
- Complete code in plan (not "add validation")
- Exact commands with expected output
- Reference relevant skills with @ syntax
- DRY, YAGNI, TDD, frequent commits

## Subagent Instructions

Spawn the planner subagent with this exact structure:

```
Task(subagent_type: "workflow-planner",
     description: "Create implementation plan",
     run_in_background: true,
     prompt: "Read the design from bots/design.md.

     Create a concrete, bite-sized implementation plan.

     Each task should be 2-5 minutes and include:
     - Exact file paths
     - Complete code snippets
     - Exact commands to run
     - Expected output
     - TDD approach (test first, then implement)

     Format each task with:
     - Files to create/modify
     - Step-by-step actions (write test, run test, implement, verify, commit)
     - Exact code to write
     - Verification steps

     Write the complete plan to bots/plan.md using the Write tool.")
```

## Execution Handoff

After subagent completes, confirm:

**"Plan complete and saved to `bots/plan.md`. Ready to move to EXECUTE phase?"**
