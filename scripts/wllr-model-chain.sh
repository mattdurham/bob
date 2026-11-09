#!/usr/bin/env bash
set -euo pipefail

# Rewrite the `model:` frontmatter of an agent prompt into a wllr fallback chain.
#
# Source agent prompts name Claude Code's model aliases ("sonnet", "haiku"),
# which are valid there and are what the claude/pi/codex installers ship. wllr
# has no such aliases: it resolves a `model:` value by trying, per candidate, an
# exact tier name, then an exact model ID, then a plain model on the active
# provider. A bare "sonnet" therefore resolves to nothing on an OpenRouter-only
# setup and the skill silently keeps whatever model was already active.
#
# wllr's chains are ordered fallbacks: the first candidate that resolves to a
# configured model wins. So the *order* is the statement of intent, and here it
# is driven by the skill's role rather than by the capability alias in the
# source:
#
#   planning skills  -> [high, low]   spend reasoning capability first
#   coding/review    -> [low, high]   spend the cheap model first
#
# Imported upstream skills declare no model at all, and most of them are
# deliberately left inheriting the session model. INJECT_PLANNING names the few
# whose whole job is reasoning about a plan or design; for those only, a
# `model: [high, low]` key is added when the frontmatter has none. This keeps the
# upstream sources pristine (they are vendored by commit) while ensuring the
# sharpest-thinking skills do not silently run on a cheap session model.
#
# Both entries are tier names, which are the only candidates guaranteed to name
# something a given machine has configured. The older concrete per-provider IDs
# (claude-sonnet-4-5, gpt-5-codex) are deliberately not used: they resolve to
# nothing on an OpenRouter-only setup, and any entry they would add sits in front
# of a tier that always resolves. Add them back to the orders below if a machine
# with native Anthropic/OpenAI access needs the same prompt to prefer those.
#
# Usage: wllr-model-chain.sh <skill-name>  < source.md > dest.md
#
# Only a bare alias value is rewritten: `model: sonnet` or `model: haiku` on its
# own line. Anything already bracketed, absent, or otherwise spelled passes
# through untouched, so this is idempotent, never rewrites a value it does not
# recognise, and never introduces a `model:` line into a skill that did not
# declare one.
#
# This runs at install time and the source prompts are left alone, because
# "sonnet"/"haiku" are correct and meaningful for the claude, pi, codex and
# goose installers.
SKILL="${1:-}"

# Role assignment is explicit rather than inferred from the skill's name, so it
# can be reviewed in one place and corrected without touching any prompt. A new
# agent that is in neither list still gets installed with the conservative
# strong-first order (never silently downgrade capability), but is reported on
# stderr so the omission is visible rather than silent.
PLANNING="planner team-planner team-brainstormer workflow-brainstormer team-spec-oracle architecture-introspector"
EXECUTION="bug-finder commit-agent explore go-presubmit-reviewer monitor-agent review-consolidator spec-doc-reviewer team-analyst team-challenger team-coder team-reviewer tester workflow-code-quality workflow-coder workflow-implementer workflow-task-reviewer"

# Imported skills whose frontmatter must gain a model chain rather than have an
# existing alias rewritten. These declare no `model:` upstream, so they would
# otherwise inherit whatever model the session happens to be using.
INJECT_PLANNING="bob-matt-grill-me bob-matt-grill-with-docs bob-matt-grilling"

# An empty name (a direct pipe with no argument) keeps the strong-first order
# rather than silently downgrading an unclassified prompt.
ORDER='[high, low]'
case " $EXECUTION " in *" $SKILL "*) ORDER='[low, high]' ;; esac

# Injecting is itself the declaration of intent, so an injected skill does not
# need to also appear in PLANNING to avoid the warning below.
INJECT=''
case " $INJECT_PLANNING " in *" $SKILL "*) INJECT=1 ;; esac

# Only report an unclassified skill when it actually declares a bare alias. The
# install loops also pipe the bob-* command skills through here, and those
# declare no `model:` at all, so they are none of this script's business.
WARN_MSG=''
case " $PLANNING $EXECUTION " in
  *" $SKILL "*) ;;
  *) if [ -n "$SKILL" ] && [ -z "$INJECT" ]; then WARN_MSG="wllr-model-chain: '$SKILL' is not classified as planning or execution; using $ORDER"; fi ;;
esac

# -0777 slurps the whole file because injecting a key needs to see the closing
# `---`, which a line-oriented pass cannot do while staying a single pass. The
# rewrite of an existing alias still applies per line.
ORDER="$ORDER" WARN_MSG="$WARN_MSG" INJECT="$INJECT" perl -0777 -pe '
  if ($ENV{INJECT} && /\A---\n(.*?)\n---\n/s && $1 !~ /^model:/m) {
    s{\A---\n}{"---\nmodel: $ENV{ORDER}\n"}e;
  }
  s{^model:\s*(?:sonnet|haiku)\s*$}{$n++; "model: $ENV{ORDER}"}gme;
  END { print STDERR "$ENV{WARN_MSG}\n" if $n && length $ENV{WARN_MSG}; }
'
