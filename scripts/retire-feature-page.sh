#!/usr/bin/env bash
set -euo pipefail
# Remove only known Bob entrypoints; preserve customized files and companions.
skills_dir=${1:?usage: retire-feature-page.sh skills-directory}
old_dir="$skills_dir/bob-generate-feature-page"
[[ ! -L "$old_dir" && -f "$old_dir/SKILL.md" && ! -L "$old_dir/SKILL.md" ]] || exit 0
if command -v shasum >/dev/null 2>&1; then
    digest=$(shasum -a 256 "$old_dir/SKILL.md")
else
    digest=$(sha256sum "$old_dir/SKILL.md")
fi
case "${digest%% *}" in
    2d3af599da00f6c83118c38aa5d45e33106aab1f225f44419b1eef9c8b59a91b|53b0fd948961d58346ae4fc17932c2911bd0c3760a0a255d245dd72202f2e471|582391c65347e293305f847d6c50058cd24cbf9e537b57c8c41f8033ec41481e|62e0975ae4b9588816c0dbf721714b0bd2d9ff17845cd007c0848d6c20e86eca|6acbd4d74f7128bd5071abd6d4c530777df216c5b0cc2d1c57d9f90c52aceb81|d245f05c045e0c157664bf3adecc2e8b92be7a7f5ccf64432d3f3066b98bd3bd)
        rm -- "$old_dir/SKILL.md"
        rmdir -- "$old_dir" 2>/dev/null || true
        ;;
esac
