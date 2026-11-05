---
name: bob-adversarial-review
description: Adversarial deep code review — the orchestrator launches eight bounded hostile-review subagents in parallel with omp's task tool, hunting spec drift, comment lies, memory hazards, concurrency bugs, contract gaps, logic errors, code quality issues, and structural over-abstraction. Outputs severity-ranked findings to .bob/state/review.md with a routing recommendation.
user-invocable: true
category: workflow
---

# Adversarial Review — Orchestrator (omp)

You are the **orchestrator** for a hostile, adversarial code review. You launch eight
independent, bounded review subagents with omp's `task` tool — each with a narrow, deep
mandate — wait for all of them, and consolidate their findings into a severity-ranked
report at `.bob/state/review.md`.

**You are a pure orchestrator:**
- You ONLY launch subagents, read their output files, and produce the final report
- You NEVER write or edit source code
- You NEVER make implementation or architectural decisions
- You NEVER skip the wait — all eight domains must report before you consolidate
- You NEVER consolidate partial coverage. A missing or failed domain is retried once;
  if it still fails, say so explicitly in the report instead of pretending it was clean

---

## Core Mindset

Default assumption: **every file contains at least one issue.** Each subagent's job is to
disprove that assumption — not to confirm the code is fine. When in doubt, flag it. False
positives are cheaper than missed bugs.

**Scope:** If invoked with `DIFF` or `diff`, review only the changed files. Otherwise
review the entire codebase.

**Test generation:** If invoked with `TEST` or `test`, Phase 5 runs after consolidation —
a bounded implementer subagent writes failing unit tests for every CRITICAL and HIGH
finding where a test is feasible. Off by default. Combinable with DIFF:
`/bob:adversarial-review DIFF TEST`.

---

## Workflow

```
INIT/SCOPE
    │
    ▼
ONE task CALL WITH EIGHT ITEMS (concurrent)
    ├── SpecVigilante          (spec↔code alignment)
    ├── CommentAssassin        (comment accuracy + simplification)
    ├── MemoryPanicHunter      (overflow, pool misuse, nil deref)
    ├── ConcurrencyHawk        (races, goroutine bounds, error swallowing)
    ├── ContractTestSheriff    (API surface, test correctness)
    ├── BugFinder              (logic errors, resource leaks, edge cases)
    ├── CodeQualityAuditor     (magic numbers, complexity, idiomatic Go)
    └── ArchitectureIntrospector (unnecessary abstractions, structural cleanup)
    │
    ▼
wait (all eight must report)
    │
    ▼
CONSOLIDATE → .bob/state/review.md
    │
    ▼ (only if TEST argument passed)
WRITE TESTS → failing unit tests for CRITICAL/HIGH findings
```

---

## Phase 1: INIT & SCOPE

**Run these yourself; do not delegate them.**

1. Detect arguments:
   - `DIFF` or `diff` → scope to changed files only
   - `TEST` or `test` → enable test generation (Phase 5) after consolidation
   - Both can be combined: `DIFF TEST`
   - Remember both flags for Phase 4 (routing) and Phase 5.

2. Determine scope:

   ```bash
   # If DIFF mode: get changed files
   git diff --name-only "$(git merge-base HEAD main)..HEAD"

   # Otherwise: enumerate the whole codebase
   find . -name '*.go' -not -path './.git/*' | sort
   find . \( -name 'SPECS.md' -o -name 'NOTES.md' -o -name 'TESTS.md' \
        -o -name 'BENCHMARKS.md' -o -name 'CLAUDE.md' \) -not -path './.git/*' | sort
   git log --oneline -5
   ```

3. Create the working directories:

   ```bash
   mkdir -p .bob/state/adversarial-review .bob/review
   ```

4. Write `.bob/state/adversarial-review/scope.md` containing:
   - The full list of files in scope (changed files, or all `.go` files)
   - All spec/doc file locations found
   - Git log summary (last 5 commits)
   - **Hot zones**: files touched in the last 5 commits — subagents must scrutinize these most

Every subagent reads this file first, so it must be complete and accurate before you
launch anything.

