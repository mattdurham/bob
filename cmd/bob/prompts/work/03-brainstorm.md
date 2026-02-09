# BRAINSTORM Phase

You are currently in the **BRAINSTORM** phase of the workflow.

## Your Goal
Explore approaches and clarify requirements before creating an implementation plan.

## What To Do

### 1. Clarify Requirements
- Ask the user questions if anything is unclear
- Understand the "why" behind the request
- Identify constraints and edge cases
- Document assumptions

### 2. Research Existing Patterns
- Search the codebase for similar implementations using Glob/Grep
- Read relevant files to understand current architecture
- Identify patterns to follow or avoid
- Note dependencies and related code

### 3. Consider Multiple Approaches
- Brainstorm at least 2-3 different ways to solve this
- Consider trade-offs:
  - Complexity vs simplicity
  - Performance implications
  - Maintainability
  - Consistency with existing patterns
- Think about what could go wrong

### 4. Document Findings
Write your thoughts to `bots/brainstorm.md`:

```markdown
# Brainstorm: <Task Description>

## Requirements
- [Key requirement 1]
- [Key requirement 2]
- [Constraints or edge cases]

## Existing Patterns
- [Similar code found at file.go:123]
- [Pattern we should follow]
- [Pattern to avoid]

## Approaches Considered

### Approach 1: <Name>
**Pros:**
- [Pro 1]
- [Pro 2]

**Cons:**
- [Con 1]
- [Con 2]

### Approach 2: <Name>
...

## Recommended Approach
[Which approach and why]

## Risks/Open Questions
- [Risk or question 1]
- [Risk or question 2]
```

## DO NOT
- ❌ Do not automatically move to the next phase
- ❌ Do not start writing code yet
- ❌ Do not create the implementation plan yet
- ❌ Do not commit anything

## When You're Done
Once you have thoroughly explored the problem space and documented your findings:

1. Share key findings with the user
2. Ask if they have feedback on your brainstorming
3. Only after user acknowledges, report your progress:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "BRAINSTORM",
       metadata: {
           "brainstormComplete": true,
           "approachesConsidered": 3
       }
   )
   ```

