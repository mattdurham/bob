# EXECUTE Phase

You are currently in the **EXECUTE** phase of the workflow.

## Your Goal
Implement the planned changes following TDD principles.

## What To Do

### 1. Follow TDD (Test-Driven Development)
**For each feature:**
1. Write the test first
2. Run the test - verify it fails
3. Implement the code to make it pass
4. Run the test - verify it passes
5. Refactor if needed

### 2. Write Tests First
```bash
# Create test file
touch path/to/feature_test.go

# Write test cases
# Run tests - they should fail
go test ./...
```

### 3. Implement Features
- Follow your plan from bots/plan.md
- Write clean, maintainable code
- Add comments for complex logic
- Follow existing code patterns
- Keep functions small (cyclomatic complexity < 40)

### 4. Consider Using Subagents
For parallel work or complex tasks:
```
Use Task tool with subagent_type="coder"
Prompt: "Implement [specific feature] in [file]. Follow TDD: write tests first..."
```

### 5. Verify As You Go
After each significant change:
```bash
go fmt ./...
go test ./...
```

### 6. Document Decisions
Add code comments explaining:
- Why you chose this approach
- Complex algorithms or logic
- Important assumptions
- Known limitations

## DO NOT
- ❌ Do not skip writing tests
- ❌ Do not write implementation before tests (violates TDD)
- ❌ Do not commit yet
- ❌ Do not automatically move to next phase
- ❌ Do not create functions with complexity > 40

## When You're Done
Once implementation is complete:

1. Verify all new code has tests
2. Run quick check:
   ```bash
   go fmt ./...
   go test ./...
   ```
3. Tell user: "Implementation complete, ready for testing phase"
4. Report your progress:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "TEST",
       metadata: {
           "executionComplete": true,
           "filesModified": ["file1.go", "file2.go"]
       }
   )
   ```

## Looping Back Here
If looping from MONITOR or TEST phases:
- Review the failures/issues
- Fix the problems
- Re-run tests to verify fixes

## Next Phase
After reporting progress, you'll move to **TEST** phase for comprehensive testing.
