# Belayin' Pin Bob

```
                                     |    |    |
                                    )_)  )_)  )_)
                                   )___))___))___)\
                                  )____)____)_____)\\
                                _____|____|____|____\\\__
                       ---------\                   /---------
                         ^^^^^ ^^^^^^^^^^^^^^^^^^^^^
                           ^^^^      ^^^^     ^^^    ^^
                                ^^^^      ^^^
```

Workflow orchestration for Claude Code, Goose, Pi, Codex, and wllr through
skills and subagents.

## What is Bob?

Bob coordinates AI agent workflows for feature development. Skills invoke specialized subagents, pass state through `.bob/` artifacts, and enforce quality gates automatically.

## Spec-Driven Development

Bob treats **SPECS.md as the source of truth** for module behavior. Every workflow is spec-aware:

- **`/bob:work`** reads existing specs before making changes. If a request contradicts a contract or invariant in SPECS.md, the workflow will question it — specs can be changed, but only deliberately. Code changes to spec-driven modules must be reflected in the corresponding spec docs.

- **`/bob:explore`** prioritizes spec docs when analyzing a codebase. For spec-driven modules, it reads SPECS.md and NOTES.md first to understand contracts and design decisions before diving into implementation code. Uses concurrent analysis and adversarial challenge phases for deep, reliable exploration.

A spec-driven module is any directory containing SPECS.md, NOTES.md, TESTS.md, BENCHMARKS.md, or `.go` files with this comment:

```go
// NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.
```

## Quick Start

```bash
git clone https://github.com/mattdurham/bob.git
cd bob
make install
```

This installs workflow skills to `~/.claude/skills/` and subagents to
`~/.claude/agents/`. When the `goose` CLI is on `PATH`, it also installs
Goose-native skills to `~/.agents/skills/` and custom agents to
`~/.agents/agents/`. Restart or open a new agent session after installation.

The default install publishes both normal skills and `-simple` siblings. These
simple files are alternate workflow specifications; they may still use agents when
their workflow requires them. The `bob-adversarial-review-simple` variant is the
exception: it is explicitly a single-reviewer, no-subagent review. To install only
the simple variants, use `SPEC=simple`:

```bash
make install SPEC=simple
```

For example, this provides both `/bob:adversarial-review` and
`/bob:adversarial-review-simple`; the latter writes findings to
`.bob/state/review.md` without spawning subagents. The same naming applies to the
Pi, Codex, wllr, and Goose install targets.

## Goose

Bob provides Goose-native versions of every orchestration-heavy workflow. The
parent skill owns phase routing and delegates bounded work through Goose's
Summon extension using `delegate` and `load`; custom agents never recursively
delegate.

The default `make install` detects Goose automatically. To install Goose support
explicitly, without requiring the Goose executable to be present:

```bash
make install-goose-skills
make install-goose-skills SPEC=simple
make install-goose-skills GOOSE_HOME=/custom/.agents
```

`GOOSE_HOME` defaults to `~/.agents`. Skills are installed under
`~/.agents/skills/<name>/SKILL.md`, and specialist agents are installed under
`~/.agents/agents/<name>.md`. Existing unrelated skills and agents are left
untouched.

Goose's Skills platform extension is enabled by default. Full Bob workflows
also require the Summon extension and autonomous permission mode; Goose does
not allow subagents in manual approval, smart approval, or chat-only modes.
Enable Summon with `goose configure` if it has been disabled. Start a new Goose
session, then use `goose skills list` or `/skills` to inspect the installed
skills and invoke one by name, such as `bob-work`.

