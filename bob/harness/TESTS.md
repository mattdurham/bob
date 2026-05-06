# harness — Test Specifications

## Existing Tests

### model_test.go

| Test | Scenario | Setup | Assertions |
|---|---|---|---|
| `TestModel_Init_ReturnsCmd` | Init returns a non-nil Cmd | `newTestModel()` | `Init()` result is non-nil |
| `TestModel_View_ReturnsNonEmpty` | View produces content | `newTestModel()` | `View().Content != ""` |
| `TestModel_Update_TokenMsg_AppendsToChat` | Tokens accumulate in `chat.current` | `m.streaming = true`; send two `TokenMsg` | `chat.current` equals concatenated tokens after each append |
| `TestModel_Update_StreamDoneMsg_ClearsStreaming` | StreamDoneMsg ends stream | `m.streaming = true`, token appended | `streaming == false` after `StreamDoneMsg{Err: nil}` |
| `TestModel_Update_StreamDoneMsg_Error_ShowsError` | Non-canceled error adds notification | `m.streaming = true` | `streaming == false`; `chat.messages` is non-empty (error notification added) |
| `TestModel_Update_StreamDoneMsg_ContextCanceled_NoError` | `context.Canceled` is not shown as error | `m.streaming = true`, partial token | `streaming == false`; no system notification message in chat |
| `TestModel_Update_ReloadMsg_TriggersExtensionReload` | ReloadMsg returns a Cmd that resolves to NotifyMsg | `newTestModel()` (nil host) | `cmd != nil`; executing cmd yields `NotifyMsg` |
| `TestModel_Update_ClearMsg_ClearsHistory` | clearMsg empties history and chat | history + chat pre-populated | `len(m.history) == 0`; `len(m.chat.messages) == 0` |
| `TestModel_Update_SetModelMsg` | setModelMsg updates activeModel and statusBar | `newTestModel()` | `m.activeModel` and `m.statusBar.modelName` equal new model name |
| `TestModel_Update_CommandMsg_Clear` | /clear command dispatched via registry | history pre-populated | `len(m.history) == 0` after processing |
| `TestModel_Update_CommandMsg_UnknownCommand` | Unknown command yields NotifyMsg | `newTestModel()` | `cmd != nil`; msg is `NotifyMsg` with non-empty text |
| `TestModel_Update_SubmitMsg_StartsStream` | SubmitMsg starts streaming | `newTestModel()` with mock provider | `m.streaming == true`; `cmd != nil` |
| `TestModel_Update_SubmitMsg_IgnoredWhileStreaming` | SubmitMsg dropped while streaming | `m.streaming = true` | No panic; model consistent |
| `TestModel_Update_NotifyMsg` | NotifyMsg adds to chat | `newTestModel()` | `len(m.chat.messages) > 0` |
| `TestModel_Update_StatusUpdateMsg` | StatusUpdateMsg updates statusBar | `newTestModel()` | `m.statusBar.statuses["foo"] == "bar"` |
| `TestModel_Update_WindowSizeMsg` | WindowSizeMsg updates dimensions | `newTestModel()` | `m.width == 120`, `m.height == 40` |
| `TestModel_NilLangModel_StreamError` | nil langModel returns StreamDoneMsg with error | `New(nil, "none", nil)` | `streaming == true` before cmd runs; `StreamDoneMsg.Err != nil` |

---

### chat_test.go

| Test | Scenario | Setup | Assertions |
|---|---|---|---|
| `TestChatView_AppendToken` | Tokens accumulate in `current` | `NewChatView(80, 20)` | `c.current == "hello world"` after two appends |
| `TestChatView_FinalizeMessage_SetsRole` | FinalizeMessage seals message with RoleAssistant | Token appended | 1 message; role == `RoleAssistant`; content matches; `current == ""` |
| `TestChatView_FinalizeMessage_Empty_NoOp` | FinalizeMessage on empty current is no-op | No tokens appended | `len(messages) == 0` |
| `TestChatView_AddUserMessage` | AddUserMessage creates user-role message | `NewChatView(80, 20)` | 1 message; role == `RoleUser`; content matches |
| `TestChatView_MessageOrder` | Messages appear in insertion order | user then assistant | `messages[0].role == RoleUser`; `messages[1].role == RoleAssistant` |
| `TestChatView_Clear` | Clear resets messages and current | user + in-progress token | `len(messages) == 0`; `current == ""` |
| `TestChatView_View_NonEmpty` | View does not panic | messages + in-progress token | No panic |
| `TestChatView_RenderUserMessage_ContainsContent` | Viewport content includes user message text | `AddUserMessage("unique-test-content")` | `vp.View()` contains the unique string |

---

### commands_test.go

