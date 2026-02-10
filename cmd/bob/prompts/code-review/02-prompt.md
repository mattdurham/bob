# PROMPT Phase

You are currently in the **PROMPT** phase of the code-review workflow.

## State Management

**IMPORTANT:** You can store and retrieve state throughout this workflow:
- **Store state**: Include any important information in the `metadata` field when calling `workflow_report_progress`
- **Retrieve state**: Call `workflow_get_guidance` at the start of any step to retrieve all stored state
- **State examples**: review scope, specific concerns, files to review

## Your Goal

Ask the user what they want to review.

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

### 1. Ask User What They Want To Review

Use AskUserQuestion or direct conversation to understand:
- What code they want reviewed (specific files, PRs, commits, branches)
- What aspects to focus on (security, performance, style, logic)
- Any particular concerns or areas of interest

Example questions:
- "What code would you like me to review?"
- "Are there specific concerns or aspects to focus on?"
- "Which files or changes should be reviewed?"

### 2. Clarify Review Scope

If needed, ask about:
- Specific files or directories to review
- Review depth (comprehensive vs. targeted)
- Any existing issues or concerns to verify
- Output format preferences

### 3. Store Review Scope

Once you understand the review scope:

```
workflow_report_progress(
    worktreePath: "<worktreePath>",
    currentStep: "PROMPT",
    metadata: {
        "reviewScope": "<what to review>",
        "focus": "<specific concerns>",
        "files": "<specific files if mentioned>"
    }
)
```

## DO NOT
- ❌ Do not create a worktree yet
- ❌ Do not start reviewing code
- ❌ Do not make assumptions about review scope

## When You're Done

Once you have clear review scope:

1. Summarize the review plan to the user
2. Report your progress with the scope in metadata
3. Let the workflow advance to WORKTREE where a worktree will be created

## End Step

Ask bob what to do next based on the metadata you provided with workflow_get_guidance.
