# Claude Configuration for Bob

**Bob** is a workflow orchestration system implemented entirely through **Claude skills** and **subagents** — intelligent workflow coordination through specialized Claude agents.

## What Bob Provides

Bob gives Claude access to:

- **Workflow Skills** - User-invocable workflows (bob:work, bob:work-teams, bob:explore, bob:explore-teams)
- **Subagent Orchestration** - Specialized agents for each workflow phase
- **State Management** - Persistent workflow artifacts in `.bob/` directory
- **Git Worktrees** - Isolated development environments

## Available Workflows

Invoke these workflows with slash commands:

1. **`/bob:work`** - Simple direct workflow — no agents, no ceremony (INIT → WORKTREE → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMMIT → COMPLETE)
2. **`/bob:work-teams`** - Team-based development workflow with concurrent agents (INIT → BRAINSTORM → PLAN → EXECUTE → REVIEW → COMMIT → MONITOR)
4. **`/bob:explore`** - Read-only codebase exploration
5. **`/bob:explore-teams`** - Team-based exploration with adversarial challenge (INIT → DISCOVER → ANALYZE → CHALLENGE → DOCUMENT)
6. **`/bob:design`** - Create or apply spec-driven module structure (SPECS.md, NOTES.md, TESTS.md, BENCHMARKS.md)
7. **`/bob:audit`** - Verify spec invariants and optionally analyze Go structural health (call graphs, complexity, coupling) (read-only)

See individual skill files in `skills/*/SKILL.md` for detailed documentation.

## How Skills Work

**Skills are orchestration layers:**

```
Skill (/bob:work-teams)
  ↓
Creates worktree & .bob directory
  ↓
Spawns specialized subagents:
  - Explore agent (research patterns)
  - workflow-planner agent (create plan)
  - workflow-coder agents (implement)
  - workflow-tester agent (run tests)
  - workflow-reviewer agent (code review)
  ↓
Manages artifacts in .bob/:
  - .bob/state/brainstorm.md
  - .bob/state/plan.md
  - .bob/state/test-results.md
  - .bob/state/review.md
  ↓
Result: Complete, high-quality implementation
```

**Skills don't do the work themselves** - they coordinate specialized Task tool agents.

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

After installation, skills are available via slash commands: `/bob:work`, `/bob:work-teams`, `/bob:explore`, `/bob:explore-teams`, etc.

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

The skill will:
1. Create isolated git worktree
2. Set up `.bob/` directory for artifacts
3. Spawn specialized subagents for each phase
4. Guide you through the complete workflow
5. Enforce quality gates and loop-back rules

## Workflow Artifacts

All workflows store artifacts in `.bob/` directory within the worktree:

```
.bob/
  state/             # Workflow progress (created by /bob:work, /bob:work-teams, etc.)
    brainstorm.md    # Research and approach decisions
    plan.md          # Detailed implementation plan
    test-results.md  # Test execution results
    review.md        # Consolidated code review findings
```

These files persist across Claude sessions and serve as context for subsequent phases.

## Subagent Specialization

Each workflow phase uses specialized agents:

| Phase | Agent Type | Purpose |
|-------|-----------|---------|
| BRAINSTORM | Explore | Research existing patterns |
| BRAINSTORM | brainstorming | Creative ideation |
| PLAN | workflow-planner | Create implementation plan |
| EXECUTE | workflow-coder | Implement code |
| TEST | workflow-tester | Run tests and checks |
| REVIEW | review-consolidator | Multi-domain review: security, bugs, errors, quality, performance, Go idioms, architecture, docs |

## Flow Control Rules

Workflows enforce strict flow control:

**Loop-back paths:**
- **REVIEW → BRAINSTORM**: Critical/high severity issues require re-thinking
- **REVIEW → EXECUTE**: Medium/low severity issues need quick fixes
- **TEST → EXECUTE**: Test failures require code changes
- **MONITOR → BRAINSTORM**: CI failures or PR feedback (always re-brainstorm!)

