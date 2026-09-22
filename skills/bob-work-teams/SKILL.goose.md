---
name: bob-work-teams
description: Goose-native concurrent development workflow with parent-owned coordination
---

# Bob Work Teams for Goose

Use Goose Summon delegates as a bounded team. The parent owns the plan, barriers,
retries, and `.bob/state/team-status.md`; workers never coordinate each other.

```text
coder_a = delegate(source: "team-coder", instructions: "Assignment A, owned paths, worktree, status artifact", async: true)
coder_b = delegate(source: "team-coder", instructions: "Assignment B, disjoint owned paths, worktree, status artifact", async: true)
load(source: "<coder_a task id>")
load(source: "<coder_b task id>")
```

Follow `INIT → WORKTREE → BRAINSTORM → PLAN → EXECUTE → TEST → REVIEW → COMMIT → MONITOR → COMPLETE`.

- Delegate brainstorming to `team-brainstormer` and planning to `team-planner`.
- If spec files exist, delegate a bounded invariant scan to `team-spec-oracle`
  before planning and a final spec check after implementation.
- Partition the plan into disjoint path sets. Start async `team-coder` jobs only
  when their writes cannot overlap. Load every job and validate its status file.
- After all writers load, run bounded `team-reviewer` jobs asynchronously by
  change set. Fixes with overlapping paths are serialized.
- Delegate the full test gate to `tester` and require
  `.bob/state/test-results.md`.
- Delegate consolidation to `review-consolidator` and require
  `.bob/state/review.md`. CRITICAL/HIGH findings route to BRAINSTORM;
  MEDIUM/LOW or test failures route to EXECUTE. REVIEW is mandatory.
- Delegate commit/publish to `commit-agent`, then CI and feedback checks to
  `monitor-agent`. Preserve configured publication approval. CI or PR feedback
  routes to BRAINSTORM.

For every job record assignment, owned files, returned task id, load status,
review status, and retry count in `.bob/state/team-status.md`. A failed load or
missing artifact is a failure, not implicit completion.
