// shipmate is a hook-based OTEL annotation daemon for Claude Code.
//
// Subcommands:
//
//	shipmate start  --upstream <url> [--session-id <id>] [--headers K=V,...] [--log-dir <dir>]
//	shipmate record [--session-id <id>]   # reads hook stdin JSON
//	shipmate memory --session-id <id> <text>
//	shipmate stop   [--session-id <id>]   # reads hook stdin JSON when --session-id omitted
//
// start and stop read session_id from hook stdin JSON when --session-id is not provided.
//
// Internal (not user-facing):
//
//	shipmate start --daemon --session-id <id> --upstream <url> [--headers K=V,...]
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/mattdurham/bob/internal/shipmate/client"
	"github.com/mattdurham/bob/internal/shipmate/daemon"
	"github.com/mattdurham/bob/internal/shipmate/hook"
	"github.com/mattdurham/bob/internal/shipmate/recorder"
	"github.com/mattdurham/bob/internal/shipmate/scorer"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

// execCommand is a package-level var so tests can stub out exec.Command.
var execCommand = exec.Command

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "start":
		runStart(os.Args[2:])
	case "record":
		runRecord(os.Args[2:])
	case "memory":
		runMemory(os.Args[2:])
	case "stop":
		runStop(os.Args[2:])
	case "configure":
		runConfigure(os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `shipmate — hook-based OTEL annotation daemon for Claude Code

Usage:
  shipmate start      --upstream <url> [--session-id <id>] [--headers K=V,...] [--log-dir <dir>]
  shipmate record     [--session-id <id>]
  shipmate memory     --session-id <id> <text>
  shipmate stop       [--session-id <id>]
  shipmate configure  --upstream <url> [--headers K=V,...] [--settings <path>]

start and stop read session_id from hook stdin JSON when --session-id is not provided.
configure registers shipmate hooks in ~/.claude/settings.json.`)
}

// runConfigure registers shipmate hooks in ~/.claude/settings.json.
func runConfigure(args []string) {
	fs := flag.NewFlagSet("configure", flag.ExitOnError)
	upstream := fs.String("upstream", "http://localhost:4318", "Upstream OTLP HTTP endpoint")
	headers := fs.String("headers", "", "Optional comma-separated Key=Value headers for upstream")
	settingsPath := fs.String("settings", "", "Path to settings.json (default ~/.claude/settings.json)")
	fs.Parse(args) //nolint:errcheck

	if *settingsPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("shipmate configure: get home dir: %v", err)
		}
		*settingsPath = filepath.Join(home, ".claude", "settings.json")
	}

	self, err := os.Executable()
	if err != nil {
		log.Fatalf("shipmate configure: resolve executable: %v", err)
	}

	// Read existing settings or start fresh.
	var settings map[string]any
	data, err := os.ReadFile(*settingsPath)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("shipmate configure: read %s: %v", *settingsPath, err)
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &settings); err != nil {
			log.Fatalf("shipmate configure: parse %s: %v", *settingsPath, err)
		}
	}
	if settings == nil {
		settings = map[string]any{}
	}

	// Remove old MCP server entry if present.
	if mcpServers, ok := settings["mcpServers"].(map[string]any); ok {
		if _, has := mcpServers["shipmate"]; has {
			delete(mcpServers, "shipmate")
			fmt.Println("  removed old mcpServers.shipmate entry")
		}
	}

	// Remove old OTEL proxy env vars.
	oldEnvKeys := []string{
		"OTEL_TRACES_EXPORTER",
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_PROTOCOL",
	}
	if env, ok := settings["env"].(map[string]any); ok {
		for _, key := range oldEnvKeys {
			if _, has := env[key]; has {
				delete(env, key)
				fmt.Printf("  removed env.%s (was for old gRPC proxy)\n", key)
			}
		}
	}

	// Ensure telemetry env vars are set.
	env, _ := settings["env"].(map[string]any)
	if env == nil {
		env = map[string]any{}
	}
	if _, ok := env["CLAUDE_CODE_ENABLE_TELEMETRY"]; !ok {
		env["CLAUDE_CODE_ENABLE_TELEMETRY"] = "1"
	}
	if _, ok := env["CLAUDE_CODE_ENHANCED_TELEMETRY_BETA"]; !ok {
		env["CLAUDE_CODE_ENHANCED_TELEMETRY_BETA"] = "1"
	}
	settings["env"] = env

	// Build hook commands.
	startCmd := self + " start --upstream " + *upstream
	if *headers != "" {
		startCmd += " --headers " + *headers
	}
	stopCmd := self + " stop"
	recordCmd := self + " record"

	hook := func(cmd string) map[string]any {
		return map[string]any{"hooks": []any{map[string]any{"type": "command", "command": cmd}}}
	}

	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
	}
	hooks["SessionStart"] = []any{hook(startCmd)}
	hooks["Stop"] = []any{hook(stopCmd)}
	for _, event := range []string{"PostToolUse", "SubagentStart", "SubagentStop", "TaskCreated", "TaskCompleted"} {
		hooks[event] = []any{hook(recordCmd)}
	}
	settings["hooks"] = hooks

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		log.Fatalf("shipmate configure: marshal settings: %v", err)
	}
	out = append(out, '\n')

	if err := os.WriteFile(*settingsPath, out, 0o644); err != nil {
		log.Fatalf("shipmate configure: write %s: %v", *settingsPath, err)
	}

	fmt.Printf("  shipmate hooks registered in %s\n", *settingsPath)
	fmt.Printf("  upstream endpoint: %s\n", *upstream)
	fmt.Println("  restart Claude Code to activate")
}