**Never skip REVIEW** - Quality gate enforced even if tests pass.

## Git Worktrees

⚠️ **CRITICAL**: All workflows create isolated git worktrees BEFORE any file operations.

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

**Creating a spec-driven module:** Use `/bob:design`
- New module: scaffold all docs and stub files
- Existing module: generate docs from analysis, add NOTE headers to .go files

**NOTES.md rules:**
- Each entry: `## N. Title`, `*Added: YYYY-MM-DD*`, **Decision:**, **Rationale:**, **Consequence:**
- Never delete entries — add `*Addendum (date):*` if a decision is reversed
- New decisions go in new numbered sections at the end

## Spec Modes

Bob supports two spec modes, chosen at install time via `make install SPEC=full|simple`.
The installed skills and agents reference **only** the chosen mode — no dual-mode detection.

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

**Creating a simple spec module:** Use `/bob:design`

### Maintaining Spec Mode Variants

Each skill and agent that references spec documentation has two files:
- `SKILL.md` — full spec mode (references SPECS.md, NOTES.md, TESTS.md, BENCHMARKS.md)
- `SKILL.simple.md` — simple spec mode (references CLAUDE.md only)

The Makefile copies the appropriate variant at install time. **When editing a skill or agent,
update both `SKILL.md` and `SKILL.simple.md`** if the change affects non-spec content (workflow
logic, phase structure, tool usage, etc.). Only spec-related sections differ between variants.

Files with simple variants:
- `agents/`: workflow-brainstormer, planner, workflow-implementer, workflow-coder, tester, review-consolidator
- `skills/`: brainstorming, explore, explore-teams, work, work-teams, writing-plans
- `skills/design-simple/` — complete alternative to `skills/design/` for simple mode

## Best Practices

**Orchestration:**
- Let subagents do the work
- Pass context via `.bob/*.md` files
- Clear input/output for each phase
- Chain agents together systematically

**Flow Control:**
- Enforce loop-back rules strictly
- MONITOR → BRAINSTORM (not REVIEW or EXECUTE)
- Never skip REVIEW phase
- Always validate test passage

**Quality:**
- TDD throughout (tests first)
- Comprehensive multi-domain code review (single consolidated reviewer)
- Fix issues properly (re-brainstorm if needed)
- Maintain code quality standards

## Example Session

```
You: /bob:work-teams "Add rate limiting to API"

Claude: I'll orchestrate the work-teams workflow...

[INIT Phase]
Verifying experimental agent teams flag...
Creating .bob directory...
✓ Ready to brainstorm

[BRAINSTORM Phase]
Spawning brainstorming skill...
Creating worktree at ../bob-worktrees/add-rate-limiting...
Spawning Explore agent to research patterns...
✓ Research complete, findings in .bob/state/brainstorm.md

[PLAN Phase]
Spawning workflow-planner agent...
✓ Implementation plan in .bob/state/plan.md
✓ Created 8 tasks in task list

[SPAWN TEAM]
✓ Spawned coder-1, coder-2, reviewer-1, reviewer-2

[EXECUTE + REVIEW - Concurrent]
coder-1: "Claimed task 1: Implement rate limiter"
coder-2: "Claimed task 2: Add config"
reviewer-1: "Reviewing task 1..."
✓ All tasks complete and reviewed

[REVIEW Phase]
Spawning review-consolidator...
  ✓ Security, bugs, errors, quality, performance, Go idioms, architecture, docs
✓ 3 medium issues found in .bob/state/review.md

[Loop to EXECUTE]
Teammates fixing medium issues...
...

[COMMIT Phase]
Shutting down teammates...
✓ PR created: https://github.com/user/repo/pull/123

[MONITOR Phase]
Checking CI status...
✓ All checks passing

[COMPLETE]
Workflow complete!
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
