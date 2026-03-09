---
name: bob:go-analyze
description: Go codebase analysis through AST, call graphs, complexity scoring, and module coupling — INIT → DISCOVER → ANALYZE → SCORE → REPORT → COMPLETE
user-invocable: true
category: workflow
---

# Go Codebase Analysis Workflow

You orchestrate a **read-only structural analysis** of a Go codebase. You examine it through the lens of an interconnected call graph, scoring functions by complexity and coupling, then produce SPECS-aware recommendations.

## What This Does

- Builds a **function call graph** from a starting point (or whole repo)
- Scores every function using **cyclomatic complexity** (gocyclo), **cognitive complexity** (gocognit), **call depth**, **allocations** (escape analysis), and **fan-in**
- Measures **module coupling**: afferent/efferent coupling, instability, interface vs concrete-type boundaries
- Detects **spec-driven modules** and cross-references findings against stated invariants
- Produces a ranked list of hot spots and actionable, spec-aware recommendations

## Workflow Diagram

```
INIT → DISCOVER → ANALYZE → SCORE → REPORT → COMPLETE
```

**Read-only:** No code changes, no commits. Output is `.bob/state/go-analysis-report.md`.

---

## Execution Rules

**CRITICAL: All subagents MUST run in background**

- ALWAYS use `run_in_background: true` for ALL Task calls
- After spawning agents, STOP — do not poll or check status
- Wait for agent completion notification
- Never use foreground execution

---

## Phase 1: INIT

**Goal:** Understand analysis scope.

**Actions:**

1. Ask the user what to analyze using `AskUserQuestion`:

```
What would you like to analyze?

1. Whole repository — analyze every package
2. Specific package — provide an import path or directory (e.g. ./internal/ratelimit)
3. Starting function — build call graph from a single entry point (e.g. main, ServeHTTP)

Also: do you have gocyclo, gocognit, and/or staticcheck installed?
(The analysis degrades gracefully without them — go vet and go build -gcflags='-m' always work)
```

2. Record:
   - `SCOPE`: `repo` | `package:<path>` | `function:<pkg>.<name>`
   - `HAS_GOCYCLO`: yes/no
   - `HAS_GOCOGNIT`: yes/no
   - `HAS_STATICCHECK`: yes/no (optional, used for additional signal)
   - `HAS_CALLGRAPH`: check if `golang.org/x/tools/cmd/callgraph` is available

3. Create state directory:
```bash
mkdir -p .bob/state
```

---

## Phase 2: DISCOVER

**Goal:** Inventory all packages, detect spec-driven modules, and collect raw tool output.

Spawn a single Explore agent:

```
Task(subagent_type: "Explore",
     description: "Inventory Go packages and detect spec-driven modules",
     run_in_background: true,
     prompt: "You are inventorying a Go codebase for structural analysis.

              SCOPE: [from INIT]

              Step 1 — Package inventory
              Run: go list -json ./...
              (or scoped package if not whole repo)

              For each package record:
              - ImportPath
              - Dir
              - GoFiles count
              - Imports (direct dependencies)
              - TestImports

              Step 2 — Detect spec-driven modules
              For every directory encountered, check for:
              - SPECS.md
              - NOTES.md
              - TESTS.md
              - BENCHMARKS.md
              - .go files containing: // NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.

              For each spec-driven module found, read SPECS.md and extract:
              - All stated invariants (numbered list)
              - Complexity or coupling constraints mentioned
              - Interface contracts

              Step 3 — Run complexity tools (use whatever is available)

              If gocyclo available:
                Run: gocyclo -over 1 [scope]
                Capture all output (complexity score, function, file:line)

              If gocognit available:
                Run: gocognit -over 1 [scope]
                Capture all output

              Always run:
                go vet [scope]
                Capture all output (warnings, errors)

              Always run (escape/allocation analysis):
                go build -gcflags='-m=1' [scope] 2>&1
                Capture lines containing 'escapes to heap', 'does not escape', 'inlining call'

              If callgraph tool available:
                Run: callgraph -algo=cha [scope] 2>&1 | head -2000
                (CHA is fast; RTA is more precise but slow)
                Capture edge list: caller -> callee

              Step 4 — Import coupling matrix
              From go list -json output, build a dependency matrix:
              For each package P:
                Efferent coupling (Ce): count of packages P imports (excluding stdlib)
                Afferent coupling (Ca): count of packages that import P (excluding stdlib)
                Instability: Ce / (Ca + Ce)  [0=stable, 1=unstable]

              Step 5 — Interface boundary detection
              Search for interface definitions:
                grep -rn 'type .* interface' [scope dirs]
              For each interface, note:
                - Which package defines it
                - Which packages implement it (by searching for method signatures)
                - Which packages consume it (accept/return the interface type)

              Write ALL raw findings to .bob/state/go-discovery.md:

              # Go Analysis Discovery

              ## Packages
              | Package | Dir | Files | Ce | Ca | Instability |
              |---------|-----|-------|----|----|-------------|
              | ...     | ... | ...   | .. | .. | ...         |

              ## Spec-Driven Modules
              For each: path, spec files present, invariants extracted

              ## Raw Complexity Output
              ### gocyclo
              [raw output or 'not available']

              ### gocognit
              [raw output or 'not available']

              ### go vet
              [raw output]

              ### Escape Analysis (go build -gcflags='-m=1')
              [lines with 'escapes to heap', grouped by package]

              ### Call Graph Edges
              [callgraph output or 'not available']

              ## Interface Boundaries
              | Interface | Defined In | Implementors | Consumers |
              |-----------|-----------|--------------|-----------|")
```

