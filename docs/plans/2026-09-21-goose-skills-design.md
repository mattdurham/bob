# Goose Skills Support Design

## Goal

Add first-class Goose support to Bob. Bob's workflows should be discoverable as
Goose Agent Skills, delegate work to Goose custom agents, preserve the existing
`.bob/state/` coordination model, and be installed by the default `make install`
command when Goose is available.

## Architecture

Goose discovers global skills in `~/.agents/skills/<name>/SKILL.md` and custom
agents in `~/.agents/agents/<name>.md`. Bob will add an explicit
`install-goose-skills` Make target that installs both sets of files. The default
`make install` target will invoke it when the `goose` executable is present.

Workflow skills that contain runtime-specific orchestration will have committed
`SKILL.goose.md` variants. The installer will prefer those variants and fall
back to compatible generic or simple skill sources where no Goose-specific
behavior is needed. Goose variants will use the built-in Summon extension's
`delegate` and `load` tools. Independent delegates will use `async: true` and
the parent orchestrator will collect their results before advancing.

Bob's specialist prompts will be installed as Goose custom agents. The
installer will normalize their frontmatter to Goose's supported `name` and
`description` fields and render companion-file paths. Agent prompts that assume
Claude task lists, persistent teammates, or nested subagents will receive
Goose-specific variants or adapters so that each delegate receives a bounded,
explicit assignment. Because Goose subagents cannot spawn further subagents,
nested orchestration remains in the parent workflow skill.

## Workflow Semantics

The Goose port preserves Bob's phase ordering, state artifacts, review gates,
loop-back rules, and spec-driven development checks. Claude-only concepts such
as experimental agent-team flags, `TaskCreate`, `TaskList`, `TaskUpdate`, and
`SendMessage` will not appear as executable requirements in installed Goose
skills. Their coordination role will be replaced by explicit delegate prompts,
parallel delegate calls, returned task identifiers, `load`, and durable files
under `.bob/state/`.

Failures and timeouts will be surfaced by the parent workflow. A failed
parallel delegate will not be silently treated as success; the workflow will
retry when its existing loop budget permits or stop with the failed phase and
missing artifact identified.

## Installation and Compatibility

`make install-goose-skills` will accept `SPEC=simple` and a configurable
`GOOSE_HOME`, defaulting to `$(HOME)/.agents`. It will install skills and agents
without requiring the Goose binary, which keeps the target testable and useful
for custom environments. `make install` will auto-detect `goose` before invoking
the target, matching the current Codex and wllr behavior.

The install is additive and idempotent. It will not remove unrelated user
skills or agents. Generated version metadata will follow the existing Bob
version-skill behavior.

## Verification

Tests will install into a temporary `GOOSE_HOME` and verify:

- expected skill and custom-agent paths exist;
- installed skill frontmatter contains Goose-compatible names;
- Goose-specific workflows use `delegate`/`load` and omit Claude team gates;
- `SPEC=simple` installs only simple variants;
- repeated installation is safe;
- the existing CI and installer checks continue to pass.

README and Make help text will document automatic and explicit Goose
installation, the Summon/autonomous-mode prerequisite, and how to list or invoke
the installed skills and agents.
