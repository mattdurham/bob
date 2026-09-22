---
name: bob-code-review
description: Goose-native code review workflow - REVIEW → FIX → TEST → COMMIT → MONITOR
---

# Bob Code Review for Goose

The parent session owns routing. Delegate bounded jobs with Summon and validate
their artifacts after loading them.

```text
review = delegate(source: "review-consolidator", instructions: "Review the exact worktree and write .bob/state/review.md", async: true)
load(source: "<review task id>")
```

## Workflow

1. **REVIEW:** Run `review-consolidator` and, for Go changes,
   `go-presubmit-reviewer` asynchronously. Load every task and require
   `.bob/state/review.md`. Missing output is `FAILED`.
2. **ROUTE:** CRITICAL/HIGH findings write `NEEDS_BRAINSTORM` to
   `.bob/state/code-review-status.md` and return to the calling workflow.
   MEDIUM/LOW findings continue to FIX. No findings continue to TEST.
3. **FIX:** Delegate one bounded, non-overlapping fix set at a time to
   `workflow-implementer`; load it and rerun REVIEW. Never let concurrent writers
   overlap.
4. **TEST:** Delegate to `tester`, call `load(source: "<tester task id>")`, and
   require `.bob/state/test-results.md`. Failures return to FIX, maximum three
   loops.
5. **COMMIT:** If `.bob/hooks/pre-publish` exists, run it before publication.
   When `BOB_CONFIRM_BEFORE_PUSH=1`, ask once before pushing; declining writes
   `FAILED` and stops. Delegate commit and PR publication to `commit-agent` and
   load the result.
6. **MONITOR:** Delegate CI and PR feedback monitoring to `monitor-agent` and
   load it. CI failure or feedback writes `NEEDS_BRAINSTORM`; operational errors
   write `FAILED`.
7. **COMPLETE:** Write `COMPLETE` to `.bob/state/code-review-status.md` only
   after review, tests, publication, CI, and feedback gates pass.

Delegated workers cannot delegate again. The parent handles every retry and
preserves the existing publication and loop-back rules.
