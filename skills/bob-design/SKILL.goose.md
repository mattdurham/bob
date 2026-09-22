---
name: bob-design
description: Goose-native spec-driven module design and scaffold workflow
---

# Bob Design for Goose

Create or update a spec-driven module with `SPECS.md`, `NOTES.md`, `TESTS.md`,
`BENCHMARKS.md`, and the required invariant note in every affected Go file.

For an existing module, delegate bounded read-only analysis to `Explore`, then
call `load(source: "<analysis task id>")` and validate its artifact. Resolve
ownership, public contracts, invariants, dependencies, test strategy, and
benchmark obligations before implementation.

Delegate the approved, explicit file set to `workflow-implementer`:

```text
implementation = delegate(source: "workflow-implementer", instructions: "Create the named module files and exact four spec documents; do not delegate", async: true)
load(source: "<implementation task id>")
```

Validate that all promised files exist, every invariant is testable, docs agree
with code, tests cover public contracts, benchmarks describe meaningful
baselines, and affected Go files include the spec-update note. A failed load or
missing file stops the workflow. Do not create extra abstractions or unrelated
features.