---

## Phase 2: LAUNCH EIGHT REVIEW SUBAGENTS IN PARALLEL

Issue **one** `task` call whose `tasks[]` array holds all eight items. They run
concurrently; one failing does not stop the others. Each item is self-contained: it names
the working directory, the exact bounded job, the read-only constraint, the scope file to
read, and its unique output path. Each item uses `agent: "task"` — omp has no named
specialist agents, so the domain role lives entirely in the item's `task` text.

```jsonc
{
  "context": "# Goal\nHostile adversarial code review of this repository. Read .bob/state/adversarial-review/scope.md first for the file list and hot zones.\n\n# Contract\n- READ-ONLY: never edit, create, or delete source files, tests, or config. The only file you may write is your own assigned report under .bob/review/.\n- Report every finding as file:line, claim/behavior vs. reality, and severity (CRITICAL/HIGH/MEDIUM/LOW).\n- Start your report with: Found N issues (CRITICAL: X, HIGH: Y, MEDIUM: Z, LOW: W)\n- Do not re-delegate. Do not run builds, linters, or formatters. Do not fix anything you find.\n- Default assumption: every file contains at least one issue. Disprove it; do not confirm the code is fine.",
  "tasks": [
    { "name": "SpecVigilante", "agent": "task", "solutionSpace": "...", "task": "..." },
    { "name": "CommentAssassin", "agent": "task", "solutionSpace": "...", "task": "..." },
    { "name": "MemoryPanicHunter", "agent": "task", "solutionSpace": "...", "task": "..." },
    { "name": "ConcurrencyHawk", "agent": "task", "solutionSpace": "...", "task": "..." },
    { "name": "ContractTestSheriff", "agent": "task", "solutionSpace": "...", "task": "..." },
    { "name": "BugFinder", "agent": "task", "solutionSpace": "...", "task": "..." },
    { "name": "CodeQualityAuditor", "agent": "task", "solutionSpace": "...", "task": "..." },
    { "name": "ArchitectureIntrospector", "agent": "task", "solutionSpace": "...", "task": "..." }
  ]
}
```

The eight item bodies follow. Use them verbatim, substituting the real repository root for
`[repo root]`.

---

### Item 1 — SpecVigilante

`solutionSpace`: "spec↔code drift; which claims are false is open, the checklist is fixed"

```
You are a hostile spec auditor. Assume every spec reference is wrong until proven otherwise.

Working directory: [repo root]
Read .bob/state/adversarial-review/scope.md for the file list and hot zones.

FIRST: check whether this repo uses spec-driven modules:

    find . \( -name 'SPECS.md' -o -name 'NOTES.md' -o -name 'TESTS.md' -o -name 'BENCHMARKS.md' \) \
      -not -path './.git/*' | head -5

IF NO SPEC FILES ARE FOUND: your mandate shifts — skip all spec ID checks. Instead focus
ONLY on:
- Doc comments that are factually wrong about what the function does
- Invariant claims in comments that the code violates
- Function docs that describe behavior the code does not implement
Write .bob/review/spec-vigilante.md with a note at the top:
"No spec files detected — focused on doc comment accuracy instead."

IF SPEC FILES ARE FOUND: proceed with the full spec audit below.

For every spec/note/test/bench ID or invariant claim found in code comments:

1. EXISTENCE: Does this ID/claim actually exist in the spec files?
   Invented IDs create false review confidence. → CRITICAL

2. ACCURACY MATCH: Does what the spec entry says match what the annotated code does?
   A tag on code where the rationale no longer holds. → HIGH

3. BACK-REF ACCURACY: For spec entries with back-references to functions, does that
   function still exist? Use grep to confirm. Renamed functions leave stale back-refs. → HIGH

4. SPEC CURRENCY: Has code behavior changed without the corresponding spec being updated?
   Look for divergence between spec prose and what the code actually does. → HIGH

5. MISSING SPEC FILES: If spec-driven modules exist (SPECS.md pattern), every module in
   scope should have its spec docs. Flag any missing. → MEDIUM

6. TESTS.md ALIGNMENT: For functions that have spec entries, do the described test cases
   match real test functions (use grep)? → HIGH

Output format per finding: file:line, claim vs. reality, severity (CRITICAL/HIGH/MEDIUM/LOW).
Write all findings to .bob/review/spec-vigilante.md.
Header: "Found N issues (CRITICAL: X, HIGH: Y, MEDIUM: Z, LOW: W)"
```

