#!/usr/bin/env python3
"""Configure shipmate hooks in ~/.claude/settings.json.

Removes the old MCP server entry (if present) and registers Claude Code hooks
that drive the shipmate daemon:

  SessionStart  -> shipmate start --session-id <id> --upstream <url>
  PostToolUse   -> shipmate record
  SubagentStart -> shipmate record
  SubagentStop  -> shipmate record
  TaskCreated   -> shipmate record
  TaskCompleted -> shipmate record
  Stop          -> shipmate stop --session-id <id>
"""
import json
import os
import shlex
import shutil
import sys

if shutil.which("jq") is None:
    print("  WARNING: jq is not installed — shipmate hooks require jq to extract session_id")
    print("  Install jq: https://jqlang.github.io/jq/download/")

SHIPMATE_BIN = os.path.expanduser("~/.local/bin/shipmate")
SETTINGS_PATH = os.path.expanduser("~/.claude/settings.json")

upstream = os.environ.get("SHIPMATE_UPSTREAM_ENDPOINT", "http://localhost:4318")
headers  = os.environ.get("SHIPMATE_UPSTREAM_HEADERS", "")
user     = os.environ.get("SHIPMATE_UPSTREAM_USER", "")
token    = os.environ.get("SHIPMATE_UPSTREAM_TOKEN", "")

if not os.path.exists(SETTINGS_PATH):
    print(f"  settings.json not found at {SETTINGS_PATH}, creating minimal file")
    settings = {}
else:
    with open(SETTINGS_PATH) as f:
        settings = json.load(f)

# Remove old MCP server entry if present.
mcp_servers = settings.get("mcpServers", {})
if "shipmate" in mcp_servers:
    del mcp_servers["shipmate"]
    settings["mcpServers"] = mcp_servers
    print("  removed old mcpServers.shipmate entry")

# Remove old OTEL proxy env vars (these were for the gRPC proxy; not needed now).
env = settings.get("env", {})
for key in [
    "OTEL_TRACES_EXPORTER",
    "OTEL_EXPORTER_OTLP_ENDPOINT",
    "OTEL_EXPORTER_OTLP_PROTOCOL",
]:
    if key in env:
        del env[key]
        print(f"  removed env.{key} (was for old gRPC proxy)")

# Keep Claude Code telemetry enabled.
env.setdefault("CLAUDE_CODE_ENABLE_TELEMETRY", "1")
env.setdefault("CLAUDE_CODE_ENHANCED_TELEMETRY_BETA", "1")
settings["env"] = env

# Build hook commands.
# SessionStart: extract session_id from stdin JSON via jq, then start the daemon.
start_cmd = (
    f"jq -r '.session_id' | xargs -I{{}} "
    f"{shlex.quote(SHIPMATE_BIN)} start "
    f"--session-id {{}} "
    f"--upstream {shlex.quote(upstream)}"
)
if headers:
    start_cmd += f" --headers {shlex.quote(headers)}"

# SHIPMATE_UPSTREAM_USER/TOKEN bake the env vars into the hook environment
# so the daemon child process can read them to build the Basic auth header.
env_prefix = ""
if user:
    env_prefix += f"SHIPMATE_UPSTREAM_USER={shlex.quote(user)} "
if token:
    env_prefix += f"SHIPMATE_UPSTREAM_TOKEN={shlex.quote(token)} "
if env_prefix:
    start_cmd = env_prefix + start_cmd

# Stop: extract session_id from stdin JSON and send stop.
stop_cmd = (
    f"jq -r '.session_id' | xargs -I{{}} "
    f"{shlex.quote(SHIPMATE_BIN)} stop --session-id {{}}"
)

# Record: read full stdin JSON (hook package handles extraction).
record_cmd = shlex.quote(SHIPMATE_BIN) + " record"

# Register hooks.
hooks = settings.setdefault("hooks", {})
hooks["SessionStart"] = [{"hooks": [{"type": "command", "command": start_cmd}]}]
hooks["Stop"] = [{"hooks": [{"type": "command", "command": stop_cmd}]}]
for event in ["PostToolUse", "SubagentStart", "SubagentStop", "TaskCreated", "TaskCompleted"]:
    hooks[event] = [{"hooks": [{"type": "command", "command": record_cmd}]}]

with open(SETTINGS_PATH, "w") as f:
    json.dump(settings, f, indent=2)
    f.write("\n")

print(f"  shipmate hooks registered in {SETTINGS_PATH}")
print(f"  upstream endpoint: {upstream}")
if user:
    print(f"  auth: Basic ({user}:***)")
print("  restart Claude Code to activate")
