#!/usr/bin/env bash
set -euo pipefail
# Retire only known, unmodified Bob entrypoints after their replacement exists.
skills_dir=${1:?usage: retire-bob-design.sh skills-directory}
[[ -f "$skills_dir/bob-specs/SKILL.md" ]] || exit 0
old_dir="$skills_dir/bob-design"
[[ ! -L "$old_dir" && -f "$old_dir/SKILL.md" && ! -L "$old_dir/SKILL.md" ]] || exit 0
if command -v shasum >/dev/null 2>&1; then
    digest=$(shasum -a 256 "$old_dir/SKILL.md")
else
    digest=$(sha256sum "$old_dir/SKILL.md")
fi
case "${digest%% *}" in
    7bdb7c73a3061b9253ce5526a173fc57bf81039afd3a749a73a1865381b07afc|1933a1c2326e87e1aec454588e7bfaffa72ced9ab37609656ff21bc9af28b399|a5721cd672c75db8c4e83d6db8123973e9c239c74d92bc263c53aa2d88cee7db|b7ad737e056e88c26f1ba8b43cd1a0497eb79a69b05181fdb0066bbb6a5b3b22)
        rm -- "$old_dir/SKILL.md"
        rmdir -- "$old_dir" 2>/dev/null || true
        ;;
esac