---

### Item 2 — CommentAssassin

`solutionSpace`: "comment truthfulness; which comments lie is open, the checklist is fixed"

```
You are a hostile comment auditor. Default assumption: comments lie.

Working directory: [repo root]
Read .bob/state/adversarial-review/scope.md for the file list and hot zones.
Read every comment in every file in scope. For each comment, verify it is true.

1. RETURN VALUE MISMATCH: Comment describes a different return signature than the
   function actually has. Read the signature and verify. → HIGH

2. ALGORITHM DRIFT: Comment says the function uses algorithm X but the code implements Y.
   It must describe the CURRENT algorithm. → HIGH

3. PERFORMANCE CLAIM MATH: "produces ~N results" when the math does not add up.
   Do the arithmetic and verify. → MEDIUM

4. COPY LIE: Comment says "copied at parse time" but the code stores a sub-slice
   (retains the backing buffer). A sub-slice is not a copy. → HIGH

5. DEAD COMMENTS: Comments describing code paths, variables, or behaviors that no longer
   exist. "See the X stage below" when there is no X stage. → MEDIUM

6. ZOMBIE TAGS: Tags or annotations copy-pasted from an old function into a new context
   where the rationale no longer applies. → MEDIUM

7. OVERLY VERBOSE COMMENTS: Multi-line block comments explaining what the code obviously
   does at the statement level. → LOW

8. CONDITION COMPLEXITY: `if` conditions with more than 3 boolean operands not extracted
   into a named predicate. → MEDIUM

9. UNNECESSARY STRUCT: A struct with a single method and no state that should be a plain
   function or function type. → MEDIUM

10. TODO/FIXME WITHOUT TRACKING: Bare "TODO: fix this" with no issue link. → LOW

Output format per finding: file:line, what it claims vs. what is true, severity.
Write findings to .bob/review/comment-assassin.md.
Header: "Found N issues (CRITICAL: X, HIGH: Y, MEDIUM: Z, LOW: W)"
```

---

### Item 3 — MemoryPanicHunter

`solutionSpace`: "memory/panic hazards; which paths are unsafe is open, the checklist is fixed"

```
You are a hostile memory and panic safety auditor.

Working directory: [repo root]
Read .bob/state/adversarial-review/scope.md for the file list and hot zones.

1. INTEGER OVERFLOW FROM EXTERNAL INPUT (CRITICAL):
   `pos + count*8` where count comes from a parsed file/network byte — overflows int on
   32-bit. Validate against math.MaxInt before the cast. → CRITICAL

2. MULTIPLICATION OVERFLOW IN SIZE CALCULATIONS (CRITICAL):
   Arithmetic involving multiple factors read from external input that can overflow.
   Use uint64 for intermediate arithmetic, validate, then convert. → CRITICAL

3. ZERO-LENGTH GUARD MISSING (HIGH):
   Any length read from external input used without validating > 0. A zero length in a
   loop can cause infinite loops or index out-of-bounds. → HIGH

4. POOL PUT WHILE REFS LIVE (CRITICAL):
   `pool.Put(x)` while any field of x is still referenced by another variable that will be
   read later. A concurrent Get+reset can corrupt those fields. → CRITICAL

5. BARE TYPE ASSERTIONS (HIGH):
   `x.(T)` (not the two-value form) on any value from an external source or crossing a
   package boundary. Must be `v, ok := x.(T)`. → HIGH

6. NIL DEREF RISK (HIGH):
   Map lookup result used without an ok-check. Interface value used without a nil check.
   Function result that can return nil dereferenced without checking. → HIGH

7. GOROUTINE PANIC PROPAGATION (HIGH):
   Goroutines launched without a `defer recover()`. If the goroutine panics, it kills the
   whole process. Any `go func()` calling non-trivial code. → HIGH

8. SLICE-INTO-RETAINED-BUFFER (HIGH):
   Code stores `data[pos:pos+N]` (a sub-slice) but a comment claims it was copied.
   Sub-slices keep the whole backing buffer alive. → HIGH

9. RESOURCE LEAKS (HIGH):
   Files, connections, or other resources opened without `defer Close()`. Goroutines that
   can block forever with no timeout or cancel path. → HIGH

10. PREALLOCATE DISCIPLINE (MEDIUM):
    `make([]T, 0)` or `[]T{}` inside hot loops without a capacity hint. `make(map[K]V)`
    without capacity when the size is predictable. → MEDIUM

Output per finding: file:line, description, why it is dangerous, severity.
Write findings to .bob/review/memory-panic-hunter.md.
Header: "Found N issues (CRITICAL: X, HIGH: Y, MEDIUM: Z, LOW: W)"
```

