#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
    echo "usage: $0 <claude|pi|codex|wllr|goose> <source>" >&2
    exit 2
fi

runtime=$1
source_file=$2
script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
case "$runtime" in
    claude)
        sed 's/bob-matt-/bob:matt-/g' "$source_file"
        exit 0
        ;;
    codex|pi|wllr|goose) ;;
    *) echo "unsupported harness: $runtime" >&2; exit 2 ;;
esac

render() {
    if [[ "$runtime" == goose ]]; then
        name=$(sed -n 's/^name: //p' "$source_file" | head -1)
        bash "$script_dir/render-goose-skill.sh" "$source_file" "$name"
    else
        awk -v runtime="$runtime" '
        NR == 1 && $0 == "---" { frontmatter = 1; print; next }
        frontmatter && $0 == "---" {
            if (runtime == "pi" || runtime == "wllr") print "user-invocable: true"
            frontmatter = 0; print; next
        }
        frontmatter && runtime == "codex" && /^(disable-model-invocation|argument-hint):/ { next }
        { print }
        ' "$source_file"
    fi
}

# These harnesses load skills by reading their entrypoint instead of a Claude Skill tool.
render | sed \
    -e 's/[Cc]all the Skill tool twice, for/Read and follow both skills,/g' \
    -e 's/[Cc]all the Skill tool with/Read and follow the skill/g' \
    -e 's/calls the Skill tool with/reads and follows the skill/g' \
    -e 's/call the Skill tool for/read and follow/g' \
    -e 's/call the Skill tool/read and follow the named skill/g'

if [[ "$runtime" == goose ]] && grep -q '^disable-model-invocation: true' "$source_file"; then
    printf '\nUse this skill only when the user explicitly requests it.\n'
fi

# Generic upstream worker instructions use the harness native delegation mechanism.
if grep -Eiq 'sub.?agents?|background agent' "$source_file"; then
    case "$runtime" in
        codex)
            printf '\nUse native collaboration tools for the independent workers described here. Give each worker its task, working directory, relevant skill paths, and expected result; collect its result before dependent work.\n'
            ;;
        pi|wllr)
            printf '\nUse the native subagent tool for the independent workers described here. Give each worker its task, working directory, relevant skill paths, and expected result; collect its result before dependent work.\n'
            ;;
        goose)
            printf '\nThe parent session owns delegation through Summon: start independent jobs with delegate(source: "Explore", instructions: "bounded task, working directory, skill paths, expected result", async: true), and collect every task id with load(source: "task id"). Use workflow-implementer for code changes. Delegates cannot delegate again; the parent performs any further worker dispatch requested by a referenced skill.\n'
            ;;
    esac
fi
