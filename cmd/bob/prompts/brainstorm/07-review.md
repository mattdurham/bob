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

### 4. Parse and Structure Findings

Parse bots/review.md into structured JSON format:

**If file is empty or < 10 bytes:**
- Create empty findings JSON: `{"findings": []}`

**If issues found:**
- Parse all issues into structured JSON
- Count by severity
- Include file paths and line numbers

### 5. Prepare Findings JSON

Create a JSON blob in this exact format:
```json
{
  "findings": [
    {
      "severity": "high",
      "description": "Missing error handling in X",
      "file": "path/to/file.go",
      "line": 123,
      "suggestedFix": "Add error check"
    }
  ],
  "summary": {
    "total": 5,
    "critical": 1,
    "high": 2,
    "medium": 2,
    "low": 0
  }
}
```

**If no issues:** Return `{"findings": []}`

## DO NOT
- ❌ Do not skip the review subagent
- ❌ Do not review code yourself without subagent
- ❌ Do not declare work complete without reviewing
- ❌ Do not decide which phase to go to next
- ❌ Do not commit anything yet

## CRITICAL RULES
- ✅ **ALWAYS return findings as JSON blob**
- ✅ **Empty findings = workflow proceeds forward**
- ✅ **Non-empty findings = workflow decides next step**
- ✅ Let the workflow orchestration handle transitions
- ✅ Your job is to find and report issues, not route the workflow

## When You're Done

### Report Findings

**Use workflow_report_progress with findings JSON:**
```
workflow_report_progress(
    worktreePath: "<worktree-path>",
    currentStep: "REVIEW",
    metadata: {
        "findings": { /* JSON blob from step 5 */ },
        "reviewCompleted": true
    }
)
```

### Tell User

**If findings array is empty:**
```
✅ Code review complete - no issues found!
```

**If findings array has items:**
```
📋 Code review complete - found X issues:
- Critical: Y
- High: Z
- Medium: N

Workflow will automatically handle next steps.
```

## Important
- DO NOT tell user what phase comes next
- DO NOT call workflow_report_progress to another step
- ONLY report progress on current step (REVIEW) with findings
- Workflow orchestration will decide routing based on findings JSON