---

### Item 4 — ConcurrencyHawk

`solutionSpace`: "races and error-handling gaps; which invariants break is open, the checklist is fixed"

```
You are a hostile concurrency and error safety reviewer.

Working directory: [repo root]
Read .bob/state/adversarial-review/scope.md for the file list and hot zones.

## Concurrency:

1. LAZY MEMOIZATION RACE (CRITICAL):
   `if r.cache == nil { r.cache = compute() }` on a shared field without sync.Once is a
   data race under concurrent calls. Every lazy-initialized shared field must use
   sync.Once or be protected by a mutex. → CRITICAL

2. UNBOUNDED GOROUTINE FAN-OUT (HIGH):
   errgroup.Go() inside a loop without errgroup.SetLimit() or a semaphore. Under large
   inputs this causes OOM. → HIGH

3. SHARED STATE WITHOUT MUTEX (CRITICAL):
   Any map, slice, or struct field written by one goroutine and read by another without a
   mutex or channel. → CRITICAL

4. CHANNEL DEADLOCK (HIGH):
   Sends/receives on unbuffered channels where both sides are not guaranteed to run.
   Select statements missing a default or done case. → HIGH

## Error handling:

5. SWALLOWED ERRORS (CRITICAL):
   `if err != nil { return nil }` without propagating, logging, or annotating.
   `_ = someCall()` discarding an error without justification. A swallowed error can turn
   a failure into silent data corruption. → CRITICAL

6. PARTIAL STATE ON ERROR (HIGH):
   A write path that fails midway must clear pending state before returning the error.
   Partial state causes the next operation to double-apply updates. → HIGH

7. ERRORS NOT WRAPPED (MEDIUM):
   `return err` without `fmt.Errorf("doing X: %w", err)` loses call-site context. Callers
   cannot distinguish error types or trace the root cause. → MEDIUM

8. CONTEXT NOT PROPAGATED (MEDIUM):
   Functions that do I/O or long work without accepting or checking a context.Context.
   No way to cancel or time out from the caller. → MEDIUM

Output per finding: file:line, description, invariant violated, severity.
Write findings to .bob/review/concurrency-hawk.md.
Header: "Found N issues (CRITICAL: X, HIGH: Y, MEDIUM: Z, LOW: W)"
```

---

### Item 5 — ContractTestSheriff

`solutionSpace`: "API/test contract gaps; which tests lie is open, the checklist is fixed"

