---
name: bob-ship
description: Publish an already verified and reviewed changeset, honor configured push controls, and report CI and PR feedback.
---

# Publish a Reviewed Change

Confirm the requested changeset has current passing verification in `.bob/state/test-results.md` and a clean final review in `.bob/state/review.md`. Compare the reviewed scope and revision with live state; missing, stale, failed, or unresolved results block publication. Preserve unrelated user changes. Do not repair code in this skill: return failures or any hook changes to the calling workflow for verification and review before publishing.

Before committing, pushing, or creating/updating a PR, read and follow [the publication procedure](references/publication-procedure.md). The parent owns publication. Work directly or use `delegate(source: "commit-agent", instructions: "bounded job from the publication procedure", async: true)` and collect it with `load(source: "<task id>")`; workers never delegate further. Respect a caller's single-agent boundary. Read `BOB_CONFIRM_BEFORE_PUSH` from the environment; when exactly `1`, prepare locally, show every unpushed commit and the exact proposed PR title and body, then wait for the specified `push` reply. Revalidate the approved state before publication. Keep the existing publication gate and recovery controls on every push; never bypass a failed gate. Collect and validate `.bob/state/commit.md`. Preserve failed publication artifacts and report what actually reached the remote. A declined publication ends this invocation without retrying.

After successful publication, inspect CI and PR feedback directly or through `delegate(source: "monitor-agent", instructions: "bounded CI and PR check with result path", async: true)` followed by `load(source: "<task id>")`. Repeat checks after a pause while CI is pending; report operational failures rather than assuming readiness. Write `.bob/state/ship-status.md` with branch, commit, PR URL, CI state, and reason: `COMPLETE` only after publication succeeds and monitoring is `READY`; `NEEDS_BRAINSTORM` for failed CI or requested changes; `FAILED` for operational failure or declined publication. Return needed code changes to the caller, keep the review skill's status intact, and report the result. Merge only with the caller's final user approval.
