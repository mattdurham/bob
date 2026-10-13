---
name: bob:audit
description: Check documented guarantees against implementation and tests; report violations and unverified claims without modifying the codebase.
---

# Audit Documented Guarantees

Audit the requested module or directory; when no scope is supplied, discover
documented guarantees across the repository. Read applicable SPECS.md,
CLAUDE.md, AGENTS.md, NOTES.md, TESTS.md, and BENCHMARKS.md, and trace each
stated invariant, contract, or binding design decision through the
implementation and relevant tests. Record its source and scope. If documentation
is absent, contradictory, or inaccessible, report that limit rather than
inventing authoritative guarantees.

For every guarantee, mark it satisfied, violated, or unverified. Cite the
implementation and test evidence with exact locations; for violations, describe
the triggering conditions and impact; for unverified claims, explain what
evidence is missing. A passing test or lack of an obvious defect alone does not
establish a guarantee. Perform the audit yourself without subagents or teams.

Keep the codebase and external systems unchanged: use file reads and inspection
commands, without repairs, test generation, builds, test runs, commits, or
publication. Write only audit artifacts under .bob/state/, with the final report
at .bob/state/audit-report.md. Include the scope, reviewed revision, every
guarantee and verdict, coverage limits, and suggested remedies. Keep proposed
invariants clearly labeled as suggestions; adding or changing specifications is
a separate task. Summarize verified violations and unverified guarantees, and
never claim full compliance when coverage is incomplete. Do not run automatic
codebase health scans or compute complexity or coupling scores.
