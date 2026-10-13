#!/usr/bin/env bash
set -euo pipefail
skills_dir=${1:?usage: retire-explore.sh skills-directory}
while IFS=$'\t' read -r old known_hashes; do
    old_dir="$skills_dir/$old"
    [[ ! -L "$old_dir" && -f "$old_dir/SKILL.md" && ! -L "$old_dir/SKILL.md" ]] || continue
    if command -v shasum >/dev/null 2>&1; then digest=$(shasum -a 256 "$old_dir/SKILL.md"); else digest=$(sha256sum "$old_dir/SKILL.md"); fi
    digest=${digest%% *}
    case "|$known_hashes|" in
        *"|$digest|"*) rm -- "$old_dir/SKILL.md"; rmdir -- "$old_dir" 2>/dev/null || true ;;
    esac
done <<'RETIRED'
bob-explore	0a77c7d1cbc3a8a16f041a2c0ac483b8edb3fc2f9b20fe2162c84fb43999882f|2e63765faa2938e3f03fc8c569b8881604c31211d1a8988c9fc78ce21d4ef382|4cb849dce513f07f638cc7c0c4e33e7b7adf7af1848bb270a480582a7d804bee|52e92cc61f24715d8bcedba57ffe985c186f83b9f0f6694cbf14a60b0832ab85|ce09c48aaf5ae09e8aa487618b31e8c4c8dd19d574bfe2a5ea71b5264d5e4ca1
bob-explore-simple	187aaab9864cd23c602e10e5ee255d402ec1f938f678027884ec164ede745b67|4bceae99c26a506529eea5c82dad89aba5c54a097af1b0d19862d2f8de0faedc|8a7f7c8146a36610faca107a989321ca90a7957118970681193a291a66a660f6|ff49e798e66621e4ba57e6fdf34f7bd7970d84d5714c4316d1145f4a1d4cec39
bob-explore-teams	2d1f425f9ae1bafee64833c83efff441c0ce6873501dbfb9d18012507bcd146d|3e1f005adfbd4cbfd28cb027c46c62b45328f3f0c6d3ee5708565f9f87f061c9|667d7e0d8e65577b437a2766a92d41c90084de64b7847cb1ba336d651eb96479|7787910b35c642e8c06a2dcbc6eab55278d485c8a52e024b26b4e384a0e39baf|da1a8a0370ff41ae8cbfef28ed8ca21fc5997ee2f446a6059e8d40ea2fb39ec1|fc7e85f26e6d50d8c8c81d5738f07029ae30e84fcd9a31968567642b86543642
RETIRED
