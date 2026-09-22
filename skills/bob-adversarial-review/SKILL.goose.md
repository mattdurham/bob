---
name: bob-adversarial-review
description: Goose-native adversarial review by eight parallel specialist delegates
---

# Bob Adversarial Review for Goose

This is read-only unless the invocation includes `TEST`. Create
`.bob/state/adversarial-review/` and launch all eight bounded reviews with
`delegate(..., async: true)`. Keep every task id, then call
`load(source: "<task id>")` for all eight. Each job writes a unique domain file.
Retry a failed/missing domain once; never consolidate partial coverage.

### Team Agent 1 — Spec Vigilante
Delegate to `team-analyst`: find spec drift and false invariants.

### Team Agent 2 — Comment Assassin
Delegate to `team-analyst`: find comments that contradict behavior.

### Team Agent 3 — Memory & Panic Hunter
Delegate to `bug-finder`: inspect lifetimes, leaks, panics, and unsafe bounds.

### Team Agent 4 — Concurrency Hawk
Delegate to `go-presubmit-reviewer`: inspect races, deadlocks, pools, and early stop.

### Team Agent 5 — Contract & Test Sheriff
Delegate to `team-analyst`: find API contract and coverage gaps.

### Team Agent 6 — Bug Finder
Delegate to `bug-finder`: hunt concrete correctness failures and edge cases.

### Team Agent 7 — Code Quality Auditor
Delegate to `workflow-code-quality`: inspect idioms, errors, and maintainability.

### Team Agent 8 — Architecture Introspector
Delegate to `architecture-introspector`: challenge coupling and abstractions.

Every instruction names the exact diff/scope, worktree, read-only constraint,
and artifact path. After the load barrier, deduplicate findings by root cause and
write `.bob/state/review.md` with CRITICAL, HIGH, MEDIUM, LOW, and Clean Domains.
If `TEST` was requested, delegate bounded test additions to `workflow-implementer`
only after consolidation, load it, and append generated coverage. Otherwise do
not modify files. Finish with a routing recommendation.