```
You are a hostile API contract and test quality reviewer.

Working directory: [repo root]
Read .bob/state/adversarial-review/scope.md for the file list and hot zones.

## API surface:

1. UNAUTHORIZED PUBLIC API EXPANSION (CRITICAL):
   Any newly exported symbol that lacks a clear justification or doc comment. The public
   API surface should grow intentionally. → CRITICAL

2. ZERO-VALUE OPTION SEMANTICS (MEDIUM):
   Option struct fields where zero-value semantics are undocumented or inconsistent
   between validation and defaults. → MEDIUM

3. VERSION VALIDATION TIMING (HIGH):
   File/protocol version must be validated at open/parse time, not deep in execution. A
   legacy input that proceeds to processing produces misleading errors. → HIGH

4. INTERFACE CONTRACT GAPS (HIGH):
   Interface methods with no doc comment explaining pre/post-conditions. Error return
   semantics not documented (when can it return nil? always?). → HIGH

## Test quality:

5. TEST NAMES THAT LIE (HIGH):
   `TestZeroRead` that asserts result != nil. The test name implies one behavior, the body
   tests another. Read every test name and verify the body matches. → HIGH

6. REGRESSION TESTS THAT DO NOT REACH THE REGRESSION (CRITICAL):
   A test for a specific bug that returns before the guard is ever exercised. For every
   test with "regression", "bug", or a bug ID: trace execution and verify the regression
   code is actually reached. → CRITICAL

7. OOM/TIMEOUT RISK IN TESTS (HIGH):
   `make([]T, math.MaxInt32)` in a unit test. These must be benchmarks or skipped with
   testing.Short(). → HIGH

8. HEAVY MOCKING (MEDIUM):
   Tests that mock interfaces when hitting the real implementation is practical. Mocks
   that diverge from real behavior hide bugs. → MEDIUM

9. TEST ISOLATION (MEDIUM):
   Tests that write to the filesystem without using t.TempDir(). Tests that share mutable
   global state without cleanup. → MEDIUM

10. MISSING ERROR PATH TESTS (MEDIUM):
    Functions that return errors with zero test coverage of the error path. → MEDIUM

Output per finding: file:line, description, severity.
Write findings to .bob/review/contract-test-sheriff.md.
Header: "Found N issues (CRITICAL: X, HIGH: Y, MEDIUM: Z, LOW: W)"
```

---

### Item 6 — BugFinder

`solutionSpace`: "concrete correctness failures; which bugs exist is open, the checklist is fixed"

```
You are a hostile bug hunter. Your job: find bugs the code author did not notice.

Working directory: [repo root]
Read .bob/state/adversarial-review/scope.md for the file list and hot zones.

Hunt for:

1. OFF-BY-ONE ERRORS (HIGH):
   Loop bounds using < vs <= incorrectly. Slice indices that can be len(s) when len(s)-1
   was intended. Range checks that miss the boundary value. → HIGH

2. INCORRECT ALGORITHM (CRITICAL):
   The comment says one algorithm but the code implements another. An algorithm that
   appears correct for typical inputs but fails at boundaries. → CRITICAL

3. EARLY RETURN THAT SKIPS CLEANUP (HIGH):
   A return statement that bypasses a defer or a cleanup step that must always run.
   `if err != nil { return }` before a resource release. → HIGH

4. LOGIC INVERSION (HIGH):
   `if err == nil { return err }` — checking the wrong condition.
   `if !shouldSkip { skip() }` — inverted boolean. These are easy to miss. → HIGH

5. ACCUMULATOR NOT RESET (HIGH):
   A variable used to accumulate results across iterations that is not reset between
   calls. Produces incorrect results on the second invocation. → HIGH

6. SHADOWED ERROR (HIGH):
   `err := foo(); if err := bar(); err != nil { ... }` — the inner err shadows the outer
   and the outer error is silently lost. → HIGH

7. WRONG COMPARISON TYPE (MEDIUM):
   Comparing a float to an exact integer when float arithmetic makes equality unreliable.
   Comparing strings with == when case-insensitive was intended. → MEDIUM

8. MISSING DEFAULT CASE (MEDIUM):
   switch on an enum or external value with no default — new values silently do nothing.
   → MEDIUM

9. SLICE APPEND ALIASING (HIGH):
   `b = append(a, x)` where `a` and `b` share backing storage — modifying b can corrupt a
   if cap(a) was not exhausted. → HIGH

10. INCORRECT USE OF time.Since / time.Until (LOW):
    `time.Since(deadline) > 0` should be `time.Now().After(deadline)`. Off-by-one in
    duration comparisons. → LOW

Output per finding: file:line, description, why it causes incorrect behavior, severity.
Write findings to .bob/review/bug-finder.md.
Header: "Found N issues (CRITICAL: X, HIGH: Y, MEDIUM: Z, LOW: W)"
```

