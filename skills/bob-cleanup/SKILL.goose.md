---
name: bob-cleanup
description: Goose-native code cleanup workflow with parallel discovery and bounded fixes
---

# Bob Cleanup for Goose

Never introduce new functionality. Preserve behavior and public contracts.

1. **DISCOVER:** Launch `bug-finder`, `workflow-code-quality`,
   `architecture-introspector`, and `spec-doc-reviewer` with
   `delegate(..., async: true)`. Each receives a bounded scope and unique
   `.bob/state/cleanup-*.md` output. Collect all with
   `load(source: "<task id>")`; retry a missing domain once.
2. **PLAN:** Parent deduplicates evidence into `.bob/state/cleanup-plan.md`. If
   no actionable issue remains, proceed directly to FINAL REVIEW.
3. **CLEANUP LOOP:** Partition non-overlapping paths and delegate bounded
   `team-coder` fixes asynchronously. Load all writers, serialize overlaps, and
   run read-only specialist reviews after the write barrier. Limit fixes to
   simplification, bug repair, dead-code removal, and truthful documentation.
4. **TEST:** Delegate to `tester`, load the result, and require
   `.bob/state/test-results.md`. Failures return to the cleanup loop, max three.
5. **FINAL REVIEW:** Delegate to `review-consolidator`, load it, and require
   `.bob/state/review.md`. REVIEW is mandatory; unresolved issues loop back.
6. **COMMIT:** Delegate to `commit-agent` only after behavior-preserving tests
   and final review pass.

The parent owns task ids, file ownership, barriers, and retry counts.
