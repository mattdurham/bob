# MONITOR Phase

You are currently in the **MONITOR** phase of the code review workflow.

## Your Goal
Push PR and monitor until merge.

## What To Do

### 1. Push Branch
```bash
git push -u origin review-fix-<timestamp>
```

### 2. Create PR
```bash
gh pr create --title "fix: address code review issues" --body "$(cat <<'EOF'
## Summary
Fixed issues found during comprehensive code review:
- [Issue 1]
- [Issue 2]
- [Issue 3]

## Changes
- [File/component changed]
- Added tests for all bug fixes

## Test Plan
- [x] All existing tests pass
- [x] New tests added for bug fixes
- [x] Linting clean
- [x] Complexity check passed

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
    // Get PR number from command line or git branch
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
    fmt.Println("\n🔄 Looping back to REVIEW phase...")
}
```

### 4. Monitor Actively

Use the validation code above, or manually check every 2-3 minutes:
```bash
gh pr checks
gh pr view --json reviewThreads
gh pr status
```

Watch for:
- ❌ CI failures → loop back to REVIEW
- 💬 Unresolved conversations → loop back to REVIEW
- ✅ All checks passed + all conversations resolved → proceed to merge

### 5. Decision Logic

**If ANY of these are true, loop back to REVIEW:**
- GitHub Actions checks fail
- Conversations are unresolved
- New feedback from reviewers

**Report progress to loop back:**
```
workflow_report_progress(
    worktreePath: "<worktree-path>",
    currentStep: "MONITOR",
    metadata: {
        "loopReason": "validation_failed",
        "failedChecks": [...],
        "unresolvedThreads": [...],
        "iteration": 3
    }
)
```

**If ALL are true, proceed to merge:**
- ✅ All GitHub Actions checks pass
- ✅ All conversations resolved
- ✅ Approved by reviewers

### 6. Auto-Merge When Ready
When validation passes:
```bash
gh pr merge --auto --squash
```

### 7. After Merge
```bash
git worktree remove <worktree-path>
git branch -d review-fix-<timestamp>
```

## DO NOT
- ❌ Do not stop monitoring after creating PR
- ❌ Do not merge without approval
- ❌ Do not ignore CI failures
- ❌ Do not merge with unresolved conversations
- ❌ Do not skip validation checks

## When You're Done

### If Validation Passes and Merged:
1. Tell user: "PR merged successfully! ✓"
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

### If Validation Fails (CI/Conversations):
1. Loop back to REVIEW phase
2. Report progress with details:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "MONITOR",
       metadata: {
           "loopReason": "validation_failed",
           "failedChecks": ["check1", "check2"],
           "unresolvedThreads": ["file.go:123"],
           "iteration": 3
       }
   )
   ```

## Next Phase
- Move to **REVIEW** if validation fails (loop back)
- Move to **COMPLETE** after successful merge
