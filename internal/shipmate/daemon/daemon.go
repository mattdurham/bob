// Package daemon implements the shipmate Unix socket server.
// It accepts NDJSON commands from hook clients. The only meaningful command is
// "stop", which triggers a final transcript export with scoring before shutdown.
// Serve returns when ctx is cancelled or a "stop" command is received.
package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/mattdurham/bob/internal/shipmate/hook"
	"github.com/mattdurham/bob/internal/shipmate/transcript"
)

// pidPath returns the PID file path for a given socket path.
func pidPath(sockPath string) string {
	return strings.TrimSuffix(sockPath, ".sock") + ".pid"
}

// killExisting reads the PID file at pidPath and sends SIGTERM to the process.
// Errors are logged but not fatal — a missing or stale PID file is normal.
func killExisting(pp string) {
	data, err := os.ReadFile(pp)
	if err != nil {
		return // no PID file; nothing to kill
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return
	}
	if err := proc.Signal(syscall.SIGTERM); err == nil {
		log.Printf("shipmate: daemon: killed previous daemon pid=%d", pid)
	}
}

// Serve listens on a Unix socket at sockPath. On each 5-minute tick it calls
// te.ExportNew to forward any new unscored turns to the upstream. When a "stop"
// command arrives (or ctx is cancelled), Serve calls te.ExportAndScore for the
// final scored export, then shuts down te and removes the socket file.
func Serve(ctx context.Context, sockPath string, te *transcript.TurnExporter) error {
	pp := pidPath(sockPath)
	killExisting(pp)
	os.Remove(sockPath) //nolint:errcheck
	if err := os.WriteFile(pp, []byte(fmt.Sprintf("%d\n", os.Getpid())), 0o644); err != nil {
		log.Printf("shipmate: daemon: write pid file: %v", err)
	}

	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		return err
	}

	stopCh := make(chan struct{})
	var stopOnce sync.Once
	shutdownDone := make(chan struct{})

	go func() {
		defer close(shutdownDone)
		select {
		case <-ctx.Done():
		case <-stopCh:
		}
		_ = ln.Close()
	}()

	// Periodic export goroutine.
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				peekCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
				if err := te.ExportNew(peekCtx); err != nil {
					log.Printf("shipmate: daemon: periodic export: %v", err)
				}
				cancel()
			case <-stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	var wg sync.WaitGroup
	for {
		conn, err := ln.Accept()
		if err != nil {
			break
		}
		wg.Add(1)
		go func(c net.Conn) {
			defer wg.Done()
			handleConn(c, stopCh, &stopOnce)
		}(conn)
	}
	wg.Wait()
	<-shutdownDone

	// Final scored export.
	scoreCtx, scoreCancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer scoreCancel()
	if err := te.ExportAndScore(scoreCtx); err != nil {
		log.Printf("shipmate: daemon: score: %v", err)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := te.Shutdown(shutdownCtx); err != nil {
		log.Printf("shipmate: daemon: shutdown: %v", err)
	}

	os.Remove(sockPath) //nolint:errcheck
	os.Remove(pp)       //nolint:errcheck
	return nil
}

// handleConn reads a single JSON command from conn and dispatches it.
func handleConn(conn net.Conn, stopCh chan struct{}, stopOnce *sync.Once) {
	defer func() { _ = conn.Close() }()

	var cmd hook.Command
	if err := json.NewDecoder(conn).Decode(&cmd); err != nil {
		log.Printf("shipmate: daemon: decode command: %v", err)
		return
	}

	switch cmd.Type {
	case "stop":
		stopOnce.Do(func() { close(stopCh) })
	default:
		log.Printf("shipmate: daemon: unknown command type %q", cmd.Type)
	}
}
