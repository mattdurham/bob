# TEST Phase

You are currently in the **TEST** phase of the workflow.

## Your Goal
Run all tests and checks to verify the code is production-ready.

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

### 1. Format Code
```bash
go fmt ./...
```
- Must be clean (no output)
- This is MANDATORY before proceeding

### 2. Run Linter
```bash
golangci-lint run
```
- Must pass with no warnings or errors
- Fix any issues found
- Document suppressions if absolutely necessary

### 3. Run All Tests
```bash
go test ./...
```
- ALL tests must pass
- No skipped tests without good reason
- If tests fail, loop back to EXECUTE to fix

### 4. Check Test Coverage
```bash
go test -cover ./...
```
- Report coverage percentage
- New code should have good coverage (aim for >80%)
- Identify any untested code paths

### 5. Check Cyclomatic Complexity
```bash
gocyclo -over 40 .
```
- No functions should exceed complexity 40
- If any do, refactor or get user approval
- Document any necessary high-complexity functions

### 6. Run Build
```bash
go build ./...
```
- Verify everything compiles
- No warnings

## Comprehensive Checklist

```
✅ go fmt ./... - clean
✅ golangci-lint run - clean
✅ go test ./... - all passing
✅ go test -cover ./... - X% coverage
✅ gocyclo -over 40 . - no violations
✅ go build ./... - successful
```

## DO NOT
- ❌ Do not declare work complete if ANY test fails
- ❌ Do not skip any of the checks above
- ❌ Do not automatically move to next phase without ALL checks passing
- ❌ Do not commit failing code

## CRITICAL: TESTS CANNOT BE SKIPPED, REMOVED, OR AVOIDED

**ABSOLUTE REQUIREMENTS:**
- ⛔ **NEVER skip tests** - Tests must always run, no exceptions
- ⛔ **NEVER remove failing tests** - Fix the code or the test, never delete tests
- ⛔ **NEVER disable tests** - Do not use skip flags, build tags, or comments to avoid tests
- ⛔ **NEVER bypass test phase** - Even if "no code changed", all tests must pass
- ⛔ **ALL tests must pass** - 100% pass rate required, no skipped or failing tests allowed
- ⛔ **Tests are MANDATORY** - Testing is not optional, negotiable, or skippable

**If tests fail:**
1. Fix the code that broke the tests
2. Fix the test if it's incorrect
3. Loop back to EXECUTE to make corrections
4. NEVER remove, skip, or disable the failing test

**Even if you made no code changes:**
- All tests still must be run
- All tests still must pass
- No shortcuts or assumptions

## When You're Done

### If ALL Checks Pass:
1. Tell user: "All tests and checks passing ✓"
2. Show the checklist with results
3. Report your progress:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "TEST",
       metadata: {
           "allTestsPass": true,
           "coverage": "85%",
           "lintClean": true
       }
   )
   ```

### If ANY Check Fails:
1. Tell user: "Tests/checks failed, fixing issues"
2. Loop back to EXECUTE:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "EXECUTE",
       metadata: {
           "loopReason": "test failures",
           "failedCheck": "<which check failed>"
       }
   )
   ```


## End Step
Ask bob what to do next based on the metadata you provided with bob_workflow_get_guidance.