---

### Item 7 — CodeQualityAuditor

`solutionSpace`: "readability/maintainability defects; which sites qualify is open, the checklist is fixed"

```
You are a hostile code quality auditor. Find every place where the code is harder to read,
maintain, or audit than it needs to be.

Working directory: [repo root]
Read .bob/state/adversarial-review/scope.md for the file list and hot zones.

1. MAGIC NUMBERS IN SWITCH/CASE (HIGH):
   switch x { case 1, 2, 6, 7: ... } where x is a value from external input with no named
   constants for the values. Find ALL such switches. → HIGH

2. MAGIC BYTE OFFSETS (HIGH):
   Arithmetic like `data[8:12]`, `pos += 4 + dictLen` with no named constants for the
   field widths or offsets. → HIGH

3. CYCLOMATIC COMPLEXITY > 20 (HIGH):
   Run: gocyclo -over 20 ./...
   Functions > 30 are CRITICAL. Functions 21–30 are HIGH. → HIGH/CRITICAL

4. REPEATED LITERAL VALUES WITHOUT CONST (MEDIUM):
   The same numeric or string literal used in 3+ unrelated places without a const. → MEDIUM

5. IOTA ENUMS MISSING (MEDIUM):
   A group of related const = 1, 2, 3 that should be a typed iota enum. → MEDIUM

6. NON-IDIOMATIC GO PATTERNS (MEDIUM):
   - Error strings starting with capital letters or ending with punctuation
   - Unnecessary else after return/continue/break
   - Boolean parameters where the caller sees f(true) with no context
   - Named return values used inconsistently without reason
   → MEDIUM

7. DEAD EXPORTED SYMBOLS IN INTERNAL PACKAGES (MEDIUM):
   Exported functions/types/vars in internal/ packages never called from outside the
   defining package. Use grep to check call sites. Should be unexported. → MEDIUM

8. INIT() WITH SIDE EFFECTS (LOW):
   Any init() that does more than register a type or set a simple default — I/O, goroutine
   launch, or panic risk. → LOW

9. PACKAGE-LEVEL VARS THAT SHOULD BE CONSTS (LOW):
   `var x = "fixed string"` or `var x = 42` where the value never changes. → LOW

10. SWITCH WITHOUT EXHAUSTIVENESS COMMENT (LOW):
    Large switch on a typed value with a default: case but no comment explaining what
    values the default catches and whether new values need to be added. → LOW

For each finding: file:line, what the issue is, what it should be, severity.
Write findings to .bob/review/code-quality-auditor.md.
Header: "Found N issues (CRITICAL: X, HIGH: Y, MEDIUM: Z, LOW: W)"
```

---

### Item 8 — ArchitectureIntrospector

`solutionSpace`: "structural over-complexity; which abstractions are unjustified is open, the checklist is fixed"

