# INIT Phase

You are currently in the **INIT** phase of the test-bob workflow.

## Your Goal
Initialize the test workflow.

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

1. Tell user:
```
🧪 Test Bob Classification Workflow

This workflow tests the Claude API classification system.
You'll be prompted to make statements, and Bob will classify them as true or false.
If false, we loop back. If true, we complete.

Ready to test!
```

2. Report progress to PROMPT:
   ```
   workflow_report_progress(
       worktreePath: "<worktreePath>",
       currentStep: "INIT",
       metadata: {
           "testStarted": true
       }
   )
   ```

