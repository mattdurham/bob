#!/usr/bin/env bash
set -euo pipefail

# Render a Bob skill for the omp harness.
# omp (native .omp provider) reads `<skills-root>/<name>/SKILL.md` with a
# `name` + `description` frontmatter. This renderer normalizes a shared
# SKILL.md/SKILL.simple.md into that shape:
#   - drops Claude-only frontmatter (requires_experimental, user-invocable, …)
#   - strips Claude experimental-agent-teams guidance
#   - rewrites the Claude "Skill tool" vocabulary to read-and-follow
#   - appends the harness delegation idiom when the body mentions subagents
# A dedicated SKILL.omp.md is written by hand and installed verbatim instead of
# being passed through here.

if [[ $# -ne 1 ]]; then
    echo "usage: $0 <source-skill.md>" >&2
    exit 2
fi
source_file=$1

perl -0pe '
  s/^requires_experimental: agent_teams\n//mg;
  s/\n<experimental_feature>\n.*?\n<\/experimental_feature>\n/\n/gsm;
  s/ using Claude Code\x27s experimental agent teams feature//g;
  s/Claude Code\x27s experimental agent teams/native agent teams/g;
  s/experimental agent teams/native agent teams/g;
  s/Verifying native agent teams flag\.\.\.\n//g;
  s/^\s*If running in Claude Code: check if CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 is set\.\n//mg;
  s/^\s*Check if CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 is set\.?\n//mg;
  s/^.*CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS.*\n//mg;
  s/\n## Prerequisites\n\s*(?=\n## )/\n/g;
' "$source_file" \
| awk '
BEGIN { delimiters = 0; name = ""; description = ""; attribution = ""; metadata = 0 }
delimiters < 2 && $0 == "---" {
    delimiters++
    if (delimiters == 2) {
        if (name == "" || description == "") {
            print "missing name or description in " FILENAME > "/dev/stderr"
            exit 1
        }
        print "---"
        print name
        print description
        printf "%s", attribution
        print "---"
    }
    next
}
delimiters == 1 && /^name:[[:space:]]*/ { name = $0; next }
delimiters == 1 && /^description:[[:space:]]*/ { description = $0; next }
delimiters == 1 && /^(license|metadata):/ { attribution = attribution $0 "\n"; metadata = /^metadata:/; next }
delimiters == 1 && metadata && /^[[:space:]]/ { attribution = attribution $0 "\n"; next }
delimiters == 1 && /^[^[:space:]]/ { metadata = 0; next }
delimiters >= 2 { print }
' \
| sed \
    -e 's/[Cc]all the Skill tool twice, for/Read and follow both skills,/g' \
    -e 's/[Cc]all the Skill tool with/Read and follow the skill/g' \
    -e 's/calls the Skill tool with/reads and follows the skill/g' \
    -e 's/call the Skill tool for/read and follow/g' \
    -e 's/call the Skill tool/read and follow the named skill/g'

# Generic upstream worker instructions use the harness native delegation mechanism.
if grep -Eiq 'sub.?agents?|background agent' "$source_file"; then
    printf '\nUse the task tool for the independent workers described here. Launch them in one call as parallel tasks[] items, each self-contained: working directory, the exact bounded job, read-only constraint, unique output path, and any companion or skill paths to read; agent: "task" for general or implementation work, "reviewer"/"scout" for read-only analysis. Results auto-deliver; call wait and collect every item before dependent work. Workers do not re-delegate; the parent performs any further dispatch.\n'
fi