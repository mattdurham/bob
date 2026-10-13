#!/usr/bin/env bash
set -euo pipefail
skills_dir=${1:?usage: retire-imported-skill-names.sh skills-directory}
repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
while IFS=$'\t' read -r old new known_hashes; do
    [[ -z "$old" || "$old" == \#* ]] && continue
    [[ -f "$skills_dir/$new/SKILL.md" && -f "$skills_dir/$old/SKILL.md" ]] || continue
    if command -v shasum >/dev/null 2>&1; then
        digest=$(shasum -a 256 "$skills_dir/$old/SKILL.md")
    else
        digest=$(sha256sum "$skills_dir/$old/SKILL.md")
    fi
    digest=${digest%% *}
    case "|$known_hashes|" in
        *"|$digest|"*)
            if [[ -L "$skills_dir/$old" ]]; then
                rm -- "$skills_dir/$old"
            elif [[ ! -L "$skills_dir/$old/SKILL.md" ]]; then
                rm -- "$skills_dir/$old/SKILL.md"
                rmdir -- "$skills_dir/$old" 2>/dev/null || true
            fi
            ;;
    esac
done < "$repo_root/skills/imported-skill-renames.tsv"
