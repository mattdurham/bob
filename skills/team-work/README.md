# Bob Team Work Workflow

**Concurrent, collaborative development using Claude Code's experimental agent teams.**

## What Is This?

`/bob:team-work` is a development workflow where multiple **teammate agents** work in parallel:
- **2 coder agents** claim tasks and implement features concurrently
- **2 reviewer agents** review completed tasks incrementally as they finish
- **Team lead** (you) coordinates through a shared task list and direct messaging

## Key Benefits

✅ **Faster feedback loops** - Reviews happen as code is written, not in batch at end
✅ **True parallelism** - Multiple coders and reviewers work simultaneously
✅ **Incremental quality** - Issues found early, not after all code complete
✅ **Direct communication** - Teammates message each other and team lead
✅ **Visual display** - Split panes show all teammates working (tmux/iTerm2)

## Prerequisites

### 1. Enable Experimental Feature

```bash
# Quick install (recommended)
make enable-agent-teams

# Or manually add to ~/.claude/settings.json:
{
  "env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  },
  "teammateMode": "auto"
}
```

### 2. Optional: Install tmux for Split Panes

```bash
# macOS
brew install tmux

# Linux
sudo apt-get install tmux
```

Without tmux, teammates run "in-process" (cycle through with Shift+Down).

### 3. Install Bob

```bash
make install
```

## Usage

### Basic Workflow

```bash
claude

# In Claude Code:
/bob:team-work "Add rate limiting to API"
```

The workflow will:
1. **INIT** - Verify experimental flag enabled
2. **WORKTREE** - Create isolated git worktree
3. **BRAINSTORM** - Research patterns and design
4. **PLAN** - Create implementation plan AND task list
5. **SPAWN TEAM** - Create agent team with 2 coders + 2 reviewers
6. **EXECUTE + REVIEW** - Teammates work concurrently
7. **TEST** - Run full test suite
8. **REVIEW** - Final comprehensive review
9. **COMMIT** - Shut down teammates, create PR
10. **MONITOR** - Check CI/CD
11. **COMPLETE** - Clean up team, merge PR

### What You'll See

**In-process mode** (default without tmux):
```
Current agent: team lead
- coder-1 (Shift+Down to view)
- coder-2
- reviewer-1
- reviewer-2

Press Shift+Down to cycle through teammates
Type to message selected teammate
```

**Split pane mode** (with tmux):
```
┌─────────────────┬─────────────────┐
│ Team Lead       │ coder-1         │
│                 │ Implementing... │
├─────────────────┼─────────────────┤
│ coder-2         │ reviewer-1      │
│ Implementing... │ Reviewing...    │
└─────────────────┴─────────────────┘
```

### Example Session

```
You: /bob:team-work "Add user authentication"

[INIT]
✓ Experimental flag verified

[WORKTREE]
✓ Created worktree at ../bob-worktrees/add-user-auth

[BRAINSTORM]
✓ Research complete → .bob/state/brainstorm.md

[PLAN]
✓ Plan complete → .bob/state/plan.md
✓ Created 8 tasks in task list

[SPAWN TEAM]
Team Lead: Creating agent team...
✓ Spawned coder-1
✓ Spawned coder-2
✓ Spawned reviewer-1
✓ Spawned reviewer-2

[EXECUTE + REVIEW - Concurrent]
Team Lead: "Broadcast: Work is starting!"

coder-1: "Claimed task 1: Implement authenticate() function"
coder-2: "Claimed task 2: Add JWT config"

[3 minutes later]
coder-1: "Completed task 1"
reviewer-1: "Reviewing task 1..."

[2 minutes later]
reviewer-1: "Approved task 1 - tests pass, looks good"
coder-1: "Claimed task 3: Implement token validation"

[continues until all tasks complete and approved]

[TEST]
✓ All tests passing

[REVIEW]
✓ Final review clean

[COMMIT]
Shutting down teammates...
✓ PR #142 created

[MONITOR]
✓ CI passing

[COMPLETE]
Ready to merge? [yes/no]
```

## How It Works

### Task List Coordination

