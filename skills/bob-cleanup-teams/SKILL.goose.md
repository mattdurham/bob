---
name: bob-cleanup-teams
description: Goose-native concurrent cleanup workflow - DISCOVER → PLAN → CLEANUP LOOP → TEST → FINAL REVIEW → COMMIT
---

# Bob Cleanup Teams for Goose

Never introduce new functionality. The parent coordinates bounded delegates;
there is no persistent shared queue.

Launch four discovery specialists (`bug-finder`, `workflow-code-quality`,
`architecture-introspector`, `spec-doc-reviewer`) through
`delegate(..., async: true)` with unique artifacts. Retain every id and call
`load(source: "<task id>")` for each. Consolidate verified findings into
`.bob/state/cleanup-plan.md`.

During CLEANUP LOOP, assign disjoint path sets to async `team-coder` delegates
and serialize overlaps. Load every coder before starting read-only reviews.
Track assignment, paths, task id, load state, review state, and retries in
`.bob/state/team-status.md`. Then run TEST via `tester` and FINAL REVIEW via
`review-consolidator`, loading and validating `.bob/state/test-results.md` and
`.bob/state/review.md`. Failures loop to bounded fixes, maximum three rounds.
When no findings exist, skip writes but never skip FINAL REVIEW. COMMIT through
`commit-agent` only after all gates pass.
