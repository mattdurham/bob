---
name: bob-specs
description: Create or update a module’s living specifications for spec-driven development, grounded in its contracts, implementation, and verification requirements.
---

# Module Specifications

Create or update the requested module’s living specifications for spec-driven
development. Infer whether the module is new or existing from the request and
repository; ask only for missing purpose, path, responsibility boundaries, or
material design decisions. Read existing code, tests, benchmarks, and repository
guidance before describing guarantees. Distinguish observed behavior from
intended contracts and unresolved proposals; confirm unresolved decisions before
treating them as authoritative. Preserve existing documentation and append
design decisions rather than replacing their history.

Maintain the spec-driven structure: SPECS.md defines ownership, public
contracts, invariants, errors, and edge cases; NOTES.md records dated decisions,
rationale, consequences, and superseded decisions; TESTS.md describes
verification scenarios, setup, and assertions; BENCHMARKS.md records benchmark
scenarios, metrics, and evidenced baselines. Follow the repository’s established
format. If tests or benchmarks are absent, say so and describe the proposed
verification without inventing results, targets, or guarantees. For a new
module, scaffold only the agreed declarations or stubs needed to express its
contract, without implementing the feature. Preserve an existing README or
create a concise usage guide when absent.

Add the reminder `// NOTE: Any changes to this file must be reflected in the
corresponding SPECS.md or NOTES.md.` to applicable Go files with implementation
logic, preserving build tags and existing headers. Check that documents agree
with the agreed contract and existing implementation, and that proposed test and
benchmark scenarios match the guarantees they verify. Report created or updated
files, unresolved decisions, and proposed verification. Document changes to
contracts in SPECS.md, rationale in NOTES.md, test scenarios in TESTS.md, and
performance baselines in BENCHMARKS.md. Work directly or delegate bounded
analysis and scaffolding when useful; give each worker the exact scope and
preserve unrelated changes.