See the [Goose Agent Skills guide](https://goose-docs.ai/docs/guides/context-engineering/using-skills/)
and [Goose subagent guide](https://goose-docs.ai/docs/guides/context-engineering/subagents/)
for runtime details.

## Workflows

### `/bob:work` — Concurrent Agent Team Workflow

```
INIT → WORKTREE → BRAINSTORM → PLAN → SPAWN TEAM → EXECUTE ↔ REVIEW → COMMIT → MONITOR → COMPLETE
```

Multiple coder and reviewer teammates work in parallel through a shared task list. Requires `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`.

### `/bob:audit` — Spec Audit

```
INIT → DISCOVER → AUDIT → REPORT → COMPLETE
```

Verify code satisfies stated invariants in spec-driven modules. Read-only — reports drift but doesn't fix it.

### `/bob:explore` — Team-Based Exploration with Adversarial Challenge

```
INIT → DISCOVER → ANALYZE (4 agents) → CHALLENGE (5 agents) → DOCUMENT → COMPLETE
                     ↑                       ↓
                     └───────────────────────┘
                          (any FAIL, max 2 loops)
```

Concurrent specialist agents for codebase exploration. ANALYZE spawns 4 agents (structure, flow, patterns, dependencies). CHALLENGE spawns adversarial agents that stress-test the analysis. Failures loop back to re-analyze. No code changes.

## Loop-Back Rules

All work workflows enforce these routing rules:

| Trigger | Route to | Reason |
|---------|----------|--------|
| CRITICAL/HIGH review issues | BRAINSTORM | Re-think the approach |
| MEDIUM/LOW review issues | EXECUTE | Targeted fixes |
| Test failures | EXECUTE | Fix the code |
| CI failures or PR feedback | BRAINSTORM | Always re-brainstorm |

REVIEW is mandatory — it cannot be skipped even if tests pass.

## Subagents

| Agent | Phase | Purpose |
|-------|-------|---------|
| workflow-brainstormer | BRAINSTORM | Research and creative ideation |
| workflow-planner | PLAN | Implementation planning |
| workflow-coder | EXECUTE | Code implementation (TDD) |
| workflow-implementer | EXECUTE | Used by workflow-coder and design |
| workflow-tester | TEST | Test execution and quality checks |
| review-consolidator | REVIEW | Multi-domain code review |
| commit-agent | COMMIT | Git operations and PR creation |
| monitor-agent | MONITOR | CI/CD and PR monitoring |
| team-coder | EXECUTE | Concurrent coder teammate |
| team-reviewer | REVIEW | Concurrent reviewer teammate |
| Explore | DISCOVER | Codebase exploration |

## Per-Repo Verification Commands

By default the TEST phases run bob's Go toolchain (`make ci`, or `go test`/`go fmt`/`golangci-lint`/`gocyclo` individually). Repos in other languages can override this with a `.bob/config` file at the repository root containing one `verify:` line per command:

```
verify: pnpm install --frozen-lockfile
verify: pnpm test
verify: pnpm exec tsc --noEmit
```

A line counts only if it starts with `verify: ` (colon + space) and the remainder holds at least one non-whitespace byte; the check runs under `LC_ALL=C`, so whitespace is read under the C locale. Examples of ignored lines: a `verify: ` whose remainder is only spaces and tabs, a bare `verify:` with nothing after it, a `verify:` followed by a tab instead of the required space, and `verify:foo` with no space. Write `.bob/config` as NUL-free text with LF line endings and plain spaces. Commands run exactly as written, serially, from the repository root; any nonzero exit fails verification; when at least one `verify:` line counts the default Go steps are skipped. Activation is presence-based: the tester reads the file from the working tree on every run, whether or not it is committed.

To commit the file in a repo that ignores `.bob` paths: `!.bob/config` works only when it comes after a `.bob/*` rule. If the directory itself is ignored (`.bob/`), git never descends into it, so re-include the parent first:

```
!.bob/
.bob/*
!.bob/config
```

## Pre-Publish Gate Hook (opt-in, per repo)

A repo can veto bob's publication step by shipping an executable
`.bob/hooks/pre-publish` (normally committed with executable mode 100755 — a
committed hook is compared byte-for-byte against HEAD, and one locally
altered, deleted, or replaced by a symlink is refused; any symlink at the
hook path is refused). Before the standard publication path
(`/bob:code-review`, including when the `/bob:work` variants invoke it) pushes
or creates a PR, commit-agent runs the hook from the repo top level with no
arguments and stdin reading end-of-file — a bob-specific contract, not git's pre-push
interface. Exit 0 lets the push proceed; any other exit (or a broken hook, or
a hook that overruns its 90-second budget) blocks both the push and the PR
(when the hook ran, the last 8 KiB of its output are quoted in the failure
report). Repos without a hook get the plain push through the same script (which uses
an explicit refspec, so `remote.origin.push` remaps are ignored). Hooks must not commit, move the branch, or modify files.
Suggested uses: size limits, scope checks, secret scans.

Manual pushes and bob-stage-prs' own publication steps do not run this gate.
If your repo gitignores `.bob/`, add the hook with
`git add -f .bob/hooks/pre-publish` (or re-include the path) so the gate
travels with the repo. On macOS the hook budget needs GNU coreutils
(`timeout` or `gtimeout`).

## Git Worktrees

All work workflows create isolated git worktrees before any file operations:

```
repo/
repo-worktrees/
  ├── add-auth/          # Feature worktree
  │   ├── .bob/state/    # Workflow artifacts
  │   └── ...
  └── fix-parser/
      ├── .bob/state/
      └── ...
```

## Confirm Before Push (opt-in)

By default bob publishes autonomously at COMMIT. To have `/bob:code-review`
(and the `/bob:work` variants, which route COMMIT through it) pause and show
you every commit the push will publish and the proposed PR body verbatim
before anything is pushed, add to
`~/.claude/settings.json`:

```json
{
  "env": {
    "BOB_CONFIRM_BEFORE_PUSH": "1"
  }
}
```

Or set an environment variable (Pi and other harnesses read only the
environment):
```bash
export BOB_CONFIRM_BEFORE_PUSH=1
```

Restart Claude Code after editing settings. With the flag set to exactly `1`,
the workflow commits locally, presents every commit the push will publish
(with the new commit's details; a resumed publish-failure preview shows the
fields its report carries), the proposed PR title, and the proposed PR body
verbatim, and asks
`Push this branch and create or update its PR with the title and body shown above? [push / stop]`
— `stop` neither creates nor updates a pull request and pushes nothing new.

Scope: the standard publication path (`/bob:code-review`, including when the
`/bob:work` variants invoke it). `bob-stage-prs` keeps its own built-in
confirmation step. The `*-simple` skill variants do not support the flag and
may ignore it. Run one bob workflow per checkout — the pause holds its state
in `.bob/state` between the two passes.

## Installation

```bash
make install                # Everything (skills + agents + LSP)
make install-skills         # Skills only
make install-agents         # Subagents only
make enable-agent-teams     # Enable /bob:work
make hooks                  # Optional: pre-commit quality checks
```

## Requirements

- Claude Code CLI
- Git

Optional: Go, golangci-lint, gocyclo (for Go-specific features)

---

*Bob - Captain of Your Agents*
