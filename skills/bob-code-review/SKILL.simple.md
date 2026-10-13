---
name: bob:code-review
description: Review changes without modifying them; report findings and draft remediation plans or comments.
user-invocable: true
category: workflow
---

# Review Changes

Review the requested changes and relevant surrounding code against the request,
repository standards, and documented contracts. Inspect the current verification
configuration, including ignored `.bob/config`, for ineffective checks; missing
custom checks alone are not a defect. Use the specified diff, branch,
PR, or revision range; without one, inspect staged, unstaged, and untracked
work. Record the scope and current revision in the report. Check correctness,
edge cases, security, resource ownership, concurrency, tests, and
maintainability; apply Go guidance only to relevant Go changes. Distinguish
demonstrated defects from suggestions and cite exact locations and evidence.

This skill is read-only for the codebase and external systems. Read files and
use inspection commands only. Do not edit source, tests, configuration, or
project documentation; run builds, tests, formatters, or fixes; commit, push,
create PRs; or publish comments. A request to fix findings produces a proposed
remediation plan. Draft comments for the user to review. Perform the review
yourself without subagents or teams.

Write findings to `.bob/state/review.md`, ordered by severity, with proposed
remedies and any limits on the review. When useful or requested, draft comments
in `.bob/state/review-comments.md` or an actionable plan in
`.bob/state/review-plan.md`, including affected files, dependencies, and
proposed verification. These review artifacts are the only permitted file
writes; keep existing implementation plans intact. Write
`.bob/state/code-review-status.md` with `COMPLETE` for no findings,
`NEEDS_BRAINSTORM` for CRITICAL/HIGH findings, `NEEDS_EXECUTE` for MEDIUM/LOW
findings, or `FAILED` for incomplete review. Include the reason and report path.
`COMPLETE` means review complete, with no implementation or publication
performed; the caller owns subsequent actions.
