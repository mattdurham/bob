# INIT Phase

You are currently in the **INIT** phase of the test-bob workflow.

## Your Goal
Initialize the test workflow.

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
       currentStep: "PROMPT",
       metadata: {
           "testStarted": true
       }
   )
   ```

