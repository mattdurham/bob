#!/usr/bin/env bash
set -euo pipefail
skills_dir=${1:?usage: retire-work-teams.sh skills-directory}
[[ -f "$skills_dir/bob-work/SKILL.md" || -f "$skills_dir/bob-work-simple/SKILL.md" ]] || exit 0
old_dir="$skills_dir/bob-work-teams"
[[ ! -L "$old_dir" && -f "$old_dir/SKILL.md" && ! -L "$old_dir/SKILL.md" ]] || exit 0
if command -v shasum >/dev/null 2>&1; then digest=$(shasum -a 256 "$old_dir/SKILL.md"); else digest=$(sha256sum "$old_dir/SKILL.md"); fi
digest=${digest%% *}
case "|134d50509d980c1769f4afb640196e492b6cef5b77e410c7080d7dfa7ddf0064|16e1e3580e5f5b961531a4b4a7a8740c7e8d4bb628ad13e1e37cb338526f9f0f|2ecab887cb1e294d2ef281c71732887a8642b0094c2fe3b5a647ffe182a949c7|45aa76d70268b151da757abc469c465c218f2b20d3bf7903aeb97021dc6f2ec9|4dbee07665c43ae370a90e42591b5d5447d36834267c50f55b000e96d816cd61|51ea84d9a5a70528c71df30f7da6d78adbd80cdfce6541401dcb54716d696c75|6802b38c55262ed279ea2cb9b9f5b7a3d1cb0ab03cb78156bfbe3f1993d61bd7|7365a280f83d910c3888dd20b3607377c388010310b619ee2f52bd80a10e69e3|743aa85ecffdc9a96468323e3ab08ec2054bea8ba359bd1671efb23b702f1619|83402e951a6133358003b99b7d058f9ae478ce5e8ec9b7de479d4ca34538dd4a|8f80e0ce182aa4fb5b00942ca2f43c9e2a718f48af92c6878929e0ca83e1f65a|b7949be1fde42da2a229cbd4b149d3d36167f354c5300450ed63f9b9ed0e2694|c6d445eeb37446e91514dec12f1e8fb7304b5ee77b5d3398855039378b6bfce8|daffa7c17fac0e69664049594a04753a610e7450658e353dfb57afc58c7f06c4|e877a7cb028ccfb0c6784cbaafb6646cecbba6437ba4c0d37dda557dd70efef7|" in
    *"|$digest|"*) rm -- "$old_dir/SKILL.md"; rmdir -- "$old_dir" 2>/dev/null || true ;;
esac
