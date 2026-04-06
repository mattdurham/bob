// Package daemon implements the shipmate Unix socket server.
// It accepts NDJSON commands from hook clients and dispatches them to a Recorder.
// Each connection carries exactly one command. Serve returns when ctx is cancelled
// or a "stop" command is received.
package daemon

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/mattdurham/bob/internal/shipmate/hook"
	"github.com/mattdurham/bob/internal/shipmate/recorder"
)

// Scorer is an optional interface that can be implemented by the span exporter
// to score buffered spans at session end (e.g. scorer.BufferingExporter).
// If the recorder's exporter implements Scorer, Serve calls Score() before flush.
type Scorer interface {
	Score(ctx context.Context) error
	Peek(ctx context.Context)
}

// Serve listens on a Unix socket at sockPath, dispatching commands to rec.
// It removes any stale socket file at sockPath before binding.
// Serve blocks until ctx is cancelled or a "stop" command is received.
// On return, Serve flushes and shuts down rec. If sc is non-nil, Serve calls
// sc.Score() before flushing to annotate buffered spans with quality scores.
func Serve(ctx context.Context, sockPath string, rec *recorder.Recorder, sc Scorer) error {
	// Remove stale socket file if present.
	os.Remove(sockPath) //nolint:errcheck

	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		return err
	}

	stopCh := make(chan struct{})
	shutdownDone := make(chan struct{})

	// Shutdown goroutine: close the listener when ctx is done or stop is received.
	go func() {
		defer close(shutdownDone)
		select {
		case <-ctx.Done():
		case <-stopCh:
		}
		_ = ln.Close()
	}()

	// Periodic flush goroutine: export a snapshot every 5 minutes without clearing the buffer.
	if sc != nil {
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					peekCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					sc.Peek(peekCtx)
					cancel()
				case <-stopCh:
					return
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	var wg sync.WaitGroup
	for {
		conn, err := ln.Accept()
		if err != nil {
			// Listener closed — either ctx cancelled or stop command received.
			break
		}
		wg.Add(1)
		go func(c net.Conn) {
			defer wg.Done()
			handleConn(ctx, c, rec, stopCh)
		}(conn)
	}
	wg.Wait()

	// Wait for the shutdown goroutine to finish before proceeding with cleanup.
	<-shutdownDone

	// Score buffered spans (non-fatal) before flushing.
	if sc != nil {
		// 35s gives the scorer's own 30s budget plus 5s margin for context propagation.
		scoreCtx, scoreCancel := context.WithTimeout(context.Background(), 35*time.Second)
		defer scoreCancel()
		if err := sc.Score(scoreCtx); err != nil {
			log.Printf("shipmate: daemon: score: %v", err)
		}
	}

	// Flush all queued spans before exiting.
	flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rec.ForceFlush(flushCtx); err != nil {
		log.Printf("shipmate: daemon: flush: %v", err)
	}
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := rec.Shutdown(shutdownCtx); err != nil {
		log.Printf("shipmate: daemon: shutdown: %v", err)
	}
	os.Remove(sockPath) //nolint:errcheck
	return nil
}

// handleConn reads a single JSON command from conn and dispatches it.
func handleConn(ctx context.Context, conn net.Conn, rec *recorder.Recorder, stopCh chan struct{}) {
	defer func() { _ = conn.Close() }()

	var cmd hook.Command
	if err := json.NewDecoder(conn).Decode(&cmd); err != nil {
		log.Printf("shipmate: daemon: decode command: %v", err)
		return
	}

	switch cmd.Type {
	case "record":
		dispatchRecord(ctx, rec, cmd)
	case "memory":
		dispatchMemory(ctx, rec, cmd)
	case "stop":
		// Signal the accept loop to stop; guard against double-close.
		select {
		case <-stopCh:
			// already closed
		default:
			close(stopCh)
		}
	default:
		log.Printf("shipmate: daemon: unknown command type %q", cmd.Type)
	}
}

func dispatchRecord(ctx context.Context, rec *recorder.Recorder, cmd hook.Command) {
	log.Printf("shipmate: record event=%q session=%s attrs=%v", cmd.HookEvent, cmd.SessionID, cmd.Attrs)
	args := recorder.RecordArgs{
		Name:       cmd.HookEvent,
		Agent:      "hook",
		Text:       cmd.HookEvent,
		SessionID:  cmd.SessionID,
		Attributes: cmd.Attrs,
	}
	if err := rec.Record(ctx, args); err != nil {
		log.Printf("shipmate: daemon: record: %v", err)
	}
}

func dispatchMemory(ctx context.Context, rec *recorder.Recorder, cmd hook.Command) {
	log.Printf("shipmate: memory session=%s text=%q", cmd.SessionID, cmd.Text)
	args := recorder.RecordArgs{
		Name:      "memory",
		Agent:     "hook",
		Text:      cmd.Text,
		SessionID: cmd.SessionID,
	}
	if err := rec.Record(ctx, args); err != nil {
		log.Printf("shipmate: daemon: memory: %v", err)
	}
}
