# Providers — LLM Backend Interface

Bob communicates with language model backends through the `Provider` interface
defined in `bob/provider`. Swapping providers or models requires no changes to
the harness or extension code.

---

## Provider Interface

```go
// bob/provider/provider.go

type Provider interface {
    Name() string
    Models() []string
    Stream(ctx context.Context, req Request, fn StreamCallback) error
}
```

### `Name() string`

Returns a short, human-readable identifier for the provider (e.g. `"anthropic"`
or `"mock"`). The name is displayed in the status bar.

### `Models() []string`

Returns the ordered list of model names the provider supports. The first element
is used as the default model when no explicit model is configured. The list is
also used to populate auto-completion (if any) and validation warnings.

### `Stream(ctx context.Context, req Request, fn StreamCallback) error`

Sends `req` to the LLM backend and delivers the streamed response one token at
a time by calling `fn` for each token.

- Blocks until the stream is fully consumed or an error occurs.
- Context cancellation (e.g. Ctrl+C in the TUI) must abort the stream promptly.
- If `fn` returns a non-nil error, `Stream` must stop and return that error.
- Returns `nil` on success, or a wrapped error describing what went wrong.

---

## Request Struct

```go
type Request struct {
    Model        string        // Active model identifier.
    SystemPrompt string        // Optional system prompt.
    Messages     []sdk.Message // Full conversation history.
    Tools        []sdk.Tool    // Tools available for this request.
    MaxTokens    int           // 0 means use provider default (4096 for Anthropic).
}
```

| Field          | Type           | Description                                                   |
|----------------|----------------|---------------------------------------------------------------|
| `Model`        | string         | Model name, e.g. `"claude-sonnet-4-5"`.                      |
| `SystemPrompt` | string         | Prepended to the conversation as a system-level instruction.  |
| `Messages`     | []sdk.Message  | Ordered conversation history (user and assistant turns).      |
| `Tools`        | []sdk.Tool     | Tool definitions registered by extensions.                    |
| `MaxTokens`    | int            | Maximum tokens in the response. Provider default if `0`.      |

`sdk.Message` fields:

| Field     | Type        | Description                        |
|-----------|-------------|------------------------------------|
| `Role`    | sdk.Role    | `"user"` or `"assistant"`.         |
| `Content` | string      | The message text.                  |

---

## StreamCallback Contract

```go
type StreamCallback func(token string) error
```

- Called once per streamed token (may be a single character or a short string,
  depending on the backend).
- **Returning a non-nil error aborts the stream.** `Stream` propagates the error
  as its return value.
- The callback is called synchronously within `Stream`; it is safe to update
  shared state only if that state is otherwise protected.
- Context cancellation is checked between tokens inside `Stream` independently
  of the callback; callers do not need to check `ctx.Done()` inside the callback
  (though doing so is harmless).

---

## Built-in Providers

### Anthropic (`bob/provider/anthropic`)

Backed by the official [anthropic-sdk-go](https://github.com/anthropics/anthropic-sdk-go)
client. Uses the Anthropic Messages API with streaming.

**Authentication:** set the `ANTHROPIC_API_KEY` environment variable. Bob
refuses to start with `BOB_PROVIDER=anthropic` and an empty key.

**Supported models (v1):**

| Model name              | Notes            |
|-------------------------|------------------|
| `claude-opus-4-5`       | Highest capability |
| `claude-sonnet-4-5`     | Default           |
| `claude-haiku-3-5`      | Fastest           |

The default `MaxTokens` when `Request.MaxTokens` is `0` is `4096`.

**Tools:** The Anthropic provider converts `sdk.Tool` values to
`anthropic.ToolParam`. The `input_schema` field must be a valid JSON Schema
`"object"` type with a `properties` map.

**Instantiation:**

```go
import "github.com/mattdurham/bob/bob/provider/anthropic"

p := anthropic.New(os.Getenv("ANTHROPIC_API_KEY"))
```

For tests, use `anthropic.NewWithTransport(apiKey, transport)` to inject a
custom `http.RoundTripper`.

---

## Runtime Selection

Bob reads provider and model from environment variables at startup.

| Variable              | Default              | Description                         |
|-----------------------|----------------------|-------------------------------------|
| `BOB_PROVIDER`        | `anthropic`          | Provider name to use.               |
| `BOB_MODEL`           | `claude-sonnet-4-5`  | Default model. Can be overridden at runtime with `/model <name>`. |
| `ANTHROPIC_API_KEY`   | (none)               | Required when `BOB_PROVIDER=anthropic`. |
| `BOB_EXTENSIONS_DIR`  | (none)               | Directory scanned for `.wasm` files.|

The active model can be changed at runtime without restarting using the `/model`
slash command:

```
/model claude-haiku-3-5
```

---

## Custom Provider

Implement the `Provider` interface and pass the value to `harness.New`:

```go
import (
    "github.com/mattdurham/bob/bob/harness"
    "github.com/mattdurham/bob/bob/extension"
)

type MyProvider struct{}

func (p *MyProvider) Name() string          { return "myprovider" }
func (p *MyProvider) Models() []string      { return []string{"my-model-v1"} }
func (p *MyProvider) Stream(ctx context.Context, req provider.Request, fn provider.StreamCallback) error {
    // Iterate tokens and call fn(token) for each.
    return nil
}

// Wire it up:
host := extension.NewHost(nil)
model := harness.New(&MyProvider{}, host)
```

No registration step is needed. The harness uses whatever provider is passed to
`harness.New(p, h)`.

---

## Mock Provider (`bob/provider/mock`)

`bob/provider/mock` contains a scripted provider for use in tests. It emits a
configured list of tokens and optionally returns a terminal error.

```go
import "github.com/mattdurham/bob/bob/provider/mock"

p := &mock.Provider{
    Tokens: []string{"Hello", ", ", "world", "!"},
    Err:    nil, // returned after last token
}
```

**Fields:**

| Field            | Type              | Description                                                  |
|------------------|-------------------|--------------------------------------------------------------|
| `Tokens`         | `[]string`        | Tokens emitted in order.                                     |
| `Err`            | `error`           | Returned from `Stream` after all tokens are emitted.         |
| `StreamErr`      | `error`           | Returned mid-stream (after `StreamErrAfter` tokens).         |
| `StreamErrAfter` | `int`             | Token index at which `StreamErr` is returned.                |
| `CallCount`      | `int`             | Incremented on each `Stream` call. Useful for assertions.    |
| `LastRequest`    | `provider.Request`| The most recent request; inspect in assertions.              |

Context cancellation is honoured between tokens. The mock provider respects
`ctx.Done()` in the same way as the real providers.
