---
name: bob:audit
description: Verify code satisfies stated invariants in spec-driven modules — DISCOVER → AUDIT → REPORT
user-invocable: true
category: workflow
---

# Spec Audit Workflow

You orchestrate a **read-only audit** that verifies code satisfies the invariants stated in spec documents. No code changes — just a report of where reality diverges from the spec.

## When to Use

- After `/bob:design apply` to verify freshly-generated specs match the existing code
- Periodically to catch spec drift across many PRs
- Before major refactors to confirm your understanding of current guarantees
- As a health check on spec-driven modules

## Workflow Diagram

```
INIT → DISCOVER → AUDIT → REPORT → COMPLETE
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

**Goal:** Understand scope of the audit.

**Actions:**

1. Ask the user what to audit using `AskUserQuestion`:

```
What would you like to audit?

1. All spec-driven modules — scan the entire repo for modules with SPECS.md
2. Specific directory — audit one module (provide the path)
```

2. If the user provides a path, verify it contains spec files (SPECS.md, NOTES.md, TESTS.md, or BENCHMARKS.md). If not, tell them and suggest `/bob:design` to create specs first.

3. Create `.bob/state/` if it doesn't exist:
```bash
mkdir -p .bob/state
```

---

## Phase 2: DISCOVER

**Goal:** Find all spec-driven modules in scope.

**Actions:**

Spawn an Explore agent to find spec-driven modules:

```
Task(subagent_type: "Explore",
     description: "Find all spec-driven modules",
     run_in_background: true,
     prompt: "Find all spec-driven modules in the repository (or in the specific directory if scoped).

              A module is spec-driven if its directory contains any of:
              - SPECS.md
              - NOTES.md
              - TESTS.md
              - BENCHMARKS.md
              - .go files with: // NOTE: Any changes to this file must be reflected in the corresponding specs.md or NOTES.md.

              For EACH module found:
              1. List the directory path
              2. List which spec files are present
              3. Read SPECS.md and extract every stated invariant, contract, and behavioral guarantee
              4. Read NOTES.md and extract key design decisions
              5. Count the number of .go files in the module

              Write findings to .bob/state/audit-discovery.md with this format:

              # Audit Discovery

              ## Module: `path/to/module/`

              **Spec files:** SPECS.md, NOTES.md, TESTS.md, BENCHMARKS.md
              **Go files:** N

              ### Invariants from SPECS.md
              1. [invariant text]
              2. [invariant text]
              ...

              ### Design Decisions from NOTES.md
              - Decision N: [summary]
              ...

              [Repeat for each module]

              ## Summary
              - Total modules: N
              - Total invariants to verify: N")
```

**Output:** `.bob/state/audit-discovery.md`

---

## Phase 3: AUDIT

**Goal:** For each module, verify that the code satisfies every stated invariant.

**Actions:**

Read `.bob/state/audit-discovery.md` to get the list of modules and their invariants.

For each module (or batch modules into a single agent if there are few), spawn an Explore agent:

```
Task(subagent_type: "Explore",
     description: "Audit module path/to/module against its SPECS.md invariants",
     run_in_background: true,
     prompt: "You are auditing whether code satisfies its stated spec invariants.

              Module: path/to/module/

              INVARIANTS TO VERIFY (from SPECS.md):
              1. [invariant text]
              2. [invariant text]
              ...

              DESIGN DECISIONS TO CHECK (from NOTES.md):
              - Decision N: [summary]
              ...

              For EACH invariant:
              1. Read the relevant code thoroughly
              2. Determine: does the code actually satisfy this guarantee?
              3. Record your finding as one of:
                 - PASS: Code satisfies the invariant. [Brief evidence]
                 - FAIL: Code violates the invariant. [Specific violation with file:line]
                 - PARTIAL: Code satisfies the invariant in most cases but has gaps. [Which cases fail]
                 - UNTESTABLE: Cannot determine from static analysis alone. [Why]

              For EACH design decision:
              1. Check whether the codebase still follows the decision
              2. Flag any code that contradicts a stated decision

              Also check for:
              - Code behaviors that have no corresponding invariant in SPECS.md (undocumented guarantees)
              - Invariants in SPECS.md that reference types/functions that no longer exist (stale specs)
              - NOTES.md decisions that reference removed or renamed code

              Write findings to .bob/state/audit-module-<module-name>.md with this format:

              # Audit: `path/to/module/`

              ## Invariant Verification

              | # | Invariant | Verdict | Evidence |
              |---|-----------|---------|----------|
              | 1 | [text] | PASS/FAIL/PARTIAL/UNTESTABLE | [brief evidence or file:line] |
              | 2 | [text] | ... | ... |

              ## Design Decision Compliance

              | Decision | Status | Notes |
              |----------|--------|-------|
              | N: [summary] | FOLLOWS/VIOLATES | [evidence] |

              ## Undocumented Behaviors
              [Code guarantees not captured in SPECS.md]

              ## Stale Spec References
              [Invariants or decisions referencing removed/renamed code]

              ## Summary
              - Invariants verified: N
              - PASS: N
              - FAIL: N
              - PARTIAL: N
              - UNTESTABLE: N")
```

If there are multiple modules, spawn agents in parallel (one per module or batched sensibly).

**Output:** `.bob/state/audit-module-*.md` (one per module)

---

## Phase 4: REPORT

**Goal:** Consolidate all audit findings into a single report.

**Actions:**

Read all `.bob/state/audit-module-*.md` files and write a consolidated report:

```
Task(subagent_type: "Explore",
     description: "Consolidate audit findings into final report",
     run_in_background: true,
     prompt: "Read all .bob/state/audit-module-*.md files and consolidate into a single report.

              Write to .bob/state/audit-report.md:

              # Spec Audit Report

              Generated: [ISO timestamp]

              ## Executive Summary

              - Modules audited: N
              - Total invariants verified: N
              - PASS: N | FAIL: N | PARTIAL: N | UNTESTABLE: N
              - Design decisions checked: N (N violations)
              - Stale spec references found: N

              ## Overall Health: [HEALTHY / NEEDS ATTENTION / CRITICAL]

              Use HEALTHY if no FAIL findings.
              Use NEEDS ATTENTION if any FAIL or multiple PARTIAL findings.
              Use CRITICAL if >30% of invariants FAIL.

              ---

              ## Failures (Code Violates Spec)

              [List every FAIL finding across all modules, grouped by module]

              ### `path/to/module/`

              **Invariant N:** [text]
              **Violation:** [what the code does wrong, file:line]
              **Impact:** [what could go wrong because of this]

              ---

              ## Partial Compliance

              [List every PARTIAL finding]

              ---

              ## Stale Specs

              [List every stale reference — invariants about removed code, decisions about renamed types]

              ---

              ## Undocumented Behaviors

              [List behaviors found in code but not captured in specs — candidates for new invariants]

              ---

              ## Recommendations

              For each FAIL:
              - Fix the code to satisfy the invariant, OR
              - Update SPECS.md if the invariant is wrong (add NOTES.md entry explaining why)

              For each stale reference:
              - Remove or update the spec entry

              For each undocumented behavior:
              - Consider adding to SPECS.md if it's a guarantee callers rely on")
```

**Output:** `.bob/state/audit-report.md`

---

## Phase 5: COMPLETE

**Goal:** Present the audit results to the user.

**Actions:**

Read `.bob/state/audit-report.md` and present the summary:

```
Spec audit complete!

Results: .bob/state/audit-report.md

[Paste the Executive Summary section]

[If HEALTHY]: All invariants verified. Specs and code are in sync.

[If NEEDS ATTENTION]: Found N issues where code diverges from specs.
  Review .bob/state/audit-report.md for details.
  Fix the code or update the specs — use /bob:work-agents for code fixes
  or edit SPECS.md directly for spec corrections.

[If CRITICAL]: Significant spec drift detected. N invariants violated.
  Review .bob/state/audit-report.md and prioritize fixes.
```

---

## Notes

- This workflow is **read-only** — it never modifies code or spec files
- Findings from the audit can feed into `/bob:work-agents` to fix violations
- Run periodically (e.g., before releases) to catch spec drift
- The audit checks code against specs, not specs against code — if you want to generate specs from code, use `/bob:design apply`
