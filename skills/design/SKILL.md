---
name: bob:design
description: Create or apply a spec-driven module structure — SPECS.md, NOTES.md, TESTS.md, BENCHMARKS.md, and // NOTE invariant on all .go files
user-invocable: true
category: workflow
---

# New Specs Workflow

You are setting up a **spec-driven module** — a pattern where every package carries living
specification documents alongside its code. Any change to a `.go` file in the module MUST
be reflected in the corresponding `SPECS.md` or `NOTES.md`.

## The Pattern

Each module following this pattern contains:

```
<module>/
  SPECS.md          # Interface contracts, behavior semantics, invariants
  NOTES.md          # Design rationale, decisions, alternatives considered (dated entries)
  TESTS.md          # Test specifications: scenarios, setup, assertions, coverage targets
  BENCHMARKS.md     # Benchmark specs: metric targets, variants, required custom metrics
  README.md         # Usage guide and examples (created if absent)
  *.go              # Every .go file with implementation logic carries the NOTE invariant
```

**The NOTE invariant** on every `.go` file with logic:
```go
// NOTE: Any changes to this file must be reflected in the corresponding specs.md or NOTES.md.
```

This comment goes at the top of the file body (after package declaration, before imports).

---

## Workflow Diagram

```
INIT → GATHER → [ANALYZE] → SCAFFOLD → COMPLETE
                (existing only)
```

---

## Execution Rules

**CRITICAL: All subagents MUST run in background**

- ✅ **ALWAYS use `run_in_background: true`** for ALL Task calls
- ✅ **After spawning agents, STOP** - do not poll or check status
- ✅ **Wait for agent completion notification** - you'll be notified automatically
- ❌ **Never use foreground execution** - it blocks the workflow

---

## Phase 1: INIT

**Goal:** Understand what the user wants to do.

Greet the user and ask which mode they want using `AskUserQuestion`:

```
Which would you like to do?

1. New module — create a spec-driven module from scratch
   (I'll ask what you want to build and scaffold all documentation and stub files)

2. Apply to existing — add the spec-driven structure to an existing implementation
   (Point me at the directory and I'll generate SPECS.md, NOTES.md, TESTS.md,
    BENCHMARKS.md, and add NOTE headers to existing .go files)
```

---

## Phase 2: GATHER

**Goal:** Collect everything needed to scaffold or apply.

### Mode: New Module

Ask the user ONE QUESTION AT A TIME using `AskUserQuestion`:

1. **What are you building?**
   Get a clear description: purpose, what it owns (responsibility boundary), what it
   explicitly does NOT own. Push back on vague answers — "fast" means what? "handles X"
   — at what layer?

2. **What is the module path?**
   Example: `internal/modules/queryplanner`, `pkg/ratelimit`, `cmd/processor`
   Also confirm the Go package name from the last path segment.

After gathering, create `.bob/` and write context:

```
mkdir -p .bob/state
```

Write `.bob/state/design-context.md`:
```
## Mode
new

## Module Path
<path>

## Package Name
<name>

## Description
<description from user>

## Responsibility Boundary
Owns: <what this module owns>
Does NOT own: <what this module explicitly delegates>
```

### Mode: Apply to Existing

Ask the user:

1. **What is the path to the existing implementation?**
   Example: `internal/modules/queryplanner`, `pkg/parser`

Write `.bob/state/design-context.md`:
```
## Mode
apply

## Target Path
<path>
```

---

## Phase 3: ANALYZE (Apply mode only)

**Goal:** Understand the existing implementation deeply enough to write precise specs and notes.

Spawn an Explore agent:

```
Task(subagent_type: "Explore",
     description: "Analyze existing implementation for spec generation",
     run_in_background: true,
     prompt: "Analyze the Go package at the path in .bob/state/design-context.md.

              For each .go file, extract:
              - Package doc comment and stated purpose
              - Every exported type, interface, function, method, and constant
              - Method contracts: inputs, outputs, error conditions, nil-safety
              - Any invariants stated in comments (sorted, non-empty, etc.)
              - Algorithms and data structures used
              - Concurrency: is the type safe for concurrent use?

              Also determine:
              - The module's responsibility boundary (what it owns vs delegates)
              - Key design decisions visible in the code structure
              - Performance characteristics (O(n log n) sorts, allocs in hot paths, etc.)
              - External dependencies (other packages this one imports)
              - Whether benchmarks exist (files ending in _bench_test.go or _test.go with Benchmark*)

              Write findings to .bob/state/design-analysis.md with these sections:
              ## Module Purpose
              ## Responsibility Boundary (owns / does not own)
              ## Exported API (interfaces, types, functions, methods — with full signatures)
              ## Behavioral Invariants
              ## Error Conditions
              ## Design Decisions (visible from code structure)
              ## Performance Notes
              ## External Dependencies
              ## Existing Tests (test function names and what they cover)
              ## Existing Benchmarks (benchmark function names if any)")
```

