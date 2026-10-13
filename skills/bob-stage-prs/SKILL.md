---
name: bob-stage-prs
description: Split a changeset into an approved stack of coherent pull requests and safely resume or advance it after merges.
---

# Stage Reviewable Changes

Split the requested changeset into coherent, dependency-ordered pull requests,
or resume its recorded stack. Read [the Git
procedures](references/git-procedures.md) before modifying branches, committing,
publishing, or advancing existing PRs. Inspect the source snapshot, target,
working tree, history, and dependencies. Prefer reviewable units around 200
implementation lines, with their tests and docs; let coherence and independently
working intermediate states determine boundaries. Work directly or delegate
bounded tasks when useful, retaining ownership of approval and persistent state.

For a fresh stack, write .bob/state/stage-prs-plan.md with scope, source and
base SHAs, unit branches, dependencies, verification, publication actions, and
merge order. Present the concrete breakdown and wait for approval before
modifying Git state or publishing. For an existing approved stack, reconcile the
recorded state with live branches, remote tips, and PR status before resuming
the next incomplete unit or advancing affected descendants after merges. Obtain
fresh approval for changed scope or actions outside the prior approval. Preserve
the source branch and unrelated work; a missing artifact or stale state is a
blocker, not authorization to recreate the stack.

Create or advance approved units in dependency order using isolated worktrees
and the recorded change boundaries. Verify and review every intermediate unit
before publication, and check the aggregate stack against the approved source
scope. Save branch tips, cuts, verification, push/PR results, and failures after
each step; reuse matching existing PRs rather than duplicating them. Stop on
conflicts, failed verification, remote movement, or declined publication. Report
.bob/state/stage-prs-status.md with actual PR URLs, branch/base state, merge
order, and the next resumable action, distinguishing a published stack from one
that has fully merged.
