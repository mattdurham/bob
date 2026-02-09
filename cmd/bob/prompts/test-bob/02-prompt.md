# PROMPT Phase

You are currently in the **PROMPT** phase of the test-bob workflow.

## Your Goal
Prompt user for a true/false statement and pass it to the classification system.

## What To Do

### 1. Ask User for Statement

Use AskUserQuestion to get a statement from the user:
```
Ask: "Make a statement (true or false) to test Bob's classification:"

Options:
- "The sky is blue" (true statement)
- "2 + 2 = 5" (false statement)
- "Water is wet" (true statement)
- "Cats are dogs" (false statement)
- <User can provide custom statement>
```

### 2. Prepare Statement as Findings

Take the user's statement and format it as findings text:
```
Statement: <user's statement>

This is a factual statement that needs to be classified as true or false.
If the statement is TRUE, there are NO issues (empty findings = advance).
If the statement is FALSE, there ARE issues (findings exist = loop back).
```

### 3. Report Progress with Statement

**CRITICAL:** Report on current step (PROMPT) with findings:
```
workflow_report_progress(
    worktreePath: "<worktreePath>",
    currentStep: "PROMPT",
    metadata: {
        "findings": "<formatted statement from step 2>",
        "userStatement": "<original user statement>",
        "promptCompleted": true
    }
)
```

### 4. Tell User

```
📤 Statement submitted to Claude API for classification...
Waiting for orchestration to analyze and route...
```

## CRITICAL RULES
- ✅ **ALWAYS pass statement as findings text in metadata**
- ✅ Report on current step (PROMPT), not next step
- ✅ Let Claude API classify the statement
- ✅ Format statement to clearly indicate true = no issues, false = issues
- ✅ Orchestration will decide routing based on classification

## How It Works

1. User provides a statement
2. You format it and pass as "findings"
3. Claude API classifies:
   - TRUE statement → "no issues" → advance to COMPLETE
   - FALSE statement → "has issues" → loop back to PROMPT
4. You'll see the loop in action!

## Example Flow

**User says:** "The sky is blue"
**Formatted findings:**
```
Statement: The sky is blue

This is a factual statement that needs to be classified as true or false.
If the statement is TRUE, there are NO issues (empty findings = advance).
If the statement is FALSE, there ARE issues (findings exist = loop back).

Classification: This statement is TRUE. The sky is indeed blue.
Therefore: NO ISSUES - can advance.
```

**Result:** Claude classifies as "no issues" → advances to COMPLETE

---

**User says:** "2 + 2 = 5"
**Formatted findings:**
```
Statement: 2 + 2 = 5

This is a factual statement that needs to be classified as true or false.
If the statement is TRUE, there are NO issues (empty findings = advance).
If the statement is FALSE, there ARE issues (findings exist = loop back).

Classification: This statement is FALSE. 2 + 2 = 4, not 5.
Therefore: ISSUES EXIST - must loop back.
```

**Result:** Claude classifies as "has issues" → loops back to PROMPT

## Important
- DO NOT decide routing yourself
- DO NOT tell user what happens next
- ONLY pass the statement as findings
- Trust the orchestration to classify and route
