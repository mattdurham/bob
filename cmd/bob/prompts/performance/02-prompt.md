# PROMPT Phase

You are currently in the **PROMPT** phase of the performance workflow.

## State Management

**IMPORTANT:** You can store and retrieve state throughout this workflow:
- **Store state**: Include any important information in the `metadata` field when calling `workflow_report_progress`
- **Retrieve state**: Call `workflow_get_guidance` at the start of any step to retrieve all stored state
- **State examples**: performance goals, target metrics, areas to optimize

## Your Goal

Ask the user what they want to optimize.

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

### 1. Ask User What They Want To Optimize

Use AskUserQuestion or direct conversation to understand:
- What performance issues they're experiencing
- Which areas need optimization (speed, memory, throughput, latency)
- Any specific target metrics or goals
- Critical operations or code paths

Example questions:
- "What performance issues would you like to address?"
- "Which areas of the system need optimization?"
- "Do you have specific performance targets or metrics?"

### 2. Gather Performance Context

If needed, clarify:
- Current performance baseline (if known)
- Target performance metrics
- Specific operations or code paths that are slow
- Any constraints (memory limits, time budgets)
- Critical vs. nice-to-have optimizations

### 3. Store Performance Goals

Once you understand the optimization goals:

```
workflow_report_progress(
    worktreePath: "<worktreePath>",
    currentStep: "PROMPT",
    metadata: {
        "performanceGoal": "<what to optimize>",
        "targetMetrics": "<specific targets>",
        "criticalAreas": "<highest priority areas>"
    }
)
```

## DO NOT
- ❌ Do not create a worktree yet
- ❌ Do not start benchmarking or optimizing
- ❌ Do not make assumptions about what needs optimization

## When You're Done

Once you have clear performance goals:

1. Summarize the optimization plan to the user
2. Report your progress with the goals in metadata
3. Let the workflow advance to WORKTREE where a worktree will be created

## End Step

Ask bob what to do next based on the metadata you provided with bob_workflow_get_guidance.
