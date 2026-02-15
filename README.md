# 🏴‍☠️ Belayin' Pin Bob
## Captain of Your Agents

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

**Belayin' Pin Bob** - Your trusty captain for orchestrating AI agent workflows!

> *"A belayin' pin is what keeps the ship's riggin' in order. Bob keeps your agents in line!"*

## What is Bob?

Bob is a workflow orchestration system implemented entirely through **Claude skills** and **specialized subagents**. Just like a ship's captain uses a belayin' pin to secure the ship's lines and rigging, Bob keeps your AI agent workflows organized, coordinated, and running smoothly.

**No MCP servers, no daemons—just intelligent workflow coordination through Claude agents.**

## Features

- 🎯 **Workflow Orchestration** - Multi-step workflows with loop-back rules
- 🤖 **Specialized Subagents** - 12+ domain-expert agents for each workflow phase
- 🌳 **Git Worktrees** - Isolated development environments
- 📁 **Artifact Management** - Persistent state in `.bob/` directory
- 🔄 **Flow Control** - Automatic routing based on severity levels
- 🔍 **9-Agent Parallel Review** - Comprehensive multi-perspective code review
- 📊 **Quality Gates** - TDD, testing, security, performance checks
- 🏴‍☠️ **Captain of Your Agents** - Keep your AI workflows in line!

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/mattdurham/bob.git
cd bob

# Install everything (skills + agents + Go LSP)
make install
```

This installs:
- **Workflow skills** → `~/.claude/skills/` (work, code-review, performance, explore, brainstorming)
- **Specialized subagents** → `~/.claude/agents/` (12+ domain experts)
- **Go LSP plugin** (if available)

After installation, restart Claude to activate all components.

### Using Bob Workflows

Simply invoke workflows with slash commands:

```
/work "Add user authentication feature"
```

Available workflows:
- **`/work`** - Full development workflow (INIT → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMMIT → MONITOR)
- **`/code-review`** - Code review and fixes
- **`/performance`** - Performance optimization
- **`/explore`** - Read-only codebase exploration
- **`/brainstorming`** - Creative ideation

## Workflows

### Work Workflow

Complete feature development from idea to merged PR:

```
INIT → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMMIT → MONITOR → COMPLETE
          ↑                                      ↓               ↓
          └──────────────────────────────────────┴───────────────┘
                        (loop back on issues)
