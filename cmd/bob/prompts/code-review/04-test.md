# TEST Phase

You are currently in the **TEST** phase of the code review workflow.

## Your Goal
Run all checks to verify fixes are correct.

## What To Do

### Run All Checks
```bash
# Format
go fmt ./...

# Lint
golangci-lint run

# Test (NO benchmarks)
go test ./...

# Coverage
go test -cover ./...

# Complexity
gocyclo -over 40 .

# Build
go build ./...
```

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
- ❌ Do not skip any checks
- ❌ Do not proceed if tests fail
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
3. Loop back to FIX to make corrections
4. NEVER remove, skip, or disable the failing test

**Even if you made no code changes:**
- All tests still must be run
- All tests still must pass
- No shortcuts or assumptions

## When You're Done

### If ALL Checks Pass:
1. Tell user: "All checks passing ✓"
2. Report progress back to REVIEW for final verification:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "REVIEW",
       metadata: {
           "allTestsPass": true,
           "iteration": 2,
           "readyForFinalReview": true
       }
   )
   ```

### If Checks Fail:
1. Tell user: "Checks failed, fixing issues"
2. Loop back to FIX:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "FIX",
       metadata: {
           "testsFailed": true,
           "failedCheck": "<which check>"
       }
   )
   ```

## Next Phase
Loop back to **REVIEW** for final verification that all issues are resolved.
