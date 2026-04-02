# Shipmate MCP Server — Design

*Date: 2026-04-02*

## Overview

Shipmate is an MCP server that acts as an enriching OTLP proxy for Claude Code sessions. It
receives OTEL spans from Claude Code, forwards them to a downstream collector, and lets agents
emit their own synthetic spans into the same trace via a single MCP tool call.

## Architecture

```
Claude Code ──[OTLP gRPC :4317]──► Shipmate
                                      │
                                      ├── receives spans, tracks session.id
                                      ├── forwards all spans to upstream OTLP
                                      │
                                      └──[OTLP]──► Jaeger / Grafana Tempo / etc.

Agent (MCP tool call)
  shipmate_record(name, agent, text, attributes)
      │
      └── Shipmate creates synthetic span:
            session.id    = (from last Claude Code span)
            service.name  = "shipmate"
            span name     = name arg
            shipmate.agent = agent arg
            shipmate.text  = text arg
            + attributes as span.key = value
          Emits to upstream alongside Claude Code's spans
```

Shipmate is a single process with two roles:
1. **OTLP receiver/forwarder** — transparent proxy on `:4317`
2. **MCP tool server** — stdio transport, exposes `shipmate_record`

## MCP Tool

### `shipmate_record`

Fire-and-forget — creates a complete, already-ended span. No span lifecycle to manage.

**Parameters:**

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Span name, e.g. `"implement rate limiter"` |
| `agent` | string | Agent identity, e.g. `"coder-1"` |
| `text` | string | Free-form description or message |
| `attributes` | map[string]string | Key-value pairs exported as span attributes |

**Example:**
```json
{
  "name": "implement rate limiter",
  "agent": "coder-1",
  "text": "Implemented token bucket in internal/ratelimit/bucket.go",
  "attributes": {
    "task_id": "42",
    "repo": "github.com/mattdurham/bob",
    "file": "internal/ratelimit/bucket.go"
  }
}
```

**Resulting span attributes:**
```
service.name     = "shipmate"
session.id       = "<from Claude Code>"
shipmate.agent   = "coder-1"
shipmate.text    = "Implemented token bucket in internal/ratelimit/bucket.go"
task_id          = "42"
repo             = "github.com/mattdurham/bob"
file             = "internal/ratelimit/bucket.go"
```

## Span Schema

Synthetic spans use these fixed attributes (set by shipmate) plus user-supplied attributes:

| Attribute | Source | Notes |
|-----------|--------|-------|
| `service.name` | shipmate | Always `"shipmate"` — distinguishes from `claude-code` spans |
| `session.id` | Claude Code | Extracted from incoming spans; links to Claude Code trace |
| `shipmate.agent` | `agent` arg | Agent identity for filtering/grouping |
| `shipmate.text` | `text` arg | Free-form annotation |
| span name | `name` arg | Displayed as span name in backend |
| `<key>` | `attributes[key]` | Each key exported as a span attribute directly |

If no Claude Code span has been received yet, `session.id` is omitted and the synthetic span
starts a new trace. This is the fallback when shipmate starts before Claude Code sends anything.

## Configuration

All configuration via environment variables. No config file.

| Variable | Required | Description |
|----------|----------|-------------|
| `SHIPMATE_OTLP_LISTEN_ADDR` | No | gRPC listen address (default `:4317`) |
| `SHIPMATE_UPSTREAM_ENDPOINT` | **Yes** | Upstream OTLP endpoint, e.g. `https://tempo.example.com:4317` |
| `SHIPMATE_UPSTREAM_HEADERS` | No | Comma-separated `Key=Value` pairs for auth headers |

Shipmate **refuses to start** if `SHIPMATE_UPSTREAM_ENDPOINT` is not set. No silent data loss.

**Example:**
```bash
SHIPMATE_UPSTREAM_ENDPOINT=https://tempo.grafana.net:443 \
SHIPMATE_UPSTREAM_HEADERS="Authorization=Bearer glc_..." \
shipmate
```

**Claude Code side:**
```bash
CLAUDE_CODE_ENABLE_TELEMETRY=1 \
CLAUDE_CODE_ENHANCED_TELEMETRY_BETA=1 \
OTEL_TRACES_EXPORTER=otlp \
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 \
claude
```

## Repository Structure

```
cmd/shipmate/
  main.go                    # entry point, env config, wire up server

internal/shipmate/
  server/
    server.go                # MCP tool registration, session.id tracking
  proxy/
    proxy.go                 # OTLP gRPC receiver + forwarder
  recorder/
    recorder.go              # creates synthetic spans via OTEL SDK
```

Follows the same pattern as `cmd/navigator/` and `internal/navigator/`.

## Makefile Target

```makefile
install-shipmate:
    go build -o ~/.local/bin/shipmate ./cmd/shipmate
```

## MCP Configuration (settings.json)

```json
{
  "mcpServers": {
    "shipmate": {
      "command": "shipmate",
      "env": {
        "SHIPMATE_UPSTREAM_ENDPOINT": "https://...",
        "SHIPMATE_UPSTREAM_HEADERS": "Authorization=Bearer ..."
      }
    }
  }
}
```

## What This Is Not

- **Not a trace store** — shipmate does not persist spans locally (no SQLite)
- **Not a span hierarchy manager** — no parent/child span tracking; agents are distinguished by `shipmate.agent` attribute only
- **Not a metrics/logs receiver** — only handles OTEL traces
