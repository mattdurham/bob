# Imported skills

Bob includes the 27 published engineering and productivity skills from
[Matt Pocock's skills repository](https://github.com/mattpocock/skills), imported
from commit `49dd158d1076134a641b33efb035946536778336`. The source checkout was
clean at import. Miscellaneous, beta, and deprecated skills are not part of this
published set.

The imported commands use `bob:matt-<skillname>` in Claude Code and
`bob-matt-<skillname>` in Pi, Codex, wllr, and Goose. Source package directories
use the hyphenated spelling; the Claude installer renders names and cross-skill
calls with the colon namespace. Matt's code review stays separate from native
`bob-code-review`. Run `bob:matt-setup-matt-pocock-skills` in a Claude project
to configure its tracker and documentation conventions. Installation does not
run that setup. Attribution records the original upstream name and URL.

Every imported `SKILL.md` carries `license: MIT` and frontmatter metadata with
the original author, source directory URL, and upstream commit. Each package
includes the upstream MIT license. Existing attribution, including Dex Horthy's
credits in `bob-matt-pr`, is preserved along with all supporting resources and
`agents/openai.yaml` invocation policies.

## Catalog

| Engineering | Productivity |
|---|---|
| `bob:matt-ask-matt` | `bob:matt-grill-me` |
| `bob:matt-code-review` | `bob:matt-grilling` |
| `bob:matt-codebase-design` | `bob:matt-handoff` |
| `bob:matt-diagnosing-bugs` | `bob:matt-teach` |
| `bob:matt-domain-modeling` | `bob:matt-to-questionnaire` |
| `bob:matt-grill-with-docs` | `bob:matt-wait-what` |
| `bob:matt-implement` | `bob:matt-writing-for-agents` |
| `bob:matt-implement-spec` | |
| `bob:matt-improve-codebase-architecture` | |
| `bob:matt-pr` | |
| `bob:matt-prototype` | |
| `bob:matt-research` | |
| `bob:matt-retro` | |
| `bob:matt-setup-matt-pocock-skills` | |
| `bob:matt-tdd` | |
| `bob:matt-to-spec` | |
| `bob:matt-to-tickets` | |
| `bob:matt-triage` | |
| `bob:matt-wayfinder` | |
| `bob:matt-wizard` | |

The checked-in installation manifest is [imported-skills.txt](../skills/imported-skills.txt).
Supporting files live inside each skill package and are copied with it.

## Installation

`make install` installs the packages into Claude Code, then installs them for
Pi, Codex, wllr, and Goose when each CLI is present on `PATH`. Explicit targets
also work without those CLIs:

| Harness | Make target | Default skill directory |
|---|---|---|
| Claude Code | `make install-skills` | `~/.claude/skills` |
| Pi | `make install-pi-skills` or `make install-pi` | `~/.pi/agent/skills` |
| Codex | `make install-codex-skills` | `~/.codex/skills` |
| wllr | `make install-wllr-skills` | `~/.wllr/skills` |
| Goose | `make install-goose-skills` | `~/.agents/skills` |

The imported set installs with both `SPEC=full` and `SPEC=simple` under its
Bob-prefixed names. It does not create additional `-simple` siblings.

For isolated installation or custom locations, use `CLAUDE_SKILLS_DIR`,
`PI_SKILLS_DIR`, `CODEX_HOME`, `WLLR_HOME`, and `GOOSE_HOME`. Existing unrelated
skills are preserved. Known unchanged entrypoints under the original imported
names and the earlier `bob-` names are retired after their prefixed replacements install; customized legacy
entrypoints and companion files are preserved. A prior symlink at an imported package's destination is
replaced with a package copy, without modifying its link target.

## Bob interview behavior

`bob-matt-grill-me` and `bob-matt-grill-with-docs` use `bob-matt-grilling`. Bob customizes that
interview to ask one question at a time and wait for each answer, preserving
dependency order and previously confirmed decisions. This differs from the
upstream batch-question rounds. Original attribution and invocation policies
remain in place.

## Harness adaptations

[render-imported-skill.sh](../scripts/render-imported-skill.sh) generates the
installed entrypoint from the shared source. Claude keeps the shared
instructions and invocation fields. Other harnesses use skill-file loading
language instead of Claude's Skill tool. Codex keeps explicit invocation policy
in the original `agents/openai.yaml`; Pi and wllr receive user-invocable metadata.
Goose retains license and attribution frontmatter and an explicit-request
instruction for skills that upstream makes manual-only.

Worker instructions use native collaboration/subagent tools on Codex, Pi, and
wllr. Goose uses Summon `delegate`/`load`, with delegation owned by the parent
session and bounded workers. Goose's existing Bob specialist agents are
installed by its normal target; delegated work requires Summon and autonomous
mode. These adaptations preserve the requested task rather than introducing
new workflow phases.

The installer tests validate the five generated package formats, supporting
resources, attribution, repeat installation, simple mode, symlink replacement,
and automatic runtime detection. They do not execute the imported workflows
against live issue trackers or deployment systems.

## Research and exploration

`bob:matt-research` replaces `bob-explore` and `bob-explore-teams`. Bob extends
upstream research with codebase tracing, documented contracts, counterexample
checks, and explicit uncertainties. Delegation is optional. The investigation
is read-only except for its Markdown report. Known unchanged exploration
entrypoints are retired after research installs; custom copies are preserved.