**Output:** `.bob/state/design-analysis.md`

---

## Phase 4: SCAFFOLD

**Goal:** Create the spec-driven file structure.

### For New Module

Spawn a workflow-implementer agent:

```
Task(subagent_type: "workflow-implementer",
     description: "Scaffold spec-driven module files",
     run_in_background: true,
     prompt: "Create a spec-driven module structure. Context in .bob/state/design-context.md.

              Create all files at the module path. Use the exact templates below.

              --- SPECS.md TEMPLATE ---
              # <PackageName> — Interface and Behaviour Specification

              This document defines the public contracts, input/output semantics, and invariants
              for the `<package>` package. It complements NOTES.md (design rationale) and
              TESTS.md (test plan).

              ---

              ## 1. Responsibility Boundary

              <PackageName> **<one-line statement of what it does>**. It does not <delegation>.

              | Concern | Owner |
              |---------|-------|
              | <concern 1> | **<package>** |
              | <concern 2> | `<other package>` |

              ---

              ## 2. <Primary Interface or Type Name>

              ```go
              type <Interface> interface {
                  <Method>(<params>) <returns>
              }
              ```

              ### 2.1 <MethodName>

              <One sentence stating what it returns or does.>

              **Invariant:** <any guarantee the caller can rely on — sorted output, nil safety, etc.>

              ---

              ## 3. <Next Type or Struct>

              ```go
              type <Struct> struct {
                  <Field> <Type>  // <purpose>
              }
              ```

              ### 3.1 <FieldName>

              <What this field contains and its invariants.>

              ---

              ## 4. <Primary Function or Method — Plan, Execute, Process, etc.>

              ```go
              func (r *<Type>) <Method>(<params>) (<returns>, error)
              ```

              ### 4.1 <Edge Case — No Input, Empty, Nil>

              When <condition>, <what happens exactly>.

              ### 4.2 <Normal Case>

              <Step-by-step description of the algorithm or behavior.>

              ### 4.3 Safety: no false negatives

              <If applicable: what guarantees correctness — e.g., bloom filters cannot false-negative.>

              --- END SPECS.md TEMPLATE ---


              --- NOTES.md TEMPLATE ---
              # <PackageName> — Design Notes

              This document captures the non-obvious design decisions, rationale, and invariants
              for the `<package>` package. These notes complement SPECS.md and are intended to
              prevent re-introducing decisions that were deliberately reversed.

              ---

              ## 1. Why a Separate Package?
              *Added: <today's date>*

              **Decision:** <one-sentence statement of the architectural decision>

              **Rationale:** <why this is the right split — what coupling it avoids, what it enables>

              **Consequence:** <what follows from this decision — what files must/must not import>

              ---

              ## 2. <Primary Interface or Dependency Injection Decision>
              *Added: <today's date>*

              **Decision:** <type> depends on the `<Interface>` interface, not on `*<Concrete>` directly.

              **Rationale:** Go structural typing means `*<Concrete>` satisfies `<Interface>` without
              changes to the concrete package. The interface:
              - Allows the package to be tested with a lightweight stub.
              - Allows alternative backends to be plugged in.
              - Makes the dependency boundary explicit.

              **Consequence:** <file> must not import <package> — it must operate on <Interface> only.

              ---

              ## 3. <Key Algorithmic or Design Choice>
              *Added: <today's date>*

              **Decision:** <one sentence>

              **Rationale:** <why this algorithm/pattern — performance, correctness, simplicity>

              ---

              ## 4. <No Caching / No Mutable State / Stateless Design>
              *Added: <today's date>*

              **Decision:** <Type> holds no mutable state beyond a reference to <Dependency>. It
              performs no caching.

              **Rationale:** Caching belongs at <layer>. Adding a second cache layer would
              complicate invalidation without benefit. Each call is stateless and safe for
              concurrent use if the underlying <Dependency> is.

              --- END NOTES.md TEMPLATE ---


              --- TESTS.md TEMPLATE ---
              # <PackageName> — Test Specifications

              This document defines the required tests for the `<package>` package. Each test is
              described with its scenario, preconditions, steps, and expected outcomes.

              ---

              ## 1. <Primary Function> — Basic Cases

              ### <PKG>-T-01: Test<Function>NoInput

              **Scenario:** <describe the edge case>

              **Setup:**
              - <step 1>
              - <step 2>

              **Assertions:**
              - `<field> == <expected value>`
              - `len(<slice>) == <expected>`

              ---

              ### <PKG>-T-02: Test<Function>Empty

              **Scenario:** <empty/nil input case>

              **Setup:**
              - <step>

              **Assertions:**
              - `<assertion>`

              ---

              ## 2. <Primary Function> — Normal Cases

              ### <PKG>-T-03: Test<Function>NormalCase

              **Scenario:** <describe the primary happy path>

              **Setup:**
              - <step>

              **Assertions:**
              - `<assertion>`

              ---

              ## 3. Coverage Requirements

              - Package statement coverage: ≥ 80%.
              - All public functions must be exercised.
              - Edge cases (empty input, nil input) must be covered.

              --- END TESTS.md TEMPLATE ---


              --- BENCHMARKS.md TEMPLATE ---
              # <PackageName> — Benchmark Specifications

              This document defines the benchmark suite for the `<package>` package. Each benchmark
              is specified as a `testing.B` function with required custom metrics. Benchmarks measure
              cost independently of I/O by pre-building dependencies outside the timed loop.

              ---

              ## Metric Targets (Never Regress Below)

              | Metric | Good | Warning | Critical |
              |--------|------|---------|----------|
              | <Primary operation> latency | < <X> µs | <X>–<Y> µs | > <Y> µs |

              ---

              ## 1. <Primary Operation> Benchmarks

              ### BENCH-<PKG>-01: Benchmark<Function>Baseline

              Measures the baseline cost of <operation>.

              **Setup (outside timed loop):**
              - <setup steps>

              **Variants:**

              | Sub-benchmark | Input Size |
              |---------------|------------|
              | `_small`      | <N>        |
              | `_medium`     | <N>        |
              | `_large`      | <N>        |

              **Required custom metrics:**
              ```go
              b.ReportMetric(float64(b.N)/elapsed.Seconds(), \"ops/sec\")
              b.ReportMetric(float64(inputSize), \"input_size\")
              ```

              --- END BENCHMARKS.md TEMPLATE ---


              Fill in the templates using the description and responsibility boundary from
              .bob/state/design-context.md. Generate realistic, specific content — not
              generic placeholders. The SPECS.md should define the public API the user described
              even if it does not yet exist.

              Also create stub .go files:

              1. <package>.go — package file with responsibility boundary comment:
                 ```go
                 // Package <name> <one-line purpose>.
                 //
                 // Responsibility boundary:
                 //   - <package> owns: <what it owns>
                 //   - <other> owns: <what it delegates>
                 package <name>
                 ```
                 NOTE: The package file gets the responsibility comment, NOT the // NOTE invariant.

              2. <package>_impl.go (or a natural name for the core logic) — with NOTE invariant:
                 ```go
                 package <name>

                 // NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.
                 ```

              3. <package>_test.go — test stub:
                 ```go
                 package <name>_test

                 // NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

                 import (
                     'testing'
                 )

                 // TODO: Implement tests per TESTS.md
                 ```

              4. README.md — always create if it does not already exist:
                 ```markdown
                 # <PackageName>

                 Brief description.

                 ## Responsibility

                 <PackageName> <owns>. It does NOT <delegate>.

                 ## Usage

                 ## Examples
                 ```")
```

### For Apply Mode

Spawn workflow-implementer with the analysis:

```
Task(subagent_type: "workflow-implementer",
     description: "Apply spec-driven structure to existing implementation",
     run_in_background: true,
     prompt: "Apply spec-driven structure to an existing package.
              Context in .bob/state/design-context.md.
              Analysis in .bob/state/design-analysis.md.

              Do NOT modify any existing .go implementation logic.

              STEP 1: Create SPECS.md from the analysis.
              - Section 1: Responsibility Boundary table
              - Section 2+: One section per major exported interface or type
                Each interface: Go code block with full signature, then one subsection per method
                with contract, invariants, nil-safety, and error conditions.
              - Semantic sections: edge cases, normal cases, safety guarantees
              - Use numbered sections (1, 2, 3...) and subsection numbering (2.1, 2.2...)
              - Be precise about invariants: 'sorted ascending', 'never nil', 'idempotent', etc.

              STEP 2: Create NOTES.md from the analysis.
              - Each entry has: ## <N>. <Title>, *Added: <today>*, **Decision:**, **Rationale:**, **Consequence:**
              - Write one entry per significant design decision visible in the code
              - Cover: why this package exists as a separate concern, interface choices,
                algorithmic choices, statefulness decisions, delegation patterns
              - Use past tense for rationale, present tense for consequences

              STEP 3: Create TESTS.md from existing tests.
              - Name each test <PKG>-T-<NN>: <TestFunctionName>
              - For each existing test function, document: Scenario, Setup, Assertions
              - Add a Coverage Requirements section at the end

              STEP 4: Create BENCHMARKS.md.
              - If benchmarks exist (from analysis): document each as BENCH-<PKG>-<NN>
              - If no benchmarks exist: create the template with TODO stubs
              - Always include the Metric Targets table

              STEP 5: Add NOTE invariant to .go files.
              Insert exactly this comment after the package declaration (and after any package
              doc comment), before the first import:
                // NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.
              Add only to files with implementation logic (not to the package-level file that
              has the package doc comment describing responsibility boundaries — that file keeps
              its package comment as the invariant statement).
              Only add if not already present.

              STEP 6: Create README.md if one does not already exist at the target path.")
```

---

## Phase 5: COMPLETE

**Goal:** Show what was created and reinforce the invariant.

After the scaffold agent completes, present a summary:

```
Spec-driven module ready!

Created:
  ✓ <path>/SPECS.md         — Interface contracts and behavioral invariants
  ✓ <path>/NOTES.md         — Design decisions and rationale (add new entries here when you make decisions)
  ✓ <path>/TESTS.md         — Test specifications (update when adding/changing tests)
  ✓ <path>/BENCHMARKS.md    — Benchmark specs and metric targets (Never Regress Below table)
  ✓ <path>/README.md        — Usage guide (created if absent)
  [✓ .go stubs created]     — Scaffolded with NOTE invariant

The invariant:
  Any change to a .go file in this module MUST be reflected in SPECS.md or NOTES.md.
  The NOTE comment in each file is the reminder.

Working with NOTES.md:
  Each new design decision gets a dated entry:
    ## <N>. <Title>
    *Added: YYYY-MM-DD*
    **Decision:** ...
    **Rationale:** ...
    **Consequence:** ...
  Never delete old entries — mark them superseded or add an *Addendum* if a decision reverses.

Next steps:
  - Review and refine SPECS.md — make it the authoritative contract before implementing
  - Use /bob:work-agents to implement against the spec (the workflow will keep docs in sync)
```

---

## Bob's Enforcement

When Bob is working in a directory that contains `SPECS.md`, `NOTES.md`, `TESTS.md`, or
`BENCHMARKS.md`, or when any `.go` file contains the NOTE invariant comment, Bob enforces:

1. **Code changes → update docs**: Any implementation change must be reflected in SPECS.md
   (if it affects the API, contracts, or invariants) or NOTES.md (if it represents a design
   decision or rationale change).

2. **New tests → update TESTS.md**: Any new test function must have a corresponding
   specification entry in TESTS.md with scenario, setup, and assertions.

3. **New benchmarks → update BENCHMARKS.md**: Any new benchmark must be documented with
   setup, variants, and required custom metrics. The Metric Targets table must be updated
   if the benchmark establishes a new performance baseline.

4. **NOTES.md entries are append-only**: Never delete or overwrite design decision entries.
   If a decision is reversed, add an *Addendum* to the original entry explaining the reversal,
   then add a new entry for the new decision.

5. **New .go files**: Must include the NOTE invariant comment (except the package-level file
   with the responsibility boundary comment).
