# harness — Interface Contracts and Behavioral Invariants

## 1. Model State Machine

The `Model` has two mutually exclusive streaming states: **idle** and **streaming**.

- **Transition idle → streaming:** triggered by `SubmitMsg` when `streaming == false`. `startStream` sets `streaming = true`, `statusBar.statuses["stream"] = "working."` (initial value), and assigns `cancelStream`. The `streamTickMsg` handler subsequently updates the value to an animated `"working<dots> <elapsed>"` string every 400 ms.
- **Transition streaming → idle:** triggered by `StreamDoneMsg`. `streaming` is set to `false` and `cancelStream` is set to `nil`.
- **No other message may change the streaming state.**

```
idle ──(SubmitMsg)──▶ streaming ──(StreamDoneMsg)──▶ idle
```

## 2. cancelStream Invariant

- `cancelStream` is `*context.CancelFunc` (a pointer to a cancel function).
- It is **non-nil only while streaming** — set in `startStream`, cleared (set to `nil`) on `StreamDoneMsg`.
- It is stored as a pointer so it survives bubbletea value copies of `Model`.
- Callers always dereference before invoking: `(*m.cancelStream)()`.
- `abortStreamMsg` and `ctrl+c` (while streaming) are the only paths that invoke the cancel function.

## 3. History Invariant

- **User messages** are appended to `m.history` in `startStream`, synchronously with the `SubmitMsg` update cycle.
- **Assistant messages** are appended to `m.history` via `addAssistantMsgToHistoryMsg`, which is sent by the streaming goroutine *before* `StreamDoneMsg`. This guarantees that history is correct when `after_provider_response` is dispatched.
- `clearMsg` resets `m.history` to `nil`.
- History is never mutated concurrently — the goroutine captures a snapshot (`append([]sdk.Message(nil), m.history...)`) and sends messages back via `prog.Send`.

## 4. SubmitMsg While Streaming

- If `m.streaming == true` when a `SubmitMsg` arrives, it is **silently dropped** — no queuing, no error, no notification.
- The check is: `if m.streaming { return m, nil }`.

## 5. ctrl+c Behaviour

- **While streaming** (`m.streaming == true` and `m.cancelStream != nil`): invokes `(*m.cancelStream)()` and returns `(m, nil)` — no `tea.Quit`.
- **While idle** (`m.streaming == false`): returns `(m, tea.Quit)`.
- `ctrl+q` always returns `(m, tea.Quit)` regardless of streaming state.

## 6. AltScreen

- `View()` always enables AltScreen by setting `v.AltScreen = true` on the returned `tea.View`.
- This is set on the `View` struct (bubbletea v2 API), not via a `tea.EnterAltScreen` program option.

## 7. ChatView

- `AppendToken(token)`: concatenates `token` to `current`, calls `refreshContent()`, scrolls to bottom.
- `FinalizeMessage()`: if `current == ""` it is a no-op; otherwise appends `chatMessage{role: RoleAssistant, content: current}` to `messages`, resets `current = ""`, calls `refreshContent()`.
- `Clear()`: sets `messages = nil` and `current = ""`, calls `refreshContent()`.
- `refreshContent()`: rebuilds the viewport content from `messages` followed by the in-progress `current` (if non-empty).
- Message order in `messages` reflects insertion order (append-only).

## 8. InputArea

- **Enter key** submits the trimmed content:
  - Empty content → no-op.
  - `/prefix` (starts with `/`) → parses as command: emits `CommandMsg{Name, Args}`.
  - Plain text → emits `SubmitMsg{Content}`.
  - In both cases the textarea is reset before emitting.
- **Shift+Enter** inserts a newline (overriding the default Enter binding in the textarea).
- **Esc key (first press):** resets the textarea (clears content), sets internal `lastWasEsc = true`. Does not emit any message.
- **Esc key (second consecutive press):** resets the textarea, clears `lastWasEsc`, and emits `abortStreamMsg{}` to cancel any active stream.
- **Any non-Esc keypress:** clears `lastWasEsc` (no abort will fire on the next Esc).

## 9. CommandRegistry

- **Unknown command:** `Dispatch` returns a `tea.Cmd` that produces `NotifyMsg{Text: "unknown command: /<name>"}`.
- **Duplicate registration:** silently overwrites the existing entry with the same name.
- **`/help` handling:** intercepted in `Model.Update` before reaching `Registry.Dispatch` — displays `commands.HelpText()` as a notification in the chat. The `/help` command is also registered in the registry (returns a generic message), but the Model-level intercept takes precedence.
- **`/clear`:** emits `clearMsg{}`.
- **`/reload`:** emits `ReloadMsg{}`.
- **`/model <name>`:** emits `setModelMsg{Model: name}`; with no args emits `NotifyMsg` with usage hint.

