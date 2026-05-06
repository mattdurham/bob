package anthropic_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	anthropicprovider "github.com/mattdurham/bob/bob/provider/anthropic"

	"github.com/mattdurham/bob/bob/provider"
	"github.com/mattdurham/bob/bob/sdk"
)

// sseBody builds a minimal SSE response body for an Anthropic streaming response.
// tokens is the list of text tokens to emit.
func sseBody(tokens []string) string {
	var sb strings.Builder

	// message_start
	sb.WriteString("event: message_start\n")
	sb.WriteString(`data: {"type":"message_start","message":{"id":"msg_test","type":"message","role":"assistant","content":[],"model":"claude-sonnet-4-5","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":10,"output_tokens":0}}}`)
	sb.WriteString("\n\n")

	// content_block_start
	sb.WriteString("event: content_block_start\n")
	sb.WriteString(`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`)
	sb.WriteString("\n\n")

	// one content_block_delta per token
	for _, tok := range tokens {
		sb.WriteString("event: content_block_delta\n")
		fmt.Fprintf(&sb, `data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":%q}}`, tok)
		sb.WriteString("\n\n")
	}

	// content_block_stop
	sb.WriteString("event: content_block_stop\n")
	sb.WriteString(`data: {"type":"content_block_stop","index":0}`)
	sb.WriteString("\n\n")

	// message_delta with stop_reason
	sb.WriteString("event: message_delta\n")
	fmt.Fprintf(&sb, `data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":%d}}`, len(tokens))
	sb.WriteString("\n\n")

	// message_stop
	sb.WriteString("event: message_stop\n")
	sb.WriteString(`data: {"type":"message_stop"}`)
	sb.WriteString("\n\n")

	return sb.String()
}

// roundTripFunc is an http.RoundTripper backed by a function.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// newTestProvider builds an anthropic provider that uses a mock HTTP transport.
func newTestProvider(transport http.RoundTripper) *anthropicprovider.Anthropic {
	return anthropicprovider.NewWithTransport("test-api-key", transport)
}

func TestStream_TokensInOrder(t *testing.T) {
	tokens := []string{"Hello", ", ", "world", "!"}
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseBody(tokens))),
		}, nil
	})

	p := newTestProvider(transport)
	req := provider.Request{
		Model:    "claude-sonnet-4-5",
		Messages: []sdk.Message{{Role: sdk.RoleUser, Content: "Hi"}},
	}

	var got []string
	err := p.Stream(context.Background(), req, func(token string) error {
		got = append(got, token)
		return nil
	})
	if err != nil {
		t.Fatalf("Stream returned error: %v", err)
	}
	if len(got) != len(tokens) {
		t.Fatalf("got %d tokens, want %d: %v", len(got), len(tokens), got)
	}
	for i, tok := range tokens {
		if got[i] != tok {
			t.Errorf("token[%d] = %q, want %q", i, got[i], tok)
		}
	}
}

