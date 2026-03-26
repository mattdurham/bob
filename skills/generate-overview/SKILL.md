---
name: bob:generate-overview
description: Generate a markdown overview of any codebase feature or module at three levels of depth — high level (non-technical), technical overview (mixed audience), or deep dive (developer)
user-invocable: true
category: documentation
---

# Generate Overview

You produce a **markdown document** describing a software feature, module, or system at the
requested depth level. You ask two questions first, then research and write.

---

## Phase 1: ASK LEVEL

Ask the user which level they want:

> Which level of overview do you need?
>
> **1. High Level** — 1–2 pages, no code, suitable for a presentation to non-technical stakeholders. Uses plain language and analogies.
>
> **2. Technical Overview** — 3–5 pages, light on implementation detail, suitable for a mixed technical audience (e.g. product, QA, DevOps, non-developer engineers). Covers architecture, components, and key design decisions without deep code.
>
> **3. Deep Dive** — No page limit, developer audience. Includes file references, call flows, interface contracts, invariants, design rationale, and gotchas.

Wait for the user to choose 1, 2, or 3 (or equivalent words).

---

## Phase 2: ASK SUBJECT

Ask:

> What would you like me to describe? (e.g. a package name, feature, subsystem, or concept)

Wait for the user's answer.

---

## Phase 3: DISCOVER

Research the subject using all available tools. Run these in parallel where possible.

### 3a. Find spec documents

Search for SPECS.md, NOTES.md, TESTS.md, BENCHMARKS.md files related to the subject:

```
find . -path "*/vendor" -prune -o \( -name "SPECS.md" -o -name "NOTES.md" -o -name "TESTS.md" -o -name "BENCHMARKS.md" \) -print | xargs grep -l "<subject>" 2>/dev/null
```

Read every spec file found. These are the **authoritative source** for contracts, invariants,
and design decisions — read them before reading code.

### 3b. Find source code

Locate the relevant packages and files:
- Search for the subject name in file paths and package declarations
- Read key files: interfaces, main types, entry points
- For deep dive: also read implementation files and tests

If first-mate is available:
```
first-mate parse_tree
first-mate query_nodes expr='name contains "<subject>"'
first-mate call_graph function_id="<entry_point>"
```

### 3c. Assess what you have

Before writing, note:
- What spec documents exist (SPECS / NOTES / TESTS / BENCHMARKS)?
- What are the key types, interfaces, and entry points?
- What problem does this solve?
- What are the non-obvious design decisions (from NOTES.md)?

---

## Phase 4: GENERATE

Write the markdown document according to the chosen level.

---

### Level 1 — High Level (non-technical, 1–2 pages)

**Audience:** Executives, product managers, stakeholders with no engineering background.
**Goal:** They should understand *what* it does and *why it matters*, not *how*.
**Constraints:** No code. No file paths. No jargon without explanation. Use analogies.

**Document structure:**
```markdown
# <Feature/Module Name>

## What Is It?
One clear sentence. What does this thing do?

## The Problem It Solves
2–3 sentences. What was broken or missing before this existed?

## How It Works
3–5 bullet points. Plain language. Analogies welcome.
No implementation detail — describe behaviour, not mechanism.

## Key Benefits
2–4 bullets. Concrete outcomes: speed, reliability, cost, user experience.

## Summary
1 paragraph. Reinforce the value. No new information.
```

**Length:** 1–2 pages. Stop when the story is complete.

---

### Level 2 — Technical Overview (mixed audience, 3–5 pages)

**Audience:** Product engineers, QA, DevOps, tech leads from other teams. Technical but
not necessarily familiar with this codebase or language.
**Goal:** They understand the architecture, the components, how data flows, and the key
design decisions — without needing to read the code themselves.
**Constraints:** Minimal code (at most 1–2 illustrative snippets). No deep implementation
detail. Component names and package paths are fine.

**Document structure:**
```markdown
# <Feature/Module Name> — Technical Overview

## Purpose
2–4 sentences. What problem, who uses it, what it produces.

## Architecture
ASCII diagram or component list showing the major parts and how they connect.
Name each component and give it one sentence.

## Data Flow
Step-by-step: what comes in, what happens to it, what comes out.
Numbered list or short paragraphs. Focus on the path, not the code.

## Key Design Decisions
3–6 bullets drawn from NOTES.md. Each decision: what was chosen and why.
Omit decisions that are obvious from context.

## Interfaces and Contracts
The public surface: what callers provide, what they get back.
Describe in prose or a simple table. No full type signatures needed.

## Observability / Operations
How is this monitored? What can go wrong? How is it debugged?
(Omit if not applicable.)

## Known Limitations
Honest list of what this does not do, or where it may struggle.
```

**Length:** 3–5 pages. Include an ASCII diagram if it aids understanding.

---

### Level 3 — Deep Dive (developer audience, no page limit)

**Audience:** Engineers who will work on, extend, or debug this code.
**Goal:** Complete understanding: contracts, invariants, flows, edge cases, file locations,
and the *why* behind non-obvious choices.
**Constraints:** None. Include code snippets, file:line references, call graphs, interface
definitions, benchmark targets.

**Document structure:**
```markdown
# <Feature/Module Name> — Deep Dive

## Overview
3–5 sentences. Purpose, scope, and where it fits in the larger system.

## Module Structure
List of packages/files with one-line descriptions.
Format: `path/to/file.go` — what it contains.

## Interfaces and Contracts
Full interface definitions (or excerpts) from SPECS.md.
State invariants explicitly. Note preconditions and postconditions.

## Architecture
ASCII diagram of components, packages, and their dependencies.
Call graph for the primary entry point if useful.

## Implementation Walkthrough
Walk the critical path end-to-end.
Reference specific files and line numbers: `path/to/file.go:42`
Show key code snippets where the logic is non-obvious.

## Design Decisions
Each significant decision from NOTES.md:
- What was decided
- Why (rationale)
- What was rejected and why
- Consequences / trade-offs

## Invariants and Edge Cases
List invariants from SPECS.md and any additional ones evident from the code.
Note edge cases that the implementation explicitly handles.

## Testing
Coverage from TESTS.md. Key test scenarios.
How to run: `go test ./path/...`
Any notable test helpers or fixtures.

## Performance
Benchmark targets from BENCHMARKS.md (if present).
Complexity of key operations. Known hot paths.

## Gotchas and Footguns
Non-obvious behaviours. Things that will bite you. Thread-safety notes.
Common mistakes when extending this code.

## References
- Key source files with brief descriptions
- Relevant spec documents
- Related packages / dependencies
```

**Length:** As long as needed. Do not omit detail for brevity.

---

## Phase 5: OUTPUT

- Print the complete markdown document in the response.
- If the document is long (Level 3), offer to write it to a file:
  `<subject>-overview.md` or `<subject>-deep-dive.md` in the current directory.
- State which spec documents were used as sources (or note if none were found).
- State if first-mate graph data was used or if the analysis was based on static file reads.
