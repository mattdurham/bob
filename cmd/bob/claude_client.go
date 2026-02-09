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

	// If findings are too short, no issues
	if len(strings.TrimSpace(findings)) < 10 {
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
		// Regular code review findings
		prompt = fmt.Sprintf(`You are a binary classifier. Analyze the following code review findings and determine if there are any actual issues that need to be fixed.

Code Review Findings:
%s

Answer with ONLY one word: "yes" if there are issues that need fixing, or "no" if there are no issues (empty findings, or just comments with no actionable items).

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