**Output:** `.bob/state/go-discovery.md`

---

## Phase 3: ANALYZE

**Goal:** Build the call graph as a node graph, compute per-function metrics, measure coupling depth.

Spawn an Explore agent:

```
Task(subagent_type: "Explore",
     description: "Analyze call graph, function metrics, and module coupling",
     run_in_background: true,
     prompt: "Read .bob/state/go-discovery.md.

              You are building a structural model of the Go codebase.

              === CALL GRAPH ANALYSIS ===

              From the call graph edges (or by reading the source AST via go/ast if callgraph was unavailable):

              For each function node, compute:
              1. CALL DEPTH — longest path from this function to a leaf (no outgoing calls)
                 - Leaf nodes: depth 0
                 - Direct callers of leaves: depth 1
                 - etc.
              2. FAN-IN — number of distinct callers
              3. FAN-OUT — number of distinct callees
              4. IS_RECURSIVE — does any call path lead back to itself

              If callgraph output was not available, reconstruct a partial graph by:
              - Reading each .go file
              - Parsing function declarations and their call expressions
              - Building edges: (file:func) -> (callee name)
              This gives a local call graph (cross-package calls may be approximate).

              === PER-FUNCTION METRIC TABLE ===

              For every function in scope, assemble all signals:
              - Cyclomatic complexity (from gocyclo output, or 'N/A')
              - Cognitive complexity (from gocognit output, or 'N/A')
              - Call depth (computed above)
              - Fan-in / Fan-out
              - Heap allocations (count of 'escapes to heap' lines attributed to this function)
              - Is recursive

              === MODULE COUPLING ANALYSIS ===

              From the package coupling matrix:

              1. Identify HIGHLY COUPLED packages (Ce + Ca > threshold, suggest >10)
              2. Identify UNSTABLE packages (instability > 0.7) that are depended upon by stable packages
                 (violation of Stable Dependencies Principle)
              3. Identify packages with NO interface boundary to callers (all coupling via concrete types)
              4. Identify packages that import each other (import cycles — go will reject these, but note near-cycles)
              5. Detect GOD PACKAGES: high fan-in AND high fan-out AND many files

              === SPEC INVARIANT CROSS-REFERENCE ===

              For each spec-driven module:
              - Check if any of its invariants mention complexity bounds (e.g. 'functions must remain simple')
              - Check if coupling invariants exist (e.g. 'this package must not import X')
              - Flag any function in the module that appears in the high-complexity list

              Write findings to .bob/state/go-analysis.md:

              # Go Structural Analysis

              ## Function Metrics Table
              | Package | Function | File:Line | Cyclo | Cognit | CallDepth | FanIn | FanOut | HeapAllocs | Recursive |
              |---------|----------|-----------|-------|--------|-----------|-------|--------|------------|-----------|
              [one row per function, sorted by package]

              ## Call Graph Hot Paths
              List the top 10 longest call chains (entry point -> ... -> leaf), showing depth and packages crossed.

              ## Module Coupling Matrix
              | Package | Ce | Ca | Instability | Interface Boundary | Risk |
              |---------|----|----|-------------|-------------------|------|
              [sorted by instability descending]

              ## Stable Dependencies Violations
              List any case where an unstable package is imported by a stable one.

              ## God Package Candidates
              Packages with unusually high coupling in both directions.

              ## Spec Module Cross-Reference
              For each spec-driven module:
                - Which functions appear as complexity hot spots?
                - Do stated invariants mention coupling or complexity?
                - Are interface boundaries respected?")
```