// sockPath returns the Unix socket path for a given session ID.
// It rejects session IDs that contain path separators to prevent path traversal.
func sockPath(sessionID string) (string, error) {
	if strings.ContainsAny(sessionID, `/\`) {
		return "", fmt.Errorf("invalid session ID: contains path separator")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, ".local", "share", "shipmate", sessionID+".sock"), nil
}

// parseHeaders parses "Key=Value,Key2=Value2" into a string map.
// Malformed pairs are silently skipped.
func parseHeaders(raw string) map[string]string {
	headers := map[string]string{}
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}
		headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return headers
}

// runStart validates flags and re-execs self with --daemon to daemonize,
// or (when --daemon is set) runs the server loop directly.
func runStart(args []string) {
	fs := flag.NewFlagSet("start", flag.ExitOnError)
	sessionID := fs.String("session-id", "", "Session ID (default: read from hook stdin JSON)")
	upstream := fs.String("upstream", "", "Upstream OTLP HTTP endpoint, e.g. http://localhost:4318 (required)")
	headers := fs.String("headers", "", "Optional comma-separated Key=Value headers")
	logDir := fs.String("log-dir", "", "Directory for daemon log file (default ~/.local/share/shipmate)")
	isDaemon := fs.Bool("daemon", false, "Internal flag: already running as daemon child")
	fs.Parse(args) //nolint:errcheck

	if *isDaemon {
		// We are the daemon child — run the server loop.
		runDaemon(*sessionID, *upstream, *headers)
		return
	}

	// When --session-id is omitted, read it from hook stdin JSON.
	if *sessionID == "" {
		cmd, err := hook.ParseHookInput(os.Stdin)
		if err != nil {
			log.Fatalf("shipmate start: parse stdin for session_id: %v", err)
		}
		*sessionID = cmd.SessionID
	}
	if *sessionID == "" {
		log.Fatal("shipmate start: --session-id is required (or provide hook JSON on stdin)")
	}
	if *upstream == "" {
		log.Fatal("shipmate start: --upstream is required")
	}

	// Ensure log + socket directories exist.
	dir := *logDir
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("shipmate start: cannot determine home directory: %v", err)
		}
		dir = filepath.Join(home, ".local", "share", "shipmate")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatalf("shipmate start: mkdir %s: %v", dir, err)
	}

	lp := filepath.Join(dir, *sessionID+".log")
	logFile, err := os.OpenFile(lp, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatalf("shipmate start: open log %s: %v", lp, err)
	}

	// Build the daemon re-exec arguments.
	daemonArgs := []string{
		"start", "--daemon",
		"--session-id", *sessionID,
		"--upstream", *upstream,
	}
	if *headers != "" {
		daemonArgs = append(daemonArgs, "--headers", *headers)
	}
	if *logDir != "" {
		daemonArgs = append(daemonArgs, "--log-dir", *logDir)
	}

	self, err := os.Executable()
	if err != nil {
		log.Fatalf("shipmate start: resolve executable: %v", err)
	}

	cmd := execCommand(self, daemonArgs...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		log.Fatalf("shipmate start: exec daemon: %v", err)
	}
	// Parent exits immediately; child continues as daemon.
	// Close logFile explicitly — os.Exit bypasses defer, and the child has its own copy.
	if err := logFile.Close(); err != nil {
		log.Printf("shipmate start: close log: %v", err)
	}
	os.Exit(0)
}

// runDaemon is the long-running server loop executed in the daemon child process.
func runDaemon(sessionID, upstream, headers string) {
	if sessionID == "" {
		log.Fatal("shipmate daemon: --session-id is required")
	}
	if upstream == "" {
		log.Fatal("shipmate daemon: --upstream is required")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	hdrs := parseHeaders(headers)
	if user := os.Getenv("SHIPMATE_AUTH_USER"); user != "" {
		pass := os.Getenv("SHIPMATE_AUTH_PASSWORD")
		hdrs["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+pass))
	}

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(upstream),
		otlptracehttp.WithHeaders(hdrs),
	}
	if !strings.HasPrefix(upstream, "https://") {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	httpExp, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		log.Fatalf("shipmate daemon: create exporter: %v", err)
	}

	bufExp := scorer.NewBufferingExporter(httpExp)

	rec, err := recorder.New(bufExp)
	if err != nil {
		log.Fatalf("shipmate daemon: create recorder: %v", err)
	}

	sp, err := sockPath(sessionID)
	if err != nil {
		log.Fatalf("shipmate daemon: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(sp), 0o755); err != nil {
		log.Fatalf("shipmate daemon: mkdir %s: %v", filepath.Dir(sp), err)
	}

	log.Printf("shipmate daemon: session=%s socket=%s", sessionID, sp)
	if err := daemon.Serve(ctx, sp, rec, bufExp); err != nil {
		log.Fatalf("shipmate daemon: serve: %v", err)
	}
}

// runRecord reads hook stdin JSON, parses it, and sends a record command to the daemon.
func runRecord(args []string) {
	fs := flag.NewFlagSet("record", flag.ExitOnError)
	sessionIDOverride := fs.String("session-id", "", "Override session ID (default: from stdin JSON)")
	fs.Parse(args) //nolint:errcheck

	cmd, err := hook.ParseHookInput(os.Stdin)
	if err != nil {
		log.Printf("shipmate record: parse stdin: %v (trace lost)", err)
		os.Exit(0)
	}
	if *sessionIDOverride != "" {
		cmd.SessionID = *sessionIDOverride
	}
	if cmd.SessionID == "" {
		log.Printf("shipmate record: no session_id in stdin (trace lost)")
		os.Exit(0)
	}

	sp, err := sockPath(cmd.SessionID)
	if err != nil {
		log.Printf("shipmate record: %v (trace lost)", err)
		os.Exit(0)
	}
	if err := client.Send(sp, cmd); err != nil {
		log.Printf("shipmate record: send: %v", err)
	}
}

// runMemory sends a free-text memory annotation to the daemon.
func runMemory(args []string) {
	fs := flag.NewFlagSet("memory", flag.ExitOnError)
	sessionID := fs.String("session-id", os.Getenv("SHIPMATE_SESSION_ID"), "Session ID")
	fs.Parse(args) //nolint:errcheck

	text := strings.Join(fs.Args(), " ")
	if text == "" {
		log.Printf("shipmate memory: text argument is required")
		os.Exit(0)
	}
	if *sessionID == "" {
		log.Printf("shipmate memory: --session-id or SHIPMATE_SESSION_ID is required")
		os.Exit(0)
	}

	cmd := hook.Command{
		Type:      "memory",
		SessionID: *sessionID,
		Text:      text,
	}
	sp, err := sockPath(*sessionID)
	if err != nil {
		log.Printf("shipmate memory: %v", err)
		os.Exit(0)
	}
	if err := client.Send(sp, cmd); err != nil {
		log.Printf("shipmate memory: send: %v", err)
	}
}

// runStop sends a stop command to the daemon.
func runStop(args []string) {
	fs := flag.NewFlagSet("stop", flag.ExitOnError)
	sessionID := fs.String("session-id", os.Getenv("SHIPMATE_SESSION_ID"), "Session ID (default: read from hook stdin JSON)")
	fs.Parse(args) //nolint:errcheck

	// When --session-id is omitted, read it from hook stdin JSON.
	if *sessionID == "" {
		cmd, err := hook.ParseHookInput(os.Stdin)
		if err != nil {
			log.Printf("shipmate stop: parse stdin for session_id: %v (nothing to stop)", err)
			os.Exit(0)
		}
		*sessionID = cmd.SessionID
	}
	if *sessionID == "" {
		log.Printf("shipmate stop: no session ID — nothing to stop")
		os.Exit(0)
	}

	cmd := hook.Command{
		Type:      "stop",
		SessionID: *sessionID,
	}
	sp, err := sockPath(*sessionID)
	if err != nil {
		log.Printf("shipmate stop: %v", err)
		os.Exit(0)
	}
	if err := client.Send(sp, cmd); err != nil {
		log.Printf("shipmate stop: send: %v", err)
	}
}
