---
name: bob-explore-teams
description: Goose-native team exploration with parallel analysts and adversarial challengers
---

# Bob Explore Teams for Goose

Run a read-only parent-owned pipeline:
`DISCOVER → ANALYZE → CHALLENGE → DOCUMENT → COMPLETE`.

Delegate discovery to `Explore` and collect it with
`load(source: "<discovery task id>")`. Then launch four bounded
`team-analyst` jobs using `delegate(..., async: true)` for structure, flow,
patterns, and dependencies. Each writes its own `.bob/state/analysis-*.md`.
Load every result before starting four bounded `team-challenger` jobs, also
async, each tied to one analysis and one output file. Load all challengers.

For a FAIL, the parent launches only that domain's re-analysis and rechallenge.
Run max 2 challenge loops. Missing loads or artifacts are failures. Synthesize
only after the barrier, cite `file:line`, privilege spec documents, record any
unresolved caveats, and make no code or commit changes.
