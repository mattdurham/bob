package pr

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// PRStatus holds PR status information.
type PRStatus struct {
	Number int
	Title  string
	State  string
	Raw    string
}

// MonitorPR retrieves PR status via the gh CLI.
func MonitorPR(prRef string) (string, error) {
	var sb strings.Builder

	// Get PR overview
	viewArgs := []string{"pr", "view", prRef, "--json", "number,title,state,reviews,statusCheckRollup,comments"}
	viewCmd := exec.Command("gh", viewArgs...)
	viewOut, viewErr := viewCmd.CombinedOutput()

	sb.WriteString("=== PR Status ===\n")

	if viewErr != nil {
		fmt.Fprintf(&sb, "Error fetching PR: %v\n%s\n", viewErr, string(viewOut))
		return sb.String(), nil
	}

	var prData map[string]any
	if err := json.Unmarshal(viewOut, &prData); err != nil {
		sb.WriteString(string(viewOut))
		sb.WriteString("\n")
	} else {
		formatPRData(&sb, prData)
	}

	// Get checks
	prNumber := extractPRNumber(prData, prRef)
	if prNumber != "" {
		checksCmd := exec.Command("gh", "pr", "checks", prNumber)
		checksOut, checksErr := checksCmd.CombinedOutput()
		sb.WriteString("\n=== CI Checks ===\n")
		if checksErr != nil {
			fmt.Fprintf(&sb, "Error fetching checks: %v\n", checksErr)
		}
		sb.WriteString(string(checksOut))
	}

	return sb.String(), nil
}

func formatPRData(sb *strings.Builder, data map[string]any) {
	if num, ok := data["number"]; ok {
		fmt.Fprintf(sb, "PR #%v", num)
	}
	if title, ok := data["title"]; ok {
		fmt.Fprintf(sb, ": %v", title)
	}
	sb.WriteString("\n")
	if state, ok := data["state"]; ok {
		fmt.Fprintf(sb, "State: %v\n", state)
	}

	// Status checks
	if rollup, ok := data["statusCheckRollup"]; ok && rollup != nil {
		checks, _ := rollup.([]any)
		if len(checks) > 0 {
			sb.WriteString("\nStatus checks:\n")
			for _, c := range checks {
				check, _ := c.(map[string]any)
				name := fmt.Sprintf("%v", check["name"])
				status := fmt.Sprintf("%v", check["status"])
				conclusion := fmt.Sprintf("%v", check["conclusion"])
				fmt.Fprintf(sb, "  %s: %s (%s)\n", name, status, conclusion)
			}
		}
	}

	// Reviews
	if reviews, ok := data["reviews"]; ok && reviews != nil {
		reviewList, _ := reviews.([]any)
		if len(reviewList) > 0 {
			sb.WriteString("\nReviews:\n")
			for _, r := range reviewList {
				review, _ := r.(map[string]any)
				author := ""
				if a, ok := review["author"].(map[string]any); ok {
					author = fmt.Sprintf("%v", a["login"])
				}
				state := fmt.Sprintf("%v", review["state"])
				body := fmt.Sprintf("%v", review["body"])
				fmt.Fprintf(sb, "  %s: %s\n", author, state)
				if body != "" && body != "<nil>" {
					fmt.Fprintf(sb, "    %s\n", body)
				}
			}
		}
	}

	// Comments
	if comments, ok := data["comments"]; ok && comments != nil {
		commentList, _ := comments.([]any)
		if len(commentList) > 0 {
			fmt.Fprintf(sb, "\nComments: %d\n", len(commentList))
			// Show last 3 comments
			start := 0
			if len(commentList) > 3 {
				start = len(commentList) - 3
			}
			for _, c := range commentList[start:] {
				comment, _ := c.(map[string]any)
				author := ""
				if a, ok := comment["author"].(map[string]any); ok {
					author = fmt.Sprintf("%v", a["login"])
				}
				body := fmt.Sprintf("%v", comment["body"])
				if len(body) > 200 {
					body = body[:200] + "..."
				}
				fmt.Fprintf(sb, "  [%s]: %s\n", author, body)
			}
		}
	}
}

func extractPRNumber(data map[string]any, prRef string) string {
	if num, ok := data["number"]; ok {
		return fmt.Sprintf("%v", num)
	}
	// Try to extract from URL
	parts := strings.Split(prRef, "/")
	if len(parts) > 0 {
		last := parts[len(parts)-1]
		if len(last) > 0 && last[0] >= '0' && last[0] <= '9' {
			return last
		}
	}
	return prRef
}
