# MONITOR Phase

You are currently in the **MONITOR** phase of the performance workflow.

## Your Goal
Push PR and monitor until merge.

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

## ⚠️ CRITICAL: Require User Permission

**DO NOT push or create PR automatically!**

Before proceeding with ANY of the steps below:
1. Tell user: "Ready to push and create PR?"
2. **WAIT for explicit user approval**
3. Only proceed after user says yes

This is a safety measure - never push to remote or create PRs without permission.

## What To Do

### 1. Push
```bash
git push -u origin perf-opt-<timestamp>
```

### 2. Create PR with Benchmark Data
```bash
gh pr create --title "perf: optimize XYZ for 50% improvement" --body "$(cat <<'EOF'
## Summary
Performance optimization of XYZ component.

## Performance Improvements
| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| Time/op | 1234ns | 617ns | 50% faster |
| Memory/op | 512B | 256B | 50% less |
| Allocs/op | 10 | 5 | 50% fewer |

## Changes
- Used sync.Pool for buffer reuse
- Replaced O(n²) algorithm with O(n) using map lookup
- Added benchmarks to verify improvements

## Test Plan
- [x] All tests passing
- [x] Benchmarks show expected improvements
- [x] No correctness regressions

## Benchmark Details
See bots/performance.md for complete benchmark results.

🤖 Generated with Claude Code
EOF
)"
```

### 3. Automated PR Validation Check

Run this Go code to validate the PR is ready:

```go
package main

import (
    "encoding/json"
    "fmt"
    "os/exec"
    "strings"
)

type PRCheck struct {
    AllChecksPassed          bool
    AllConversationsResolved bool
    FailedChecks             []string
    UnresolvedThreads        []string
    Feedback                 string
}

func ValidatePR(prNumber string) (*PRCheck, error) {
    result := &PRCheck{
        AllChecksPassed:          true,
        AllConversationsResolved: true,
        FailedChecks:             []string{},
        UnresolvedThreads:        []string{},
    }

    // Check GitHub Actions status
    checksCmd := exec.Command("gh", "pr", "checks", prNumber, "--json", "name,conclusion")
    checksOut, err := checksCmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to get PR checks: %w", err)
    }

    var checks []struct {
        Name       string `json:"name"`
        Conclusion string `json:"conclusion"`
    }
    if err := json.Unmarshal(checksOut, &checks); err != nil {
        return nil, fmt.Errorf("failed to parse checks: %w", err)
    }

    for _, check := range checks {
        if check.Conclusion != "SUCCESS" && check.Conclusion != "SKIPPED" {
            result.AllChecksPassed = false
            result.FailedChecks = append(result.FailedChecks,
                fmt.Sprintf("%s: %s", check.Name, check.Conclusion))
        }
    }

    // Check for unresolved conversations
    reviewsCmd := exec.Command("gh", "pr", "view", prNumber, "--json", "reviewThreads")
    reviewsOut, err := reviewsCmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to get PR reviews: %w", err)
    }

    var prData struct {
        ReviewThreads []struct {
            IsResolved bool   `json:"isResolved"`
            Path       string `json:"path"`
            Line       int    `json:"line"`
        } `json:"reviewThreads"`
    }
    if err := json.Unmarshal(reviewsOut, &prData); err != nil {
        return nil, fmt.Errorf("failed to parse reviews: %w", err)
    }

    for _, thread := range prData.ReviewThreads {
        if !thread.IsResolved {
            result.AllConversationsResolved = false
            result.UnresolvedThreads = append(result.UnresolvedThreads,
                fmt.Sprintf("%s:%d", thread.Path, thread.Line))
        }
    }

    // Generate feedback if issues found
    if !result.AllChecksPassed || !result.AllConversationsResolved {
        var feedbackParts []string

        if !result.AllChecksPassed {
            feedbackParts = append(feedbackParts,
                fmt.Sprintf("❌ Failed checks:\n%s", strings.Join(result.FailedChecks, "\n")))
        }

        if !result.AllConversationsResolved {
            feedbackParts = append(feedbackParts,
                fmt.Sprintf("💬 Unresolved conversations:\n%s", strings.Join(result.UnresolvedThreads, "\n")))
        }

        result.Feedback = strings.Join(feedbackParts, "\n\n")
    }

    return result, nil
}

func main() {
    // Get current PR number
    cmd := exec.Command("gh", "pr", "view", "--json", "number")
    out, err := cmd.Output()
    if err != nil {
        fmt.Printf("Failed to get PR: %v\n", err)
        return
    }

    var pr struct {
        Number int `json:"number"`
    }
    if err := json.Unmarshal(out, &pr); err != nil {
        fmt.Printf("Failed to parse PR: %v\n", err)
        return
    }

    prNumber := fmt.Sprintf("%d", pr.Number)
    check, err := ValidatePR(prNumber)
    if err != nil {
        fmt.Printf("Validation failed: %v\n", err)
        return
    }

    if check.AllChecksPassed && check.AllConversationsResolved {
        fmt.Println("✅ PR is ready to merge!")
        return
    }

    fmt.Println("⚠️  PR not ready. Issues found:")
    fmt.Println(check.Feedback)
    fmt.Println("\n🔄 Looping back to ANALYZE phase...")
}
```

### 4. Monitor

Use the validation code above, or manually check:
```bash
gh pr checks
gh pr view --json reviewThreads
gh pr status
```

Watch for:
- ❌ CI failures → loop back to ANALYZE
- ❌ Benchmark regressions → loop back to ANALYZE
- 💬 Unresolved conversations → loop back to ANALYZE
- ✅ All checks passed + conversations resolved → proceed to merge

### 5. Decision Logic

**If ANY of these are true, loop back to ANALYZE:**
- GitHub Actions checks fail
- Benchmark regressions detected
- Conversations are unresolved

**Report progress to loop back:**
```
workflow_report_progress(
    worktreePath: "<worktree-path>",
    currentStep: "MONITOR",
    metadata: {
        "loopReason": "validation_failed",
        "findings": "Validation failed: CI checks or conversations unresolved",
        "failedChecks": [...],
        "unresolvedThreads": [...],
        "benchmarkRegression": false
    }
)
```

**If ALL are true, proceed to merge:**
- ✅ All tests passing
- ✅ Benchmarks show expected improvements
- ✅ No unresolved comments
- ✅ Approved by reviewers

### 6. Auto-Merge When Ready
When validation passes:
```bash
gh pr merge --auto --squash
```

## DO NOT
- ❌ Do not merge if benchmarks regress
- ❌ Do not ignore performance feedback
- ❌ Do not merge with unresolved conversations
- ❌ Do not skip validation checks

## When You're Done

### If Validation Passes and Merged:
1. Tell user: "Performance improvements merged! ✓"
2. Report progress:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "COMPLETE",
       metadata: {
           "merged": true,
           "allChecksPassed": true,
           "allConversationsResolved": true
       }
   )
   ```

### If Validation Fails (CI/Benchmarks/Conversations):
1. Tell user: "Issues found during validation"
2. Loop back to ANALYZE phase:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "MONITOR",
       metadata: {
           "loopReason": "validation_failed",
           "findings": "Validation failed: CI checks or conversations unresolved",
           "failedChecks": ["check1"],
           "unresolvedThreads": ["file.go:123"],
           "benchmarkRegression": false
       }
   )
   ```

## Next Phase
- Move to **ANALYZE** if validation fails (loop back)
- Move to **COMPLETE** after successful merge
