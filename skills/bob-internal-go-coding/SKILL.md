---
name: bob-internal-go-coding
description: Consult Go-specific guidance for ownership, concurrency, external sizes, and error handling when implementing or reviewing those hazards.
---

# Go-Specific Implementation Guidance

Use this reference when Go changes involve borrowed or pooled data, shared
state, concurrent I/O, external sizes, or error handling. Read the relevant
sections of [Go hazards](references/go-hazards.md) and the repository’s own
contracts and Go-version requirements. Trace ownership and lifetime before
returning pooled objects, coordinate shared state and file replacement, and
bound input-dependent concurrency with cancellation and worker completion
appropriate to the workload.

Validate external sizes, arithmetic, allocation budgets, and slice bounds before
use. Integer-typed make lengths and slice indices are legal; convert to int only
when an API requires it and the conversion is checked. Handle meaningful I/O and
cleanup failures according to the contract. Allocation measurements do not force
garbage collection, and pool presence alone does not justify an arbitrary
allocation threshold. Treat examples as guidance to investigate, not automatic
findings or unconditional style rules.

Verify relevant changed behavior with focused tests, race checks where
appropriate, or other concrete evidence. Distinguish a demonstrated defect from
a hypothetical concern, and explain its trigger, impact, and simplest
appropriate remedy. Follow project verification and test-design conventions; use
bob-matt-tdd when test-first work is requested or already part of the agreed
workflow. This skill supplies Go-specific reasoning without adding a separate
workflow, mandatory pre-commit ceremony, or speculative abstractions.