## 10. StatusBar

- Statuses are stored in a `map[string]string` keyed by an arbitrary string key.
- **"stream" key:** set to `"working."` when a stream starts (in `startStream`); updated every 400 ms to `"working<dots> <elapsed>"` by the `streamTickMsg` handler; deleted from the map on `StreamDoneMsg` (success or `context.Canceled`); set to `"error"` on non-canceled errors.
- `StatusUpdateMsg` sets or overwrites the value for `Key`.
- Keys are rendered in sorted order.

## 11. Extension Callbacks (SetProgram)

`SetProgram` wires four callbacks on `extHost`:

| Callback | Action |
|---|---|
| `OnSetStatus` | Sends `StatusUpdateMsg{Key, Value}` via `prog.Send` |
| `OnNotify` | Sends `NotifyMsg{Text}` via `prog.Send` |
| `OnSendMessage` | Sends `SubmitMsg{Content}` via `prog.Send` |
| `OnAbort` | Sends `abortStreamMsg{}` via `prog.Send` |

`abortStreamMsg` causes `Model.Update` to invoke `(*m.cancelStream)()` on the **live** model's cancel function (not a stale closure copy).

## 12. SetProgram Threading Contract

- `SetProgram` **must be called before** `prog.Run()`.
- It is **not thread-safe** — no synchronization is provided.
- After `prog.Run()` starts, only `prog.Send` is safe to call from goroutines.

## 13. SetActiveModel

- `SetActiveModel(model string)` sets both `m.activeModel` and `m.statusBar.modelName`.
- If `model == ""`, it is a **no-op** (neither field is changed).
- Intended for use after `New` and before `prog.Run()` to honour a user-configured model override.

## 14. startStream Goroutine Safety

- The streaming goroutine captures `priorHistory` — a snapshot of `m.history` **without** the current user message (`append([]sdk.Message(nil), m.history[:len(m.history)-1]...)`). The current user message is passed separately as `Prompt: content` to `AgentStreamCall`; only prior turns go in `Messages`.
- It also captures `langModel`, `extHost`, `activeModel`, and `prog` by value — none of these are mutated by the goroutine.
- All results are returned to the bubbletea event loop via `prog.Send` (thread-safe).

## 15. LanguageModel Contract

- `Model.langModel` is `fantasy.LanguageModel` (`charm.land/fantasy`) — the AI SDK interface.
- `New(langModel fantasy.LanguageModel, provName string, h *extension.Host) Model` is the
  sole constructor. `langModel` may be `nil` (for tests not exercising streaming).
- When `langModel == nil`, `startStream` immediately cancels the context and returns
  `StreamDoneMsg{Err: fmt.Errorf("no language model configured")}`.
- `Model.provName` stores the provider identifier string (e.g. `"anthropic"`) for status
  bar display. It is set at construction and never changed.
- `activeModel` is initialised from `langModel.Model()` in `New()` when `langModel != nil`.
- Streaming is performed via `fantasy.NewAgent(langModel).Stream(ctx, fantasy.AgentStreamCall{...})`.
- Text tokens arrive via `AgentStreamCall.OnTextDelta`, which sends `TokenMsg{Token: text}`
  via `prog.Send`. Empty tokens are skipped.
- Conversation history (`m.history []sdk.Message`) is converted to `[]fantasy.Message` at
  the time each stream starts, using `sdkToFantasyMessages`. This is a text-only conversion.

## 16. Message Conversion

- Message conversion: `sdk.Message` (Role+Content) is converted to `fantasy.Message` (MessageRole + `[]MessagePart{TextPart}`) by `sdkToFantasyMessages` before passing to `agent.Stream`.
- The conversion is text-only: only the `Content` field is used; no tool calls or multi-part messages are produced.

## 17. OnTextDelta Token Delivery

- `OnTextDelta` is the token delivery mechanism: `AgentStreamCall.OnTextDelta` fires per token delta; it returns `ctx.Err()` on cancellation.
- Non-empty token deltas are forwarded as `TokenMsg{Token: text}` via `prog.Send`.
- Returning a non-nil error from `OnTextDelta` signals cancellation to the fantasy agent loop.
