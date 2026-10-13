#!/usr/bin/env bash
set -euo pipefail
skills_dir=${1:?usage: retire-internal-wrappers.sh skills-directory}
# Retire known Bob entrypoints; preserve customized files and companion resources.
for name in bob-internal-brainstorming bob-internal-brainstorming-simple bob-internal-writing-plans bob-internal-writing-plans-simple; do
    old_dir="$skills_dir/$name"
    [[ ! -L "$old_dir" && -f "$old_dir/SKILL.md" && ! -L "$old_dir/SKILL.md" ]] || continue
    if command -v shasum >/dev/null 2>&1; then
        digest=$(shasum -a 256 "$old_dir/SKILL.md")
    else
        digest=$(sha256sum "$old_dir/SKILL.md")
    fi
    case "$name:${digest%% *}" in
        bob-internal-brainstorming:55321318bff7b63f5609eb5604fb1139a8e509584d717d236362b22f4509c9ed|bob-internal-brainstorming:61515ad3d1669e19b56755bcfbb150d7cc57b55cc60fee9182932c1ee780ad7b|bob-internal-brainstorming:64d0fbb5820f14ef53d68f58ae6cc8c6ae9b49ce1b39d445f69d61f056c0a9b8|bob-internal-brainstorming-simple:04bb85dee8534d5c087911f46bb8eb276b7122908e01bc47642afd42d914d12d|bob-internal-brainstorming-simple:a5284bf950d227b52e1916856c9478e44ed716a49826446e84660babedeac23d|bob-internal-brainstorming-simple:e91b5033f4cfefebd89096d75926fb161d4d4058f6505d5012ba2ecfe18e4d9a|bob-internal-writing-plans:9832e545748018fcf03a9a56da5d80e7b2ccde4db1a42174450b4a8874cc077e|bob-internal-writing-plans:a5d7b029f85364170c0f4ddf3cef69a4b9ad16d3465da2cb3551dd6bb8c8138b|bob-internal-writing-plans:ee72bc2d5b957345f76ebb8292b8a0123ee7250b3153a44b774d6385a60e885c|bob-internal-writing-plans:fd92b9635f9f8463b32d3ec83001a8b3a24451820acbdd8303928f29a00fd325|bob-internal-writing-plans-simple:0e3beeabc7161e795d9090445a48c0d523efb03a6a767fdbde13e6cb9a771802|bob-internal-writing-plans-simple:9b27fd67b05ed41e7b0401012cf2af6dd96f3482db456429d0c29fdff7ecd5b8|bob-internal-writing-plans-simple:b575fe6773052df1f5cd46597ed0198d54c4c768900a02fc57530728fa1196a3)
            rm -- "$old_dir/SKILL.md"
            rmdir -- "$old_dir" 2>/dev/null || true
            ;;
    esac
done
