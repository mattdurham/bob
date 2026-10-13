---
name: bob:cleanup
description: Simplify code and documentation directly without delegation, preserving behavior and verifying the final diff.
---

# Code Cleanup

Inspect the requested area, its tests, and documented contracts. Identify evidence-backed simplifications: dead code, needless abstraction, confusing names, inaccurate comments, and stale documentation. Record a brief worklist in `.bob/state/cleanup-plan.md` with scope and verification. Preserve user changes and public contracts; report behavior-changing fixes or disagreements between specifications and implementation for a separate decision. Do not rewrite an authoritative contract merely to match the code.

Make the smallest useful behavior-preserving changes yourself. Do not spawn subagents or teammates. Keep the work within the agreed cleanup scope; add no features or speculative abstractions. If an edit cannot be shown to preserve behavior, leave it out and explain the uncertainty.

Run relevant formatting, tests, and static checks, then review the final diff for regressions and scope drift. Fix issues introduced by the cleanup and repeat the affected checks; unresolved failures block completion. Record verification in `.bob/state/test-results.md` and review findings in `.bob/state/review.md`. Report changes, verification, and deferred opportunities. Commit only when requested. For requested publication, apply `bob-ship`'s verification, approval, and recovery controls directly without delegation. Check `BOB_CONFIRM_BEFORE_PUSH` from the environment; if exactly `1`, commit locally, show every commit the push would publish and the exact proposed PR title and body, then ask `Push this branch and create or update its PR with the title and body shown above? [push / stop]`. Require an exact `push` reply before pushing or creating/updating the PR. A declined push ends publication; any hook changes to reviewed code require verification and review again.
