# FIX Phase

You are currently in the **FIX** phase of the code review workflow.

## Your Goal
Fix all issues found during code review.

## Continuation Behavior

**IMPORTANT:** Do NOT ask continuation questions like:
- "Should I proceed?"
- "Ready to continue?"
- "Shall I move to the next step?"
- "Done. Continue?"

**AUTOMATICALLY PROCEED** after completing your tasks.

**ONLY ASK THE USER** when:
- Choosing between multiple approaches/solutions
- Clarifying unclear requirements
- Confirming potentially risky/destructive actions (deletes, force pushes, etc.)
- Making architectural or design decisions

## What To Do

### 1. Read Review Findings
```bash
cat bots/review.md
```

### 2. Fix Each Issue
For every issue:
- Understand the problem
- Implement the fix
- **CRITICAL**: Write or update tests to cover the bug
- Verify test passes after fix

### 3. Create Tests for Bugs
For every bug fixed:
```go
// Test should fail before fix, pass after fix
func TestBugFix_IssueDescription(t *testing.T) {
    // Test the bug scenario
    // This should have failed before your fix
    // And pass after your fix
}
```

Use unit tests, NOT benchmarks.

### 4. Document Fixes
Keep track in bots/review.md by marking fixed issues.

## DO NOT
- ❌ Do not skip writing tests
- ❌ Do not fix issues without tests
- ❌ Do not commit yet
- ❌ Do not automatically move forward

## When You're Done
After fixing all issues:

1. Tell user: "All issues fixed, running checks"
2. Report progress:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "FIX"
   )
   ```

**Note:** No need to pass metadata - Bob tracks progress automatically.

