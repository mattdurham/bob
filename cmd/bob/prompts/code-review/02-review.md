# REVIEW Phase

You are currently in the **REVIEW** phase of the code review workflow.

## Your Goal
Perform comprehensive code review and identify issues.

## What To Do

### 1. Spawn Review Subagent
```
Use Task tool with subagent_type="general-purpose"
Prompt: "Review ALL code in this repository as a junior engineer.
Look for:
- Bugs and logic errors
- Security vulnerabilities
- Edge cases not handled
- Missing error handling
- Code quality issues
- Performance problems
- Best practices violations

EXCLUDE benchmark files (*_bench_test.go, files with Benchmark* functions).

Write ALL findings to bots/review.md with:
- Severity (critical/high/medium/low)
- File and line number
- Issue description
- Suggested fix

If no issues, create empty bots/review.md file."
```

### 2. Wait for Review
Let subagent complete its work.

### 3. Check Results
```bash
cat bots/review.md
```

### 4. Analyze Findings
- If file is empty or < 10 bytes: No issues, proceed to COMMIT
- If issues found: Record them and loop to FIX

## DO NOT
- ❌ Do not skip review subagent
- ❌ Do not proceed if critical issues found
- ❌ Do not commit yet

## When You're Done

### If NO Issues:
1. Tell user: "Code review complete - no issues found!"
2. Report progress:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "COMMIT",
       metadata: { "reviewClean": true }
   )
   ```

### If Issues Found:
1. Tell user: "Found X issues (Y critical, Z high severity)"
2. Record issues:
   ```
   workflow_record_issues(
       worktreePath: "<worktree-path>",
       step: "REVIEW",
       issues: [/* issues from review.md */]
   )
   ```
3. Report progress to FIX:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "FIX",
       metadata: {
           "issuesFound": X,
           "iteration": 1
       }
   )
   ```

## Next Phase
- **If clean**: Move to COMMIT
- **If issues**: Loop to FIX phase