```
PLAN creates tasks:
  ├── Task 1: Implement auth function [pending]
  ├── Task 2: Add config [pending]
  └── Task 3: Write tests [pending, blocked by Task 1]

EXECUTE phase:
  coder-1 claims Task 1 → [in_progress]
  coder-2 claims Task 2 → [in_progress]

  coder-1 completes → Task 1 [completed, unreviewed]
  reviewer-1 claims Task 1 for review

  reviewer-1 approves → Task 1 [completed, reviewed, approved]

  Task 3 unblocked (Task 1 done)
  coder-1 claims Task 3 → [in_progress]
```

### Direct Messaging

Teammates communicate through messages:

```
[coder-1 → team lead]: "Completed task 1: Implement authenticate()"
[team lead → coder-1]: "Great work!"

[reviewer-1 → team lead]: "Task 1 needs fixes: missing nil check"
[team lead → coder-1]: "Check fix task 9 for issues in task 1"

[coder-1 → reviewer-1]: "Fixed the nil check, can you re-review?"
```

## Comparison with bob:work

| Aspect | bob:work | bob:team-work |
|--------|----------|---------------|
| **Execution** | Sequential | Concurrent |
| **Agents** | 1 coder, 1 reviewer | 2 coders, 2 reviewers (teammates) |
| **Review timing** | After all code complete | Incremental as code completes |
| **Feedback loops** | Long (batch) | Short (real-time) |
| **Communication** | File-based | Direct messaging + task list |
| **Display** | Hidden in terminal | Split panes (tmux) or cycling |
| **Experimental flag** | Not required | **Required** |

## When to Use

### Use bob:team-work When:

✅ **Complex features** - Multiple files, components, layers
✅ **Large implementations** - 10+ tasks, hours of work
✅ **Quality-critical** - Need thorough incremental review
✅ **Parallel-friendly** - Independent modules or components

### Use bob:work (Sequential) When:

✅ Simple, small tasks (single file, quick fixes)
✅ Exploratory work (requirements unclear)
✅ Learning new codebase
✅ Want simpler workflow

## Troubleshooting

### "Agent teams not available"

**Fix:** Run `make enable-agent-teams` or manually set environment variable:
```bash
export CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1
```

### Teammates not appearing

**Check:**
1. Experimental flag enabled?
2. Team creation successful?
3. Teammates spawned?

**Debug:**
```
"Show me current team members"
```

### Split panes not working

**Install tmux:**
```bash
# macOS
brew install tmux

# Linux
sudo apt-get install tmux
```

**Or use in-process mode:**
```json
{
  "teammateMode": "in-process"
}
```

### Teammates going idle

**If tasks remain but teammates idle:**
```
"Message coder-1: There are still pending tasks, can you claim one?"
```

## Files Created

During the workflow, these files are created in the worktree:

```
.bob/
  state/
    brainstorm.md           # Research and design
    plan.md                 # Implementation plan
    test-results.md         # Test execution results
    review.md               # Final review findings
  planning/                 # Optional (from /bob:project)
    PROJECT.md              # Project context
    REQUIREMENTS.md         # Requirements with REQ-IDs
```

## Advanced Usage

### Custom Team Size

You can adjust the number of teammates:

```
"Create an agent team with 3 coders and 3 reviewers"
```

### Specialized Reviewers

Request specific review focus:

```
"Spawn a security-focused reviewer teammate"
"Spawn a performance-focused reviewer teammate"
```

### Messaging Teammates

**Message specific teammate:**
```
"Message coder-1: Can you prioritize the authentication task?"
```

**Broadcast to all:**
```
"Broadcast to all teammates: We're 50% done, great work!"
```

### Monitoring Progress

**Check task list:**
```
"What's the current task status?"
"Show me the task list"
```

**Check specific teammate:**
```
"What is coder-1 working on?"
"Has reviewer-2 approved any tasks?"
```

## Resources

- **Design doc**: `docs/team-work-design.md`
- **Skill definition**: `skills/team-work/SKILL.md`
- **Agent definitions**:
  - `agents/team-coder/SKILL.md`
  - `agents/team-reviewer/SKILL.md`
- **Claude Code docs**: https://code.claude.com/docs/en/agent-teams

## Support

**Issues or questions?**
- Check the troubleshooting section above
- Read `docs/team-work-design.md` for detailed architecture
- Ask Claude: "How does bob:team-work work?"

---

🏴‍☠️ **Belayin' Pin Bob - Now with Full Agent Teams!**
