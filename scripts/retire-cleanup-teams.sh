#!/usr/bin/env bash
set -euo pipefail
skills_dir=${1:?usage: retire-cleanup-teams.sh skills-directory}
# Retire only once the replacement workflow is present.
[[ -f "$skills_dir/bob-cleanup/SKILL.md" ]] || exit 0
old_dir="$skills_dir/bob-cleanup-teams"
[[ ! -L "$old_dir" && -f "$old_dir/SKILL.md" && ! -L "$old_dir/SKILL.md" ]] || exit 0
if command -v shasum >/dev/null 2>&1; then digest=$(shasum -a 256 "$old_dir/SKILL.md"); else digest=$(sha256sum "$old_dir/SKILL.md"); fi
digest=${digest%% *}
case "|62672f81d02164237d748cfea0bfa90aa2ddf24f25267e9a301657bda9649f94|207138df3a25ee093cdc3ac2ccd0657fd2454a7d5010097147fbfbe8415f0236|26ae11d3ef2cee771b1dab6ef544a5e2c4a446d6faf824d098a41529dd30216b|27ec13cae7e3390a09ea51bef3cf45dd30efff9af6fa6249e913e786854609c5|96b52883cf467b73c7ca30821372fea01a3d1d9453138f33aa0d473591ba6106|d517a727a9454e8b8903422ba5f4a4bbf89b9ddf3091db7b404ee0ec2698bb45|" in
    *"|$digest|"*) rm -- "$old_dir/SKILL.md"; rmdir -- "$old_dir" 2>/dev/null || true ;;
esac