**Input:** `.bob/state/go-discovery.md`
**Output:** `.bob/state/go-analysis.md`

---

## Phase 4: SCORE

**Goal:** Compute a composite health score per function and per module, rank hot spots.

Spawn an Explore agent:

```
Task(subagent_type: "Explore",
     description: "Score functions and modules, rank hot spots",
     run_in_background: true,
     prompt: "Read .bob/state/go-analysis.md.

              === FUNCTION SCORING ===

              For each function, compute a COMPLEXITY SCORE (0-100, higher = more concerning):

              Score = min(100, (
                cyclo_points +
                cognit_points +
                depth_points +
                alloc_points +
                fanin_points
              ))

              Where:
                cyclo_points  = min(40, cyclo * 2)        -- cyclomatic complexity, cap 40
                cognit_points = min(30, cognit * 1.5)     -- cognitive complexity, cap 30
                depth_points  = min(15, call_depth * 3)   -- call depth, cap 15
                alloc_points  = min(10, heap_allocs * 2)  -- escaping allocations, cap 10
                fanin_points  = min(5, fan_in)            -- high fan-in = blast radius, cap 5

              If a metric is N/A, assign 0 for that component and note the gap.

              RISK TIERS:
                CRITICAL  (80-100): Refactor immediately
                HIGH      (60-79):  Should be broken up in next sprint
                MEDIUM    (40-59):  Worth watching, add complexity tests
                LOW       (20-39):  Acceptable, monitor over time
                CLEAN     (0-19):   No action needed

              === MODULE SCORING ===

              For each package, compute a COUPLING SCORE (0-100):

              coupling_score = min(100, (
                instability * 30 +               -- 0-30 points
                interface_penalty +              -- +20 if NO interface boundary to callers
                god_package_penalty +            -- +20 if god package candidate
                spec_violation_penalty           -- +10 if spec-driven and has CRITICAL functions
              ))

              MODULE COUPLING RISK:
                FRAGILE   (70-100): High risk — changes ripple widely
                COUPLED   (40-69):  Needs interface boundaries or decomposition
                MODERATE  (20-39):  Some coupling, manageable
                CLEAN     (0-19):   Well-bounded, low coupling

              === HOT SPOT RANKING ===

              Produce two ranked lists:

              1. TOP 20 FUNCTION HOT SPOTS (by function score, descending)
              2. TOP 10 MODULE HOT SPOTS (by coupling score, descending)

              For each hot spot note:
              - Score and tier
              - The dominant contributing metric (what drives the score)
              - Whether it is in a spec-driven module (and which invariants apply)

              Write to .bob/state/go-scores.md:

              # Go Health Scores

              ## Function Hot Spots (Top 20)
              | Rank | Package | Function | File:Line | Score | Tier | Dominant Factor | Spec Module? |
              |------|---------|----------|-----------|-------|------|----------------|--------------|

              ## Module Hot Spots (Top 10)
              | Rank | Package | Score | Tier | Ce | Ca | Interface Boundary | Spec Module? |
              |------|---------|-------|------|----|----|-------------------|--------------|

              ## Score Distribution
              | Tier | Function Count | Module Count |
              |------|---------------|--------------|
              | CRITICAL  | N | N |
              | HIGH      | N | N |
              | MEDIUM    | N | N |
              | LOW       | N | N |
              | CLEAN     | N | N |

              ## Coverage Gaps
              Note any functions/packages where metrics were N/A due to missing tools,
              and what installing those tools would reveal.")
```

