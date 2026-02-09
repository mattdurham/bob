# REVIEW Phase

You are currently in the **REVIEW** phase of the workflow.

## Your Goal
Perform thorough code review to catch bugs, issues, and quality problems.

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

### 1. Spawn Review Subagent

Use Task tool to spawn a code review agent. Pass the following text verbatim as the prompt parameter:

~~~
subagent_type: "general-purpose"
prompt: "Perform comprehensive MULTI-PASS code review to catch semantic and logical errors.

## Review Process (3 Passes)

### PASS 1: Cross-File Consistency & Semantic Correctness
Focus on catching logical errors that span multiple files.

**Check for:**
- **Config-Code Mismatches**: Do configuration files match code that reads them?
  - Config options defined but not handled in code
  - Code expects config fields that don't exist
  - Default values differ between config and code

- **API Contract Violations**: Are required fields included?
  - Function calls missing required parameters
  - Struct initialization missing required fields
  - Interface implementations missing required methods

- **Cross-File Reference Errors**: Do references match definitions?
  - Constants/enums used match those defined
  - Type names referenced actually exist
  - State transitions follow valid rules

- **State Management Consistency**: Are keys/identifiers used consistently?
  - Cache/storage keys match between set/get operations
  - Session/context keys consistent across usage
  - Map keys use same naming pattern throughout

**Validation commands to run:**
```bash
# Find cross-file references
grep -r "config\..*\|Config{" . --include="*.go"
grep -r "[a-zA-Z_][a-zA-Z0-9_]*{" . --include="*.go"  # Struct initialization

# Check for inconsistent naming (limit output to 100 matches)
grep -Eroh '\<[a-z_]*config[a-z_]*\>' . | sort -u | head -100
grep -Eroh '\<[a-z_]*state[a-z_]*\>' . | sort -u | head -100
grep -Eroh '\<[a-z_]*key[a-z_]*\>' . | sort -u | head -100
```

**Output:** Write cross-file consistency issues to bots/review-consistency.md

---

### PASS 2: Code Quality & Logic
Standard code review for bugs and quality.

**Check for:**
- Bugs and logic errors
- Security vulnerabilities
- Edge cases not handled
- Missing error handling
- Race conditions
- Resource leaks
- Performance problems
- Best practices violations

**EXCLUDE:** Benchmark files (*_bench_test.go, files with Benchmark* functions)

**Output:** Write code quality issues to bots/review-quality.md

---

### PASS 3: Documentation & Examples Alignment
Verify documentation matches implementation.

**Check for:**
- **Example Code Validity**: Do code examples in comments/docs actually work?
  - Extract code blocks from markdown/comments
  - Verify they match current API signatures
  - Check parameters are correct

- **Configuration Examples**: Are config examples valid?
  - Do they include all required fields?
  - Are values within valid ranges?
  - Do they match schema/validation rules?

- **API Documentation**: Does documentation match implementation?
  - Parameter descriptions match actual parameters
  - Return types correct
  - Error conditions documented accurately

**Validation commands to run:**
```bash
# Extract code examples from docs
find . -name "*.md" -exec grep -A20 '```' {} \;
grep -A10 "// Example:" . -r --include="*.go"

# Find function signatures to compare
grep -r "^func " . --include="*.go"
```

**Output:** Write documentation issues to bots/review-docs.md

---

## Common Semantic Error Patterns to Watch For

### Pattern 1: Config-Code Mismatch
❌ Config defines "retry_count" but code uses cfg.RetryAttempts
✅ Names should match exactly

### Pattern 2: Missing Required Fields
❌ Function requires (name, opts) but called with just (name)
✅ All required parameters must be provided

### Pattern 3: Invalid References
❌ switch case uses "COMPLETE" but enum only defines "COMPLETED"
✅ References must match definitions exactly

### Pattern 4: State Inconsistency
❌ store.Set("user_id") but retrieve with store.Get("userId")
✅ Keys must be consistent (same spelling, case, format)

### Pattern 5: Documentation Mismatch
❌ Comment says "Returns nil on success" but code returns error
✅ Documentation must match actual behavior

---

## Final Output

Consolidate ALL findings from all 3 passes into bots/review.md with this format:

```markdown
## Cross-File Consistency Issues

### Issue 1: Config-Code Mismatch
**Severity:** HIGH
**Files:** config.json:12, server.go:45
**Description:** Config option \"log_level\" defined but not handled in code
**Impact:** Configuration option silently ignored
**Fix:** Add LogLevel field to Config struct and handle in initialization

## Code Quality Issues

### Issue 2: Missing Error Handling
**Severity:** MEDIUM
**Files:** client.go:89
**Description:** HTTP request error not checked
**Impact:** Silent failures, no error reporting
**Fix:** Check err and return/log appropriately

## Documentation Issues

### Issue 3: Invalid Example Code
**Severity:** LOW
**Files:** README.md:45
**Description:** Example missing required 'opts' parameter
**Impact:** Users will copy broken code
**Fix:** Update example to include Options parameter

[... continue for all findings ...]

## Summary

**Total Issues:** 5
- **CRITICAL:** 0
- **HIGH:** 1
- **MEDIUM:** 2
- **LOW:** 2

**Recommendation:** Address all CRITICAL/HIGH issues before proceeding. Review MEDIUM issues for potential impact.
```

If NO issues found in any pass, create empty bots/review.md file."
~~~

### 2. Wait for Review Completion
- Let the subagent complete its work
- Do not interrupt or check status repeatedly
- Trust the subagent to return results

### 3. Read Review Results
```bash
cat bots/review.md
```

### 4. Parse and Structure Findings

Parse bots/review.md into structured JSON format:

**If file is empty or < 10 bytes:**
- Create empty findings JSON: `{"findings": []}`

**If issues found:**
- Parse all issues into structured JSON
- Count by severity
- Include file paths and line numbers

### 5. Read Findings Content

Read the full content of bots/review.md:
```bash
cat bots/review.md
```

Store this content to pass in metadata (even if empty).

## DO NOT
- ❌ Do not skip the review subagent
- ❌ Do not review code yourself without subagent
- ❌ Do not declare work complete without reviewing
- ❌ Do not decide which phase to go to next
- ❌ Do not commit anything yet

## CRITICAL RULES
- ✅ **ALWAYS include findings text in metadata**
- ✅ Pass full content of bots/review.md (empty string if no issues)
- ✅ Workflow orchestration will classify and route automatically
- ✅ Your job is to find and report issues, not route the workflow
- ✅ Let Claude API determine if issues exist

## When You're Done

### Report Findings

**Use workflow_report_progress with findings text:**
```
workflow_report_progress(
    worktreePath: "<worktree-path>",
    currentStep: "REVIEW",
    metadata: {
        "findings": "<full content of bots/review.md>",
        "reviewCompleted": true
    }
)
```

### Tell User

```
📋 Code review complete - findings recorded.
Workflow will analyze and route automatically.
```

## Important
- DO NOT tell user what phase comes next
- DO NOT call workflow_report_progress to another step
- ONLY report progress on current step (REVIEW) with findings text
- Claude API will classify findings and orchestration will route
- Pass findings even if empty (empty string = no issues)
