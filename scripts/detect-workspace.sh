#!/usr/bin/env bash
# detect-workspace.sh — detect OKF and spec-driven workspace mode
#
# Writes .bob/state/workspace.md with detected mode and context.
# Safe to call multiple times (idempotent).
#
# Exit codes:
#   0 — detection complete, workspace.md written
#   1 — .bob/state/ does not exist (call after worktree setup)
#
# Usage:
#   bash scripts/detect-workspace.sh
#   bash scripts/detect-workspace.sh /path/to/worktree

set -euo pipefail

WORKDIR="${1:-$(pwd)}"
STATE_DIR="$WORKDIR/.bob/state"

if [ ! -d "$STATE_DIR" ]; then
	echo "ERROR: $STATE_DIR does not exist — run after worktree setup" >&2
	exit 1
fi

# ── OKF workspace detection ───────────────────────────────────────────────────
OKF=false
FEATURE_CONCEPT=""

if [ -d "$WORKDIR/.knowledge" ] && [ -f "$WORKDIR/.knowledge/index.md" ]; then
	OKF=true

	# Find the most recently modified planned feature concept
	FEATURE_CONCEPT=$(
		find "$WORKDIR/.knowledge/features" -name "*.md" ! -name "index.md" \
			-exec grep -l "^status: planned" {} \; 2>/dev/null |
			sort | head -1
	)
fi

# ── Spec-driven workspace detection ──────────────────────────────────────────
SPEC_DRIVEN=false
SPEC_MODULES=""

# Look for spec doc files (excluding .knowledge/ itself)
SPEC_FILES=$(
	find "$WORKDIR" \
		-not -path "$WORKDIR/.knowledge/*" \
		-not -path "$WORKDIR/.git/*" \
		\( -name "SPECS.md" -o -name "NOTES.md" -o -name "TESTS.md" -o -name "BENCHMARKS.md" \) \
		2>/dev/null | head -20
)

# Look for NOTE invariant in .go files
NOTE_FILES=$(
	grep -rl "NOTE: Any changes to this file must be reflected" \
		--include="*.go" "$WORKDIR" 2>/dev/null | head -5 || true
)

if [ -n "$SPEC_FILES" ] || [ -n "$NOTE_FILES" ]; then
	SPEC_DRIVEN=true
	# Collect unique directories containing spec docs
	SPEC_MODULES=$(
		echo "$SPEC_FILES" | xargs -I{} dirname {} 2>/dev/null |
			sort -u |
			sed "s|$WORKDIR/||g" |
			head -10 || true
	)
fi

# ── Compute mode ─────────────────────────────────────────────────────────────
if $OKF && $SPEC_DRIVEN; then
	MODE=full
elif $OKF; then
	MODE=okf
elif $SPEC_DRIVEN; then
	MODE=spec-driven
else
	MODE=none
fi

# ── Write workspace.md ───────────────────────────────────────────────────────
{
	echo "---"
	echo "okf: $OKF"
	echo "spec-driven: $SPEC_DRIVEN"
	echo "mode: $MODE"
	if [ -n "$FEATURE_CONCEPT" ]; then
		echo "feature: $(echo "$FEATURE_CONCEPT" | sed "s|$WORKDIR/||g")"
	else
		echo "feature: \"\""
	fi
	echo "---"
	echo ""
	echo "# Workspace Context"
	echo ""
	echo "**Mode:** $MODE"
	echo ""

	if $OKF; then
		echo "**OKF bundle:** .knowledge/ (index at .knowledge/index.md)"
		if [ -n "$FEATURE_CONCEPT" ]; then
			FEATURE_TITLE=$(grep "^title:" "$FEATURE_CONCEPT" | sed 's/^title: //')
			echo "**Active feature:** $FEATURE_TITLE"
			echo "**Feature concept:** $(echo "$FEATURE_CONCEPT" | sed "s|$WORKDIR/||g")"
		else
			echo "**Active feature:** none (no planned feature concept found)"
		fi
	else
		echo "**OKF bundle:** not present"
	fi
	echo ""

	if $SPEC_DRIVEN; then
		echo "**Spec-driven modules:**"
		if [ -n "$SPEC_MODULES" ]; then
			echo "$SPEC_MODULES" | while read -r mod; do
				[ -n "$mod" ] && echo "- $mod"
			done
		else
			echo "- (detected via NOTE invariant in .go files)"
		fi
	else
		echo "**Spec-driven modules:** none detected"
	fi
	echo ""
	echo "## Behavior"
	echo ""
	case "$MODE" in
	full)
		echo "- Brainstormer reads: .knowledge/ index → package and decision concepts → linked SPECS.md and NOTES.md"
		echo "- Enrichment writes: OKF concepts (package enrichment, decisions) + SPECS.md/NOTES.md updates"
		;;
	okf)
		echo "- Brainstormer reads: .knowledge/ index → package and decision concepts"
		echo "- Enrichment writes: OKF concepts (package enrichment, decisions)"
		;;
	spec-driven)
		echo "- Brainstormer reads: SPECS.md + NOTES.md for in-scope modules"
		echo "- Enrichment writes: SPECS.md + NOTES.md updates"
		;;
	none)
		echo "- Brainstormer reads: cold discovery from code"
		echo "- Enrichment writes: nothing (no knowledge layer present)"
		;;
	esac
} >"$STATE_DIR/workspace.md"

echo "WORKSPACE_MODE=$MODE"
echo "WORKSPACE_OKF=$OKF"
echo "WORKSPACE_SPEC_DRIVEN=$SPEC_DRIVEN"
if [ -n "$FEATURE_CONCEPT" ]; then
	echo "WORKSPACE_FEATURE=$(echo "$FEATURE_CONCEPT" | sed "s|$WORKDIR/||g")"
fi
