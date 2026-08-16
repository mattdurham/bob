#!/bin/sh
# bob publication gate: run the repo's .bob/hooks/pre-publish (if present),
# then push, in one process. Nonzero exit blocks publication. Lines starting
# with "BOB_GATE_BLOCK:" are gate verdicts; any other failure is an ordinary
# push failure. Invoked by commit-agent Step 5 through the rendered
# agent-directory marker (see the call site in agents/commit-agent/SKILL.md).
# Hook contract (bob-specific; NOT git's pre-push interface): executed from
# the repo top level with no arguments and stdin reading end-of-file; exit 0
# allows the push,
# any other exit blocks it; it must not commit, move the branch, or modify
# files.
set -u

# The leading newline guarantees the marker starts a line even when prior
# output lacked a trailing newline; block() exits at once, so the marker is
# always the last line.
block() { printf '\nBOB_GATE_BLOCK: %s\n' "$1"; exit 1; }

cd "$(git rev-parse --show-toplevel)" || block "not inside a git repository"
BR=$(git branch --show-current) || block "cannot resolve the current branch"
if [ -z "$BR" ]; then block "detached HEAD - publication needs a branch"; fi
HEAD_PRE=$(git rev-parse HEAD) || block "cannot resolve HEAD"

HOOK=.bob/hooks/pre-publish
if [ -e "$HOOK" ] || [ -L "$HOOK" ]; then
    if [ -L "$HOOK" ] || [ ! -f "$HOOK" ] || [ ! -x "$HOOK" ]; then
        block "hook broken: $HOOK is not an executable regular file (symlinks are not allowed)"
    fi
    if ! TRACKED=$(git ls-tree -r HEAD -- "$HOOK"); then
        block "cannot inspect the committed hook (git ls-tree failed)"
    fi
    if [ -n "$TRACKED" ]; then
        case "$TRACKED" in
            100755*) ;;
            *) block "$HOOK is committed without the executable bit; commit it with mode 100755" ;;
        esac
        BLOB=$(mktemp) || block "mktemp failed"
        if ! git cat-file blob "HEAD:$HOOK" >"$BLOB" 2>/dev/null; then
            rm -f "$BLOB"
            block "cannot read the committed hook (git cat-file failed)"
        fi
        if ! cmp -s "$BLOB" "$HOOK"; then
            rm -f "$BLOB"
            block "$HOOK differs from the commit being published; commit the hook change or restore it"
        fi
        rm -f "$BLOB"
    fi
    W=$(command -v timeout || command -v gtimeout)
    if [ -z "$W" ]; then
        block "hook present but no timeout wrapper found; install coreutils (timeout/gtimeout) or remove $HOOK"
    fi
    OUT=$(mktemp) || block "mktemp failed"
    "$W" -k 10 90 "./$HOOK" </dev/null >"$OUT" 2>&1
    RC=$?
    tail -c 8192 "$OUT"
    rm -f "$OUT"
    if [ "$(git branch --show-current)" != "$BR" ] || [ "$(git rev-parse HEAD)" != "$HEAD_PRE" ]; then
        block "branch or HEAD moved while the hook ran (expected $BR at $HEAD_PRE); hooks must not commit or move the branch"
    fi
    if [ "$RC" -eq 124 ] || [ "$RC" -eq 137 ]; then
        block "hook timed out after 90s (rc=$RC); fix or remove $HOOK"
    fi
    if [ "$RC" -ne 0 ]; then
        block "hook rejected publication (rc=$RC) - fix what it reports above"
    fi
else
    if ! TRACKED=$(git ls-tree -r HEAD --name-only -- "$HOOK"); then
        block "cannot inspect the committed hook (git ls-tree failed)"
    fi
    if [ -n "$TRACKED" ]; then
        block "$HOOK is tracked in the commit being published but missing from the worktree; restore it or commit its removal"
    fi
fi

git push -u origin "refs/heads/$BR:refs/heads/$BR"
exit $?
