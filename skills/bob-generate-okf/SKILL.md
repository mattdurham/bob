---
name: bob-generate-okf
description: Analyze a codebase and interactively generate an OKF (.knowledge/) bundle — asks clarifying questions, discovers packages and decisions, then writes a complete navigable knowledge catalog.
user-invocable: true
category: documentation
---

# Generate OKF Bundle

<!-- AGENT CONDUCT: Be direct and curious. Ask sharp questions. Push back on vague answers. -->

You generate an **OKF (Open Knowledge Format) bundle** — a `.knowledge/` directory that
serves as a navigable, cross-linked knowledge catalog for a codebase.

OKF is the **catalog and navigation layer**. It references code, SPECS.md, NOTES.md, and
other spec-driven docs — it does not duplicate or replace them.

**Reference material:**

- [OKF SPEC](https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md) — the underlying format definition
- [Bob OKF Convention](docs/OKF.md) — Bob's bundle structure, types, frontmatter conventions, and agent writing rules

---

## Workflow Diagram

```
INIT → DISCOVER → INTERVIEW → GENERATE → REVIEW → COMPLETE
```

---

## Phase 1: INIT

Check whether a `.knowledge/` bundle already exists:

```bash
[ -d .knowledge ] && [ -f .knowledge/index.md ] && echo EXISTS || echo NONE
```

**If exists:** Tell the user and ask:

> A `.knowledge/` bundle already exists here. Do you want to:
>
> 1. **Extend it** — add missing concepts and update stale ones
> 2. **Regenerate it** — analyze fresh and replace everything
> 3. **Cancel**

**If not present:** Greet the user:

> I'll analyze this codebase and generate an OKF `.knowledge/` bundle — a navigable
> knowledge catalog with packages, decisions, patterns, and planned features.
>
> This takes 3 steps:
>
> 1. I'll explore the codebase structure
> 2. I'll ask you a few questions about things I can't infer from code
> 3. I'll generate the full bundle
>
> Starting discovery...

---

## Phase 2: DISCOVER

**Goal:** Map the codebase structure so you know what to ask about.

Run all discovery in parallel:

### 2a. Project type and language

```bash
# Detect project type
ls go.mod go.sum Cargo.toml package.json pyproject.toml pom.xml 2>/dev/null
cat go.mod 2>/dev/null | head -5
```

### 2b. Package/module structure

```bash
# Go: find all packages
find . -name "*.go" -not -path "./.git/*" -not -path "./vendor/*" \
  | xargs -I{} dirname {} | sort -u | head -50

# Show package declarations
grep -r "^package " --include="*.go" -l \
  | xargs grep "^// Package" 2>/dev/null | head -30
```

### 2c. Existing spec-driven docs

```bash
find . -name "SPECS.md" -o -name "NOTES.md" -o -name "CLAUDE.md" \
  | grep -v ".git" | grep -v ".knowledge" | sort
```

### 2d. Existing features / planned work

```bash
# Git log for recent work
git log --oneline -20 2>/dev/null

# Open issues / TODO files
find . -name "TODO*" -o -name "ROADMAP*" -o -name "CHANGELOG*" \
  | grep -v ".git" | head -10
```

### 2e. Cross-cutting decisions visible in code

```bash
# Architecture docs
find . -name "ARCHITECTURE*" -o -name "ADR*" -o -name "DECISIONS*" \
  | grep -v ".git" | head -10
# README
cat README.md 2>/dev/null | head -60
```

### 2f. Existing .knowledge/ (extend mode)

```bash
# If extending:
cat .knowledge/index.md 2>/dev/null
ls .knowledge/*/index.md 2>/dev/null | xargs cat 2>/dev/null
```

Write a discovery summary to `.bob/state/okf-discovery.md`:

```markdown
## Project
- Language: <go|rust|typescript|…>
- Module/package root: <go module path or equivalent>
- Top-level packages: [list with one-line description from package doc or README]

## Spec-driven docs found
- [list of SPECS.md, NOTES.md paths]

## Recent work (from git log)
- [list of commit subjects — potential features]

## Cross-cutting decisions visible
- [list of architecture docs, or patterns inferred from code structure]

## Gaps (things I could not infer)
- [list of questions I need to ask]
```

---

## Phase 3: INTERVIEW

**Goal:** Fill gaps that cannot be inferred from code. Ask ONE question at a time.
Be concrete — show what you found before asking about it.

Ask these questions, **skipping any you already answered from discovery**:

### Q1 — Project purpose

> Here's what I found: `<module path>`, `<N> packages`, README says: `"<excerpt>"`.
>
> In one or two sentences: what is this project for and who uses it?

### Q2 — Package groupings

> I found these top-level packages:
>
> ```
> <list>
> ```
>
> Are there natural groupings — e.g., "these 3 are all part of the ingestion pipeline",
> "these are internal utilities"? Or should each package be cataloged individually?

### Q3 — Planned or in-flight features

