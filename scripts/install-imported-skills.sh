#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 2 ]]; then
    echo "usage: $0 <claude|pi|codex|wllr|goose> <skills-directory>" >&2
    exit 2
fi

runtime=$1
skills_dir=$2
repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
case "$runtime" in
    claude|pi|codex|wllr|goose) ;;
    *) echo "unsupported harness: $runtime" >&2; exit 2 ;;
esac

while IFS= read -r skill || [[ -n "$skill" ]]; do
    [[ -z "$skill" || "$skill" == \#* ]] && continue
    source_dir="$repo_root/skills/$skill"
    destination="$skills_dir/$skill"
    # Replace a prior linked installation without writing through it to upstream.
    [[ ! -L "$destination" ]] || rm "$destination"
    mkdir -p "$destination"
    cp -R "$source_dir/." "$destination/"
    bash "$repo_root/scripts/render-imported-skill.sh" "$runtime" \
        "$source_dir/SKILL.md" > "$destination/SKILL.md"
    if [[ "$runtime" == claude ]]; then
        # Supporting instructions may invoke other imported skills too.
        while IFS= read -r -d '' companion; do
            [[ "$companion" != "$destination/SKILL.md" ]] || continue
            sed 's/bob-matt-/bob:matt-/g' "$companion" > "$companion.tmp"
            mv "$companion.tmp" "$companion"
        done < <(find "$destination" -type f \( -name '*.md' -o -name '*.yaml' \) -print0)
    fi
    echo "   Installed $skill ($runtime)"
done < "$repo_root/skills/imported-skills.txt"

# Retire known old names only after all prefixed packages install successfully.
bash "$repo_root/scripts/retire-imported-skill-names.sh" "$skills_dir"
