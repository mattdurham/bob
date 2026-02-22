# Team-Based Workflow Design (Experimental Agent Teams)

## Overview

`bob:team-work` is a concurrent, team-based development workflow using **Claude Code's experimental agent teams feature**. Multiple teammate agents work in parallel, coordinated through a shared task list and direct inter-agent messaging.

This design implements the full [Claude Code agent teams pattern](https://code.claude.com/docs/en/agent-teams), not just tasklist coordination.

## Prerequisites

### Enable Experimental Feature

Agent teams are disabled by default. Enable by adding to `~/.claude/settings.json`:

```json
{
  "env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  },
  "teammateMode": "auto"  // or "in-process" or "tmux"
}
```

Or set environment variable:
```bash
export CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1
```

### Optional: Split Pane Display

For visual teammate display in split panes:
- **tmux**: Install via package manager (`brew install tmux` on macOS)
- **iTerm2**: Install [it2 CLI](https://github.com/mkusaka/it2) and enable Python API

Set `teammateMode: "tmux"` to force split pane mode.

## Key Innovation: Full Agent Teams

### What Makes This Different

This implementation uses Claude Code's **full agent teams feature**, which includes:

1. **Teammate agents**: Separate Claude Code instances (not just subagents)
2. **Direct messaging**: Teammates message each other and the team lead
3. **Split pane display**: Visual representation of each teammate (optional)
4. **Team configuration**: Stored at `~/.claude/teams/<team-name>/`
5. **Mailbox system**: Automatic message delivery between agents
6. **Idle notifications**: Teammates notify team lead when done

### vs Subagents

| Feature | Subagents | Agent Teams |
|---------|-----------|-------------|
| **Context** | Own context, results return to caller | Own context, fully independent |
| **Communication** | Report results back only | Direct messaging between teammates |
| **Coordination** | Caller manages all work | Shared task list + messaging |
| **Display** | Hidden in caller output | Split panes or in-process cycling |
| **Best for** | Focused tasks, results matter | Complex work requiring collaboration |

### vs Tasklist-Only Coordination

The initial prototype used just TaskList APIs. This full implementation adds:

| Feature | Tasklist-Only | Full Agent Teams |
|---------|---------------|------------------|
| **Messaging** | Task metadata only | Direct inter-agent messages |
| **Display** | Terminal only | Split panes (tmux/iTerm2) |
| **Team management** | Manual spawning | Team lead + teammates |
| **Debugging** | Read task updates | Message teammates directly |
| **Collaboration** | Indirect via task list | Direct + task list |

## Architecture

### Team Structure

```
Team Lead (bob:team-work skill)
  ↓
  ├── Teammate: coder-1 (team-coder agent)
  ├── Teammate: coder-2 (team-coder agent)
  ├── Teammate: reviewer-1 (team-reviewer agent)
  └── Teammate: reviewer-2 (team-reviewer agent)

Coordination Mechanisms:
  1. Shared task list (TaskCreate, TaskList, TaskGet, TaskUpdate)
  2. Direct messaging (teammate → teammate, teammate → lead)
  3. Broadcast messages (lead → all teammates)
  4. Team config (~/.claude/teams/<team-name>/config.json)
```

### Communication Patterns

**Task List (Work Queue):**
- Team lead creates tasks
- Coders claim pending tasks
- Coders mark tasks complete
- Reviewers claim completed tasks
- Reviewers approve or create fix tasks

**Direct Messaging:**
- Coder completes task → messages team lead
- Reviewer approves → messages team lead
- Reviewer finds issues → messages team lead + coder
- Team lead provides guidance → messages specific teammate
- Team lead broadcasts → messages all teammates

**Example Message Flow:**
```
[Team Lead]: "Broadcast: Work is starting!"
[Coder-1]: "Claimed task 1: Implement auth"
[Coder-1]: "Completed task 1"
[Reviewer-1]: "Reviewing task 1"
[Reviewer-1]: "Approved task 1 - tests pass, looks good"
[Team Lead]: "Great work coder-1 and reviewer-1, keep going!"
```

## Workflow Phases

```
1. INIT → Verify experimental flag enabled
2. WORKTREE → Create isolated workspace
3. BRAINSTORM → Research and design
4. PLAN → Create plan.md AND task list
5. SPAWN TEAM → Create agent team, spawn 2 coders + 2 reviewers
6. EXECUTE + REVIEW → Teammates work concurrently
7. TEST → Run full test suite
8. REVIEW → Final comprehensive review
9. COMMIT → Shut down teammates, commit, create PR
10. MONITOR → Check CI
11. COMPLETE → Clean up team, merge PR
```

**Key difference:** Phase 5 (SPAWN TEAM) creates the agent team and spawns teammates before execution.

## Phase 5: SPAWN TEAM (New)

This phase is unique to agent teams and critical to the workflow:

### Step 1: Create Agent Team

Team lead tells Claude to create an agent team:
```
"I need to create an agent team for this development task.

Team structure:
- 2 coder teammates (team-coder agents)
- 2 reviewer teammates (team-reviewer agents)

Working directory: [worktree-path]

All teammates should use the Sonnet model.

Please create this team now."
```

### Step 2: Spawn Coder Teammates

Team lead spawns 2 coder teammates with detailed prompts:

```
"Spawn a teammate named 'coder-1' to implement tasks from the shared task list.

Teammate prompt:
'You are coder-1, a team-coder agent.

Your job:
1. Check TaskList for available tasks
2. Claim a task (set status: in_progress, owner: coder-1)
3. Implement using TDD
4. Mark task completed
5. Message team lead when done
6. Repeat until no more tasks

When you complete a task, send a brief message to the team lead.
If you encounter issues, message the team lead for help.

Working directory: [worktree-path]'
"
```

(Similar for coder-2)

### Step 3: Spawn Reviewer Teammates

Team lead spawns 2 reviewer teammates:

```
"Spawn a teammate named 'reviewer-1' to review completed tasks.

Teammate prompt:
'You are reviewer-1, a team-reviewer agent.

Your job:
1. Monitor TaskList for completed, unreviewed tasks
2. Claim a task for review
3. Review code quality, correctness, tests
4. Either APPROVE or CREATE FIX TASKS
5. Message team lead with result
6. Repeat until all completed tasks reviewed

When you approve a task, message the team lead.
If you find critical issues, message the team lead immediately.

Working directory: [worktree-path]'
"
```

(Similar for reviewer-2)

### Step 4: Verify Team Created

Team lead verifies all teammates spawned:
```
"Show me the current team members and their status"
```

Expected output:
- coder-1 (active)
- coder-2 (active)
- reviewer-1 (active)
- reviewer-2 (active)

## Phase 6: EXECUTE + REVIEW (Concurrent)

This is where the magic happens - coders and reviewers work simultaneously.

### Team Lead Responsibilities

1. **Broadcast kickoff:**
   ```
   "Broadcast to all teammates: Work is starting! Coders claim tasks, reviewers review completed tasks."
   ```

2. **Monitor progress:**
   ```
   TaskList()  // Check: pending, in_progress, complete, reviewed, approved
   ```

3. **Handle messages:**
   - Coder completes → acknowledge
   - Reviewer approves → acknowledge
   - Reviewer finds issues → ensure fix task claimed
   - Teammate blocked → provide guidance

4. **Facilitate collaboration:**
   ```
   "Message coder-1: reviewer-1 found issues in task 123, check fix task 456"
   ```

5. **Decide when done:**
   - All tasks complete + approved → TEST phase
   - Unapproved tasks with HIGH/CRITICAL → BRAINSTORM phase
   - Unapproved tasks with MEDIUM/LOW → stay in EXECUTE

### Example Concurrent Execution

```
Time 0:00 - [Team Lead]: "Broadcast: Starting work!"

Time 0:01 - [Coder-1]: "Claimed task 1: Implement rate limiter"
            [Coder-2]: "Claimed task 2: Add config"

Time 0:05 - [Coder-1]: "Completed task 1"
            TaskList: 6 pending, 1 in_progress, 1 complete

Time 0:06 - [Reviewer-1]: "Reviewing task 1"

Time 0:08 - [Coder-2]: "Completed task 2"
            [Coder-1]: "Claimed task 3: Implement storage"
            TaskList: 4 pending, 2 in_progress, 1 reviewing, 1 complete

Time 0:10 - [Reviewer-1]: "Approved task 1 - tests pass"
            [Reviewer-2]: "Reviewing task 2"
            TaskList: 4 pending, 2 in_progress, 1 approved, 1 reviewing

Time 0:12 - [Reviewer-2]: "Task 2 needs fixes - missing validation"
            [Team Lead]: "Message coder-2: Check fix task 9 for task 2 issues"

Time 0:13 - [Coder-2]: "Claimed fix task 9"
            TaskList: 3 pending, 3 in_progress, 1 approved, 1 needs_fix

... continues until all complete and approved ...
```

### Concurrent Benefits

**Compared to sequential:**
- Coder-1 finishes → Reviewer-1 starts immediately (not waiting for all code)
- Coder-2 continues working while Reviewer-1 reviews Coder-1's work
- Issues found early (task 1 reviewed before task 8 implemented)
- Faster feedback loops (minutes vs hours)

## New Agent Types

### team-coder

A teammate agent (not subagent) that:
- Claims tasks from shared task list
- Implements using TDD
- Marks tasks complete
- Messages team lead on completion
- Asks for help when blocked
- Works until no more tasks available

**Key difference from workflow-coder:**
- **Teammate instance**: Separate Claude Code session
- **Direct messaging**: Can message team lead and reviewers
- **Self-directed**: Claims tasks autonomously
- **Visible**: Shows in split pane or teammate list

### team-reviewer

A teammate agent that:
- Monitors completed tasks
- Claims tasks for review
- Reviews incrementally (not batch)
- Approves or creates fix tasks
- Messages team lead with results
- Works until all completed tasks reviewed

**Key difference from review-consolidator:**
- **Teammate instance**: Separate Claude Code session
- **Incremental**: Reviews as tasks complete
- **Actionable**: Creates fix tasks directly
- **Communicative**: Messages coders and team lead
- **Visible**: Shows in split pane or teammate list

## Team Management

### Monitoring Teammates

**In-process mode:**
- Press Shift+Down to cycle through teammates
- Type to message selected teammate
- Press Enter to view teammate's full session

**Split pane mode:**
- Each teammate has own pane
- Click pane to interact
- See all teammates simultaneously

### Messaging Teammates

**Message specific teammate:**
```
"Message coder-1: Great work on task 123, can you also add edge case handling?"
```

**Broadcast to all:**
```
"Broadcast to all teammates: We're 50% done, keep up the great work!"
```

### Handling Issues

**Teammate blocked:**
```
[Coder-1]: "Can't proceed on task 123, unclear what validation level needed"
[Team Lead]: "Message coder-1: Use schema validation with JSON Schema v7, example in validation.go:42"
```

**Teammate idle:**
```
TaskList shows pending tasks
[Team Lead]: "Message coder-2: There are still 3 pending tasks, can you claim task 10?"
```

**Teammates conflicting:**
```
[Coder-1]: "Working on auth.go"
[Coder-2]: "Also working on auth.go"
[Team Lead]: "Message coder-1: Focus on authenticate() function
              Message coder-2: Focus on validate() function"
```

### Shutting Down

Before committing:
```
"Ask coder-1 teammate to shut down"
"Ask coder-2 teammate to shut down"
"Ask reviewer-1 teammate to shut down"
"Ask reviewer-2 teammate to shut down"
```

Wait for confirmations, then proceed to commit.

### Cleaning Up

After workflow complete:
```
"Clean up the agent team"
```

This removes:
- Team config at `~/.claude/teams/<team-name>/`
- Shared resources
- Teammate references

**CRITICAL:** Only team lead should clean up, never teammates.

## Benefits

### 1. True Concurrency

**Sequential (bob:work):**
```
Coder implements all → Reviewer reviews all
        30 minutes           10 minutes
        [========] → [==]
Total: 40 minutes
```

**Concurrent (bob:team-work):**
```
Coder-1: Task 1 → Task 3 → Task 5
         [===] → [===] → [===]

Coder-2: Task 2 → Task 4 → Task 6
         [===] → [===] → [===]

Reviewer-1: → Review 1 → Review 3 → Review 5
               [=] →      [=] →      [=]

Reviewer-2: → Review 2 → Review 4 → Review 6
               [=] →      [=] →      [=]

Total: 15 minutes (overlapped execution)
```

### 2. Incremental Feedback

**Sequential:**
- Find issues after all code complete
- Large batch of fixes needed
- Context-switch cost high

**Concurrent:**
- Find issues as code completes
- Small, focused fixes
- Fresh context for fixes

### 3. Direct Communication

**Tasklist-only:**
- All communication via task metadata
- Indirect and delayed
- Hard to debug issues

**Agent teams:**
- Direct messaging between teammates
- Immediate feedback
- Easy collaboration and debugging

### 4. Visibility

**Sequential:**
- Hidden progress inside agent
- Unclear what's happening
- Can't intervene easily

**Agent teams:**
- See each teammate working (split panes)
- Real-time progress messages
- Message teammates to redirect

### 5. Scalability

**Sequential:**
- Fixed: 1 coder, 1 reviewer
- Can't parallelize more

**Agent teams:**
- Variable: N coders, M reviewers
- Spawn more for larger tasks
- Dynamic based on task list size

## Comparison Table

| Aspect | bob:work | bob:team-work (Agent Teams) |
|--------|----------|------------------------------|
| **Architecture** | Sequential phases | Concurrent teammates |
| **Coordination** | Phase-based | Task list + messaging |
| **Coders** | 1 subagent | 2 teammate agents |
| **Reviewers** | 1 subagent | 2 teammate agents |
| **Communication** | File-based (.bob/state/) | Messaging + task list |
| **Review timing** | After all code | Incremental as code completes |
| **Feedback loops** | Long (batch) | Short (real-time) |
| **Visibility** | Hidden | Split panes or cycling |
| **Display** | Terminal only | tmux/iTerm2 split panes |
| **Team management** | None | Spawn, message, shutdown |
| **Scalability** | Fixed | Variable (add more teammates) |
| **Experimental flag** | Not required | **Required** |

## When to Use

### Use bob:work (Sequential) When:

- Simple, small tasks (single file, quick fixes)
- Exploratory work (requirements unclear)
- Learning new codebase
- No benefit from parallelism
- Want simpler workflow

### Use bob:team-work (Agent Teams) When:

- **Complex features** (multiple files, layers, components)
- **Large implementations** (10+ tasks, hours of work)
- **Quality-critical** (need thorough incremental review)
- **Parallel-friendly** (independent modules)
- **Team efficiency** (maximize throughput)
- **Experimental flag enabled** (prerequisite)

## Troubleshooting

### Experimental Flag Not Enabled

**Error:** "Agent teams not available"

**Fix:**
```bash
# Add to ~/.claude/settings.json
{
  "env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  }
}

# Or set environment variable
export CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1
```

### Teammates Not Appearing

**Check:**
1. Experimental flag enabled?
2. Team creation message sent?
3. Teammate spawn messages sent?
4. Terminal mode correct?

**Debug:**
```
"Show me current team members"
```

### Split Panes Not Working

**Requirements:**
- tmux installed: `which tmux`
- Or iTerm2 with it2 CLI

**Settings:**
```json
{
  "teammateMode": "tmux"  // Force split pane mode
}
```

### Teammates Going Idle

**If tasks remain but teammates idle:**
```
"Message coder-1: There are still pending tasks in the task list, can you claim one?"
```

### Review Bottleneck

**If reviewers can't keep up:**
```
"Spawn another teammate named 'reviewer-3' to help with reviews"
```

### Permission Prompts

**Too many prompts interrupting teammates?**

Pre-approve operations in `~/.claude/settings.json`:
```json
{
  "permissions": {
    "autoApprove": ["read", "write", "bash"]
  }
}
```

## Installation

### 1. Enable Experimental Feature

Add to `~/.claude/settings.json`:
```json
{
  "env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  },
  "teammateMode": "auto"
}
```

### 2. Install Bob Skills and Agents

```bash
cd bob
make install
```

This installs:
- `skills/team-work/` → `~/.claude/skills/bob:team-work`
- `agents/team-coder/` → `~/.claude/agents/team-coder`
- `agents/team-reviewer/` → `~/.claude/agents/team-reviewer`

### 3. Optional: Install tmux

For split pane display:
```bash
# macOS
brew install tmux

# Ubuntu/Debian
sudo apt install tmux

# Or use iTerm2 with it2 CLI
```

### 4. Test It

```bash
claude

# In Claude Code:
/bob:team-work "Add rate limiting to API"
```

Watch the teammates work in split panes!

## Future Enhancements

### 1. Dynamic Team Scaling

Automatically spawn more teammates based on task list size:
```
If pending tasks > 10: spawn coder-3
If completed unreviewed > 10: spawn reviewer-3
```

### 2. Specialized Reviewers

Different reviewer types:
- Security reviewer (security-only issues)
- Performance reviewer (performance-only issues)
- Test reviewer (test coverage focused)

### 3. Inter-Teammate Messaging

Direct coder-to-reviewer messages:
```
[Coder-1] → [Reviewer-1]: "I added edge case handling in commit abc123, can you re-review?"
```

### 4. Progress Dashboard

Real-time visualization in terminal:
```
┌─ Team Progress ─────────────────┐
│ Tasks: 12/20 complete (60%)     │
│ Approved: 8/12 (67%)            │
│                                 │
│ Teammates:                      │
│ ● coder-1    [Task 13] ████     │
│ ● coder-2    [Task 14] ██       │
│ ● reviewer-1 [Task 10] ███      │
│ ● reviewer-2 [Idle]             │
└─────────────────────────────────┘
```

### 5. Team Persistence

Save team state for resumption:
```
/bob:team-work resume <team-id>
```

Restore teammates and task list from previous session.

## Summary

`bob:team-work` implements the **full Claude Code experimental agent teams feature** for concurrent, collaborative development:

**Key Features:**
- ✅ Teammate agents (separate Claude Code instances)
- ✅ Direct inter-agent messaging
- ✅ Split pane display (tmux/iTerm2)
- ✅ Task list coordination
- ✅ Team management (spawn, message, shutdown, cleanup)
- ✅ Concurrent execution (coders + reviewers in parallel)
- ✅ Incremental review (as code completes)
- ✅ Real-time visibility and collaboration

**Requirements:**
- ✅ `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` environment variable
- ✅ Optional: tmux or iTerm2 for split panes
- ✅ Bob skills and agents installed

**Benefits:**
- ⚡ Faster feedback loops (minutes vs hours)
- 🚀 True parallelism (N coders, M reviewers)
- 💬 Direct communication (messaging between teammates)
- 👀 Visibility (split panes show all teammates)
- 📈 Scalability (add more teammates as needed)

---

**Files:**
- `skills/team-work/SKILL.md` - Team lead orchestrator
- `agents/team-coder/SKILL.md` - Coder teammate agent
- `agents/team-reviewer/SKILL.md` - Reviewer teammate agent
- `docs/team-work-design.md` - This design document

🏴‍☠️ **Belayin' Pin Bob - Now with Full Agent Teams!**
