# COMPLETE Phase

You are currently in the **COMPLETE** phase of the exploration workflow.

## Your Goal
Finalize the exploration and hand off findings.

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

### 1. Final Review
- Review `bots/explore.md` for completeness
- Ensure all questions answered
- Verify accuracy of findings

### 2. Present Findings
Share with user:
- Executive summary
- Key discoveries
- Architecture understanding
- Answers to their questions

### 3. Next Steps Discussion
Ask user:
- Do they need deeper exploration of any area?
- Do they want to implement changes based on findings?
- Should we start a new workflow (e.g., `brainstorm`) to build something?

## Summary Report

Tell the user:
```
✅ Exploration Complete!

Summary:
- Topic: <what was explored>
- Files Reviewed: <count>
- Components Analyzed: <count>
- Documentation: bots/explore.md

Key Findings:
- <Finding 1>
- <Finding 2>
- <Finding 3>

Original Questions:
✓ <Question 1> - <Answer>
✓ <Question 2> - <Answer>

No files were modified (read-only exploration).
```

## CRITICAL RULES
- ❌ **NO FILE CHANGES** - Exploration is read-only
- ✅ **KNOWLEDGE GAINED** - You now understand the codebase better!

## When You're Done
1. Show summary to user
2. Workflow complete - no further reporting needed
3. Ready for next task!

## What's Next?
If user wants to make changes based on exploration:
- Start a new `brainstorm` workflow
- Create a worktree
- Implement the changes

If user wants more exploration:
- Start a new `explore` workflow
- Focus on different area