> From recent commits I can see:
>
> ```
> <last 10 commits>
> ```
>
> Are there any **planned or in-progress features** that aren't in the code yet?
> (These become Feature concepts in `.knowledge/features/`.)
> If none, just say "none" and I'll skip Features.

### Q4 — Cross-cutting decisions

> I found `<N>` architecture/ADR docs. Are there any **cross-cutting architectural decisions**
> that every contributor should know? For example:
>
> - "We never cache at the service layer"
> - "All state is owned by the storage package"
> - "This is stateless by design — callers own persistence"
>
> List them briefly, or say "none".

### Q5 — Recurring playbooks

> Are there any **operational playbooks** worth cataloging? E.g.:
>
> - How to triage a specific alert
> - How to roll back a deploy
> - How to handle a migration safely
>
> If none, say "none".

### Q6 — Reusable patterns

> Are there any **code patterns** you'd want new contributors to follow — patterns that
> aren't obvious from reading one file? E.g.:
>
> - "Always use X interface for dependency injection"
> - "Use the context pattern for request-scoped data"
>
> If none, say "none".

After all questions, write answers to `.bob/state/okf-interview.md`:

```markdown
## Project Purpose
<answer>

## Package Groupings
<answer>

## Planned Features
<list or "none">

## Cross-cutting Decisions
<list or "none">

## Playbooks
<list or "none">

## Patterns
<list or "none">
```

---

## Phase 4: GENERATE

**Goal:** Write the complete `.knowledge/` bundle.

### 4a. Create directory structure

```bash
mkdir -p .knowledge/{features,packages,decisions,playbooks,patterns}
```

### 4b. Generate all concepts

Read `.bob/state/okf-discovery.md` and `.bob/state/okf-interview.md` before writing.

For each concept type below, create the files, then update the subdirectory `index.md`
and the root `log.md`.

---

#### Packages

For every package identified in discovery (or grouping from interview):

**File:** `.knowledge/packages/<slug>.md`

```markdown
---
type: Go Package
title: <package name>
description: <one-sentence description — from package doc comment or inferred>
resource: ./<relative/path>
tags: [<relevant tags>]
timestamp: <ISO 8601 now>
---

<2–3 sentence summary: what this package does, why it exists, what it explicitly does NOT own.>

# Specification

<Link to SPECS.md if it exists:>
- [Contracts and invariants](<relative-path-to-SPECS.md>)
- [Design decisions](<relative-path-to-NOTES.md>)
- [Test plan](<relative-path-to-TESTS.md>)

<If no spec docs: "No spec documents yet. Run /bob:design to scaffold them.">

# Key Interfaces

<List exported interfaces, each linked if an interface concept exists:>
- `<InterfaceName>` — <one-line purpose>

# Cross-cutting Decisions

<Links to any decision concepts that apply to this package.>

# Dependencies

<Other packages/modules this package imports — list internal ones>

# Usage Patterns

_Populated as workflows run._
```

**Format rules for packages:**

- Use the package's own doc comment (`// Package <name> ...`) as the description if present
- `resource` must be a relative path from the repo root
- Tags should include language keywords from the package purpose (e.g. `[ratelimit, concurrency, stateless]`)

---

#### Decisions

For each cross-cutting decision from the interview (plus any inferred from code structure):

**File:** `.knowledge/decisions/<slug>.md`

```markdown
---
type: Decision
title: <short title — imperative or noun phrase>
description: <one-sentence summary>
tags: [<relevant packages and topics>]
timestamp: <ISO 8601 now>
---

**Decision:** <clear statement of what was decided>

**Rationale:** <why — what coupling it avoids, what it enables, what alternative was rejected>

**Consequence:** <what follows from this — what callers must/must not do>

# Applies To

<Links to package concepts this decision governs.>

# Origin

<Where this was first established — commit, PR, ADR, or "inferred from code structure".>
```

---

#### Features

For each planned or in-progress feature from the interview:

