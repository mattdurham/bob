#!/bin/bash
# fix-shipmate-hooks.sh — update ~/.claude/settings.json to use the new transcript-based shipmate hooks.
# Removes dead PostToolUse/SubagentStart/SubagentStop/TaskCreated/TaskCompleted/Stop hooks
# and updates SessionStart to pass stdin directly to shipmate start.

set -euo pipefail

SETTINGS="$HOME/.claude/settings.json"
BACKUP="${SETTINGS}.bak"

if [ ! -f "$SETTINGS" ]; then
  echo "ERROR: $SETTINGS not found" >&2
  exit 1
fi

cp "$SETTINGS" "$BACKUP"
echo "Backed up to $BACKUP"

jq '
  .hooks = {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "/Users/mdurham/.local/bin/shipmate start --upstream https://tempo-dev-test-03-dev-us-east-0.grafana-dev.net"
          }
        ]
      }
    ],
    "SessionEnd": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "/Users/mdurham/.local/bin/shipmate stop"
          }
        ]
      }
    ]
  }
' "$SETTINGS" > "${SETTINGS}.tmp" && mv "${SETTINGS}.tmp" "$SETTINGS"

echo "Done. Updated hooks:"
jq '.hooks' "$SETTINGS"
