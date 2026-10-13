---
name: bob-stage-prs
description: Stage an approved changeset as coherent, verified local commits without delegation or automatic publication.
---

# Stage Reviewable Changes

Partition the requested changes into a coherent dependency-ordered sequence of
local commits. Inspect scope, history, working tree, and existing
.bob/state/stage-plan.md before proposing new work. Read [the Git
procedures](references/git-procedures.md) for isolation, commit boundaries,
verification, and recovery, using its local-only mode. Perform the work yourself
without subagents or teams; roughly 200 implementation lines is a review-size
preference rather than a hard limit.

Write or update the concrete local plan with selected changes, dependencies,
affected files, expected verification, and current progress. Obtain approval for
a new partition before creating commits, and reconcile existing progress against
actual hashes before resuming. Keep tests and docs with their code, preserve
unrelated work, and use isolated checkouts when staging would disturb the
current workspace. Do not copy a file’s final version into an early commit when
that introduces later changes.

Create the approved sequence directly and verify each intermediate boundary and
the aggregate result. Save actual hashes and verification in
.bob/state/stage-plan.md after each unit; stop and report conflicts, failures,
or stale state. Finish with the commit order, completed hashes, and any
remaining work. Remote pushes, PR creation, branch deletion, and rewriting
published branches require separate explicit authorization; this local workflow
does not inherit the normal skill’s publication approval.
