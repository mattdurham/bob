# ANALYZE Phase

You are currently in the **ANALYZE** phase of the exploration workflow.

## Your Goal
Deeply analyze what you discovered to understand how it works.

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

### 1. Read Key Files
For each important file/component found:
- Read the full file
- Understand the logic
- Note important functions/classes
- Identify relationships between components

### 2. Trace Code Paths
Follow the execution flow:
- Entry points
- Call chains
- Data flow
- Error handling

### 3. Identify Patterns
Look for:
- Design patterns used
- Architecture decisions
- Common practices
- Anti-patterns or issues

### 4. Document Analysis
Update `bots/explore.md` with analysis:

```markdown
## Analysis

### How It Works
[Step-by-step explanation of the system]

### Code Flow
```
Entry → Component A → Component B → Result
```

### Key Functions
1. **functionName** (file.go:123)
   - Purpose: [what it does]
   - Called by: [callers]
   - Calls: [callees]

### Architecture Notes
- [Pattern or design choice 1]
- [Pattern or design choice 2]

### Interesting Findings
- [Notable implementation detail]
- [Clever solution]
- [Potential issue]
```

## CRITICAL RULES
- ❌ **NO FILE CHANGES** - Still read-only
- ❌ **NO EDITS** - Do not modify code
- ❌ **NO SUGGESTIONS YET** - Just understand, don't propose changes

## When You're Done
After thorough analysis:

1. Share key insights with user
2. Explain how the system works
3. Report progress:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "DOCUMENT",
       metadata: {
           "componentsAnalyzed": 5,
           "codePathsTraced": 3
       }
   )
   ```

