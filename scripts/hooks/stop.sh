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

# Use jq -n to construct JSON safely — prevents injection from tool/preview values
if [ -n "$tool" ]; then
    jq -n --arg tool "$tool" --arg preview "$preview" \
        '{"pending":true,"tool":$tool,"preview":$preview}' \
        > "$status_dir/bob-approval.json"
else
    printf '{"pending":false}\n' > "$status_dir/bob-approval.json"
fi
