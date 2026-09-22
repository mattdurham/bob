#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
tmp_root=$(mktemp -d)
trap 'rm -rf "$tmp_root"' EXIT

fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
assert_file() { [[ -f "$1" ]] || fail "missing file: $1"; }
assert_contains() { grep -Fq -- "$2" "$1" || fail "$1 does not contain: $2"; }
assert_not_contains() { ! grep -Fq -- "$2" "$1" || fail "$1 unexpectedly contains: $2"; }
frontmatter_keys() {
    awk 'NR == 1 && $0 == "---" { in_fm=1; next }
         in_fm && $0 == "---" { exit }
         in_fm && /^[A-Za-z0-9_-]+:/ { sub(/:.*/, ""); print }' "$1"
}

fixture="$tmp_root/fixture.md"
printf '%s\n' '---' 'name: "bob:test"' 'description: Example renderer fixture' \
    'tools: Read, Write' 'model: sonnet' '---' \
    'Body with [agent-directory]/reference.md.' > "$fixture"
bash "$repo_root/scripts/render-goose-skill.sh" "$fixture" bob-test > "$tmp_root/skill.md"
bash "$repo_root/scripts/render-goose-agent.sh" "$fixture" "$tmp_root/support" > "$tmp_root/agent.md"
[[ "$(frontmatter_keys "$tmp_root/skill.md")" == $'name\ndescription' ]] || fail "skill renderer leaked frontmatter"
assert_contains "$tmp_root/skill.md" 'name: bob-test'
assert_contains "$tmp_root/agent.md" "$tmp_root/support/reference.md"

full_home="$tmp_root/full"
make -C "$repo_root" install-goose-skills GOOSE_HOME="$full_home"

assert_file "$full_home/skills/bob-work/SKILL.md"
assert_file "$full_home/skills/bob-work-simple/SKILL.md"
assert_file "$full_home/skills/bob-version/SKILL.md"
assert_file "$full_home/agents/workflow-brainstormer.md"
assert_file "$full_home/agents/workflow-implementer.md"
assert_file "$full_home/agents/workflow-implementer/golang-pro.md"
assert_file "$full_home/agents/commit-agent/scripts/push-with-gate.sh"
[[ "$(frontmatter_keys "$full_home/skills/bob-work/SKILL.md")" == $'name\ndescription' ]] || fail "skill frontmatter is not Goose-compatible"
[[ "$(frontmatter_keys "$full_home/agents/workflow-implementer.md")" == $'name\ndescription' ]] || fail "agent frontmatter is not Goose-compatible"
assert_contains "$full_home/skills/bob-work/SKILL.md" 'delegate('
assert_contains "$full_home/skills/bob-work/SKILL.md" 'load(source:'
assert_not_contains "$full_home/skills/bob-work/SKILL.md" 'CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS'
assert_not_contains "$full_home/skills/bob-work/SKILL.md" 'TaskCreate'
assert_not_contains "$full_home/agents/workflow-brainstormer.md" 'Task('
assert_not_contains "$full_home/agents/workflow-coder.md" 'delegate('
assert_contains "$full_home/agents/workflow-implementer.md" "$full_home/agents/workflow-implementer/golang-pro.md"

orchestrated_skills=(
    bob-work bob-work-agents bob-work-teams bob-explore bob-explore-teams
    bob-audit bob-code-review bob-cleanup bob-cleanup-teams bob-design
    bob-stage-prs bob-adversarial-review bob-challenge-idea bob-operational
    bob-internal-writing-plans
)
for skill in "${orchestrated_skills[@]}"; do
    installed="$full_home/skills/$skill/SKILL.md"
    assert_file "$installed"
    assert_contains "$installed" 'delegate('
    assert_contains "$installed" 'load(source:'
    for forbidden in CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS TaskCreate TaskList TaskGet TaskUpdate SendMessage TeamCreate subagent_type: run_in_background:; do
        assert_not_contains "$installed" "$forbidden"
    done
done

bounded_agents=(
    workflow-brainstormer workflow-coder team-analyst team-challenger
    team-reviewer team-spec-oracle architecture-introspector bug-finder
    spec-doc-reviewer monitor-agent
)
for agent in "${bounded_agents[@]}"; do
    installed="$full_home/agents/$agent.md"
    assert_file "$installed"
    assert_not_contains "$installed" 'delegate('
    for forbidden in 'Task(' TaskCreate TaskList TaskGet TaskUpdate SendMessage TeamCreate; do
        assert_not_contains "$installed" "$forbidden"
    done
done

make -C "$repo_root" install-goose-skills GOOSE_HOME="$full_home"

simple_home="$tmp_root/simple"
make -C "$repo_root" install-goose-skills SPEC=simple GOOSE_HOME="$simple_home"
assert_file "$simple_home/skills/bob-work-simple/SKILL.md"
[[ ! -e "$simple_home/skills/bob-work/SKILL.md" ]] || fail "SPEC=simple installed a full skill"
assert_not_contains "$simple_home/skills/bob-adversarial-review-simple/SKILL.md" 'delegate('

fake_bin="$tmp_root/bin"
mkdir -p "$fake_bin"
printf '#!/usr/bin/env bash\nexit 0\n' > "$fake_bin/goose"
chmod +x "$fake_bin/goose"
auto_home="$tmp_root/auto"
PATH="$fake_bin:$PATH" make -C "$repo_root" install-runtime-skills \
    GOOSE_HOME="$auto_home" CODEX_HOME="$tmp_root/codex" WLLR_HOME="$tmp_root/wllr" >/dev/null
assert_file "$auto_home/skills/bob-work/SKILL.md"

absent_home="$tmp_root/absent"
PATH="/usr/bin:/bin" make -C "$repo_root" install-runtime-skills \
    GOOSE_HOME="$absent_home" CODEX_HOME="$tmp_root/codex-absent" \
    WLLR_HOME="$tmp_root/wllr-absent" > "$tmp_root/absent.log"
assert_contains "$tmp_root/absent.log" 'Goose CLI not installed — skipping Goose skills'
[[ ! -e "$absent_home" ]] || fail "runtime detection created Goose home without Goose"

for phrase in 'make install-goose-skills' 'GOOSE_HOME' '~/.agents/skills/' \
    '~/.agents/agents/' 'Summon' 'autonomous' 'goose skills list'; do
    assert_contains "$repo_root/README.md" "$phrase"
done

printf 'PASS: Goose installer contract\n'
