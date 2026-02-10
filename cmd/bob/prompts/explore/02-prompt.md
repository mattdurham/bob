# PROMPT Phase

You are currently in the **PROMPT** phase of the explore workflow.

## State Management

**IMPORTANT:** You can store and retrieve state throughout this workflow:
- **Store state**: Include any important information in the `metadata` field when calling `workflow_report_progress`
- **Retrieve state**: Call `workflow_get_guidance` at the start of any step to retrieve all stored state
- **State examples**: exploration goals, specific files/modules to explore, questions to answer

## Your Goal

Ask the user what they want to explore in the codebase.

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

### 1. Ask User What They Want To Explore

Use AskUserQuestion or direct conversation to understand:
- What part of the codebase they want to understand
- What questions they want answered
- Any specific patterns or features to investigate

Example questions:
- "What would you like to explore in this codebase?"
- "What questions do you want answered?"
- "Which modules or features should we investigate?"

### 2. Gather Exploration Scope

If needed, clarify:
- Specific files, directories, or modules of interest
- Particular patterns or architectural concerns
- Depth of exploration needed
- Output format preferences

### 3. Store Exploration Goals

Once you understand what the user wants to explore:

```
workflow_report_progress(
    worktreePath: "<worktreePath>",
    currentStep: "PROMPT",
    metadata: {
        "explorationGoal": "<what to explore>",
        "scope": "<which parts of codebase>",
        "questions": "<specific questions to answer>"
    }
)
```

## DO NOT
- ❌ Do not create a worktree yet
- ❌ Do not start exploring the codebase
- ❌ Do not make assumptions about what to investigate

## When You're Done

Once you have clear exploration goals:

1. Summarize the exploration plan to the user
2. Report your progress with the goals in metadata
3. Let the workflow advance to WORKTREE where a worktree will be created

## End Step

Ask bob what to do next based on the metadata you provided with workflow_get_guidance.
