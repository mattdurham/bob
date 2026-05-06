// Package provider defines the Provider interface and shared types for
// streaming LLM responses in the bob coding harness.
package provider

import (
	"context"

	"github.com/mattdurham/bob/bob/sdk"
)

// Request is a streaming chat request sent to a provider.
type Request struct {
	Model        string        `json:"model"`
	SystemPrompt string        `json:"system_prompt,omitempty"`
	Messages     []sdk.Message `json:"messages"`
	Tools        []sdk.Tool    `json:"tools,omitempty"`
	MaxTokens    int           `json:"max_tokens,omitempty"`
}

// StreamCallback is called once per token during streaming.
// Returning a non-nil error aborts the stream.
type StreamCallback func(token string) error

// Provider abstracts an LLM backend.
type Provider interface {
	// Name returns the provider identifier (e.g. "anthropic").
	Name() string

	// Models returns the list of supported model names.
	Models() []string

	// Stream sends a request and calls fn for each streamed token.
	// It blocks until the stream is complete or an error occurs.
	// Context cancellation stops streaming.
	Stream(ctx context.Context, req Request, fn StreamCallback) error
}
