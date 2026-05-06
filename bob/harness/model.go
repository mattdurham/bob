package harness

// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/fantasy"
	"github.com/mattdurham/bob/bob/extension"
	"github.com/mattdurham/bob/bob/sdk"
)

// Model is the root bubbletea v2 model for the bob TUI.
type Model struct {
	chat      ChatView
	input     InputArea
	statusBar StatusBar
	commands  *Registry

	langModel fantasy.LanguageModel
	provName  string // display identifier for status bar (e.g. "anthropic")
	extHost   *extension.Host

	history     []sdk.Message
	streaming   bool
	streamStart time.Time
	activeModel string

	// cancelStream is set when a stream is in progress.
	// It is stored as a pointer to a function so it survives value copies.
	cancelStream *context.CancelFunc

	width, height int

	// Loaded extension paths for reload.
	extPaths []string

	// program is set after the bubbletea program starts so goroutines can
	// send messages back. Set via SetProgram.
	program *tea.Program
}

const inputAreaHeight = 5
const statusBarHeight = 1

// New creates a Model wired to the given language model and extension host.
func New(langModel fantasy.LanguageModel, provName string, h *extension.Host) Model {
	modelName := ""
	if langModel != nil {
		modelName = langModel.Model()
	}

	m := Model{
		chat:        NewChatView(80, 20),
		input:       NewInputArea(80),
		statusBar:   NewStatusBar(provName, modelName),
		commands:    NewRegistry(),
		langModel:   langModel,
		provName:    provName,
		extHost:     h,
		activeModel: modelName,
	}

	registerBuiltins(m.commands)
	return m
}

// sdkToFantasyMessages converts the harness conversation history to fantasy message format.
// Only text content is supported; this is a text-only conversion.
func sdkToFantasyMessages(msgs []sdk.Message) []fantasy.Message {
	result := make([]fantasy.Message, 0, len(msgs))
	for _, m := range msgs {
		var role fantasy.MessageRole
		switch m.Role {
		case sdk.RoleUser:
			role = fantasy.MessageRoleUser
		case sdk.RoleAssistant:
			role = fantasy.MessageRoleAssistant
		default:
			continue
		}
		result = append(result, fantasy.Message{
			Role:    role,
			Content: []fantasy.MessagePart{fantasy.TextPart{Text: m.Content}},
		})
	}
	return result
}

// SetProgram stores the bubbletea program reference so goroutines can call Send.
func (m *Model) SetProgram(p *tea.Program) {
	m.program = p
	if m.extHost != nil {
		m.extHost.OnSetStatus = func(k, v string) {
			p.Send(StatusUpdateMsg{Key: k, Value: v})
		}
		m.extHost.OnNotify = func(text string) {
			p.Send(NotifyMsg{Text: text})
		}
		m.extHost.OnSendMessage = func(msg sdk.Message) {
			p.Send(SubmitMsg{Content: msg.Content})
		}
		m.extHost.OnAbort = func() {
			p.Send(abortStreamMsg{})
		}
	}
}

// SetActiveModel pre-sets the active model before the bubbletea program starts.
// Call this after New and before prog.Run() to honour a user-configured model.
func (m *Model) SetActiveModel(model string) {
	if model == "" {
		return
	}
	m.activeModel = model
	m.statusBar.modelName = model
}

// SetExtensionPaths sets the paths used by /reload.
func (m *Model) SetExtensionPaths(paths []string) {
	m.extPaths = paths
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.input.ta.Focus(),
		m.cmdDispatchSessionStart(),
	)
}

// cmdDispatchSessionStart sends EventSessionStart to all extensions asynchronously.
func (m Model) cmdDispatchSessionStart() tea.Cmd {
	if m.extHost == nil {
		return nil
	}
	extHost := m.extHost
	return func() tea.Msg {
		payload, _ := json.Marshal(sdk.SessionStartPayload{Reason: "new_session"})
		evt := sdk.Event{Type: sdk.EventSessionStart, Payload: payload}
		results, err := extHost.DispatchEvent(context.Background(), evt)
		return ExtensionEventResultMsg{Results: results, Err: err}
	}
}

