# Claude Configuration for Bob

**Bob** is a workflow orchestration system implemented entirely through **Claude skills** and **subagents** — intelligent workflow coordination through specialized Claude agents.

## What Bob Provides

Bob gives Claude access to:

- **Workflow Skills** - User-invocable workflows (bob:work, bob:matt-research)
- **Subagent Orchestration** - Specialized agents for each workflow phase
- **State Management** - Persistent workflow artifacts in `.bob/` directory
- **Git Worktrees** - Isolated development environments

## Available Workflows

Invoke these workflows with slash commands:

1. **`/bob:work`** - Development workflow: ask about an existing plan, use it or run `bob:matt-grill-me`, derive a coding worklist, and choose direct execution or independent coders before verification, review, and publication. Startup intake and the confirmed interview precede autonomous execution. Completed assignments may be reviewed concurrently with independent coding; final verification and review follow the write barrier.
2. **`/bob:matt-research`** - Primary-source research and codebase exploration with a grounded Markdown report.
3. **`/bob:audit`** - Check documented guarantees against implementation and tests; report satisfied, violated, and unverified guarantees without changing code or specs

4. **`/bob:specs`** - Create or update living module specifications for spec-driven development.

See individual skill files in `skills/*/SKILL.md` for detailed documentation.

## How Skills Work

Skills combine task instructions with optional delegation:

```
Request → existing plan or confirmed interview → coding worklist
        → direct work or bounded workers → verification → read-only review
        → publication controls → final merge decision
```

**Skills define the workflow contract.** They may perform work directly or
coordinate specialized agents; `bob-work` chooses from its coding worklist.

## Installation

### 1. Install Bob

Bob skills and subagents need to be in `~/.claude/`:

```bash
# From bob repository
make install
```

This installs:
- Workflow skills → `~/.claude/skills/`
- Specialized subagents → `~/.claude/agents/`
- Go LSP plugin (if available)

After installation, skills are available via slash commands: `/bob:work`, `/bob:matt-research`, etc.

**Individual component installation:**
```bash
make install-skills   # Skills only
make install-agents   # Subagents only
make install-lsp      # Go LSP only
```

**Optional components (not installed by default):**
```bash
make hooks            # Pre-commit hooks for quality checks
```

Pre-commit hooks enforce quality gates before commits:
- Run tests (`go test ./...`)
- Check linting (`golangci-lint`)
- Verify formatting (`go fmt`)

Hooks are opt-in to give you control over your workflow. Install them when you want automatic quality enforcement.

## Starting a Workflow

Simply invoke the skill:

```
/bob:work "Add user authentication feature"
```

`bob-work` asks about a plan, uses it or runs the confirmed interview, creates
or reuses a feature worktree, and derives a coding worklist. It chooses direct
execution or independent coders, verifies the integrated changes, and obtains a
read-only final review before publication and the merge decision.

## Workflow Artifacts

All workflows store artifacts in `.bob/` directory within the worktree:

```
.bob/
  state/             # Workflow progress (created by /bob:work, etc.)
    brainstorm.md    # Research and approach decisions
    plan.md          # Detailed implementation plan
    test-results.md  # Test execution results
    review.md        # Consolidated code review findings
```

These files persist across Claude sessions and serve as context for subsequent phases.

## Subagent Specialization

When delegation helps, these specialist roles are available:

| Phase | Agent Type | Purpose |
|-------|-----------|---------|
| BRAINSTORM | Explore | Research existing patterns |
| BRAINSTORM | workflow-brainstormer | Creative ideation |
| PLAN | workflow-planner | Create implementation plan |
| EXECUTE | workflow-coder | Implement code |
| TEST | workflow-tester | Run tests and checks |
| REVIEW | review-consolidator | Multi-domain review: security, bugs, errors, quality, performance, Go idioms, architecture, docs |

## Flow Control Rules

The calling workflow owns repairs and plan decisions. Reconsider the plan when
findings invalidate agreed assumptions; otherwise make bounded corrections.
Test failures and CI or PR feedback return to implementation, followed by
relevant verification and a fresh final review. Incomplete reviews or failed
publication gates block completion.

**Never skip final review**, even when tests pass.

## Conflicting Inputs (coder + planner contract)

Plans are written by agents and can disagree with the code that already exists. Both
sides of that handoff have an explicit contract:

**Planner:** before writing a plan, verify its inputs agree (plan vs. code, decision
record vs. code, decision record vs. decision record). Resolve each conflict in the
plan, or name it as open with the affected symbols and an instruction to the coder.

**Coder:** before editing, check the inputs agree. On a conflict, do not guess and do
not stall — bisect what it affects, implement everything it does not, adopt
already-landed code as the tiebreaker, and record a `DECISION NEEDED` entry (question,
both readings, affected symbols, tiebreaker applied, impact if wrong) in the status
file. A `DECISION NEEDED` entry is a successful outcome, not a failure.

**Orchestrator:** read the coder status files before TEST. Resolve every
`DECISION NEEDED` entry, record the ruling, and re-run the affected work. Routing
straight to TEST hides whether the tiebreaker was right, because TEST cannot tell you.

`make check-conflict-guidance` enforces that every coder, planner, and workflow prompt
carries this guidance. It runs as part of `make ci`.

## Git Worktrees

`bob-work` creates or reuses an isolated feature worktree before implementation.
`bob-work-simple` operates directly in the current workspace. Read-only skills
inspect their requested scope without creating a development worktree.

**Why worktrees?**
- Isolate work from main branch
- Safe experimentation
- Easy cleanup if abandoned
- Parallel development possible

