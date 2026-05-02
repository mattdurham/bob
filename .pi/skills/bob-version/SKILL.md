---
name: bob:version
description: Display Bob version and installation information
user-invocable: true
category: utility
---

# Bob Version Information

**Bob** - Workflow orchestration system implemented through Claude skills and subagents.

## Installation Details

- **Git Commit:** `ba7137c368affaedf96598337399c460effb2841`
- **Commit Date:** 2026-04-28 16:25:38
- **Branch:** main
- **Installed:** 2026-05-02 08:31:56
- **Repository:** git@github.com:mattdurham/bob.git

## Available Workflows

- `/bob:work` - Team-based development workflow with concurrent agents (INIT → WORKTREE → BRAINSTORM → PLAN → EXECUTE → REVIEW → COMMIT → MONITOR)
- `/bob:explore` - Team-based exploration with adversarial challenge (INIT → DISCOVER → ANALYZE → CHALLENGE → DOCUMENT)

## Internal Skills

- `bob:internal:brainstorming` - Interactive ideation (used by work)
- `bob:internal:writing-plans` - Implementation planning (used by work)

## Installed Components

**Skills:**       13 workflow skills installed
- work, explore, bob:version
- Internal: brainstorming, writing-plans

**Subagents:**       23 specialized subagents installed
- workflow-brainstormer, workflow-planner, workflow-coder, workflow-tester
- workflow-implementer, workflow-task-reviewer, workflow-code-quality
- review-consolidator, commit-agent, monitor-agent
- team-coder, team-reviewer, Explore

{{HOOKS_STATUS}}

## Update Bob

To update to the latest version:

```bash
cd /Users/mdurham/source/bob
git pull
make install
```

After updating, run `/bob:version` again to verify the new commit hash.

## Support

- **Repository:** https://github.com/mattdurham/bob
- **Issues:** Report bugs and feature requests on GitHub
- **Documentation:** See CLAUDE.md in the repository

---

*🏴‍☠️ Belayin' Pin Bob - Captain of Your Agents*