// Update handles all incoming messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		chatHeight := msg.Height - inputAreaHeight - statusBarHeight
		if chatHeight < 1 {
			chatHeight = 1
		}
		m.chat.SetSize(msg.Width, chatHeight)
		m.input.SetWidth(msg.Width)
		m.statusBar.SetWidth(msg.Width)
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			if m.streaming && m.cancelStream != nil {
				(*m.cancelStream)()
				return m, nil
			}
			return m, tea.Quit
		case "ctrl+q":
			return m, tea.Quit
		}

	case TokenMsg:
		m.chat.AppendToken(msg.Token)
		return m, nil

	case streamTickMsg:
		if m.streaming {
			since := time.Since(m.streamStart)
			dots := strings.Repeat(".", int(since/400/time.Millisecond)%3+1)
			m.statusBar.statuses["stream"] = fmt.Sprintf("working%-3s %s", dots, formatElapsed(since))
			cmds = append(cmds, tea.Tick(400*time.Millisecond, func(time.Time) tea.Msg { return streamTickMsg{} }))
		}
		return m, tea.Batch(cmds...)

	case StreamDoneMsg:
		m.streaming = false
		m.cancelStream = nil
		if msg.Err != nil && !errors.Is(msg.Err, context.Canceled) {
			m.chat.AddNotification(fmt.Sprintf("Error: %v", msg.Err))
			m.statusBar.statuses["stream"] = "error"
		} else {
			delete(m.statusBar.statuses, "stream")
		}
		m.chat.FinalizeMessage()
		// Finalise assistant message in history.
		// (History is maintained by addAssistantMessage cmd below.)
		cmds = append(cmds, m.cmdDispatchAfterProviderResponse())
		return m, tea.Batch(cmds...)

	case addAssistantMsgToHistoryMsg:
		m.history = append(m.history, sdk.Message{Role: sdk.RoleAssistant, Content: msg.content})
		return m, nil

	case ExtensionEventResultMsg:
		for _, r := range msg.Results {
			if r.Error != "" {
				m.chat.AddNotification(fmt.Sprintf("Extension error: %s", r.Error))
			}
		}
		return m, nil

	case ReloadMsg:
		return m, m.cmdReloadExtensions()

	case NotifyMsg:
		m.chat.AddNotification(msg.Text)
		return m, nil

	case StatusUpdateMsg:
		m.statusBar, _ = m.statusBar.Update(msg)
		return m, nil

	case SubmitMsg:
		if m.streaming {
			return m, nil
		}
		return m.startStream(msg.Content)

	case CommandMsg:
		// /help is special: show command list in chat
		if msg.Name == "help" {
			m.chat.AddNotification(m.commands.HelpText())
			return m, nil
		}
		return m, m.commands.Dispatch(msg.Name, msg.Args)

	case clearMsg:
		m.chat.Clear()
		m.history = nil
		return m, nil

	case setModelMsg:
		m.activeModel = msg.Model
		m.statusBar.modelName = msg.Model
		m.chat.AddNotification(fmt.Sprintf("Model set to: %s", msg.Model))
		return m, nil

	case abortStreamMsg:
		if m.cancelStream != nil {
			(*m.cancelStream)()
		}
		return m, nil
	}

	// Forward to sub-models.
	var inputCmd tea.Cmd
	m.input, inputCmd = m.input.Update(msg)
	cmds = append(cmds, inputCmd)

	var chatCmd tea.Cmd
	m.chat, chatCmd = m.chat.Update(msg)
	cmds = append(cmds, chatCmd)

	return m, tea.Batch(cmds...)
}

// addAssistantMsgToHistoryMsg carries a finalised assistant message to add to history.
type addAssistantMsgToHistoryMsg struct{ content string }

