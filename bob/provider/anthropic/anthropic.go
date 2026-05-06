// Package anthropic implements the provider.Provider interface backed by the
// Anthropic Messages API.
package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/mattdurham/bob/bob/provider"
	"github.com/mattdurham/bob/bob/sdk"
)

// Anthropic implements provider.Provider using the Anthropic Messages API.
type Anthropic struct {
	client anthropic.Client
}

// supportedModels is the v1 list of supported Anthropic model IDs.
var supportedModels = []string{
	"claude-opus-4-5",
	"claude-sonnet-4-5",
	"claude-haiku-3-5",
}

// New creates a new Anthropic provider with the given API key.
func New(apiKey string) *Anthropic {
	return &Anthropic{
		client: anthropic.NewClient(option.WithAPIKey(apiKey)),
	}
}

// NewWithTransport creates a new Anthropic provider with a custom HTTP transport.
// Intended for testing.
func NewWithTransport(apiKey string, transport http.RoundTripper) *Anthropic {
	return &Anthropic{
		client: anthropic.NewClient(
			option.WithAPIKey(apiKey),
			option.WithHTTPClient(&http.Client{Transport: transport}),
		),
	}
}

// Name returns "anthropic".
func (a *Anthropic) Name() string { return "anthropic" }

// Models returns the list of supported Anthropic model names.
func (a *Anthropic) Models() []string {
	out := make([]string, len(supportedModels))
	copy(out, supportedModels)
	return out
}

// Stream sends req to the Anthropic Messages API and calls fn for each text
// token in the response. It blocks until streaming completes or an error
// occurs. Context cancellation propagates immediately.
func (a *Anthropic) Stream(ctx context.Context, req provider.Request, fn provider.StreamCallback) error {
	params, err := buildParams(req)
	if err != nil {
		return fmt.Errorf("anthropic: build params: %w", err)
	}

	stream := a.client.Messages.NewStreaming(ctx, params)
	defer stream.Close() //nolint:errcheck // best-effort close on stream

	for stream.Next() {
		event := stream.Current()
		if event.Type != "content_block_delta" {
			continue
		}

		delta := event.Delta
		if delta.Type != "text_delta" {
			continue
		}

		token := delta.Text
		if token == "" {
			continue
		}

		if cbErr := fn(token); cbErr != nil {
			return cbErr
		}

		// Check context between tokens.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	if err := stream.Err(); err != nil {
		return fmt.Errorf("anthropic: stream: %w", err)
	}

	return nil
}

// buildParams converts a provider.Request into anthropic.MessageNewParams.
func buildParams(req provider.Request) (anthropic.MessageNewParams, error) {
	model := req.Model
	if model == "" {
		model = "claude-sonnet-4-5"
	}

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	messages, err := convertMessages(req.Messages)
	if err != nil {
		return anthropic.MessageNewParams{}, err
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: int64(maxTokens),
		Messages:  messages,
	}

	if req.SystemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: req.SystemPrompt},
		}
	}

	if len(req.Tools) > 0 {
		tools, err := convertTools(req.Tools)
		if err != nil {
			return anthropic.MessageNewParams{}, err
		}
		params.Tools = tools
	}

	return params, nil
}

// convertMessages converts sdk.Message slice to Anthropic message params.
func convertMessages(msgs []sdk.Message) ([]anthropic.MessageParam, error) {
	out := make([]anthropic.MessageParam, 0, len(msgs))
	for _, m := range msgs {
		var role anthropic.MessageParamRole
		switch m.Role {
		case sdk.RoleUser:
			role = anthropic.MessageParamRoleUser
		case sdk.RoleAssistant:
			role = anthropic.MessageParamRoleAssistant
		default:
			return nil, fmt.Errorf("anthropic: unknown role %q", m.Role)
		}
		out = append(out, anthropic.MessageParam{
			Role: role,
			Content: []anthropic.ContentBlockParamUnion{
				{OfText: &anthropic.TextBlockParam{Text: m.Content}},
			},
		})
	}
	return out, nil
}

// convertTools converts sdk.Tool slice to Anthropic tool params.
func convertTools(tools []sdk.Tool) ([]anthropic.ToolUnionParam, error) {
	out := make([]anthropic.ToolUnionParam, 0, len(tools))
	for _, t := range tools {
		var schema anthropic.ToolInputSchemaParam
		if len(t.InputSchema) > 0 {
			// InputSchema is raw JSON with the full schema object.
			// Unmarshal and extract properties and required separately.
			var schemaObj struct {
				Properties map[string]interface{} `json:"properties"`
				Required   []string               `json:"required"`
			}
			if err := json.Unmarshal(t.InputSchema, &schemaObj); err != nil {
				return nil, fmt.Errorf("anthropic: unmarshal tool %q input schema: %w", t.Name, err)
			}
			schema = anthropic.ToolInputSchemaParam{
				Type:       "object",
				Properties: schemaObj.Properties,
				Required:   schemaObj.Required,
			}
		} else {
			schema = anthropic.ToolInputSchemaParam{Type: "object"}
		}

		out = append(out, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        t.Name,
				Description: anthropic.String(t.Description),
				InputSchema: schema,
			},
		})
	}
	return out, nil
}
