---
name: bob:audit
description: Verify CLAUDE.md invariants and analyze codebase health — DISCOVER → AUDIT → [ANALYZE → SCORE] → REPORT → COMPLETE
user-invocable: true
category: workflow
---

# Invariant Audit Workflow

You orchestrate a **read-only audit** that verifies code satisfies the invariants stated in CLAUDE.md files. Automatically detects Go codebases and includes deep structural analysis: call graphs, complexity scoring, and module coupling. No code changes — just a report.

## When to Use

- After generating invariants to verify they match the existing code
- Periodically to catch invariant drift across many PRs
- Before major refactors to confirm your understanding of current guarantees
- To identify Go complexity and coupling hot spots

## Workflow Diagram

**Non-Go codebase:**
```
INIT → DISCOVER → AUDIT → REPORT → COMPLETE
```

**Go codebase (auto-detected via go.mod):**
```
INIT → DISCOVER → [AUDIT + ANALYZE in parallel] → SCORE → REPORT → COMPLETE
```

**Read-only:** No code changes, no commits. Output is `.bob/state/audit-report.md`.

---

## Execution Rules

**CRITICAL: All subagents MUST run in background**

- ALWAYS use `run_in_background: true` for ALL Task calls
- After spawning agents, STOP — do not poll or check status
- Wait for agent completion notification
- Never use foreground execution

---

## Phase 1: INIT

**Goal:** Understand scope and detect codebase type.

**Actions:**

1. Ask the user using `AskUserQuestion`:

```
What would you like to audit?

1. All documented modules — scan the entire repo for modules with CLAUDE.md
2. Specific directory — audit one module (provide the path)
```

2. If the user provides a specific path, verify it contains a `CLAUDE.md`. If not, tell them they need to create a CLAUDE.md first.

3. **Auto-detect Go codebase:** Check if `go.mod` exists at the repo root (or in the scoped directory).
   - If `go.mod` exists, set `GO_ANALYSIS=yes` and probe for available tools:
     ```bash
     command -v gocyclo >/dev/null 2>&1 && echo "HAS_GOCYCLO=yes" || echo "HAS_GOCYCLO=no"
     command -v gocognit >/dev/null 2>&1 && echo "HAS_GOCOGNIT=yes" || echo "HAS_GOCOGNIT=no"
     command -v callgraph >/dev/null 2>&1 && echo "HAS_CALLGRAPH=yes" || echo "HAS_CALLGRAPH=no"
     ```
   - If `go.mod` does not exist, set `GO_ANALYSIS=no`.
   - The analysis degrades gracefully without optional tools — `go vet` and escape analysis always run.

4. Record:
   - `SCOPE`: all | `path/to/module`
   - `GO_ANALYSIS`: yes/no (auto-detected)
   - `HAS_GOCYCLO`, `HAS_GOCOGNIT`, `HAS_CALLGRAPH`: yes/no (if GO_ANALYSIS)

5. Create state directory:
```bash
mkdir -p .bob/state
```

---

## Phase 2: DISCOVER

**Goal:** Find all documented modules and (if Go codebase) collect raw structural data.

**Actions:**

Always spawn a discovery agent:

```
Task(subagent_type: "Explore",
     description: "Find all documented modules with CLAUDE.md",
     run_in_background: true,
     prompt: "Find all documented modules in the repository (or in the specific directory if scoped).

              A module is documented if its directory contains a CLAUDE.md file.

              For EACH module found:
              1. List the directory path
              2. Read CLAUDE.md and extract every numbered invariant, axiom, assumption, and constraint
              3. Count the number of .go files in the module

              Write findings to .bob/state/audit-discovery.md with this format:

              # Audit Discovery

              ## Module: `path/to/module/`

              **Go files:** N

              ### Invariants from CLAUDE.md
              1. [invariant text]
              2. [invariant text]
              ...

              [Repeat for each module]

              ## Summary
              - Total modules: N
              - Total invariants to verify: N")
```

If GO_ANALYSIS, also spawn a Go inventory agent in parallel:

