#!/bin/bash
# Bob approval hook — installed to <worktree>/.claude/hooks/stop.sh
# Called by Claude Code on stop events (waiting for user input / approval).
# Writes bob-approval.json so the bob Zellij plugin can show the ⚠ indicator.

input=$(cat)

tool=$(echo "$input" | jq -r '.tool_name // empty' 2>/dev/null)
preview=$(echo "$input" | jq -r '
  .tool_input
  | to_entries
  | map(.value | tostring)
  | .[0]
  // ""
' 2>/dev/null | head -c 60 | tr '\n' ' ')

cwd=$(pwd)
project_hash=$(printf '%s' "$cwd" | sha256sum | cut -c1-8)
status_dir="$HOME/.claude/projects/$project_hash"
mkdir -p "$status_dir"

if [ -n "$tool" ]; then
    printf '{"pending":true,"tool":"%s","preview":"%s"}\n' \
        "$tool" "$preview" \
        > "$status_dir/bob-approval.json"
else
    printf '{"pending":false}\n' > "$status_dir/bob-approval.json"
fi
