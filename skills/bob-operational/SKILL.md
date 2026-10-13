---
name: bob-operational
description: Coordinate approved changes across repositories, deploy to the specified environment, and verify the deployed system’s acceptance criteria.
---

# Coordinate Changes and Deployment

Coordinate the requested change across the named repositories and target
environment. Resolve repository paths, deployment method, dependencies,
acceptance criteria, operational authority, and rollback prerequisites from the
request and available evidence; ask only for missing information that blocks the
work. Write .bob/operational/plan.md with ordered per-repository changes,
isolated worktrees where needed, ownership, exact deployment targets,
verification, stop conditions, and recovery steps. Present the concrete plan for
approval unless a matching plan is already approved; obtain an updated decision
for material changes to scope or authority. Preserve unrelated work and inspect
recorded progress before resuming.

Implement directly or delegate bounded assignments according to the actual
dependencies. Parallelize independent work with clear ownership, serialize
overlapping files or shared deployment resources, and start dependent steps only
after their prerequisites are verified. Run repository checks and mandatory code
review before deploying each affected change; for spec-driven modules, audit
applicable documented guarantees and resolve blocking violations or
deployment-critical unverified claims. Operate only within the approved
repositories, environment, and command scope, using the planned publication and
deployment gates. Record each repository’s commit or artifact identity and each
operational step’s actual result in uniquely named artifacts; the parent
consolidates .bob/operational/operate-results.md without overwriting workers’
evidence.

Verify the deployed system against the agreed acceptance criteria and record
evidence in .bob/operational/test-results.md. Distinguish code defects,
deployment failures, and unknown causes: repair the affected code and repeat its
checks and review before redeploying, correct deployment issues within the
approved plan, and stop for a decision when target, authority, or recovery is
unclear. Retry transient failures only when repeating the operation is safe and
bounded; reconcile partial effects before resuming. Use rollback only within the
approved recovery scope. Report the actual deployed versions, PRs or commits,
validation results, partial failures, and remaining actions. Mark completion
only when the requested operations and acceptance checks are satisfied; preserve
worktrees and recovery artifacts still needed.