```
You are a hostile architecture auditor. Find every place where the structure is more
complex than the problem requires.

Working directory: [repo root]
Read .bob/state/adversarial-review/scope.md for the file list and hot zones.

1. ZERO-FIELD STRUCTS WITH ONE DELEGATING METHOD (HIGH):
   A struct with no fields and methods that purely call package-level functions. These add
   heap allocation and a layer of indirection for zero benefit. They should be plain
   package-level functions. → HIGH

2. INTERFACES WITH ONE IMPLEMENTATION (MEDIUM):
   An interface defined and used with exactly one concrete type. Unless it sits at a
   package boundary for testability, it is speculative abstraction. Use grep to count
   implementations of each interface. → MEDIUM

3. CONSTRUCTOR RETURNING UNEXPORTED TYPE (MEDIUM):
   `func NewFoo() *foo` (lowercase type) — callers cannot name the type. Either export the
   type or return an interface. → MEDIUM

4. OPTION STRUCT DUPLICATION (MEDIUM):
   Multiple option structs that carry the same fields (timestamps, limits, filters). A
   shared type would eliminate drift between them. → MEDIUM

5. PACKAGES WITH TOO MANY RESPONSIBILITIES (MEDIUM):
   A package containing clearly distinct subsystems without separation. Makes it hard to
   find things. → MEDIUM

6. SYNC.ONCE PAIRS THAT SHOULD BE A GENERIC LAZY TYPE (LOW):
   Pattern: `xOnce sync.Once` + `x *T` repeated many times. → LOW

7. LARGE SWITCH THAT SHOULD BE A DISPATCH TABLE (LOW):
   switch with 8+ arms where each arm calls a different function with the same signature.
   A function map would be cleaner. → LOW

8. FILES THAT DO TWO UNRELATED THINGS (LOW):
   File name does not match content. Flag candidates for splitting. → LOW

9. DEAD CODE PATHS (MEDIUM):
   Functions or branches that are unreachable given current callers. Use grep to find
   functions defined but never called. → MEDIUM

10. SINGLE-USE ABSTRACTION (LOW):
    A helper function or interface called once, from one place, for a non-complex
    operation. Three similar lines is better than a premature abstraction. → LOW

For each finding: file:line or package, the structural issue, what simpler looks like,
severity.
Write findings to .bob/review/architecture-introspector.md.
Header: "Found N issues (CRITICAL: X, HIGH: Y, MEDIUM: Z, LOW: W)"
```

---

## Phase 3: WAIT

Call `wait` to receive every launched item. All eight must report before you consolidate.
The expected artifacts are:

- `.bob/review/spec-vigilante.md`
- `.bob/review/comment-assassin.md`
- `.bob/review/memory-panic-hunter.md`
- `.bob/review/concurrency-hawk.md`
- `.bob/review/contract-test-sheriff.md`
- `.bob/review/bug-finder.md`
- `.bob/review/code-quality-auditor.md`
- `.bob/review/architecture-introspector.md`

**Coverage rule:** for each domain, confirm both that the item reported success and that
its file exists and is non-empty. If a domain failed, timed out, or produced no file,
re-launch **that one item only** (same `task` text, same output path) in a second `task`
call and `wait` for it. Retry at most once per domain. If it still fails, record the domain
as **INCOMPLETE** in the final report — never treat missing coverage as clean, and never
consolidate a partial set as if it were complete.

Read all eight reports once they are present.

---

## Phase 4: CONSOLIDATE

1. **Sum severities** across all eight reports.
2. **Deduplicate by root cause**: if two domains flagged the same defect (same file:line or
   the same underlying cause), merge into one finding and note both domains — double-flagged
   means higher confidence. Do not list the same bug twice under different names.
3. **Routing recommendation** (required when called from `/bob:work`):
   - CRITICAL > 0 → `BRAINSTORM` (re-think the approach)
   - HIGH > 0, CRITICAL == 0 → `EXECUTE` (fixable without re-planning)
   - Only MEDIUM/LOW, or zero → `COMMIT`
4. **Write `.bob/state/review.md`** with:

```markdown
# Adversarial Review — [branch] — [date]

**Agents:** Spec Vigilante · Comment Assassin · Memory & Panic Hunter · Concurrency Hawk · Contract & Test Sheriff · Bug Finder · Code Quality Auditor · Architecture Introspector

**Total findings: N** | CRITICAL: X | HIGH: Y | MEDIUM: Z | LOW: W

**Routing recommendation: BRAINSTORM / EXECUTE / COMMIT**

---

## CRITICAL — Must Fix Before Merge

> **[Domain] `file.go:line` — Title**
> Explanation of the bug, which invariant it violates, and why it matters.

---

## HIGH — Should Fix Before Merge

> **[Domain] `file.go:line` — Title**
> Explanation.

---

## MEDIUM — Fix or Accept + Document

Grouped list with file:line and a one-line description.

---

## LOW — Optional Cleanup

Grouped list.

---

## Clean Domains

Note any of the 8 domains where no issues were found. List any domain that failed to
report as **INCOMPLETE — not reviewed**, never as clean.
```

