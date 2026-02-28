---
name: bob:design
description: Create a lightweight spec module with a single CLAUDE.md containing numbered invariants, axioms, assumptions, and constraints
user-invocable: true
category: workflow
---

# Simple Spec Workflow

You are setting up a **simple spec module** — a lightweight pattern where each package carries
a single `CLAUDE.md` file containing numbered invariants. No SPECS.md, no NOTES.md, no TESTS.md,
no BENCHMARKS.md, no NOTE invariant on `.go` files.

## The Pattern

Each module following this pattern contains:

```
<module>/
  CLAUDE.md         # Numbered invariants, axioms, assumptions, non-obvious constraints
  *.go              # No NOTE invariant comment needed
```

**CLAUDE.md rules:**
- Keep them tidy
- They contain only numbered invariants, axioms, assumptions, and non-obvious constraints
- Never add anything trivial, ephemeral, or obviously derivable from reading the code
- NEVER include copies of the code itself

**Example CLAUDE.md:**
```markdown
# ratelimit — Invariants

1. The TokenBucket interface is the sole entry point; all callers must go through it.
2. Refill is atomic — concurrent Acquire calls never see a partially refilled bucket.
3. The package never persists state — persistence is the caller's responsibility.
4. Thread-safe only if the underlying Store is thread-safe.
5. Zero-value Config means "no limit" — callers must explicitly set limits.
```

**No NOTE invariant on .go files.** Claude Code natively loads CLAUDE.md from directories
it works in, so no per-file reminder is needed.

---

## Workflow Diagram

```
INIT → GATHER → [ANALYZE] → SCAFFOLD → COMPLETE
                (existing only)
```

---

## Execution Rules

**CRITICAL: All subagents MUST run in background**

- ALWAYS use `run_in_background: true` for ALL Task calls
- After spawning agents, STOP — do not poll or check status
- Wait for agent completion notification — you'll be notified automatically
- Never use foreground execution — it blocks the workflow

---

## Phase 1: INIT

**Goal:** Understand what the user wants to do.

Greet the user and ask which mode they want using `AskUserQuestion`:

```
Which would you like to do?

1. New module — create a simple spec module from scratch
   (I'll ask what you want to build and create a CLAUDE.md with numbered invariants)

2. Apply to existing — add CLAUDE.md to an existing implementation
   (Point me at the directory and I'll generate invariants from the code)
```

---

## Phase 2: GATHER

**Goal:** Collect everything needed to scaffold or apply.

### Mode: New Module

Ask the user ONE QUESTION AT A TIME using `AskUserQuestion`:

1. **What are you building?**
   Get a clear description: purpose, what it owns (responsibility boundary), what it
   explicitly does NOT own. Push back on vague answers.

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
simple-new

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
simple-apply

## Target Path
<path>
```

---

## Phase 3: ANALYZE (Apply mode only)

**Goal:** Understand the existing implementation deeply enough to extract precise invariants.

Spawn an Explore agent:

```
Task(subagent_type: "Explore",
     description: "Analyze existing implementation for invariant extraction",
     run_in_background: true,
     prompt: "Analyze the Go package at the path in .bob/state/design-context.md.

              Extract the non-obvious properties of this code:
              - Invariants: what must always be true (sorting guarantees, nil-safety, idempotency)
              - Axioms: foundational assumptions the code relies on
              - Assumptions: what the code assumes about its callers or dependencies
              - Constraints: non-obvious limitations (thread-safety conditions, capacity limits)
              - Responsibility boundary: what this package owns vs delegates

              Focus on things a maintainer needs to know that are NOT obvious from reading
              the code. Skip anything trivial or derivable.

              Write findings to .bob/state/design-analysis.md with:
              ## Module Purpose (one sentence)
              ## Responsibility Boundary (owns / does not own)
              ## Invariants (numbered list)
              ## Assumptions (numbered list)
              ## Constraints (numbered list)")
```

**Output:** `.bob/state/design-analysis.md`

---

## Phase 4: SCAFFOLD

**Goal:** Create the CLAUDE.md file.

### For New Module

Spawn a workflow-implementer agent:

```
Task(subagent_type: "workflow-implementer",
     description: "Scaffold simple spec module",
     run_in_background: true,
     prompt: "Create a simple spec module. Context in .bob/state/design-context.md.

              Create the module directory if it doesn't exist.

              Create CLAUDE.md at the module path with this format:

              # <PackageName> — Invariants

              1. <invariant — one sentence stating something that must always be true>
              2. <axiom — a foundational assumption>
              ...

              Rules for CLAUDE.md content:
              - Only numbered invariants, axioms, assumptions, and non-obvious constraints
              - Never add anything trivial, ephemeral, or obviously derivable from reading the code
              - NEVER include copies of the code itself
              - Each item is one clear sentence
              - Be specific: 'sorted ascending by score' not 'sorted'
              - Be honest: if thread-safety depends on a condition, state the condition

              Generate realistic, specific invariants from the description and responsibility
              boundary in .bob/state/design-context.md. Aim for 3-8 items.

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

              2. <package>_test.go — test stub:
                 ```go
                 package <name>_test

                 import (
                     \"testing\"
                 )

                 // TODO: Implement tests
                 ```")
```

### For Apply Mode

Spawn workflow-implementer with the analysis:

```
Task(subagent_type: "workflow-implementer",
     description: "Apply simple spec to existing implementation",
     run_in_background: true,
     prompt: "Apply simple spec structure to an existing package.
              Context in .bob/state/design-context.md.
              Analysis in .bob/state/design-analysis.md.

              Do NOT modify any existing .go implementation logic.

              Create CLAUDE.md at the target path with this format:

              # <PackageName> — Invariants

              1. <invariant>
              2. <axiom>
              ...

              Rules for CLAUDE.md content:
              - Only numbered invariants, axioms, assumptions, and non-obvious constraints
              - Never add anything trivial, ephemeral, or obviously derivable from reading the code
              - NEVER include copies of the code itself
              - Each item is one clear sentence
              - Derive items from .bob/state/design-analysis.md

              Focus on what a maintainer needs to know. Skip what's obvious from reading
              the code. Aim for 3-12 items depending on complexity.")
```

---

## Phase 5: COMPLETE

**Goal:** Show what was created and explain how it works.

After the scaffold agent completes, present a summary:

```
Simple spec module ready!

Created:
  ✓ <path>/CLAUDE.md — Numbered invariants and constraints

How it works:
  Claude Code automatically loads CLAUDE.md when working in this directory.
  All Bob workflow agents detect and respect the invariants.
  When code changes affect an invariant, update the numbered list.

Rules:
  - Keep it tidy — only non-obvious things
  - Never add trivial or ephemeral items
  - Never include copies of code
  - Update when invariants change, delete when they no longer apply

Next steps:
  - Review CLAUDE.md and refine the invariants
  - Use /bob:work-agents to implement (the workflow will check invariants)
```
