// Package client sends a single NDJSON command to the shipmate daemon over a Unix socket.
// Send always returns nil — hook processes must exit 0 even if the daemon is unreachable.
package client

import (
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/mattdurham/bob/internal/shipmate/hook"
)

// maxRetries and retrySleep are package-level vars so tests can override them.
var (
	maxRetries = 10
	retrySleep = 50 * time.Millisecond
)

// Send connects to the Unix socket at sockPath, writes cmd as a single JSON line,
// and closes the connection. On ENOENT or ECONNREFUSED it retries up to maxRetries
// times with retrySleep between attempts. Returns nil after exhausting retries —
// hooks must not fail Claude Code with a non-zero exit.
func Send(sockPath string, cmd hook.Command) error {
	var conn net.Conn
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		conn, lastErr = net.Dial("unix", sockPath)
		if lastErr == nil {
			break
		}
		time.Sleep(retrySleep)
	}
	if lastErr != nil {
		log.Printf("shipmate: client: could not connect to %s after %d attempts: %v (trace lost)",
			sockPath, maxRetries, lastErr)
		return nil // intentionally swallowed — hooks must exit 0
	}
	defer func() { _ = conn.Close() }()

	if err := json.NewEncoder(conn).Encode(cmd); err != nil {
		log.Printf("shipmate: client: encode command: %v", err)
	}
	return nil
}
