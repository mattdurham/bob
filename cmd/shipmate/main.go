// shipmate is a hook-based OTEL annotation daemon for Claude Code.
//
// Subcommands:
//
//	shipmate start  --session-id <id> --upstream <url> [--headers K=V,...] [--log-dir <dir>] [--transcript-path <path>]
//	shipmate stop   --session-id <id>
//
// Internal (not user-facing):
//
//	shipmate start --daemon --session-id <id> --upstream <url> [--headers K=V,...] [--transcript-path <path>]
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/mattdurham/bob/internal/shipmate/daemon"
	"github.com/mattdurham/bob/internal/shipmate/hook"
	"github.com/mattdurham/bob/internal/shipmate/transcript"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"google.golang.org/grpc/credentials"
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
	case "stop":
		runStop(os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `shipmate — hook-based OTEL annotation daemon for Claude Code

Usage:
  shipmate start  --session-id <id> --upstream <url> [--headers K=V,...] [--log-dir <dir>] [--transcript-path <path>]
  shipmate stop   --session-id <id>`)
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
	sessionID := fs.String("session-id", "", "Session ID (required)")
	upstream := fs.String("upstream", "", "Upstream OTLP gRPC endpoint, e.g. http://localhost:4317 (required)")
	headers := fs.String("headers", "", "Optional comma-separated Key=Value headers")
	logDir := fs.String("log-dir", "", "Directory for daemon log file (default ~/.local/share/shipmate)")
	transcriptPath := fs.String("transcript-path", "", "Path to Claude Code JSONL transcript")
	isDaemon := fs.Bool("daemon", false, "Internal flag: already running as daemon child")
	fs.Parse(args) //nolint:errcheck

	if *isDaemon {
		// We are the daemon child — run the server loop.
		runDaemon(*sessionID, *upstream, *headers, *transcriptPath)
		return
	}

	// When --session-id or --transcript-path is omitted, read from hook stdin JSON.
	if *sessionID == "" || *transcriptPath == "" {
		cmd, err := hook.ParseHookInput(os.Stdin)
		if err != nil {
			log.Fatalf("shipmate start: parse stdin: %v", err)
		}
		if *sessionID == "" {
			*sessionID = cmd.SessionID
		}
		if *transcriptPath == "" {
			*transcriptPath = cmd.Attrs["transcript_path"]
		}
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
	if *transcriptPath != "" {
		daemonArgs = append(daemonArgs, "--transcript-path", *transcriptPath)
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
	if err := logFile.Close(); err != nil {
		log.Printf("shipmate start: close log: %v", err)
	}
	os.Exit(0)
}

// buildHeaders merges explicit headers with Basic auth from env vars.
// SHIPMATE_UPSTREAM_USER + SHIPMATE_UPSTREAM_TOKEN → "Authorization: Basic <base64(user:token)>"
// Explicit --headers take precedence over the constructed Basic auth header.
func buildHeaders(raw string) map[string]string {
	h := parseHeaders(raw)
	user := os.Getenv("SHIPMATE_UPSTREAM_USER")
	token := os.Getenv("SHIPMATE_UPSTREAM_TOKEN")
	if user != "" && token != "" {
		encoded := base64.StdEncoding.EncodeToString([]byte(user + ":" + token))
		if _, exists := h["Authorization"]; !exists {
			h["Authorization"] = "Basic " + encoded
		}
	}
	return h
}

// runDaemon is the long-running server loop executed in the daemon child process.
func runDaemon(sessionID, upstream, headers, transcriptPath string) {
	if sessionID == "" {
		log.Fatal("shipmate daemon: --session-id is required")
	}
	if upstream == "" {
		log.Fatal("shipmate daemon: --upstream is required")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Strip scheme, path, and add default port — gRPC expects host:port only.
	endpoint := upstream
	useTLS := true
	if after, ok := strings.CutPrefix(endpoint, "https://"); ok {
		endpoint = after
	} else if after, ok := strings.CutPrefix(endpoint, "http://"); ok {
		endpoint = after
		useTLS = false
	}
	if i := strings.IndexByte(endpoint, '/'); i >= 0 {
		endpoint = endpoint[:i]
	}
	if !strings.Contains(endpoint, ":") {
		if useTLS {
			endpoint += ":443"
		} else {
			endpoint += ":80"
		}
	}

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithHeaders(buildHeaders(headers)),
	}
	if useTLS {
		opts = append(opts, otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
	} else {
		opts = append(opts, otlptracegrpc.WithInsecure()) //nolint:staticcheck
	}

	grpcExp, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		log.Fatalf("shipmate daemon: create exporter: %v", err)
	}

	te := transcript.NewTurnExporter(transcriptPath, sessionID, grpcExp)

	sp, err := sockPath(sessionID)
	if err != nil {
		log.Fatalf("shipmate daemon: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(sp), 0o755); err != nil {
		log.Fatalf("shipmate daemon: mkdir %s: %v", filepath.Dir(sp), err)
	}

	log.Printf("shipmate daemon: session=%s transcript=%s socket=%s", sessionID, transcriptPath, sp)
	if err := daemon.Serve(ctx, sp, te); err != nil {
		log.Fatalf("shipmate daemon: serve: %v", err)
	}
}

// runStop sends a stop command to the daemon.
func runStop(args []string) {
	fs := flag.NewFlagSet("stop", flag.ExitOnError)
	sessionID := fs.String("session-id", os.Getenv("SHIPMATE_SESSION_ID"), "Session ID")
	fs.Parse(args) //nolint:errcheck

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

	conn, err := net.Dial("unix", sp)
	if err != nil {
		log.Printf("shipmate stop: dial %s: %v", sp, err)
		os.Exit(0)
	}
	defer conn.Close()
	if err := json.NewEncoder(conn).Encode(cmd); err != nil {
		log.Printf("shipmate stop: send: %v", err)
	}
}
