package harness

import (
	"context"

	"charm.land/fantasy"
)

// mockLM is a test double for fantasy.LanguageModel that emits a fixed
// list of text tokens and then finishes.
type mockLM struct {
	tokens    []string
	streamErr error
	provider  string
	modelID   string
	callCount int
}

// compile-time interface check
var _ fantasy.LanguageModel = (*mockLM)(nil)

func newMockLM(tokens ...string) *mockLM {
	return &mockLM{
		tokens:   tokens,
		provider: "mock",
		modelID:  "mock-model-1",
	}
}

func (m *mockLM) Generate(_ context.Context, _ fantasy.Call) (*fantasy.Response, error) {
	return &fantasy.Response{}, nil
}

func (m *mockLM) Stream(ctx context.Context, _ fantasy.Call) (fantasy.StreamResponse, error) {
	m.callCount++
	tokens := m.tokens
	streamErr := m.streamErr
	return func(yield func(fantasy.StreamPart) bool) {
		for _, tok := range tokens {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if !yield(fantasy.StreamPart{
				Type:  fantasy.StreamPartTypeTextDelta,
				Delta: tok,
			}) {
				return
			}
		}
		if streamErr != nil {
			yield(fantasy.StreamPart{Type: fantasy.StreamPartTypeError, Error: streamErr})
			return
		}
		yield(fantasy.StreamPart{
			Type:         fantasy.StreamPartTypeFinish,
			FinishReason: fantasy.FinishReasonStop,
		})
	}, nil
}

func (m *mockLM) GenerateObject(_ context.Context, _ fantasy.ObjectCall) (*fantasy.ObjectResponse, error) {
	return nil, nil
}

func (m *mockLM) StreamObject(_ context.Context, _ fantasy.ObjectCall) (fantasy.ObjectStreamResponse, error) {
	return nil, nil
}

func (m *mockLM) Provider() string { return m.provider }
func (m *mockLM) Model() string    { return m.modelID }
