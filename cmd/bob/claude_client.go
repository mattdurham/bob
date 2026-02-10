package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// ClaudeClient handles interactions with Claude API
type ClaudeClient struct {
	APIKey     string
	HTTPClient *http.Client
}

// ClaudeRequest represents the request structure for Claude API
type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []ClaudeMessage `json:"messages"`
}

// ClaudeMessage represents a message in the conversation
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeResponse represents the response from Claude API
type ClaudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
}

// NewClaudeClient creates a new Claude API client
func NewClaudeClient() *ClaudeClient {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		// Try alternate environment variable names
		apiKey = os.Getenv("CLAUDE_API_KEY")
	}

	return &ClaudeClient{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ClassifyFindings uses Claude API to determine if findings contain issues
// Returns (hasIssues bool, err error)
func (c *ClaudeClient) ClassifyFindings(findings string) (bool, error) {
	// If API key is not set, fall back to simple heuristics
	if c.APIKey == "" {
		return c.fallbackClassification(findings), nil
	}

	// Issue #8 fixed: Don't skip short findings - let model detect lazy reviews
	// Empty findings = no issues, but short non-empty findings might be lazy
	if len(strings.TrimSpace(findings)) == 0 {
		return false, nil
	}

	// Construct the classification prompt
	// Detect if this is a test-bob statement or code review findings
	var prompt string
	if strings.Contains(findings, "Statement:") && strings.Contains(findings, "classified as true or false") {
		// test-bob workflow: classify true/false statements
		prompt = fmt.Sprintf(`You are a binary classifier. Analyze the following statement and determine if it is factually TRUE or FALSE.

%s

Answer with ONLY one word: "yes" if the statement is FALSE (issues exist), or "no" if the statement is TRUE (no issues).

Answer:`, findings)
	} else {
		// Regular code review findings - STRICT validation
		prompt = fmt.Sprintf(`You are a balanced classifier for code review validation.

Analyze the following code review and determine if there are ACTUAL ISSUES that need to be fixed.

Code Review Findings:
%s

BALANCED RULES (Issue #7 fixed: Check for issues FIRST, then approval):

1. Answer "yes" (has issues) if ANY of these are true (CHECK FIRST):
   - Lists actual bugs, errors, or problems to fix
   - Contains severity markers (HIGH, MEDIUM, CRITICAL) with REAL issues
   - Has TODO items or action items that block progress
   - Test failures, build errors, or broken functionality
   - Security vulnerabilities or bugs listed
   - Review is LAZY/INSUFFICIENT (< 20 chars, just "OK"/"LGTM", no analysis)

2. Answer "no" (no issues) ONLY if ALL of these are true:
   - No bugs, errors, or problems listed (checked rule 1 first)
   - Explicitly states "Total Issues: 0" or "no issues found" OR
   - Contains approval (APPROVE, LGTM, Ready to merge) with meaningful analysis OR
   - Lists only positive findings with checkmarks (✅) and no issues
   - Summary clearly indicates clean/passing status

3. IGNORE these (they are NOT issues by themselves):
   - Positive checkmarks (✅ Tests pass, ✅ Code compiles, ✅ Clean)
   - Verification statements (Code works, No regressions)
   - Documentation of what was checked
   - Summary of passing checks

NOTE: If review has BOTH approval text AND actual issues, answer "yes" (issues take priority)

EXAMPLES:
- "Total Issues: 0. All tests pass." → Answer: "no"
- "✅ Code compiles ✅ Tests pass ✅ Ready to merge" → Answer: "no"
- "Bug in line 45: null pointer" → Answer: "yes"
- "CRITICAL: Security vulnerability found" → Answer: "yes"
- "OK" → Answer: "yes" (too short)

Answer with ONLY one word: "yes" if there are actual issues OR if review is insufficient, "no" if clean and approved.

Answer:`, findings)
	}

	// Make the API call
	req := ClaudeRequest{
		Model:     "claude-haiku-4-5-20251001", // Use Haiku for speed and cost
		MaxTokens: 10,                          // We only need "yes" or "no"
		Messages: []ClaudeMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(reqBody))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return false, fmt.Errorf("failed to call Claude API: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("Warning: failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("API error from Claude (status %d): %s", resp.StatusCode, string(body))
	}

	var claudeResp ClaudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return false, fmt.Errorf("empty response from Claude API")
	}

	// Parse yes/no from response
	answer := strings.ToLower(strings.TrimSpace(claudeResp.Content[0].Text))
	hasIssues := strings.Contains(answer, "yes")

	return hasIssues, nil
}

// fallbackClassification provides simple heuristics when API is not available
func (c *ClaudeClient) fallbackClassification(findings string) bool {
	findings = strings.TrimSpace(findings)

	// Empty or very short = no issues
	if len(findings) < 10 {
		return false
	}

	// Check for common issue indicators
	lowerFindings := strings.ToLower(findings)
	issueIndicators := []string{
		"error", "bug", "issue", "problem", "warning",
		"critical", "high", "medium", "severity",
		"fix", "missing", "incorrect", "invalid",
		"vulnerability", "security", "unsafe",
	}

	for _, indicator := range issueIndicators {
		if strings.Contains(lowerFindings, indicator) {
			return true
		}
	}

	// If it contains structured sections, likely has issues
	if strings.Contains(findings, "##") || strings.Contains(findings, "###") {
		return true
	}

	return false
}
