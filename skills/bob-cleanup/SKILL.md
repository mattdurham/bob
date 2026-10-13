---
name: bob-cleanup
description: Simplify code and documentation while preserving behavior and public contracts, then verify and review the changes.
---

# Code Cleanup

Inspect the requested area, its tests, and documented contracts. Identify evidence-backed simplifications: dead code, needless abstraction, confusing names, inaccurate comments, and stale documentation. Record a brief worklist in `.bob/state/cleanup-plan.md` with scope and verification. Preserve user changes and public contracts; report behavior-changing fixes or disagreements between specifications and implementation for a separate decision. Do not rewrite an authoritative contract merely to match the code.

Make the smallest useful behavior-preserving changes. Work directly, or delegate independent tasks when their scope warrants it. Give writers clear, disjoint file ownership, serialize overlapping edits, and collect every result before final verification. Keep the work within the agreed cleanup scope; add no features or speculative abstractions. If an edit cannot be shown to preserve behavior, leave it out and explain the uncertainty.

Run relevant formatting, tests, and static checks, then review the final diff for regressions and scope drift. Fix issues introduced by the cleanup and repeat the affected checks; unresolved failures block completion. Record verification in `.bob/state/test-results.md` and review findings in `.bob/state/review.md`. Report what changed, what was verified, and deferred opportunities. Commit only when requested. If publication is requested, follow `bob-ship` after verification and review; preserve its configured push approval and recovery controls.
