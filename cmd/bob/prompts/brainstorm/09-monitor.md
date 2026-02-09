# MONITOR Phase

You are currently in the **MONITOR** phase of the workflow.

## Your Goal
Push the branch, create a PR, and actively monitor it until merge.

## ⚠️ CRITICAL: Require User Permission

**DO NOT push or create PR automatically!**

Before proceeding with ANY of the steps below:
1. Tell user: "Ready to push and create PR?"
2. **WAIT for explicit user approval**
3. Only proceed after user says yes

This is a safety measure - never push to remote or create PRs without permission.

## What To Do

### 1. Push Branch
```bash
git push -u origin <branch-name>
```

### 2. Create Pull Request
```bash
gh pr create --title "<PR title>" --body "$(cat <<'EOF'
## Summary
- [Key change 1]
- [Key change 2]
- [Key change 3]

## Test Plan
- [x] Unit tests pass
- [x] Linting clean
- [x] Manual testing: [what you tested]

## Related Issues
Closes #[issue number if applicable]

🤖 Generated with Claude Code
EOF
)"
```

### 3. Get PR URL
```bash
gh pr view --web
```
Save the PR URL and share with user.

### 4. Automated PR Validation Check

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
    fmt.Println("\n🔄 Looping back to REVIEW phase...")
}
```

### 5. ACTIVELY MONITOR (Continuous Loop)
**This is critical - you must stay engaged until merge:**

Use the validation code above, or manually check every 2-3 minutes:
```bash
# Check CI/Actions status
gh pr checks

# Check for comments
gh pr view --json reviewThreads

# Check PR status
gh pr status
```

#### Watch For:
- ❌ **CI failures** - loop back to REVIEW
- 💬 **Unresolved conversations** - loop back to REVIEW
- ✅ **All checks passed + conversations resolved** - proceed to merge
- ⚠️ **Change requests** - address feedback

### 6. Decision Logic

**If ANY of these are true, loop back to REVIEW:**
- GitHub Actions checks fail
- Conversations are unresolved
- Change requests from reviewers

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
        "iteration": 3
    }
)
```

**If ALL are true, proceed to merge:**
- ✅ All CI checks passing (green)
- ✅ At least one approval
- ✅ No pending change requests
- ✅ No unresolved comments

### 7. Auto-Merge When Ready
**When validation passes:**
```bash
gh pr merge --auto --squash
```

### 8. After Merge
```bash
# Verify merge
gh pr status

# Clean up
git worktree remove <worktree-path>
git branch -d <branch-name>
```

## DO NOT
- ❌ Do not stop monitoring after creating PR
- ❌ Do not merge without approvals
- ❌ Do not merge with failing checks
- ❌ Do not ignore comments or feedback
- ❌ Do not wait to be asked - proactively check status
- ❌ Do not merge with unresolved conversations
- ❌ Do not skip validation checks

## When You're Done

### If Validation Passes and Merged:
1. Tell user: "PR merged successfully! ✓"
2. Report progress to COMPLETE:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "COMPLETE",
       metadata: {
           "merged": true,
           "allChecksPassed": true,
           "allConversationsResolved": true,
           "prUrl": "<url>"
       }
   )
   ```

### If Validation Fails (CI/Conversations):
1. Tell user: "Issues found during CI/review"
2. Loop back to REVIEW phase:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "MONITOR",
       metadata: {
           "loopReason": "validation_failed",
           "findings": "Validation failed: CI checks or conversations unresolved",
           "failedChecks": ["check1", "check2"],
           "unresolvedThreads": ["file.go:123"],
           "iteration": 3
       }
   )
   ```

## Next Phase
- Move to **REVIEW** if validation fails (loop back)
- Move to **COMPLETE** after successful merge
