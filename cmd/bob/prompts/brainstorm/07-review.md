# REVIEW Phase

You are currently in the **REVIEW** phase of the workflow.

## Your Goal
Perform thorough code review to catch bugs, issues, and quality problems.

## What To Do

### 1. Spawn Review Subagent
Use Task tool to spawn a code review agent:
```
subagent_type: "general-purpose"
prompt: "Perform comprehensive code review of all changes in this worktree.
Review for:
- Bugs and logic errors
- Security vulnerabilities
- Edge cases not handled
- Code quality issues
- Best practices violations
- Missing error handling
- Potential performance problems

Write ALL findings to bots/review.md with:
- Severity (critical/high/medium/low)
- File and line number
- Issue description
- Suggested fix

If no issues found, create empty bots/review.md file."
```

### 2. Wait for Review Completion
- Let the subagent complete its work
- Do not interrupt or check status repeatedly
- Trust the subagent to return results

### 3. Read Review Results
```bash
cat bots/review.md
```

### 4. Analyze Findings

**If file is empty or < 10 bytes:**
- No issues found
- Ready to proceed to COMMIT

**If issues found:**
- Count critical/high severity issues
- Determine if issues require code changes
- Decide whether to loop back to PLAN

### 5. Use workflow_record_issues Tool
Record the issues found:
```
workflow_record_issues(
    worktreePath: "<worktree-path>",
    step: "REVIEW",
    issues: [
        {
            severity: "high",
            description: "Missing error handling in X",
            file: "path/to/file.go",
            line: 123
        },
        // ... more issues
    ]
)
```

This will tell you if you should loop back.

## DO NOT
- ❌ Do not skip the review subagent
- ❌ Do not review code yourself without subagent
- ❌ Do not declare work complete without reviewing
- ❌ Do not automatically move forward if issues exist
- ❌ Do not commit anything yet

## When You're Done

### Scenario A: No Issues Found
1. Tell user: "Code review complete - no issues found! ✓"
2. Report progress to COMMIT:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "COMMIT",
       metadata: {
           "reviewClean": true
       }
   )
   ```

### Scenario B: Issues Found
1. Tell user: "Found X issues during review (Y critical, Z high severity)"
2. Show preview of issues from bots/review.md
3. Record issues (use workflow_record_issues tool)
4. Tell user: "Looping back to PLAN to address issues"
5. Report progress back to PLAN:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "PLAN",
       metadata: {
           "loopReason": "review issues",
           "issueCount": X,
           "iteration": 2
       }
   )
   ```

## Next Phase
- **If clean**: Move to COMMIT phase
- **If issues**: Loop back to PLAN phase to address them
