#!/usr/bin/env bash
set -euo pipefail
skills_dir=${1:?usage: retire-version.sh skills-directory}
old_dir="$skills_dir/bob-version"
[[ ! -L "$old_dir" && -f "$old_dir/SKILL.md" && ! -L "$old_dir/SKILL.md" ]] || exit 0
# Generated values differ per installation; identify Bob's reserved utility format.
if grep -Eq '^name: (bob-version|bob:version)$' "$old_dir/SKILL.md" &&
   grep -Fxq '# Bob Version Information' "$old_dir/SKILL.md" &&
   grep -Fxq '## Installation Details' "$old_dir/SKILL.md" &&
   grep -Fxq '## Installed Components' "$old_dir/SKILL.md"; then
    rm -- "$old_dir/SKILL.md"
    rmdir -- "$old_dir" 2>/dev/null || true
fi