**Input:** `.bob/state/go-analysis.md`
**Output:** `.bob/state/go-scores.md`

---

## Phase 5: REPORT

**Goal:** Produce a consolidated, SPECS-aware report with actionable recommendations.

Spawn an Explore agent:

```
Task(subagent_type: "Explore",
     description: "Consolidate findings and produce SPECS-aware recommendations",
     run_in_background: true,
     prompt: "Read .bob/state/go-discovery.md, .bob/state/go-analysis.md, and .bob/state/go-scores.md.

              Write a final report to .bob/state/go-analysis-report.md:

              # Go Codebase Analysis Report

              Generated: [ISO timestamp]
              Scope: [what was analyzed]

              ---

              ## Executive Summary

              - Packages analyzed: N
              - Functions analyzed: N
              - Spec-driven modules found: N
              - CRITICAL functions: N
              - CRITICAL modules: N
              - Stable Dependencies Principle violations: N
              - God package candidates: N
              - Tools available: [gocyclo yes/no, gocognit yes/no, callgraph yes/no]

              ### Overall Codebase Health: [HEALTHY / NEEDS ATTENTION / CRITICAL]
              HEALTHY:         No CRITICAL or HIGH functions; no FRAGILE modules
              NEEDS ATTENTION: Some HIGH functions or COUPLED modules
              CRITICAL:        Any CRITICAL functions or FRAGILE modules, or SDP violations

              ---

              ## Critical Hot Spots

              ### CRITICAL Functions
              For each (score 80-100):

              **`package.FunctionName`** — Score: N/100 (CRITICAL)
              File: `path/file.go:line`
              Dominant factor: [cyclomatic | cognitive | call depth | allocations | blast radius]
              Metrics: cyclo=N, cognit=N, depth=N, heap_allocs=N, fan_in=N

              If in a spec-driven module:
              > **Spec context:** Module `path/` has SPECS.md.
              > Relevant invariants: [list the invariants that this function is supposed to satisfy]
              > Risk: complexity in this function may make these invariants difficult to verify.

              Recommendation:
              - [Specific, actionable refactoring suggestion — e.g. extract strategy, split by concern, add interface]
              - [If spec-driven: suggest adding a complexity invariant to SPECS.md]

              ### FRAGILE Modules
              For each (coupling score 70-100):

              **`package/path`** — Coupling Score: N/100 (FRAGILE)
              Ce=N (depends on N packages), Ca=N (depended on by N packages), Instability=0.N
              Interface boundary: [NONE | PARTIAL | FULL]

              If in a spec-driven module:
              > **Spec context:** Relevant invariants from SPECS.md: [list]

              Recommendation:
              - [Define interface at boundary X]
              - [Extract sub-package Y to reduce Ce]
              - [If spec-driven: add coupling constraints to SPECS.md]

              ---

              ## Stable Dependencies Principle Violations

              The Stable Dependencies Principle (SDP) states: packages should depend in the
              direction of stability (depend on MORE stable packages, not less stable ones).

              For each violation:
              **`stable-package` imports `unstable-package`**
              stable-package instability: 0.N (stable)
              unstable-package instability: 0.N (unstable)
              Risk: changes to `unstable-package` force changes in `stable-package`
              Fix: introduce an interface or invert the dependency

              ---

              ## Module Coupling Overview

              [Include the full Module Coupling Matrix table from go-analysis.md]

              **Interface Boundary Assessment:**
              List packages that have NO interface boundary to their callers, and which packages
              are calling them directly via concrete types. These are tight couplings that resist change.

              ---

              ## Call Graph Structure

              **Deepest Call Chains:**
              [Top 5 longest chains, showing path and depth]

              **High Fan-In Functions (Blast Radius):**
              Functions called from many places — changes here ripple widely.
              | Function | Fan-In | Packages Affected |
              |----------|--------|------------------|

              **Recursive Functions:**
              [List any recursive functions — flag if unbounded recursion possible]

              ---

              ## Allocation Hot Spots

              Functions responsible for significant heap allocations:
              | Function | Heap Escapes | Notes |
              |----------|-------------|-------|

              Heap escapes reduce GC efficiency. In hot paths (high fan-in, deep call chains),
              these are especially costly.

              ---

              ## Spec-Driven Module Analysis

              For each spec-driven module, a dedicated section:

              ### `path/to/module/` (spec-driven)
              Spec files: [SPECS.md, NOTES.md, ...]

              **Stated invariants relevant to complexity/coupling:**
              [List invariants that mention functions, complexity, or coupling constraints]

              **Findings:**
              - Functions in this module with score >= 40: [list]
              - Coupling score: N/100
              - Interface boundary to callers: [yes/no/partial]

              **Invariant gaps (undocumented structural properties):**
              [Properties observed in code that should be captured as invariants]
              Example: 'All exported functions are thin wrappers — no business logic' if observed.

              **Recommendations:**
              - [Specific actions tied to spec invariants]
              - Proposed new invariants for SPECS.md (require user approval before adding):
                N+1. [proposed invariant text] — Evidence: [file:line or pattern]

              ---

              ## Prioritized Recommendations

              Ordered by impact (highest first):

              ### P1 — Immediate Action
              [CRITICAL functions and FRAGILE modules]
              For each: what to do, where, why, and whether a spec needs updating.

              ### P2 — Next Sprint
              [HIGH functions and COUPLED modules]

              ### P3 — Backlog
              [MEDIUM functions, MODERATE modules, SDP violations without immediate risk]

              ### P4 — Monitoring
              [LOW functions, tool gaps to fill]

              ---

              ## Tool Gap Analysis

              If any tools were unavailable:

              | Tool | What It Would Reveal | Install |
              |------|---------------------|---------|
              | gocyclo | Cyclomatic complexity per function | go install github.com/fzipp/gocyclo/cmd/gocyclo@latest |
              | gocognit | Cognitive complexity per function | go install github.com/uudashr/gocognit/v2/cmd/gocognit@latest |
              | callgraph | Precise cross-package call graph | go install golang.org/x/tools/cmd/callgraph@latest |
              | staticcheck | Bug patterns, deprecated API usage | go install honnef.co/go/tools/cmd/staticcheck@latest |

              Run this analysis again after installing missing tools for complete coverage.")
```