```

**Key phases:**
- **BRAINSTORM**: Research patterns, create worktree, document approach
- **PLAN**: Generate detailed implementation plan
- **EXECUTE**: Implement code following TDD
- **TEST**: Run comprehensive test suite
- **REVIEW**: 9 specialized agents review in parallel
- **COMMIT**: Create commit and PR
- **MONITOR**: Watch CI and handle feedback

**Loop-back rules:**
- **CRITICAL/HIGH issues** → Loop to BRAINSTORM (re-think approach)
- **MEDIUM/LOW issues** → Loop to EXECUTE (quick fixes)
- **CI failures** → Loop to BRAINSTORM (always re-brainstorm!)

### Code Review Workflow

Iterative code review until clean:

```
REVIEW → FIX → TEST → (loop until clean) → COMMIT
```

### Performance Workflow

Optimize code for speed and efficiency:

```
BENCHMARK → ANALYZE → OPTIMIZE → VERIFY → COMMIT
```

### Explore Workflow

Read-only codebase exploration:

```
DISCOVER → ANALYZE → DOCUMENT
```

## Specialized Subagents

Bob coordinates 12+ specialized agents:

| Agent | Purpose |
|-------|---------|
| **workflow-planner** | Creates detailed implementation plans |
| **workflow-coder** | Implements code with TDD approach |
| **workflow-tester** | Runs tests and quality checks |
| **workflow-reviewer** | Multi-pass code review |
| **security-reviewer** | Security vulnerability scanning |
| **performance-analyzer** | Performance bottleneck analysis |
| **docs-reviewer** | Documentation accuracy validation |
| **architect-reviewer** | Architecture and design evaluation |
| **code-reviewer** | Deep code quality review |
| **golang-pro** | Go-specific idiomatic review |
| **debugger** | Bug diagnosis and debugging |
| **error-detective** | Error pattern analysis |

## 9-Agent Parallel Review

The REVIEW phase spawns 9 specialized reviewers in parallel:

1. **Code Quality** - Logic, bugs, best practices
2. **Security** - OWASP Top 10, vulnerabilities
3. **Performance** - Algorithmic complexity, bottlenecks
4. **Documentation** - Accuracy, completeness
5. **Architecture** - Design patterns, scalability
6. **Code Quality Deep** - Comprehensive analysis
7. **Go-Specific** - Idiomatic patterns, concurrency
8. **Debugging** - Potential bugs, race conditions
9. **Error Patterns** - Error handling consistency

Results are consolidated into `.bob/review.md` with severity-based routing:
- **CRITICAL/HIGH** → Loop to BRAINSTORM
- **MEDIUM/LOW** → Loop to EXECUTE
- **No issues** → Continue to COMMIT

## Git Worktrees

⚠️ **All workflows create isolated git worktrees BEFORE any file operations.**

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

**Benefits:**
- Isolate work from main branch
- Safe experimentation
- Easy cleanup if abandoned
- Parallel development possible

## Workflow Artifacts

All workflows store state in `.bob/` directory:

- `.bob/brainstorm.md` - Research and approach
- `.bob/plan.md` - Implementation plan
- `.bob/test-results.md` - Test results
- `.bob/review.md` - Consolidated code review
- `.bob/review-*.md` - Individual agent reviews

These files persist across Claude sessions and serve as context for subsequent phases.

## Installation Options

**Full installation** (recommended):
```bash
make install
```

**Component installation:**
```bash
make install-skills   # Workflow skills only
make install-agents   # Subagents only
make install-lsp      # Go LSP only
```

**Install to another repo:**
```bash
make install-guidance PATH=/path/to/repo
```

This copies `CLAUDE.md` and `AGENTS.md` to configure the repo for Bob workflows.

## Requirements

- **Claude Code CLI** - For running workflows
- **Git** - For worktree management
- **Filesystem MCP server** - For file operations (installed automatically by Claude)

Optional:
- **Go** - For Go-specific features
- **gopls** - For Go LSP integration

## Example Session

```
You: /work "Add rate limiting to API"

Claude: I'll orchestrate the work workflow...

[INIT Phase]
Creating .bob directory...
✓ Ready to brainstorm

[BRAINSTORM Phase]
Spawning brainstorming skill...
Creating worktree at ../bob-worktrees/add-rate-limiting...
Spawning Explore agent to research patterns...
✓ Research complete, findings in .bob/brainstorm.md

[PLAN Phase]
Spawning workflow-planner agent...
✓ Implementation plan in .bob/plan.md

[EXECUTE Phase]
Spawning workflow-coder agent...
✓ Code implementation complete

[TEST Phase]
Spawning workflow-tester agent...
✓ All tests passing, results in .bob/test-results.md

[REVIEW Phase]
Spawning 9 parallel reviewers...
  ✓ Code quality review
  ✓ Security review
  ✓ Performance review
  ✓ Documentation review
  ✓ Architecture review
  ✓ Code quality deep review
  ✓ Go-specific review
  ✓ Debugging review
  ✓ Error pattern review
Consolidating findings...
✓ 3 medium issues found in .bob/review.md

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

Add custom workflows by creating skill files in `skills/`:

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
- Comprehensive multi-agent review (9 specialized reviewers)
- Fix issues properly (re-brainstorm if needed)
- Maintain code quality standards

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

## Contributing

Contributions welcome! Please:
1. Follow existing skill format
2. Add tests for new agents
3. Update documentation
4. Submit PR with clear description

## License

MIT License - see LICENSE file for details

---

*🏴‍☠️ Belayin' Pin Bob - Captain of Your Agents*

**"Keep your agents in line, and your code shipshape!"**
