# PROMPT Phase

You are currently in the **PROMPT** phase of the work workflow.

## State Management

**IMPORTANT:** You can store and retrieve state throughout this workflow:
- **Store state**: Include any important information in the `metadata` field when calling `workflow_report_progress`
- **Retrieve state**: Call `workflow_get_guidance` at the start of any step to retrieve all stored state
- **State examples**: file paths, configuration values, decisions made, task descriptions

## Your Goal

Ask the user what they want to work on and gather requirements.

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

### 1. Ask User What They Want To Do

Use AskUserQuestion or direct conversation to understand:
- What feature/fix/change they want to implement
- Any specific requirements or constraints
- The scope of the work

Example questions:
- "What would you like to work on?"
- "What feature or fix should we implement?"
- "Can you describe what you want to accomplish?"

### 2. Gather Additional Context

If needed, ask follow-up questions to clarify:
- Which parts of the codebase are involved
- Any architectural decisions needed
- Dependencies or related work
- Expected outcomes

### 3. Store Task Description

Once you understand what the user wants, store it for the next step:

```
workflow_report_progress(
    worktreePath: "<worktreePath>",
    currentStep: "PROMPT",
    metadata: {
        "taskDescription": "<clear description of what to do>",
        "scope": "<scope of work>",
        "requirements": "<any specific requirements>"
    }
)
```

## DO NOT
- ❌ Do not create a worktree yet
- ❌ Do not start planning or coding
- ❌ Do not make assumptions about what to build

## When You're Done

Once you have a clear understanding of the task:

1. Summarize back to the user what you understood
2. Report your progress with the task description in metadata
3. Let the workflow advance to WORKTREE where a worktree will be created

## End Step

Ask bob what to do next based on the metadata you provided with bob_workflow_get_guidance.
