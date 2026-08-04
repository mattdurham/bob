#!/usr/bin/env bash
set -euo pipefail

# pi and wllr have native team support. Keep Claude Code's team-flag guidance in
# the source skills, but remove it from generated native-environment installs.
perl -0pe '
  s/^requires_experimental: agent_teams\n//mg;
  s/\n<experimental_feature>\n.*?\n<\/experimental_feature>\n/\n/gsm;
  s/ using Claude Code'\''s experimental agent teams feature//g;
  s/Claude Code'\''s experimental agent teams/native agent teams/g;
  s/experimental agent teams/native agent teams/g;
  s/Verifying native agent teams flag\.\.\.\n//g;
  s/^\s*If running in Claude Code: check if CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 is set\.\n//mg;
  s/^\s*Check if CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 is set\.?\n//mg;
  s/^.*CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS.*\n//mg;
  s/\n## Prerequisites\n\s*(?=\n## )/\n/g;
' 
