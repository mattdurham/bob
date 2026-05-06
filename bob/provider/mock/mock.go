// Package mock provides a scripted Provider implementation for tests.
package mock

import (
	"context"
	"fmt"

	"github.com/mattdurham/bob/bob/provider"
)

// Provider is a mock provider that returns pre-configured tokens.
// It is safe for use in tests.
type Provider struct {
	// Tokens is the ordered list of tokens to emit.
	Tokens []string

	// Err is an optional error to return after all tokens are emitted.
	// If set, it is returned from Stream after the last token.
	Err error

	// StreamErr is an optional error to return mid-stream before all tokens.
	// If non-nil, it is returned after emitting StreamErrAfter tokens.
	StreamErr      error
	StreamErrAfter int

	// CallCount is incremented each time Stream is called.
	CallCount int

	// LastRequest is the most recent request passed to Stream.
	LastRequest provider.Request
}

// Name returns "mock".
func (p *Provider) Name() string { return "mock" }

// Models returns a fixed list of test model names.
func (p *Provider) Models() []string {
	return []string{"mock-model-1", "mock-model-2"}
}

// Stream emits the configured tokens one by one, then returns p.Err.
// If p.StreamErr is set, it is returned after p.StreamErrAfter tokens.
func (p *Provider) Stream(ctx context.Context, req provider.Request, fn provider.StreamCallback) error {
	p.CallCount++
	p.LastRequest = req

	for i, tok := range p.Tokens {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if p.StreamErr != nil && i == p.StreamErrAfter {
			return p.StreamErr
		}

		if err := fn(tok); err != nil {
			return fmt.Errorf("stream callback: %w", err)
		}
	}

	return p.Err
}
