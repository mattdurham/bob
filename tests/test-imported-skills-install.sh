#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
tmp_root=$(mktemp -d)
trap 'rm -rf "$tmp_root"' EXIT

fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
assert_file() { [[ -f "$1" ]] || fail "missing file: $1"; }
assert_contains() { grep -Fq -- "$2" "$1" || fail "$1 does not contain: $2"; }
assert_not_contains() { ! grep -Fq -- "$2" "$1" || fail "$1 unexpectedly contains: $2"; }

skills=()
while IFS= read -r skill || [[ -n "$skill" ]]; do
    [[ -z "$skill" || "$skill" == \#* ]] || skills+=("$skill")
done < "$repo_root/skills/imported-skills.txt"
[[ ${#skills[@]} -eq 27 ]] || fail "expected the 27 published upstream skills"

install_harnesses() {
    make -s -C "$repo_root" install-skills install-pi-skills install-codex-skills \
        install-wllr-skills install-goose-skills SPEC="$1" \
        CLAUDE_SKILLS_DIR="$2/claude/skills" PI_SKILLS_DIR="$2/pi/skills" \
        CODEX_HOME="$2/codex" WLLR_HOME="$2/wllr" GOOSE_HOME="$2/goose" \
        > "$tmp_root/install.log" 2>&1 || { cat "$tmp_root/install.log"; fail "harness installation failed"; }
}

check_packages() {
    local harness skill source_file relative_file installed
    for harness in claude pi codex wllr goose; do
        for skill in "${skills[@]}"; do
            installed="$1/$harness/skills/$skill"
            assert_file "$installed/SKILL.md"
            name=$skill
            [[ "$harness" != claude ]] || name=${skill/bob-matt-/bob:matt-}
            assert_contains "$installed/SKILL.md" "name: $name"
            assert_contains "$installed/SKILL.md" 'author: Matt Pocock'
            assert_contains "$installed/SKILL.md" 'source: https://github.com/mattpocock/skills/tree/'
            assert_contains "$installed/SKILL.md" 'upstream-commit: "49dd158d1076134a641b33efb035946536778336"'
            assert_contains "$installed/SKILL.md" 'license: MIT'
            while IFS= read -r source_file; do
                relative_file=${source_file#"$repo_root/skills/$skill/"}
                [[ "$relative_file" == SKILL.md ]] && continue
                assert_file "$installed/$relative_file"
                if [[ "$harness" == claude && ( "$relative_file" == *.md || "$relative_file" == *.yaml ) ]]; then
                    sed 's/bob-matt-/bob:matt-/g' "$source_file" > "$tmp_root/expected-companion"
                    cmp -s "$tmp_root/expected-companion" "$installed/$relative_file" || fail "changed companion: $harness/$skill/$relative_file"
                else
                    cmp -s "$source_file" "$installed/$relative_file" || fail "changed companion: $harness/$skill/$relative_file"
                fi
            done < <(find "$repo_root/skills/$skill" -type f)
        done
        assert_contains "$1/$harness/skills/bob-matt-pr/SKILL.md" 'author: Dex Horthy'
        assert_contains "$1/$harness/skills/bob-matt-grill-me/agents/openai.yaml" 'allow_implicit_invocation: false'
        assert_contains "$1/$harness/skills/bob-matt-tdd/LICENSE" 'Copyright (c) 2026 Matt Pocock'
    done
}

# A space in the destination exercises quoting in both Make and shell helpers.
full_root="$tmp_root/full install"
mkdir -p "$full_root/codex/skills/unrelated"
printf 'keep me\n' > "$full_root/codex/skills/unrelated/SKILL.md"
# An unchanged legacy wrapper must retire; unrelated skills must survive.
mkdir -p "$full_root/claude/skills/grill-me" "$full_root/claude/skills/code-review"
sed -e 's/^name: bob-matt-grill-me$/name: grill-me/' -e 's/"bob-matt-grilling"/"grilling"/' \
    "$repo_root/skills/bob-matt-grill-me/SKILL.md" > "$full_root/claude/skills/grill-me/SKILL.md"
printf 'custom code review\n' > "$full_root/claude/skills/code-review/SKILL.md"
for harness in claude pi codex wllr goose; do
    mkdir -p "$full_root/$harness/skills/bob-grill-me" "$full_root/$harness/skills/bob-research"
    sed 's/bob-matt-/bob-/g' "$repo_root/skills/bob-matt-grill-me/SKILL.md" > "$tmp_root/old-prefixed-wrapper"
    bash "$repo_root/scripts/render-imported-skill.sh" "$harness" "$tmp_root/old-prefixed-wrapper" \
        > "$full_root/$harness/skills/bob-grill-me/SKILL.md"
    printf 'custom research\n' > "$full_root/$harness/skills/bob-research/SKILL.md"
done
install_harnesses full "$full_root"
check_packages "$full_root"
for harness in claude pi codex wllr goose; do
    [[ ! -e "$full_root/$harness/skills/bob-grill-me/SKILL.md" ]] || fail "earlier Bob prefix remained installed: $harness"
    assert_contains "$full_root/$harness/skills/bob-research/SKILL.md" 'custom research'
    for retired in bob-explore bob-explore-simple bob-explore-teams bob-cleanup-teams bob-work-agents bob-work-teams; do
        [[ ! -e "$full_root/$harness/skills/$retired/SKILL.md" ]] || fail "removed exploration skill installed: $harness/$retired"
    done
done
[[ ! -e "$full_root/claude/skills/grill-me/SKILL.md" ]] || fail "legacy wrapper remained installed"
assert_contains "$full_root/claude/skills/code-review/SKILL.md" 'custom code review'
assert_contains "$full_root/claude/skills/bob-code-review/SKILL.md" 'Review changes without modifying them'
assert_contains "$full_root/claude/skills/bob-matt-code-review/SKILL.md" 'author: Matt Pocock'
assert_contains "$full_root/claude/skills/bob-matt-implement/SKILL.md" '"bob:matt-code-review"'
assert_contains "$full_root/claude/skills/bob-matt-grill-me/SKILL.md" '"bob:matt-grilling"'

assert_contains "$full_root/codex/skills/unrelated/SKILL.md" 'keep me'
bash "$repo_root/scripts/render-imported-skill.sh" claude "$repo_root/skills/bob-matt-implement/SKILL.md" > "$tmp_root/expected-claude"
cmp -s "$tmp_root/expected-claude" "$full_root/claude/skills/bob-matt-implement/SKILL.md" || fail "Claude import changed shared instructions"
assert_not_contains "$full_root/codex/skills/bob-matt-grill-me/SKILL.md" 'disable-model-invocation:'
assert_not_contains "$full_root/codex/skills/bob-matt-teach/SKILL.md" 'argument-hint:'
assert_not_contains "$full_root/codex/skills/bob-matt-implement/SKILL.md" 'Skill tool'
assert_contains "$full_root/pi/skills/bob-matt-implement/SKILL.md" 'user-invocable: true'
assert_contains "$full_root/wllr/skills/bob-matt-implement/SKILL.md" 'user-invocable: true'
assert_contains "$full_root/goose/skills/bob-matt-implement-spec/SKILL.md" 'delegate('
assert_contains "$full_root/goose/skills/bob-matt-implement-spec/SKILL.md" 'load(source:'
assert_not_contains "$full_root/goose/skills/bob-matt-implement/SKILL.md" 'Skill tool'
assert_contains "$full_root/goose/skills/bob-matt-grill-me/SKILL.md" 'only when the user explicitly requests it'
assert_file "$full_root/goose/agents/workflow-implementer.md"
assert_file "$full_root/goose/agents/explore.md"

# Reinstalling must keep complete packages and leave unrelated skills in place.
install_harnesses full "$full_root"
assert_contains "$full_root/codex/skills/unrelated/SKILL.md" 'keep me'

simple_root="$tmp_root/simple"
install_harnesses simple "$simple_root"
check_packages "$simple_root"
for harness in claude pi codex wllr goose; do
    assert_file "$simple_root/$harness/skills/bob-ship/SKILL.md"
    assert_file "$simple_root/$harness/skills/bob-ship/references/publication-procedure.md"
    cmp -s "$repo_root/skills/bob-ship/references/publication-procedure.md" "$simple_root/$harness/skills/bob-ship/references/publication-procedure.md" || fail "publication procedure changed: $harness"
done
assert_file "$simple_root/codex/skills/bob-work-simple/SKILL.md"
[[ ! -e "$simple_root/codex/skills/bob-work/SKILL.md" ]] || fail "simple mode installed a normal Bob work skill"

# An earlier skills.sh/local-link install must not be modified through a symlink.
linked_root="$tmp_root/linked"
mkdir -p "$linked_root/skills" "$linked_root/upstream"
printf 'original upstream\n' > "$linked_root/upstream/SKILL.md"
ln -s "$linked_root/upstream" "$linked_root/skills/bob-matt-tdd"
bash "$repo_root/scripts/install-imported-skills.sh" codex "$linked_root/skills" > /dev/null
[[ ! -L "$linked_root/skills/bob-matt-tdd" ]] || fail "linked installation was not replaced"
assert_contains "$linked_root/upstream/SKILL.md" 'original upstream'
assert_contains "$linked_root/skills/bob-matt-tdd/SKILL.md" 'author: Matt Pocock'

# Default runtime detection includes Pi as well as the previously supported CLIs.
fake_bin="$tmp_root/bin"
mkdir -p "$fake_bin"
for harness in pi codex wllr goose; do
    printf '#!/usr/bin/env bash\nexit 0\n' > "$fake_bin/$harness"
    chmod +x "$fake_bin/$harness"
done
auto_root="$tmp_root/auto"
PATH="$fake_bin:$PATH" make -s -C "$repo_root" install-runtime-skills \
    PI_SKILLS_DIR="$auto_root/pi/skills" CODEX_HOME="$auto_root/codex" \
    WLLR_HOME="$auto_root/wllr" GOOSE_HOME="$auto_root/goose" \
    > "$tmp_root/auto.log" 2>&1 || { cat "$tmp_root/auto.log"; fail "automatic installation failed"; }
for harness in pi codex wllr goose; do
    assert_file "$auto_root/$harness/skills/bob-matt-tdd/SKILL.md"
    assert_file "$auto_root/$harness/skills/bob-matt-teach/LEARNING-RECORD-FORMAT.md"
done

printf 'PASS: Imported skills install into all five harnesses with attribution and companions\n'
