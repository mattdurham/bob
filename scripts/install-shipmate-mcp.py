#!/usr/bin/env python3
"""Register shipmate in ~/.claude/settings.json after binary installation."""
import json
import os
import sys

path = os.path.expanduser("~/.claude/settings.json")
if not os.path.exists(path):
    print(f"  settings.json not found at {path}, skipping MCP registration")
    sys.exit(0)

with open(path) as f:
    s = json.load(f)

s.setdefault("mcpServers", {})["shipmate"] = {
    "command": os.path.expanduser("~/.local/bin/shipmate"),
    "env": {
        "SHIPMATE_UPSTREAM_ENDPOINT": os.environ.get("SHIPMATE_UPSTREAM_ENDPOINT", "http://localhost:4318"),
        "SHIPMATE_UPSTREAM_HEADERS": os.environ.get("SHIPMATE_UPSTREAM_HEADERS", ""),
    },
}

s.setdefault("env", {}).update({
    "CLAUDE_CODE_ENABLE_TELEMETRY": "1",
    "CLAUDE_CODE_ENHANCED_TELEMETRY_BETA": "1",
    "OTEL_TRACES_EXPORTER": "otlp",
    "OTEL_EXPORTER_OTLP_ENDPOINT": "http://localhost:4317",
    "OTEL_EXPORTER_OTLP_PROTOCOL": "grpc",
})

allow = s.setdefault("permissions", {}).setdefault("allow", [])
if "mcp__shipmate__shipmate_record" not in allow:
    allow.append("mcp__shipmate__shipmate_record")

with open(path, "w") as f:
    json.dump(s, f, indent=2)
    f.write("\n")

print("  shipmate registered in settings.json (restart Claude Code to activate)")
print(f"  upstream endpoint: {os.environ.get('SHIPMATE_UPSTREAM_ENDPOINT', 'http://localhost:4318')}")
