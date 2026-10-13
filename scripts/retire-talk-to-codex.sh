#!/usr/bin/env bash
set -euo pipefail
skills_dir=${1:?usage: retire-talk-to-codex.sh skills-directory}
old_dir="$skills_dir/talk-to-codex"
[[ ! -L "$old_dir" && -f "$old_dir/SKILL.md" && ! -L "$old_dir/SKILL.md" ]] || exit 0
if command -v shasum >/dev/null 2>&1; then
    digest=$(shasum -a 256 "$old_dir/SKILL.md")
else
    digest=$(sha256sum "$old_dir/SKILL.md")
fi
case "${digest%% *}" in
    5b0b230012c61d89d8296314f87a83070f1967ff356716bbab1e41a0304523a2|64243cb77f75956e4ef74efc72d9ae0f96633f20030be1e0d58c5d5b0415762a|76ee7685a87a0e24077e45bedcc9ab5c93b03d58f690f807f64d030c15eb7ceb)
        rm -- "$old_dir/SKILL.md"
        rmdir -- "$old_dir" 2>/dev/null || true
        ;;
esac