**Input:** `.bob/state/go-discovery.md`, `.bob/state/go-analysis.md`, `.bob/state/go-scores.md`
**Output:** `.bob/state/go-analysis-report.md`

---

## Phase 6: COMPLETE

**Goal:** Present findings and handle proposed new spec invariants.

Read `.bob/state/go-analysis-report.md` and present:

```
Go analysis complete!

Report: .bob/state/go-analysis-report.md

[Paste the Executive Summary section]

[If CRITICAL]: N critical hot spots found. Immediate refactoring recommended.
  See .bob/state/go-analysis-report.md → "Critical Hot Spots"
  Use /bob:work-agents to fix — pass the report as context.

[If NEEDS ATTENTION]: Some high-complexity or tightly-coupled areas found.
  Review .bob/state/go-analysis-report.md for prioritized recommendations.

[If HEALTHY]: Codebase structure is healthy. No critical issues found.
```

### Proposed New Spec Invariants

If the report contains proposed new invariants for any spec-driven module, present them:

```
The analysis found N structural properties that are not yet captured as invariants:

[For each, numbered]:
  N. [proposed invariant text]
     Module: path/to/module/
     Evidence: [file:line or pattern]
     Target: SPECS.md

Which of these would you like to add? (e.g., "1,3", "all", "none")
```

Use `AskUserQuestion`. **NEVER add invariants to spec files without explicit user approval.**

For each approved invariant:
- Add to the appropriate SPECS.md using the next available number
- Keep wording as proposed unless the user requests changes

---

## Notes

- **Read-only** unless the user approves proposed spec invariants
- Run after major refactors to verify structural health has improved
- Pair with `/bob:audit` to check both structural health (this skill) and spec compliance (audit)
- The scoring weights are opinionated defaults — high cognitive complexity is weighted heavily because it predicts maintenance burden better than cyclomatic complexity alone
- Instability scores near 0.5 are not inherently bad — the concern is when an unstable package is imported by a stable one
- Missing tools degrade results gracefully: go vet + escape analysis always run; gocyclo/gocognit/callgraph add depth