**Structure:**
```
~/source/bob/                    # Main repo
~/source/bob-worktrees/
  ├── add-auth/                  # Feature 1 worktree
  │   ├── .bob/                  # Workflow artifacts
  │   └── ...                    # Feature code
  └── fix-parser/                # Feature 2 worktree
      ├── .bob/
      └── ...
```

## Spec-Driven Module Pattern

Some packages follow a **spec-driven** pattern where living specification documents accompany
the code. Bob detects and enforces this pattern automatically.

**Detection:** A module is spec-driven if its directory contains any of:
- `SPECS.md` — interface contracts, behavioral invariants, edge cases
- `NOTES.md` — design decisions with dated entries (append-only)
- `TESTS.md` — test specifications: scenario, setup, assertions
- `BENCHMARKS.md` — benchmark specs with Metric Targets table
- `.go` files with the NOTE invariant: `// NOTE: Any changes to this file must be reflected in the corresponding specs.md or NOTES.md.`

**The invariant:** Any change to a `.go` file in a spec-driven module MUST be reflected in
`SPECS.md` (API/contract changes) or `NOTES.md` (design decisions).

**Bob's enforcement during workflows:**
- BRAINSTORM: detect spec-driven modules in scope, read their invariants, and constrain approach selection
- EXECUTE: workflow-coder passes actual invariants from SPECS.md to implementer as hard constraints
- REVIEW: review-consolidator verifies code satisfies stated invariants in SPECS.md (primary) and checks that spec docs were updated (secondary)

**Creating a spec-driven module:**
- New module: create SPECS.md, NOTES.md, TESTS.md, BENCHMARKS.md, and stub files manually
- Existing module: generate docs from analysis, add NOTE headers to .go files

**NOTES.md rules:**
- Each entry: `## N. Title`, `*Added: YYYY-MM-DD*`, **Decision:**, **Rationale:**, **Consequence:**
- Never delete entries — add `*Addendum (date):*` if a decision is reversed
- New decisions go in new numbered sections at the end

## Spec Modes

Bob supports two spec modes, chosen at install time via `make install SPEC=full|simple`.
Full installation includes normal skills and available single-agent variants;
simple installation selects the single-agent variants. The shared `bob-ship`
controls and Matt imports install in both modes. Honor the project’s actual
documented contracts; existing specialized agents also retain mode-specific
specification guidance.

### Full Spec Mode (default: `make install` or `make install SPEC=full`)

Each module carries 4 living specification documents: `SPECS.md`, `NOTES.md`, `TESTS.md`,
`BENCHMARKS.md`, plus a `// NOTE` invariant on `.go` files. See "Spec-Driven Module Pattern"
above for details.

### Simple Spec Mode (`make install SPEC=simple`)

Each module carries a single `CLAUDE.md` file containing numbered invariants.

**CLAUDE.md rules:**
- Keep them tidy
- They contain only numbered invariants, axioms, assumptions, and non-obvious constraints
- Never add anything trivial, ephemeral, or obviously derivable from reading the code
- NEVER include copies of the code itself
- No NOTE invariant on `.go` files — Claude Code natively loads CLAUDE.md

**Example CLAUDE.md:**
```markdown
# ratelimit — Invariants

1. The TokenBucket interface is the sole entry point; all callers must go through it.
2. Refill is atomic — concurrent Acquire calls never see a partially refilled bucket.
3. The package never persists state — persistence is the caller's responsibility.
4. Thread-safe only if the underlying Store is thread-safe.
```

**Enforcement during workflows:**
- BRAINSTORM: detect CLAUDE.md modules, read their invariants, and constrain approach selection
- EXECUTE: pass actual invariants from CLAUDE.md to implementer as hard constraints
- REVIEW: verify code satisfies stated invariants in CLAUDE.md (primary) and check that docs were updated (secondary)

**Creating a simple spec module:** Create a CLAUDE.md with numbered invariants manually.

### Maintaining Spec Mode Variants

Where variants exist, `SKILL.md` provides the normal workflow and
`SKILL.simple.md` provides its direct, single-agent counterpart. Update both
when changing their shared behavior; preserve their execution boundaries and
project-specific specification guidance.

## Best Practices

- Derive assignments from the worklist and delegate only when useful.
- Give workers clear scope, file ownership, and unique result paths.
- Collect writers and early reviewers before integrated verification.
- Preserve contracts, user changes, and recorded decisions.
- Require current verification and a clean final review before publication.
- Return CI and PR feedback to the caller; honor push and merge approval.

## Example Session

```
You: /bob:work "Add rate limiting to API"
Bob: Do you have an existing implementation plan? [yes/no]
You: no

Bob follows bob:matt-grill-me, asking one question at a time.
You confirm the requirements and decisions.

Bob creates the coding worklist in an isolated worktree.
Independent assignments may use concurrent coders and early read-only reviews.
All workers finish before verification and final integrated review.

After passing checks and a clean review, bob:ship prepares the PR and
honors configured push approval. Bob reports CI and asks for the merge decision.
```

## Customization

Add custom workflows by creating new skill files in `skills/`:

```
skills/
  my-workflow/
    SKILL.md    # Workflow definition with frontmatter
```

Frontmatter format:
```yaml
---
name: my-workflow
description: Brief description
user-invocable: true
category: workflow
---
```

## Troubleshooting

**Skills not appearing:**
1. Check skills installed: `ls ~/.claude/skills/`
2. Verify frontmatter has required `name` field
3. Restart Claude Code

**Worktree creation fails:**
1. Check you're in a git repository: `git status`
2. Verify git is configured: `git config user.name`
3. Check disk space: `df -h`

**Subagents failing:**
1. Check subagent has necessary tools in skill definition
2. Check disk space: `df -h`

---

*🏴‍☠️ Belayin' Pin Bob - Captain of Your Agents*