```
Task(subagent_type: "Explore",
     description: "Inventory Go packages and collect structural data",
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

              Step 2 — Detect documented modules
              For every directory encountered, check for a CLAUDE.md file.
              A directory with CLAUDE.md is a documented module.

              For each documented module found, read CLAUDE.md and extract:
              - All numbered invariants
              - Any complexity or coupling constraints mentioned
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

              ## Documented Modules (CLAUDE.md)
              For each: path, invariants extracted from CLAUDE.md

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

**Output:** `.bob/state/audit-discovery.md` (always), `.bob/state/go-discovery.md` (if GO_ANALYSIS)

---

## Phase 3: AUDIT (and ANALYZE if GO_ANALYSIS)

**Goal:** Verify invariants per module. If GO_ANALYSIS, build call graph and coupling model in parallel.

**Actions:**

Read `.bob/state/audit-discovery.md`. For each module (or batch if few), spawn an audit agent:

```
Task(subagent_type: "Explore",
     description: "Audit module path/to/module against its CLAUDE.md invariants",
     run_in_background: true,
     prompt: "You are auditing whether code satisfies its stated invariants.

              Module: path/to/module/

              INVARIANTS TO VERIFY (from CLAUDE.md):
              1. [invariant text]
              2. [invariant text]
              ...

              For EACH invariant:
              1. Read the relevant code thoroughly
              2. Determine: does the code actually satisfy this guarantee?
              3. Record your finding as one of:
                 - PASS: Code satisfies the invariant. [Brief evidence]
                 - FAIL: Code violates the invariant. [Specific violation with file:line]
                 - PARTIAL: Code satisfies the invariant in most cases but has gaps. [Which cases fail]
                 - UNTESTABLE: Cannot determine from static analysis alone. [Why]

              Also check for:
              - Code behaviors that have no corresponding invariant in CLAUDE.md (undocumented guarantees)
              - Invariants in CLAUDE.md that reference types/functions that no longer exist (stale invariants)

              IMPORTANT — Candidate Invariant Discovery:
              Actively scan the code for patterns that SHOULD be documented as invariants but are not.
              Look for:
              - Thread-safety guarantees (mutexes, atomics, channel patterns)
              - Nil/error handling contracts (functions that never return nil, always return errors)
              - Ordering guarantees (initialization order, cleanup order)
              - Concurrency boundaries (what is safe to call concurrently)
              - Resource lifecycle rules (who creates, who closes, ownership transfers)
              - Interface compliance assumptions (which concrete types satisfy which interfaces)
              - Architectural boundaries (packages that must not import each other)

              For each candidate, write it as a proposed invariant in clear, numbered format.

              Write findings to .bob/state/audit-module-<module-name>.md with this format:

              # Audit: `path/to/module/`

              ## Invariant Verification

              | # | Invariant | Verdict | Evidence |
              |---|-----------|---------|----------|
              | 1 | [text] | PASS/FAIL/PARTIAL/UNTESTABLE | [brief evidence or file:line] |
              | 2 | [text] | ... | ... |

              ## Undocumented Behaviors
              [Code guarantees not captured in CLAUDE.md]

              ## Proposed New Invariants
              These are candidate invariants discovered by scanning the code.
              Each should be reviewed by the user before adding to CLAUDE.md.

              | # | Proposed Invariant | Evidence | Rationale |
              |---|-------------------|----------|-----------|
              | 1 | [proposed invariant text] | [file:line or pattern] | [why this should be an invariant] |

              ## Stale Invariants
              [Invariants referencing removed/renamed code]

              ## Summary
              - Invariants verified: N
              - PASS: N
              - FAIL: N
              - PARTIAL: N
              - UNTESTABLE: N
              - Candidate new invariants: N")
```

If multiple modules, spawn agents in parallel (one per module or batched sensibly).

If GO_ANALYSIS, also spawn the structural analysis agent in parallel with the audit agents:

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

              === CLAUDE.md INVARIANT CROSS-REFERENCE ===

              For each documented module (CLAUDE.md present):
              - Check if any invariants mention complexity bounds (e.g. 'functions must remain simple')
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

              ## CLAUDE.md Cross-Reference
              For each documented module:
                - Which functions appear as complexity hot spots?
                - Do stated invariants mention coupling or complexity?
                - Are interface boundaries respected?")
```

**Output:** `.bob/state/audit-module-*.md` (always), `.bob/state/go-analysis.md` (if GO_ANALYSIS)

---

## Phase 4: SCORE (only if GO_ANALYSIS)

**Goal:** Compute composite health scores per function and module, rank hot spots.

Skip this phase if GO_ANALYSIS is no.

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
                spec_violation_penalty           -- +10 if documented and has CRITICAL functions
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
              - Whether it is in a documented module (CLAUDE.md present) and which invariants apply

              Write to .bob/state/go-scores.md:

              # Go Health Scores

              ## Function Hot Spots (Top 20)
              | Rank | Package | Function | File:Line | Score | Tier | Dominant Factor | Documented? |
              |------|---------|----------|-----------|-------|------|----------------|-------------|

              ## Module Hot Spots (Top 10)
              | Rank | Package | Score | Tier | Ce | Ca | Interface Boundary | Documented? |
              |------|---------|-------|------|----|----|-------------------|-------------|

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

**Goal:** Consolidate all findings into a single report.

**Actions:**

Spawn a consolidation agent:

```
Task(subagent_type: "Explore",
     description: "Consolidate audit findings into final report",
     run_in_background: true,
     prompt: "Read all .bob/state/audit-module-*.md files.
              If .bob/state/go-scores.md exists, also read .bob/state/go-analysis.md and .bob/state/go-scores.md.

              Write to .bob/state/audit-report.md:

              # Audit Report

              Generated: [ISO timestamp]

              ## Executive Summary

              ### Invariant Compliance
              - Modules audited: N
              - Total invariants verified: N
              - PASS: N | FAIL: N | PARTIAL: N | UNTESTABLE: N
              - Stale invariant references found: N

              ### Go Structural Health (if applicable)
              - Packages analyzed: N
              - Functions analyzed: N
              - CRITICAL functions: N
              - FRAGILE modules: N
              - Stable Dependencies Principle violations: N
              - Tools available: [gocyclo yes/no, gocognit yes/no, callgraph yes/no]

              ## Overall Health: [HEALTHY / NEEDS ATTENTION / CRITICAL]

              HEALTHY if: no FAIL findings AND (no GO_ANALYSIS or no CRITICAL/FRAGILE)
              NEEDS ATTENTION if: any FAIL, PARTIAL findings, HIGH functions, or COUPLED modules
              CRITICAL if: >30% invariants FAIL, or any CRITICAL functions or FRAGILE modules

              ---

              ## Invariant Compliance Findings

              ### Failures (Code Violates Invariant)

              [List every FAIL finding across all modules, grouped by module]

              #### `path/to/module/`

              **Invariant N:** [text]
              **Violation:** [what the code does wrong, file:line]
              **Impact:** [what could go wrong because of this]

              ### Partial Compliance

              [List every PARTIAL finding]

              ### Stale Invariants

              [List every stale reference — invariants about removed code or renamed types]

              ### Undocumented Behaviors

              [List behaviors found in code but not captured in CLAUDE.md]

              ---

              ## Go Structural Findings (if applicable)

              ### Critical Hot Spots

              #### CRITICAL Functions (score 80-100)
              For each:

              **`package.FunctionName`** — Score: N/100 (CRITICAL)
              File: `path/file.go:line`
              Dominant factor: [cyclomatic | cognitive | call depth | allocations | blast radius]
              Metrics: cyclo=N, cognit=N, depth=N, heap_allocs=N, fan_in=N

              If in a documented module (CLAUDE.md present):
              > **Invariant context:** Relevant invariants: [list]
              > Risk: complexity may make these invariants difficult to verify.

              Recommendation: [specific, actionable refactoring suggestion]

              #### FRAGILE Modules (coupling score 70-100)
              For each:

              **`package/path`** — Coupling Score: N/100 (FRAGILE)
              Ce=N, Ca=N, Instability=0.N, Interface boundary: [NONE | PARTIAL | FULL]

              Recommendation: [define interface at boundary X, extract sub-package Y]

              ### Stable Dependencies Principle Violations

              For each:
              **`stable-package` imports `unstable-package`**
              Fix: introduce an interface or invert the dependency

              ### Module Coupling Overview

              [Full Module Coupling Matrix table from go-analysis.md]

              ### Call Graph Structure

              **Deepest Call Chains:** [Top 5 longest chains]

              **High Fan-In Functions (Blast Radius):**
              | Function | Fan-In | Packages Affected |
              |----------|--------|------------------|

              **Recursive Functions:** [list; flag if unbounded recursion possible]

              ### Allocation Hot Spots

              | Function | Heap Escapes | Notes |
              |----------|-------------|-------|

              ### Tool Gap Analysis

              If any tools were unavailable:
              | Tool | What It Would Reveal | Install |
              |------|---------------------|---------|
              | gocyclo | Cyclomatic complexity per function | go install github.com/fzipp/gocyclo/cmd/gocyclo@latest |
              | gocognit | Cognitive complexity per function | go install github.com/uudashr/gocognit/v2/cmd/gocognit@latest |
              | callgraph | Precise cross-package call graph | go install golang.org/x/tools/cmd/callgraph@latest |
              | staticcheck | Bug patterns, deprecated API usage | go install honnef.co/go/tools/cmd/staticcheck@latest |

              ---

              ## Proposed New Invariants

              These invariants were discovered by scanning the code but are NOT yet documented.
              They require user approval before being added.

              ### `path/to/module/`

              | # | Proposed Invariant | Evidence | Rationale |
              |---|-------------------|----------|-----------|
              | 1 | [text] | [file:line] | [why this matters] |

              [Group by module]

              ---

              ## Recommendations

              ### Immediate (invariant failures and CRITICAL structural issues)
              For each FAIL: fix the code OR update CLAUDE.md if the invariant is wrong
              For each CRITICAL function: [specific refactoring action]
              For each FRAGILE module: [specific decoupling action]

              ### Short-term (partial compliance and HIGH structural issues)

              ### Backlog (stale invariants, MEDIUM functions, SDP violations)")
```

**Output:** `.bob/state/audit-report.md`

---

## Phase 6: COMPLETE

**Goal:** Present results and handle proposed new invariants.

**Actions:**

Read `.bob/state/audit-report.md` and present the summary:

```
Audit complete!

Results: .bob/state/audit-report.md

[Paste the Executive Summary section]

[If HEALTHY]: All invariants verified. CLAUDE.md and code are in sync.

[If NEEDS ATTENTION]: Found N issues where code diverges from invariants.
  Review .bob/state/audit-report.md for details.
  Fix the code or update CLAUDE.md — use /bob:work for code fixes.

[If CRITICAL]: Significant invariant drift or structural issues detected.
  Review .bob/state/audit-report.md and prioritize fixes.
```

### Proposed New Invariants Review

If the audit report contains proposed new invariants, present them to the user for approval:

```
The audit discovered N candidate invariants that are not yet documented:

[For each candidate, numbered]:
  N. [proposed invariant text]
     Evidence: [file:line or pattern]
     Module: path/to/module/
     Target: CLAUDE.md

Which of these would you like to add? (e.g., "1,3,5", "all", or "none")
```

Use `AskUserQuestion` to get the user's selection. **NEVER add invariants to CLAUDE.md without explicit user approval.**

For each approved invariant:
- Add it to the appropriate module's CLAUDE.md
- Use the next available number in the existing invariant list
- Keep the wording as proposed unless the user requests changes

Skip any the user declines. If the user says "none", proceed to completion without changes.

---

## Notes

- This workflow is **read-only** unless the user approves proposed new invariants
- CLAUDE.md files are only modified with explicit user consent
- Findings can feed into `/bob:work` to fix violations
- Run periodically (e.g., before releases) to catch invariant drift
- The audit checks code against invariants, not invariants against code
- Go structural analysis scoring weights are opinionated: cognitive complexity is weighted heavily because it predicts maintenance burden better than cyclomatic complexity alone
- Instability scores near 0.5 are not inherently bad — the concern is when an unstable package is imported by a stable one
- Missing tools degrade gracefully: go vet + escape analysis always run; gocyclo/gocognit/callgraph add depth
