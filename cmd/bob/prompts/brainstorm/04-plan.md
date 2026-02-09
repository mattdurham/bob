# PLAN Phase

You are currently in the **PLAN** phase of the workflow.

## Your Goal
Create a detailed, actionable implementation plan based on your brainstorming.

## What To Do

### 1. Break Down Into Tasks
- List concrete, specific tasks in order
- Identify which files need to be changed
- Identify new files that need to be created
- Consider dependencies between tasks

### 2. Plan Your Test Strategy
- Decide what tests need to be written
- Plan to write tests BEFORE implementation (TDD)
- Identify edge cases to test
- Consider test coverage goals

### 3. Check Dependencies
- What existing code will you use?
- What libraries might you need?
- Check license compatibility for new dependencies (see user's principles)
- Document any concerns

### 4. Document The Plan
Write your plan to `bots/plan.md`:

```markdown
# Implementation Plan: <Task Description>

## Overview
[Brief summary of what you're going to build]

## Files to Modify
1. `path/to/file1.go` - [what changes]
2. `path/to/file2.go` - [what changes]

## Files to Create
1. `path/to/newfile.go` - [purpose]

## Implementation Steps

### Step 1: Write Tests
- [ ] Create `path/to/file_test.go`
- [ ] Test case: [description]
- [ ] Test case: [description]
- [ ] Verify tests fail (TDD)

### Step 2: Implement Feature
- [ ] Task 1: [specific action]
- [ ] Task 2: [specific action]
- [ ] Task 3: [specific action]

### Step 3: Verify
- [ ] Run tests - all should pass
- [ ] Run linters
- [ ] Check cyclomatic complexity

## Edge Cases to Handle
- [Edge case 1]
- [Edge case 2]

## Risks/Concerns
- [Concern 1 and mitigation]
- [Concern 2 and mitigation]

## Dependencies
- [Library/package needed]
- [License: MIT/Apache/etc.]
```

### 5. Consider Cyclomatic Complexity
- Keep functions simple (< 40 complexity)
- Break down complex logic into smaller functions
- Plan for readability and maintainability

## DO NOT
- ❌ Do not automatically move to EXECUTE phase
- ❌ Do not start implementing yet
- ❌ Do not write code (except perhaps pseudocode in plan.md)
- ❌ Do not commit anything

## When You're Done
Once you have a clear, detailed plan:

1. Share the plan with the user
2. Ask if they want any changes to the approach
3. Only after user approves or acknowledges, report your progress:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "EXECUTE",
       metadata: {
           "planComplete": true,
           "filesAffected": 5
       }
   )
   ```

## Looping Back Here
If you loop back from REVIEW or MONITOR phases:
- Review the issues in `bots/review.md`
- Update this plan to address the issues
- Document what changed and why

## Next Phase
After reporting progress, you'll move to **EXECUTE** phase to implement the plan.
