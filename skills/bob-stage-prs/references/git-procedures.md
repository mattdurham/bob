# Git Procedures for Staging and Advancing a PR Stack

## Persistent State and Approval

Use `.bob/state/stage-prs-plan.md` for the approved stack and
`.bob/state/stage-prs-status.md` for the current result. Keep per-unit execution
and advancement records in `stage-prs-execute-N.md` and `stage-prs-advance-N.md`.
The local-only simple variant keeps its plan in `.bob/state/stage-plan.md`.
Read compatible existing state before creating anything; reconcile older
records from live Git and PR state without assuming missing fields.

Record the repository, remote, source branch and immutable source SHA, target
branch and starting SHA, changes included, dependencies, approval, and permitted
actions. For each unit record its identity, title, branch, base, original parent
cut SHA, current head SHA, selected change range or patches, verification,
publication state, PR URL, and latest observed remote SHA. Keep original cuts
separate from updated bases. Record actual completed actions after each step;
retain partial results and failures. A draft plan alone is not approval.

Present the stack with titles, bases, scope, review size, test/doc coverage,
verification, and merge order. Explain whether approval includes later
advancement and explicit lease-protected updates to the stack's own branches.
Wait for approval before creating commits or branches for a fresh stack, pushing,
or creating PRs. Material regrouping, a changed source snapshot, or expanded
publication actions require updated approval. Never treat elapsed time, old
files, or another workflow's approval as consent. A decline ends this invocation.

## Inspect and Partition

Determine the actual source and target from the request, existing stack, and
repository; do not assume `main`, `master`, or `origin`. Inspect status, history,
merge bases, and changed content with read-only Git commands. Identify dirty
or untracked source work and explain what the committed snapshot excludes;
clarify scope if the request may include it. Never stash, reset, or commit
unrelated changes automatically. Inspect complex histories before selecting a
range; do not suggest rewriting the source branch just to simplify staging.

Group changes into coherent dependency-ordered units, with their corresponding
tests and docs. Roughly 200 implementation lines is a review-size preference,
not a limit or justification for arbitrary splits. A single file can span
multiple units. Preserve necessary foundational changes in each earlier unit.

## Create and Verify Each Unit

Use isolated clean worktrees or an equivalent checkout that preserves the
user's current branch and uncommitted work. Give each branch a unique name;
never replace a pre-existing unrelated branch. Create the first branch from
the approved target and each dependent branch from its immediate predecessor.
Save the parent cut before adding the unit's changes.

Apply the selected commits or patches for this unit. Do not copy final versions
of whole files unless every included change belongs to this unit; that can
silently include later dependencies. Include deletions, renames, generated
updates, tests, and docs belonging to the unit. Stage only explicit files or
hunks. Commit normally, preserving hooks and signatures; never amend unrelated
commits, bypass hooks, or force-push the source or target branch.

Verify each intermediate branch against its own parent, including relevant
builds/tests and documented contracts. Run the repo's applicable verification
commands from its root. A green final stack does not verify its intermediate
units. Review each unit's diff for unintended later changes. Check that the
final stack's aggregate tracked changes match the approved source scope, or
record and obtain approval for deliberate differences. Verification failures
stop publication; revise the partition or implementation within approved scope
and reverify all affected units.

Push approved branches using explicit remote/branch refspecs; do not overwrite
unknown remote work. Create or reuse the PR for the recorded branch and correct
base, and verify its URL, head, and base. Publish PR bodies using structured
arguments or a body file, preserving literal content. Create later PRs as drafts
unless the user requested otherwise. Record each push and PR result separately;
a successful push with a failed PR create is partial publication, not rollback.
Stop on failure and preserve resumable state. Do not skip a failed dependency.

## Resume Without Duplicating Work

Read the saved plan and inspect actual branch tips, remote tips, and every
recorded PR's OPEN/MERGED/CLOSED state, base, and head. Fetch only the refs needed
for these checks. Verify that the repository, source snapshot, branch ownership,
and approved scope still match. Record discrepancies rather than guessing.
Reuse existing branches, commits, and PRs only when their identity and content
match the recorded unit. A closed-unmerged PR is a blocker, not a merged unit.

If PR creation failed after pushing, inspect for a matching existing PR before
retrying creation. If some units are not yet created, resume the first incomplete
approved unit after reconciling its dependencies. If open units need no action,
report their status. If all units merged, report the stack merged. Distinguish
an entirely published stack from an entirely merged stack. Persist the next
resumable action after every invocation.

## Advance After a Merge

Confirm that the predecessor really merged into the intended target. Read the
child's saved old parent cut and save all affected descendant boundary SHAs
before rewriting anything. Use a clean isolated checkout; keep recoverable
local references to the old stack tips. Check that the saved cut is an ancestor
of the recorded child and identify exactly the commits belonging to that child.
Missing or contradictory cuts block automatic advancement; do not infer them
solely from whether the current target is an ancestor of the child.

For the first unmerged child, replay only its own commits onto the verified
new target tip. The bounded operation is:

```text
git rebase --onto <new-parent-tip> <saved-old-parent-cut> <child-branch>
```

Using the saved cut excludes the merged parent's old commits even after a
squash merge. Verify the result against the unit's recorded scope and rerun
applicable verification before publication. Do not blindly accept one side of
a conflict. Resolve only when the intended content is clear; otherwise abort
the in-progress rebase and report the affected files and decision needed.

Then update affected unmerged descendants in dependency order: replay each
unit's own commits onto its immediate predecessor's newly verified tip, using
that descendant's saved old parent cut. A descendant whose predecessor remains
open still bases on that predecessor, not directly on the target branch.
Preserve each unit's intended delta and reverify affected units.

When a rewritten remote branch must be updated, require approval covering that
operation and compare its live remote SHA with the recorded expected SHA. Use
an explicit lease and explicit refspec:

```text
git push --force-with-lease=refs/heads/<branch>:<expected-remote-sha> <remote> <new-head-sha>:refs/heads/<branch>
```

A lease rejection is a blocker: inspect the new remote work and obtain a
revised decision, rather than refreshing the lease and overwriting it. Never
use plain `--force`. Confirm or update the PR's actual base explicitly; do not
assume GitHub retargeted it. Persist new tips and publication results after each
successful unit, and stop after any partial failure.

## Local-Only Simple Mode

Create the approved coherent commit sequence directly without delegation,
preserving unrelated work and verifying every boundary. Report hashes and the
remaining work in its local plan. Remote pushes, PR creation, branch deletion,
and lease-protected rewriting require separate explicit authorization. A local
commit plan is not approval to run the normal publication procedure.
