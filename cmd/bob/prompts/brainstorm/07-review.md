# REVIEW Phase

You are currently in the **REVIEW** phase of the workflow.

## Your Goal
Perform thorough code review to catch bugs, issues, and quality problems.

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

### 5. Read Findings Content

Read the full content of bots/review.md:
```bash
cat bots/review.md
```

Store this content to pass in metadata (even if empty).

## DO NOT
- ❌ Do not skip the review subagent
- ❌ Do not review code yourself without subagent
- ❌ Do not declare work complete without reviewing
- ❌ Do not decide which phase to go to next
- ❌ Do not commit anything yet

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
