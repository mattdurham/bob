// Package agent runs an embedded Haiku loop that can query the thoughts store.
// It acts as a senior developer: it decides what to search for, retrieves
// relevant findings, and synthesises an opinionated answer.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/mattdurham/bob/internal/navigator/store"
)

const systemPrompt = `You are a senior developer with deep knowledge accumulated from past code reviews,
debugging sessions, PR feedback, and architectural decisions on this codebase.

Your knowledge base contains findings recorded by other agents over time.
Use the search_thoughts tool to look up relevant past findings before answering.
You may call it multiple times — search from different angles if the first result is insufficient.

When answering:
- Draw on specific findings from your knowledge base, citing thought IDs where helpful
- Be concrete and opinionated: give a recommendation, not just a list of facts
- Distinguish confidence levels: "we verified this in production" vs "this was observed but not confirmed"
- If your knowledge base does not have enough context, say so clearly — do not guess`

var searchTool = anthropic.ToolUnionParam{
	OfTool: &anthropic.ToolParam{
		Name:        "search_thoughts",
		Description: anthropic.String("Search accumulated knowledge for past findings, fixes, decisions, and patterns"),
		InputSchema: anthropic.ToolInputSchemaParam{
			Type: "object",
			Properties: map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "What to search for",
				},
				"scope": map[string]any{
					"type":        "string",
					"description": "Optional: narrow to a package or file path, e.g. pkg/store",
				},
				"limit": map[string]any{
					"type":        "integer",
					"description": "Max results to return (default 10)",
				},
			},
			Required: []string{"query"},
		},
	},
}

// Agent runs the Haiku agentic loop against the thoughts store.
type Agent struct {
	client anthropic.Client
	store  *store.Store
}

// New creates an Agent with the given Anthropic API key and store.
func New(apiKey string, s *store.Store) *Agent {
	return &Agent{
		client: anthropic.NewClient(option.WithAPIKey(apiKey)),
		store:  s,
	}
}

type searchParams struct {
	Query string `json:"query"`
	Scope string `json:"scope"`
	Limit int    `json:"limit"`
}

// Consult asks the agent a question. The agent iteratively searches the store
// and returns a synthesised answer grounded in accumulated findings.
func (a *Agent) Consult(ctx context.Context, question, scope string) (string, error) {
	// Seed the scope hint into the question if provided.
	userMsg := question
	if scope != "" {
		userMsg = fmt.Sprintf("[scope: %s]\n\n%s", scope, question)
	}

	messages := []anthropic.MessageParam{
		{
			Role: anthropic.MessageParamRoleUser,
			Content: []anthropic.ContentBlockParamUnion{
				{OfText: &anthropic.TextBlockParam{Text: userMsg}},
			},
		},
	}

	const maxIterations = 6
	for i := range maxIterations {
		_ = i
		resp, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.ModelClaudeHaiku4_5_20251001,
			MaxTokens: 2048,
			System:    []anthropic.TextBlockParam{{Text: systemPrompt}},
			Tools:     []anthropic.ToolUnionParam{searchTool},
			Messages:  messages,
		})
		if err != nil {
			return "", fmt.Errorf("claude: %w", err)
		}

		switch resp.StopReason {
		case anthropic.StopReasonEndTurn:
			return extractText(resp.Content), nil

		case anthropic.StopReasonToolUse:
			// Append the assistant turn.
			messages = append(messages, resp.ToParam())

			// Execute each tool call and collect results.
			var results []anthropic.ContentBlockParamUnion
			for _, block := range resp.Content {
				if block.Type != "tool_use" {
					continue
				}
				output, isErr := a.runSearch(block.Input)
				results = append(results, anthropic.NewToolResultBlock(block.ID, output, isErr))
			}
			messages = append(messages, anthropic.MessageParam{
				Role:    anthropic.MessageParamRoleUser,
				Content: results,
			})

		default:
			return "", fmt.Errorf("unexpected stop reason: %s", resp.StopReason)
		}
	}

	return "", fmt.Errorf("agent did not converge after %d iterations", maxIterations)
}

func (a *Agent) runSearch(input json.RawMessage) (output string, isErr bool) {
	var p searchParams
	if err := json.Unmarshal(input, &p); err != nil {
		return fmt.Sprintf("invalid params: %v", err), true
	}

	thoughts, err := a.store.Search(p.Query, p.Scope, p.Limit)
	if err != nil {
		return fmt.Sprintf("search error: %v", err), true
	}
	if len(thoughts) == 0 {
		return "no findings matched that query", false
	}

	var sb strings.Builder
	for _, t := range thoughts {
		fmt.Fprintf(&sb, "[thought-%d]", t.ID)
		if t.Repo != "" {
			fmt.Fprintf(&sb, " repo:%s", t.Repo)
		}
		fmt.Fprintf(&sb, " scope:%s confidence:%s source:%s\n%s\n\n",
			t.Scope, t.Confidence, t.Source, t.Content)
	}
	return sb.String(), false
}

func extractText(blocks []anthropic.ContentBlockUnion) string {
	var sb strings.Builder
	for _, b := range blocks {
		if b.Type == "text" {
			sb.WriteString(b.Text)
		}
	}
	return sb.String()
}
