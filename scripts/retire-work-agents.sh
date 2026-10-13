#!/usr/bin/env bash
set -euo pipefail
skills_dir=${1:?usage: retire-work-agents.sh skills-directory}
[[ -f "$skills_dir/bob-work/SKILL.md" || -f "$skills_dir/bob-work-simple/SKILL.md" ]] || exit 0
old_dir="$skills_dir/bob-work-agents"
[[ ! -L "$old_dir" && -f "$old_dir/SKILL.md" && ! -L "$old_dir/SKILL.md" ]] || exit 0
if command -v shasum >/dev/null 2>&1; then digest=$(shasum -a 256 "$old_dir/SKILL.md"); else digest=$(sha256sum "$old_dir/SKILL.md"); fi
digest=${digest%% *}
case "|1a6a6b67be0e932b91b4212b4535bf6f0155c609500fc4f2e3c271daee93a452|2b8d507bf54fc5dbee4d9bad82e85c5694d7ff2487202737513ba347725b0ff4|32634ccada88e0676215d9e68dfc6ffedd89bae9ae1a69130f436d3c4689aa73|660227206bb2495a06e5a30749eec793ea7ae1395fa6f7dd0e5a7a9b44d79c98|78e8fcd0980735719758c911ec2e1e36555e77b4433351b3e3715fcee19f1b36|7d3713edeea59a2f9492bbd7236694502639ce4f4031c3a1457add708629626b|97c501355a30720080c8104b1b15f854b41a9da614ecb7f52f2998e804374889|983da2aff1a0ed553baa3d78f53efee45e8ce3a7bed13021a6fba16f07fbddcf|9f7cf5dc84cef6fed1ef40b2a4ad527e8e67f3c75fe88d37e9a661483975f25f|c5fcc74b0c23d883573d6d2776b6fc587c1736dad4cc1ee0e5824e68edbcdd31|cab2d1bd44fc1890f61f1a9842212234e6ddb82630f251bf6b79d4ff8c2ac9fa|" in
    *"|$digest|"*) rm -- "$old_dir/SKILL.md"; rmdir -- "$old_dir" 2>/dev/null || true ;;
esac
