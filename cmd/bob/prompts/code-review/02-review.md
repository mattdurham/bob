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

### 4. Parse and Structure Findings

Parse bots/review.md into structured JSON format:

**If file is empty or < 10 bytes:**
- Create empty findings JSON: `{"findings": []}`

**If issues found:**
- Parse all issues into structured JSON
- Count by severity
- Include file paths and line numbers

### 5. Read Findings Content

Read the full content of bots/review.md:
```bash
cat bots/review.md
```

Store this content to pass in metadata (even if empty).

## DO NOT
- ❌ Do not skip review subagent
- ❌ Do not decide which phase to go to next
- ❌ Do not commit yet

## CRITICAL RULES
- ✅ **ALWAYS include findings text in metadata**
- ✅ Pass full content of bots/review.md (empty string if no issues)
- ✅ Workflow orchestration will classify and route automatically
- ✅ Your job is to find and report issues, not route the workflow
- ✅ Let Claude API determine if issues exist

## When You're Done

### Report Findings

**Use workflow_report_progress with findings text:**
```
workflow_report_progress(
    worktreePath: "<worktree-path>",
    currentStep: "REVIEW",
    metadata: {
        "findings": "<full content of bots/review.md>",
        "reviewCompleted": true
    }
)
```

### Tell User

```
📋 Code review complete - findings recorded.
Workflow will analyze and route automatically.
```

## Important
- DO NOT tell user what phase comes next
- DO NOT call workflow_report_progress to another step
- ONLY report progress on current step (REVIEW) with findings text
- Claude API will classify findings and orchestration will route
- Pass findings even if empty (empty string = no issues)
