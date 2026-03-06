---
name: bob:audit
description: Verify code satisfies stated invariants in CLAUDE.md modules — DISCOVER → AUDIT → REPORT
user-invocable: true
category: workflow
---

# Invariant Audit Workflow

You orchestrate a **read-only audit** that verifies code satisfies the invariants stated in CLAUDE.md files. No code changes — just a report of where reality diverges from the spec.

## When to Use

- After `/bob:design apply` to verify freshly-generated invariants match the existing code
- Periodically to catch invariant drift across many PRs
- Before major refactors to confirm your understanding of current guarantees
- As a health check on documented modules

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

1. All documented modules — scan the entire repo for modules with CLAUDE.md
2. Specific directory — audit one module (provide the path)
```

2. If the user provides a path, verify it contains a `CLAUDE.md`. If not, tell them and suggest `/bob:design` to create one first.

3. Create `.bob/state/` if it doesn't exist:
```bash
mkdir -p .bob/state
```

---

## Phase 2: DISCOVER

**Goal:** Find all documented modules in scope.

**Actions:**

Spawn an Explore agent to find documented modules:

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

**Output:** `.bob/state/audit-discovery.md`

---

## Phase 3: AUDIT

**Goal:** For each module, verify that the code satisfies every stated invariant.

**Actions:**

Read `.bob/state/audit-discovery.md` to get the list of modules and their invariants.

For each module (or batch modules into a single agent if there are few), spawn an Explore agent:

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

              # Invariant Audit Report

              Generated: [ISO timestamp]

              ## Executive Summary

              - Modules audited: N
              - Total invariants verified: N
              - PASS: N | FAIL: N | PARTIAL: N | UNTESTABLE: N
              - Stale invariant references found: N

              ## Overall Health: [HEALTHY / NEEDS ATTENTION / CRITICAL]

              Use HEALTHY if no FAIL findings.
              Use NEEDS ATTENTION if any FAIL or multiple PARTIAL findings.
              Use CRITICAL if >30% of invariants FAIL.

              ---

              ## Failures (Code Violates Invariant)

              [List every FAIL finding across all modules, grouped by module]

              ### `path/to/module/`

              **Invariant N:** [text]
              **Violation:** [what the code does wrong, file:line]
              **Impact:** [what could go wrong because of this]

              ---

              ## Partial Compliance

              [List every PARTIAL finding]

              ---

              ## Stale Invariants

              [List every stale reference — invariants about removed code or renamed types]

              ---

              ## Undocumented Behaviors

              [List behaviors found in code but not captured in CLAUDE.md]

              ---

              ## Proposed New Invariants

              These invariants were discovered by scanning the code but are NOT yet documented
              in any CLAUDE.md file. They require user approval before being added.

              ### `path/to/module/`

              | # | Proposed Invariant | Evidence | Rationale |
              |---|-------------------|----------|-----------|
              | 1 | [text] | [file:line] | [why this matters] |

              [Group by module, list all candidates]

              ---

              ## Recommendations

              For each FAIL:
              - Fix the code to satisfy the invariant, OR
              - Update CLAUDE.md if the invariant is wrong

              For each stale invariant:
              - Remove from CLAUDE.md

              For each undocumented behavior:
              - Consider adding to CLAUDE.md if it's a guarantee callers rely on")
```

**Output:** `.bob/state/audit-report.md`

---

## Phase 5: COMPLETE

**Goal:** Present the audit results to the user.

**Actions:**

Read `.bob/state/audit-report.md` and present the summary:

```
Audit complete!

Results: .bob/state/audit-report.md

[Paste the Executive Summary section]

[If HEALTHY]: All invariants verified. CLAUDE.md and code are in sync.

[If NEEDS ATTENTION]: Found N issues where code diverges from invariants.
  Review .bob/state/audit-report.md for details.
  Fix the code or update CLAUDE.md — use /bob:work-agents for code fixes.

[If CRITICAL]: Significant invariant drift detected. N invariants violated.
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
- Findings from the audit can feed into `/bob:work-agents` to fix violations
- Run periodically (e.g., before releases) to catch invariant drift
- The audit checks code against invariants, not invariants against code — if you want to generate invariants from code, use `/bob:design apply`
