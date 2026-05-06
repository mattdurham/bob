package harness

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestRegistry_Register_And_List(t *testing.T) {
	r := NewRegistry()
	r.Register(Command{Name: "foo", Desc: "do foo", Handler: func(_ []string) tea.Cmd { return nil }})
	r.Register(Command{Name: "bar", Desc: "do bar", Handler: func(_ []string) tea.Cmd { return nil }})

	cmds := r.List()
	if len(cmds) != 2 {
		t.Errorf("expected 2 commands, got %d", len(cmds))
	}
	// Should be sorted.
	if cmds[0].Name != "bar" || cmds[1].Name != "foo" {
		t.Errorf("unexpected order: %v %v", cmds[0].Name, cmds[1].Name)
	}
}

func TestRegistry_Dispatch_KnownCommand(t *testing.T) {
	r := NewRegistry()
	called := false
	r.Register(Command{
		Name: "greet",
		Desc: "say hello",
		Handler: func(args []string) tea.Cmd {
			called = true
			return nil
		},
	})

	cmd := r.Dispatch("greet", nil)
	if cmd != nil {
		cmd() // Execute the returned Cmd.
	}
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestRegistry_Dispatch_UnknownCommand(t *testing.T) {
	r := NewRegistry()
	cmd := r.Dispatch("nonexistent", nil)
	if cmd == nil {
		t.Fatal("expected non-nil Cmd for unknown command")
	}
	msg := cmd()
	notify, ok := msg.(NotifyMsg)
	if !ok {
		t.Fatalf("expected NotifyMsg, got %T", msg)
	}
	if notify.Text == "" {
		t.Error("expected non-empty error message")
	}
}

func TestBuiltinHelp(t *testing.T) {
	r := NewRegistry()
	registerBuiltins(r)
	// /help is handled directly in model.go, but let's also confirm it's registered.
	cmds := r.List()
	names := make(map[string]bool, len(cmds))
	for _, c := range cmds {
		names[c.Name] = true
	}
	for _, expected := range []string{"help", "clear", "reload", "model"} {
		if !names[expected] {
			t.Errorf("expected builtin command %q to be registered", expected)
		}
	}
}

func TestBuiltinClear_EmitsMsg(t *testing.T) {
	r := NewRegistry()
	registerBuiltins(r)

	cmd := r.Dispatch("clear", nil)
	if cmd == nil {
		t.Fatal("expected non-nil Cmd")
	}
	msg := cmd()
	if _, ok := msg.(clearMsg); !ok {
		t.Errorf("expected clearMsg, got %T", msg)
	}
}

func TestBuiltinReload_EmitsMsg(t *testing.T) {
	r := NewRegistry()
	registerBuiltins(r)

	cmd := r.Dispatch("reload", nil)
	if cmd == nil {
		t.Fatal("expected non-nil Cmd")
	}
	msg := cmd()
	if _, ok := msg.(ReloadMsg); !ok {
		t.Errorf("expected ReloadMsg, got %T", msg)
	}
}

func TestBuiltinModel_EmitsMsg(t *testing.T) {
	r := NewRegistry()
	registerBuiltins(r)

	cmd := r.Dispatch("model", []string{"claude-haiku-3-5"})
	if cmd == nil {
		t.Fatal("expected non-nil Cmd")
	}
	msg := cmd()
	setModel, ok := msg.(setModelMsg)
	if !ok {
		t.Fatalf("expected setModelMsg, got %T", msg)
	}
	if setModel.Model != "claude-haiku-3-5" {
		t.Errorf("expected model name %q, got %q", "claude-haiku-3-5", setModel.Model)
	}
}

func TestBuiltinModel_NoArgs(t *testing.T) {
	r := NewRegistry()
	registerBuiltins(r)

	cmd := r.Dispatch("model", nil)
	if cmd == nil {
		t.Fatal("expected non-nil Cmd")
	}
	msg := cmd()
	if _, ok := msg.(NotifyMsg); !ok {
		t.Errorf("expected NotifyMsg for missing model arg, got %T", msg)
	}
}

func TestRegistry_ExtensionCommand_Callable(t *testing.T) {
	r := NewRegistry()
	var gotArgs []string
	r.Register(Command{
		Name: "ext-hello",
		Desc: "extension hello command",
		Handler: func(args []string) tea.Cmd {
			gotArgs = args
			return nil
		},
	})

	r.Dispatch("ext-hello", []string{"world"})
	if len(gotArgs) != 1 || gotArgs[0] != "world" {
		t.Errorf("expected args [world], got %v", gotArgs)
	}
}