// startStream processes user input: adds to history, starts streaming.
func (m Model) startStream(content string) (tea.Model, tea.Cmd) {
	m.chat.AddUserMessage(content)
	m.history = append(m.history, sdk.Message{Role: sdk.RoleUser, Content: content})
	m.streaming = true
	m.streamStart = time.Now()
	m.statusBar.statuses["stream"] = "working."

	ctx, cancel := context.WithCancel(context.Background())
	m.cancelStream = &cancel

	langModel := m.langModel
	extHost := m.extHost
	activeModel := m.activeModel
	// Snapshot history WITHOUT the current user message — the current message
	// goes in Prompt; prior history goes in Messages.
	priorHistory := append([]sdk.Message(nil), m.history[:len(m.history)-1]...)
	prog := m.program

	cmd := func() tea.Msg {
		// Dispatch before_agent_start.
		if extHost != nil {
			payload, _ := json.Marshal(sdk.BeforeAgentStartPayload{
				Prompt: content,
			})
			evt := sdk.Event{Type: sdk.EventBeforeAgentStart, Payload: payload}
			_, _ = extHost.DispatchEvent(ctx, evt)
		}

		// Dispatch before_provider_request.
		if extHost != nil {
			allHistory := append(priorHistory, sdk.Message{Role: sdk.RoleUser, Content: content})
			payload, _ := json.Marshal(sdk.BeforeProviderRequestPayload{
				Messages: allHistory,
				Model:    activeModel,
			})
			evt := sdk.Event{Type: sdk.EventBeforeProviderRequest, Payload: payload}
			_, _ = extHost.DispatchEvent(ctx, evt)
		}

		if langModel == nil {
			cancel()
			return StreamDoneMsg{Err: fmt.Errorf("no language model configured")}
		}

		var collected strings.Builder
		agent := fantasy.NewAgent(langModel)
		_, err := agent.Stream(ctx, fantasy.AgentStreamCall{
			Messages: sdkToFantasyMessages(priorHistory),
			Prompt:   content,
			OnTextDelta: func(id, text string) error {
				if text == "" {
					return nil
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
				collected.WriteString(text)
				if prog != nil {
					prog.Send(TokenMsg{Token: text})
				}
				return nil
			},
		})
		cancel()

		// Send history update before StreamDoneMsg so history is correct when
		// after_provider_response is dispatched.
		// Only send if content was actually collected — avoid empty assistant messages
		// on early cancellation (e.g. ctrl+c before any tokens arrive).
		if (err == nil || errors.Is(err, context.Canceled)) && collected.Len() > 0 {
			if prog != nil {
				prog.Send(addAssistantMsgToHistoryMsg{content: collected.String()})
			}
		}

		return StreamDoneMsg{Err: err}
	}

	tick := tea.Tick(400*time.Millisecond, func(time.Time) tea.Msg { return streamTickMsg{} })
	return m, tea.Batch(cmd, tick)
}

func (m Model) cmdDispatchAfterProviderResponse() tea.Cmd {
	if m.extHost == nil {
		return nil
	}
	extHost := m.extHost
	return func() tea.Msg {
		payload, _ := json.Marshal(sdk.AfterProviderResponsePayload{})
		evt := sdk.Event{Type: sdk.EventAfterProviderResponse, Payload: payload}
		results, err := extHost.DispatchEvent(context.Background(), evt)
		return ExtensionEventResultMsg{Results: results, Err: err}
	}
}

func (m Model) cmdReloadExtensions() tea.Cmd {
	if m.extHost == nil {
		return func() tea.Msg { return NotifyMsg{Text: "No extension host configured."} }
	}
	paths := m.extPaths
	extHost := m.extHost
	return func() tea.Msg {
		if err := extHost.Reload(context.Background(), paths); err != nil {
			return NotifyMsg{Text: fmt.Sprintf("Reload error: %v", err)}
		}
		return NotifyMsg{Text: "Extensions reloaded."}
	}
}

// View renders the full TUI.
func (m Model) View() tea.View {
	var sb strings.Builder

	// Chat view.
	sb.WriteString(m.chat.View())
	sb.WriteString("\n")

	// Status bar.
	sb.WriteString(m.statusBar.View())
	sb.WriteString("\n")

	// Input area.
	sb.WriteString(m.input.View())

	v := tea.NewView(sb.String())
	v.AltScreen = true
	return v
}