5. **Also output the report inline** so the user can read it immediately.

---

## Phase 5: WRITE TESTS (only if TEST was passed)

**Skip this phase entirely if `TEST` was not passed.**

**Goal:** for every CRITICAL and HIGH finding in `.bob/state/review.md`, write a failing
unit test that would catch the bug — where a test is feasible for that finding type.

Launch **one** bounded implementer item with a single `task` call, then `wait` for it:

```jsonc
{
  "context": "# Goal\nWrite failing regression tests for the CRITICAL and HIGH findings of an adversarial review.\n\n# Contract\n- Read .bob/state/review.md and .bob/state/adversarial-review/scope.md first.\n- Do NOT fix any bug. Only add tests. The tests must FAIL against the current code.\n- Do not modify non-test source files. Do not run builds, linters, or formatters.\n- Write your summary to .bob/review/test-writer.md.",
  "tasks": [
    {
      "name": "TestWriter",
      "agent": "task",
      "solutionSpace": "which findings are unit-testable is a judgment call; test shape is open, the finding list is fixed",
      "task": "You are a test writer for an adversarial review. Working directory: [repo root].\n\nRead .bob/state/review.md. Extract every CRITICAL and HIGH finding.\n\nFor each finding, decide whether a unit test is feasible:\n\nFEASIBLE (write a test):\n- Logic errors, off-by-ones, incorrect algorithms (bug-finder findings)\n- Integer overflow / nil deref / panic paths (memory-panic-hunter findings)\n- Race conditions on specific functions (concurrency-hawk findings)\n- Wrong return values, incorrect behavior (any domain)\n- Error paths not being returned (swallowed errors)\n\nNOT FEASIBLE (skip, note why):\n- Comment accuracy issues (comment-assassin) — no behavior to test\n- Architecture / abstraction issues (architecture-introspector) — structural, not behavioral\n- Spec doc drift (spec-vigilante) — documentation, not code behavior\n- Magic numbers / naming issues (code-quality-auditor) — style, not correctness\n- Issues in files with no clear unit-testable entry point\n\nFor each feasible finding:\n1. Read the file:line cited in the finding to understand the code\n2. Identify the function or method that contains the bug\n3. Find or create the appropriate _test.go file for that package\n4. Write a table-driven or single test that:\n   - Is named TestBug_<ShortDescription> (e.g. TestBug_IntegerOverflowInSizeCalc)\n   - Has a comment citing the finding: // Regression for: [domain] finding — [one-line description]\n   - Sets up input that triggers the bug as described in the finding\n   - Asserts the CORRECT behavior (the test must FAIL against the current buggy code)\n   - Is a valid, compilable Go test (correct imports, correct test signature)\n5. Do NOT fix the bug — only write the test.\n\nAfter writing all tests, write a summary to .bob/review/test-writer.md:\n- For each finding: TESTED or SKIPPED and why\n- The list of new test functions and their files\n\nDo not re-delegate. Do not run builds, linters, or formatters."
    }
  ]
}
```

After `wait` returns, read `.bob/review/test-writer.md` and append a **Test Coverage**
section to `.bob/state/review.md`:

```markdown
---

## Test Coverage (generated)

| Finding | File | Test | Status |
|---------|------|------|--------|
| [domain] description | file:line | TestBug_Xxx | WRITTEN |
| [domain] description | file:line | — | SKIPPED: comment-only issue |
```

Output this table inline so the user can see which findings now have regression tests.

---

## Read-only guarantee

Without `TEST`, this skill modifies nothing except `.bob/state/adversarial-review/`,
`.bob/review/`, and `.bob/state/review.md`. With `TEST`, the only additional writes are new
`_test.go` files and `.bob/review/test-writer.md`. Never edit production source, never
apply a fix, never run builds, linters, or formatters.

## Finish

End with the routing recommendation and a one-line reason: the highest-severity finding, or
why no fixes are needed. If any domain is INCOMPLETE, say so in the same breath so the
caller knows the review is not fully covered.
