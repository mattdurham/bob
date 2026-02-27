# Claude Configuration for Bob

**Bob** is a workflow orchestration system implemented entirely through **Claude skills** and **subagents**. No MCP servers, no daemons—just intelligent workflow coordination through specialized Claude agents.

## What Bob Provides

Bob gives Claude access to:

- **Workflow Skills** - User-invocable workflows (bob:work, bob:code-review, bob:performance, bob:explore)
- **Subagent Orchestration** - Specialized agents for each workflow phase
- **State Management** - Persistent workflow artifacts in `.bob/` directory
- **Git Worktrees** - Isolated development environments

## Available Workflows

Invoke these workflows with slash commands:

1. **`/bob:project`** - Project initialization (INIT → DISCOVER → QUESTION → RESEARCH → DEFINE → COMPLETE)
2. **`/bob:work`** - Full development workflow (INIT → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMMIT → MONITOR)
3. **`/bob:code-review`** - Code review and fixes (REVIEW → FIX → TEST → loop until clean)
4. **`/bob:performance`** - Performance optimization (BENCHMARK → ANALYZE → OPTIMIZE → VERIFY)
5. **`/bob:explore`** - Read-only codebase exploration
6. **`/bob:new-specs`** - Create or apply spec-driven module structure (SPECS.md, NOTES.md, TESTS.md, BENCHMARKS.md)

See individual skill files in `skills/*/SKILL.md` for detailed documentation.

## How Skills Work

**Skills are orchestration layers:**

```
Skill (/bob:work)
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
  - .bob/planning/PROJECT.md (from /bob:project)
  - .bob/planning/REQUIREMENTS.md (from /bob:project)
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

After installation, skills are available via slash commands: `/bob:work`, `/bob:code-review`, etc.

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

### 2. Required MCP Server (Filesystem Only)

Bob workflows require the filesystem MCP server for file operations.

**Quick install (recommended):**
```bash
# Default directories ($HOME/source and /tmp)
make install-mcp

# Custom directories (comma-delimited)
make install-mcp DIRS="/home/matt/projects,/home/matt/work,/tmp"
```

**Manual install:**
```bash
# Check if already installed
claude mcp list

# Install manually if needed
claude mcp add filesystem -- npx -y @modelcontextprotocol/server-filesystem "$HOME/source" /tmp
```

That's it! No Bob-specific MCP server needed—everything runs through skills and subagents.

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
  planning/          # Project context (created by /bob:project)
    PROJECT.md       # Vision, scope, technical decisions
    REQUIREMENTS.md  # Traceable requirements with REQ-IDs
    CODEBASE.md      # Existing code analysis (brownfield)
    RESEARCH.md      # Technology research (optional)
  state/             # Workflow progress (created by /bob:work, etc.)
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

**Bob's enforcement during `/bob:work`:**
- BRAINSTORM: detect spec-driven modules in scope and note them in the brainstorm
- EXECUTE: workflow-coder updates docs alongside code — no code change without doc update
- REVIEW: review-consolidator checks that spec docs were updated with code changes

**Creating a spec-driven module:** Use `/bob:new-specs`
- New module: scaffold all docs and stub files
- Existing module: generate docs from analysis, add NOTE headers to .go files

**NOTES.md rules:**
- Each entry: `## N. Title`, `*Added: YYYY-MM-DD*`, **Decision:**, **Rationale:**, **Consequence:**
- Never delete entries — add `*Addendum (date):*` if a decision is reversed
- New decisions go in new numbered sections at the end

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
You: /bob:work "Add rate limiting to API"

Claude: I'll orchestrate the work workflow...

[INIT Phase]
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

[EXECUTE Phase]
Spawning workflow-coder agent...
✓ Code implementation complete

[TEST Phase]
Spawning workflow-tester agent...
✓ All tests passing, results in .bob/state/test-results.md

[REVIEW Phase]
Spawning review-consolidator...
  ✓ Security, bugs, errors, quality, performance, Go idioms, architecture, docs
✓ 3 medium issues found in .bob/state/review.md

[Loop to EXECUTE]
Spawning workflow-coder to fix medium issues...
...

[COMMIT Phase]
Creating commit and PR...
✓ PR created: https://github.com/user/repo/pull/123

[MONITOR Phase]
Checking CI status...
✓ All checks passing

[COMPLETE]
🎉 Workflow complete!
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
1. Check filesystem MCP server: `claude mcp list`
2. Verify allowed directories include your repo
3. Check subagent has necessary tools in skill definition

## Migration from MCP-Based Bob

If you previously used Bob as an MCP server:

**Old approach:**
- Bob binary running as MCP server
- `bob.workflow_register()` tool calls
- Manual progress tracking
- Complex state management

**New approach (current):**
- Pure skill-based orchestration
- No MCP server needed
- Automatic subagent coordination
- Simple artifact-based state in `.bob/`

To migrate:
1. Remove Bob MCP server from config: `claude mcp remove bob`
2. Install Bob skills: `make install-skills`
3. Use slash commands: `/bob:work`, `/bob:code-review`, etc.

---

*🏴‍☠️ Belayin' Pin Bob - Captain of Your Agents*
