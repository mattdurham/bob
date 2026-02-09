# DOCUMENT Phase

You are currently in the **DOCUMENT** phase of the exploration workflow.

## Your Goal
Create comprehensive documentation of your exploration findings.

## What To Do

### 1. Finalize bots/explore.md
Complete the exploration document with:

```markdown
# Exploration Report: <Topic>

## Executive Summary
[2-3 sentence overview of what you found]

## File Structure
```
directory/
├── component1/
│   ├── file1.go
│   └── file2.go
└── component2/
    └── file3.go
```

## Architecture Overview
[High-level architecture diagram or description]

## Key Components

### Component 1: <Name>
- **Location**: path/to/files
- **Purpose**: [what it does]
- **Dependencies**: [what it uses]
- **Key Functions**:
  - `Function1` - [description]
  - `Function2` - [description]

### Component 2: <Name>
...

## Code Flows

### Flow 1: <Scenario>
```
User Request → Handler → Processor → Database → Response
```

### Flow 2: <Scenario>
...

## Patterns & Practices
- **Pattern 1**: [description and locations]
- **Pattern 2**: [description and locations]

## Potential Improvements
[Optional: If you noticed opportunities]
- [Improvement 1]
- [Improvement 2]

## Questions Answered
- ✓ [Original question 1] - [Answer]
- ✓ [Original question 2] - [Answer]

## Open Questions
- [Question that needs further exploration]
- [Question for the user]

## References
- [File references]
- [External documentation]
```

### 2. Create Summary
Prepare a verbal summary for the user covering:
- What you explored
- Key findings
- How it works
- Answer to original questions

## CRITICAL RULES
- ❌ **NO FILE CHANGES** - Documentation goes only in bots/explore.md
- ❌ **NO CODE EDITS** - This is exploration, not implementation

## When You're Done
After documentation is complete:

1. Present findings to user
2. Ask if they have follow-up questions
3. Report progress:
   ```
   workflow_report_progress(
       worktreePath: "<current-directory>",
       currentStep: "COMPLETE",
       metadata: {
           "documentationComplete": true,
           "questionsAnswered": 5
       }
   )
   ```