| Test | Scenario | Setup | Assertions |
|---|---|---|---|
| `TestRegistry_Register_And_List` | Register two commands; List returns sorted | Two commands registered | `len == 2`; sorted alphabetically |
| `TestRegistry_Dispatch_KnownCommand` | Dispatch calls handler | Handler sets `called = true` | `called == true` |
| `TestRegistry_Dispatch_UnknownCommand` | Dispatch unknown name yields NotifyMsg | Empty registry | `cmd != nil`; msg is `NotifyMsg` with non-empty text |
| `TestBuiltinHelp` | All four builtins are registered | `registerBuiltins(r)` | help, clear, reload, model all present |
| `TestBuiltinClear_EmitsMsg` | /clear emits `clearMsg` | `registerBuiltins(r)` | msg is `clearMsg{}` |
| `TestBuiltinReload_EmitsMsg` | /reload emits `ReloadMsg` | `registerBuiltins(r)` | msg is `ReloadMsg{}` |
| `TestBuiltinModel_EmitsMsg` | /model sets model name | args `["claude-haiku-3-5"]` | msg is `setModelMsg{Model: "claude-haiku-3-5"}` |
| `TestBuiltinModel_NoArgs` | /model with no args yields usage hint | no args | msg is `NotifyMsg` |
| `TestRegistry_ExtensionCommand_Callable` | Extension-registered command receives args | custom handler captures args | `gotArgs == ["world"]` |

---

### integration_test.go

| Test | Scenario | Setup | Assertions |
|---|---|---|---|
| `TestIntegration_FullStreamingFlow` | Full submit → stream → done cycle | mock provider with tokens | streaming true then false; history has 1 user message; `CallCount == 1` |
| `TestIntegration_UserMessageInHistory` | User message recorded immediately on SubmitMsg | mock provider | `history[0].Role == RoleUser`; `history[0].Content == "hello world"`; `chat.messages` has 1 entry |
| `TestIntegration_CtrlC_CancelsStream` | ctrl+c while streaming invokes cancel but does not quit | streaming started | `streaming` remains true (StreamDoneMsg not yet arrived); `cancelStream` still set |
| `TestIntegration_NilExtensionHost_Safe` | nil extension host never panics | `New(p, nil)` | No panic on Init, SubmitMsg, stream execution |
| `TestIntegration_ExtensionHost_NoExtensions_Safe` | Real host with no extensions loaded | `extension.NewHost(nil)` | No panic; `StreamDoneMsg.Err == nil` |

---

## Missing / Recommended Tests

The following scenarios are not currently covered and should be added:

| Priority | Test Name (suggested) | Scenario |
|---|---|---|
| High | `TestModel_CtrlC_Idle_Quits` | ctrl+c while idle (`streaming == false`) returns `tea.Quit` |
| High | `TestModel_CtrlQ_AlwaysQuits` | ctrl+q always returns `tea.Quit` regardless of streaming state |
| High | `TestModel_AbortStreamMsg_CancelsStream` | `abortStreamMsg{}` while streaming invokes cancel without quitting |
| High | `TestModel_AbortStreamMsg_Idle_NoOp` | `abortStreamMsg{}` while idle (cancelStream nil) does not panic |
| High | `TestModel_AddAssistantMsgToHistoryMsg` | Processing `addAssistantMsgToHistoryMsg` appends to `m.history` with `RoleAssistant` |
| Medium | `TestModel_SetActiveModel_SetsFields` | `SetActiveModel("foo")` updates both `m.activeModel` and `m.statusBar.modelName` |
| Medium | `TestModel_SetActiveModel_Empty_NoOp` | `SetActiveModel("")` leaves both fields unchanged |
| Medium | `TestModel_SetProgram_WiresCallbacks` | After `SetProgram`, `extHost.OnAbort()` sends `abortStreamMsg` to program |
| Medium | `TestModel_StreamDoneMsg_SetsStreamStatusError` | On non-canceled error, `statusBar.statuses["stream"] == "error"` |
| Medium | `TestModel_StreamDoneMsg_DeletesStreamStatus` | On success, `statusBar.statuses` does not contain key `"stream"` |
| Medium | `TestModel_CommandMsg_Help_ShowsHelpText` | `CommandMsg{Name: "help"}` adds HelpText notification without delegating to registry |
| Medium | `TestChatView_SetSize_UpdatesDimensions` | `SetSize` changes width/height and calls refreshContent |
| Medium | `TestChatView_AddNotification_SystemRole` | `AddNotification` appends a message with role `"system"` |
| Low | `TestChatView_RenderMessage_LineWrapping` | Content wider than `width` is wrapped correctly (no panic; line fits width) |
| Low | `TestChatView_FinalizeMessage_CalledTwice` | Double finalize: second call is no-op (current is empty after first) |
| Low | `TestInputArea_Enter_Empty_NoOp` | Pressing Enter on empty textarea emits no message |
| Low | `TestInputArea_Enter_Command_EmitsCommandMsg` | `/foo bar` emits `CommandMsg{Name: "foo", Args: ["bar"]}` |
| Low | `TestInputArea_Enter_PlainText_EmitsSubmitMsg` | `hello world` emits `SubmitMsg{Content: "hello world"}` |
| Low | `TestInputArea_Esc_ResetsTextarea` | Esc clears content without emitting a message |
| Low | `TestStatusBar_View_StreamingStatus` | When `statuses["stream"] == "streaming…"`, view contains that string |
| Low | `TestStatusBar_Update_SetsStatus` | `StatusUpdateMsg` sets the keyed value |
| Low | `TestStatusBar_View_SortedKeys` | Multiple status keys appear in sorted order in view output |
