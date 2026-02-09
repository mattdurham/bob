# TEST Phase

You are currently in the **TEST** phase of the workflow.

## Your Goal
Run all tests and checks to verify the code is production-ready.

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

## When You're Done

### If ALL Checks Pass:
1. Tell user: "All tests and checks passing ✓"
2. Show the checklist with results
3. Report your progress:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "REVIEW",
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

## Next Phase
After reporting progress with all checks passing, you'll move to **REVIEW** phase for code review.