func TestStream_APIError(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"type":"error","error":{"type":"authentication_error","message":"invalid api key"}}`)),
		}, nil
	})

	p := newTestProvider(transport)
	req := provider.Request{
		Model:    "claude-sonnet-4-5",
		Messages: []sdk.Message{{Role: sdk.RoleUser, Content: "Hi"}},
	}

	err := p.Stream(context.Background(), req, func(token string) error { return nil })
	if err == nil {
		t.Fatal("expected error from API, got nil")
	}
}

func TestStream_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel before making any request.
	cancel()

	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		// Return a valid SSE response body; context should already be cancelled.
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseBody([]string{"should", "not", "appear"}))),
		}, nil
	})

	p := newTestProvider(transport)
	req := provider.Request{
		Model:    "claude-sonnet-4-5",
		Messages: []sdk.Message{{Role: sdk.RoleUser, Content: "Hi"}},
	}

	var count int
	_ = p.Stream(ctx, req, func(token string) error {
		count++
		return nil
	})
	// Either an error is returned or no tokens are delivered (context cancelled).
	// The important invariant: we don't hang.
	// count may be 0 if context cancels before tokens, which is correct.
	_ = count
}

func TestStream_CallbackError(t *testing.T) {
	tokens := []string{"tok1", "tok2", "tok3"}
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseBody(tokens))),
		}, nil
	})

	p := newTestProvider(transport)
	req := provider.Request{
		Model:    "claude-sonnet-4-5",
		Messages: []sdk.Message{{Role: sdk.RoleUser, Content: "Hi"}},
	}

	callbackErr := fmt.Errorf("abort")
	var count int
	err := p.Stream(context.Background(), req, func(token string) error {
		count++
		if count == 2 {
			return callbackErr
		}
		return nil
	})
	if err == nil {
		t.Fatal("expected error from callback, got nil")
	}
	if count != 2 {
		t.Errorf("callback called %d times, want 2", count)
	}
}

func TestModels_NonEmpty(t *testing.T) {
	p := newTestProvider(http.DefaultTransport)
	models := p.Models()
	if len(models) == 0 {
		t.Fatal("Models() returned empty list")
	}
}

func TestName(t *testing.T) {
	p := newTestProvider(http.DefaultTransport)
	if p.Name() != "anthropic" {
		t.Errorf("Name() = %q, want %q", p.Name(), "anthropic")
	}
}

func TestStream_WithTools(t *testing.T) {
	tokens := []string{"result"}

	var capturedBody []byte
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		var err error
		capturedBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseBody(tokens))),
		}, nil
	})

	p := newTestProvider(transport)
	req := provider.Request{
		Model: "claude-sonnet-4-5",
		Messages: []sdk.Message{{Role: sdk.RoleUser, Content: "search for stuff"}},
		Tools: []sdk.Tool{
			{
				Name:        "search",
				Description: "Search the web",
				InputSchema: json.RawMessage(`{"type":"object","properties":{"q":{"type":"string"}},"required":["q"]}`),
			},
		},
	}

	var got []string
	err := p.Stream(context.Background(), req, func(token string) error {
		got = append(got, token)
		return nil
	})
	if err != nil {
		t.Fatalf("Stream returned error: %v", err)
	}

	// Verify the request body contains the tools array with properties.
	var reqBody map[string]interface{}
	if err := json.Unmarshal(capturedBody, &reqBody); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	toolsRaw, ok := reqBody["tools"]
	if !ok {
		t.Fatal("request body missing 'tools' field")
	}
	toolsList, ok := toolsRaw.([]interface{})
	if !ok || len(toolsList) == 0 {
		t.Fatalf("'tools' is not a non-empty array: %v", toolsRaw)
	}
	tool0, ok := toolsList[0].(map[string]interface{})
	if !ok {
		t.Fatalf("tools[0] is not an object: %v", toolsList[0])
	}
	inputSchema, ok := tool0["input_schema"].(map[string]interface{})
	if !ok {
		t.Fatalf("tools[0].input_schema is not an object: %v", tool0["input_schema"])
	}
	props, ok := inputSchema["properties"]
	if !ok {
		t.Fatal("input_schema missing 'properties' field — schema was silently dropped")
	}
	propsMap, ok := props.(map[string]interface{})
	if !ok || propsMap["q"] == nil {
		t.Fatalf("properties does not contain 'q': %v", props)
	}

	if len(got) != 1 || got[0] != "result" {
		t.Errorf("got tokens %v, want [result]", got)
	}
}

func TestStream_MultipleMessages(t *testing.T) {
	tokens := []string{"reply"}
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sseBody(tokens))),
		}, nil
	})

	p := newTestProvider(transport)
	req := provider.Request{
		Model: "claude-sonnet-4-5",
		Messages: []sdk.Message{
			{Role: sdk.RoleUser, Content: "Message 1"},
			{Role: sdk.RoleAssistant, Content: "Response 1"},
			{Role: sdk.RoleUser, Content: "Message 2"},
		},
	}

	var got []string
	err := p.Stream(context.Background(), req, func(token string) error {
		got = append(got, token)
		return nil
	})
	if err != nil {
		t.Fatalf("Stream returned error: %v", err)
	}
	if len(got) != 1 || got[0] != "reply" {
		t.Errorf("got tokens %v, want [reply]", got)
	}
}