**File:** `.knowledge/features/<YYYY-MM-DD>-<slug>.md`
(Use today's date: `2026-06-30`)

```markdown
---
type: Feature
title: <feature title>
description: <one-sentence description>
status: planned
tags: [<relevant tags>]
timestamp: <ISO 8601 now>
branch: ""
pr: ""
started: ""
completed: ""
---

# Prompt

<The task as the user described it. This is what /bob:work will receive as its task.>

# Scope

## Packages Affected

<Links to package concepts that will be touched.>

## Decisions to Respect

<Links to decision concepts that constrain the implementation.>

# Acceptance Criteria

<Bullet list of measurable outcomes from the user's description.
Push for specifics: coverage %, latency targets, interface shape.>

# Workflow

_Filled in when work begins._
```

**Only create Feature concepts for things the user explicitly said are planned.**
Do not invent features from git log unless the user confirmed them.

---

#### Playbooks

For each playbook from the interview:

**File:** `.knowledge/playbooks/<slug>.md`

```markdown
---
type: Playbook
title: <title>
description: <one-sentence description>
tags: [<relevant tags>]
timestamp: <ISO 8601 now>
---

<Context: when does this playbook apply? What problem does it solve?>

# Steps

1. <step 1>
2. <step 2>
…

# Related

<Links to packages or decisions relevant to this playbook.>
```

---

#### Patterns

For each pattern from the interview:

**File:** `.knowledge/patterns/<slug>.md`

```markdown
---
type: Pattern
title: <title>
description: <one-sentence description>
tags: [<relevant tags>]
timestamp: <ISO 8601 now>
---

<When to use this pattern and why.>

# Example

```<lang>
<concise code example>
```

# Applies To

<Links to packages where this pattern is used.>
```

---

### 4c. Generate index files

#### `.knowledge/packages/index.md`

```markdown
# Packages

* [<Title>](<filename>.md) — <description>
…
```

#### `.knowledge/features/index.md`

```markdown
# Features

* [<Title>](<filename>.md) — <description> (`status: <status>`)
…
```

(Or: `_No features cataloged yet._`)

#### `.knowledge/decisions/index.md`

```markdown
# Decisions

* [<Title>](<filename>.md) — <description>
…
```

(Or: `_No decisions cataloged yet._`)

#### `.knowledge/playbooks/index.md`

```markdown
# Playbooks

* [<Title>](<filename>.md) — <description>
…
```

(Or: `_No playbooks cataloged yet._`)

#### `.knowledge/patterns/index.md`

```markdown
# Patterns

* [<Title>](<filename>.md) — <description>
…
```

(Or: `_No patterns cataloged yet._`)

---

### 4d. Root index

**`.knowledge/index.md`** — this is the entry point. Use **progressive disclosure**:

```markdown
# <Project Name> — Knowledge

This is the project knowledge catalog. Start here, then navigate to the relevant section.

## Sections

* [Packages](packages/index.md) — Go package catalog (<N> packages)
* [Features](features/index.md) — Planned and completed work (<N> features)
* [Decisions](decisions/index.md) — Cross-cutting architectural decisions (<N> decisions)
* [Playbooks](playbooks/index.md) — How to handle recurring situations (<N> playbooks)
* [Patterns](patterns/index.md) — Reusable code patterns (<N> patterns)

## Quick Reference

<3–5 bullet points: the most important things to know about this codebase.
Link to the most important package and decision concepts.>
```

---

### 4e. Log

**`.knowledge/log.md`**:

```markdown
# Knowledge Log

## 2026-06-30
* **Initialization**: Generated .knowledge/ bundle with <N> packages, <N> decisions, <N> features, <N> playbooks, <N> patterns.
```

---

## Phase 5: REVIEW

After generating, show the user a summary:

```
OKF bundle generated at .knowledge/

Created:
  ✓ packages/     — <N> package concepts
  ✓ decisions/    — <N> decision concepts
  ✓ features/     — <N> feature concepts
  ✓ playbooks/    — <N> playbook concepts
  ✓ patterns/     — <N> pattern concepts
  ✓ index.md      — project knowledge index
  ✓ log.md        — chronological update history

Key concepts:
  - <link to 2–3 most important package concepts>
  - <link to 1–2 most important decision concepts>
  - <link to any in-progress feature concepts>
```

Then ask:

> Does this look right? A few things to check:
>
> - Are any packages missing or mislabeled?
> - Are there decisions I missed?
> - Any features that should be added?
>
> Say **done** to finish, or tell me what to fix.

Make any requested corrections. Repeat until the user says done.

---

## Phase 6: COMPLETE

```
✅ OKF bundle complete!

Your .knowledge/ catalog is ready. Here's how Bob uses it:

• /bob:work reads .knowledge/features/ for planned work — drop a Feature concept there
  to give the workflow a full prompt, scope, and acceptance criteria.

• Agents read .knowledge/ before starting work — packages, decisions, and links to
  SPECS.md guide implementation without re-discovering the codebase from scratch.

• After each /bob:work run, the COMPLETE phase enriches the bundle automatically —
  updating Feature status, adding new Decisions, linking Packages Changed.

• To add a new feature to the catalog:
  Create .knowledge/features/YYYY-MM-DD-<slug>.md following the Feature format,
  then run /bob:work and point it at the feature file.

OKF reference:
  https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md
  docs/OKF.md (Bob's convention doc)
```

---

## Writing Rules

When generating concepts, follow these rules from `docs/OKF.md`:

**Signal-to-noise:** Write when there is something worth preserving, not as a log of activity.

**Promotion threshold for packages:**

- Include a package concept for every package with meaningful exported API
- Skip trivially small internal helpers unless they're referenced by multiple packages

**Promotion threshold for decisions:**

- Include a decision concept for any constraint that would surprise a new contributor
- Include if it's the kind of thing that gets re-introduced and has to be reverted

**Naming conventions:**

- Dated concepts (Features, Decisions from a specific workflow): `YYYY-MM-DD-<slug>.md`
- Evergreen concepts (Packages, Patterns, Playbooks): `<slug>.md`
- Slugs: lowercase, hyphens, no special characters

**Never invent facts:** If you cannot determine the rationale for a decision from the code,
say `"Inferred from code structure. Verify with team."` rather than fabricating a rationale.
