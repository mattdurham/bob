# Open Knowledge Format — Bob Convention

Bob uses OKF (Open Knowledge Format) for project-level knowledge management.
The `.knowledge/` bundle at the repo root is the project's living knowledge catalog —
a navigable, cross-linked map of the project's software assets, decisions, and planned work.

OKF is the **catalog and navigation layer**. It references SPECS.md, NOTES.md, and other
spec-driven docs — it does not duplicate or replace them.

See the [OKF SPEC](https://github.com/GoogleCloudPlatform/knowledge-catalog/blob/main/okf/SPEC.md)
for the underlying format.

---

## Bundle Structure

```
.knowledge/
├── index.md                        # Project knowledge index (progressive disclosure)
├── log.md                          # Chronological update history
├── features/
│   ├── index.md
│   └── YYYY-MM-DD-<slug>.md        # type: Feature — planned/in-progress/complete
├── packages/
│   ├── index.md
│   └── <name>.md                   # type: Go Package — links to SPECS.md, NOTES.md
├── decisions/
│   ├── index.md
│   └── <slug>.md                   # type: Decision — cross-cutting architectural decisions
├── playbooks/
│   ├── index.md
│   └── <slug>.md                   # type: Playbook — how to handle recurring situations
└── patterns/
    ├── index.md
    └── <slug>.md                   # type: Pattern — reusable code patterns
```

`index.md` and `log.md` are OKF reserved files. All other `.md` files are concept documents.

---

## Types

Bob uses the following `type` values. Consumers must tolerate unknown types gracefully.

| Type | Directory | Has `resource`? | Purpose |
|------|-----------|-----------------|---------|
| `Feature` | `features/` | No | A planned, in-progress, or completed unit of work |
| `Go Package` | `packages/` | Yes — relative path | A Go package in the repo |
| `Go Interface` | `packages/interfaces/` | Yes — `./path/file.go:N` | A key exported interface |
| `Decision` | `decisions/` | No | A cross-cutting architectural or design decision |
| `Playbook` | `playbooks/` | No | How to handle a recurring situation |
| `Pattern` | `patterns/` | No | A reusable code pattern |

---

## Frontmatter Conventions

### Feature

```yaml
---
type: Feature
title: Add token bucket rate limiter
description: Configurable per-route rate limiting, safe for concurrent use.
status: planned        # planned | in-progress | complete | abandoned
tags: [ratelimit, api, concurrency]
timestamp: 2026-06-28T14:00:00Z
branch: ""             # filled in at WORKTREE phase
pr: ""                 # filled in at COMMIT
started: ""            # ISO date, filled in when work begins
completed: ""          # ISO date, filled in at COMPLETE
---
```

### Go Package

```yaml
---
type: Go Package
title: ratelimit
description: Token bucket rate limiter. Stateless; callers own persistence.
resource: ./internal/ratelimit
tags: [ratelimit, concurrency, stateless]
timestamp: 2026-06-28T14:30:00Z
---
```

### Decision

```yaml
---
type: Decision
title: No caching at service layer
description: Rate limiting holds no mutable state. Caching belongs to callers.
tags: [ratelimit, architecture, state]
timestamp: 2026-06-28T14:30:00Z
---
```

### Playbook / Pattern

```yaml
---
type: Playbook
title: Handling freshness alerts
description: Steps to triage a data freshness alert on any pipeline.
tags: [oncall, incident, data]
timestamp: 2026-06-28T14:30:00Z
---
```

---

## Feature Concept — Full Format

Feature concepts are the **unit of project work**. They exist before a workflow starts,
drive it, and get enriched as it runs. They are the persistent record of why and how
the code is the way it is.

```markdown
---
type: Feature
title: Add token bucket rate limiter
description: Configurable per-route rate limiting, safe for concurrent use.
status: planned
tags: [ratelimit, api, concurrency]
timestamp: 2026-06-28T14:00:00Z
branch: ""
pr: ""
started: ""
completed: ""
---

# Prompt

Add a token bucket rate limiter to the API layer. Configurable per-route,
supports burst capacity, thread-safe. Integrate as middleware. The limiter
must be stateless — callers own any persistence.

# Scope

## Packages Affected

- [ratelimit](/knowledge/packages/ratelimit.md) — new package to create
- [api](/knowledge/packages/api.md) — middleware integration point

## Decisions to Respect

- [No caching at service layer](/knowledge/decisions/no-shared-cache.md)

## Specifications to Reference

- [api SPECS.md](../../internal/api/SPECS.md)

# Acceptance Criteria

- Token bucket with configurable rate and burst
- Thread-safe for concurrent use  
- Integrated as middleware in the API layer
- Coverage ≥ 80%

# Workflow

_(filled in by the workflow as work progresses)_

branch: —
pr: —
started: —
completed: —

# Decisions Made

_(enriched at COMPLETE — links to Decision concepts produced during this work)_

# Packages Changed

_(enriched at COMPLETE — links to package concepts for modules modified)_
```

### Feature Lifecycle

```
created (status: planned)
    ↓  workflow picks it up
status: in-progress  +  branch set
    ↓  workflow completes
status: complete  +  pr linked  +  Decisions Made  +  Packages Changed filled in
```

---

## Go Package Concept — Full Format

```markdown
---
type: Go Package
title: ratelimit
description: Token bucket rate limiter. Stateless; callers own persistence.
resource: ./internal/ratelimit
tags: [ratelimit, concurrency, stateless]
timestamp: 2026-06-28T14:30:00Z
---

Implements token bucket rate limiting with configurable rate and burst capacity.

# Specification

- [Contracts and invariants](../../internal/ratelimit/SPECS.md)
- [Design decisions](../../internal/ratelimit/NOTES.md)
- [Test plan](../../internal/ratelimit/TESTS.md)
- [Benchmark specs](../../internal/ratelimit/BENCHMARKS.md)

# Key Interfaces

- [TokenBucket](/knowledge/packages/interfaces/token-bucket.md)

# Cross-cutting Decisions

- [No caching at service layer](/knowledge/decisions/no-shared-cache.md)

# Dependencies

- [storage](/knowledge/packages/storage.md) — dependency for persistence

# Usage Patterns

_(enriched by agents during workflows)_
```

---

## Decision Concept — Full Format

```markdown
---
type: Decision
title: No caching at service layer
description: Rate limiting holds no mutable state. Caching belongs to callers.
tags: [ratelimit, architecture, state]
timestamp: 2026-06-28T14:30:00Z
---

**Decision:** TokenBucket holds no mutable state beyond a reference to its dependency.
No caching.

**Rationale:** Caching at multiple layers complicates invalidation without benefit.
Each call is stateless and safe for concurrent use if the underlying dependency is.

**Consequence:** Callers must own any caching behavior. The package never persists state.

# Applies To

- [ratelimit](/knowledge/packages/ratelimit.md)

# Origin

First documented in [ratelimit NOTES.md §3](../../internal/ratelimit/NOTES.md#3-no-caching).
Introduced by feature [add-rate-limiting](/knowledge/features/2026-06-28-add-rate-limiting.md).
```

---

## Agent Writing Rules

Agents write to `.knowledge/` during workflows. Signal-to-noise matters — write when
there is something worth preserving, not as a log of activity.

### Promotion Threshold

| Source | Write to `.knowledge/` when… | Stay in `.bob/state/` |
|--------|------------------------------|----------------------|
| Brainstormer | A cross-cutting decision is made or an approach explicitly rejected with reasoning | Exploratory notes, codebase observations |
| Implementer | A design choice that isn't captured in NOTES.md, or that crosses module boundaries | Implementation details |
| Reviewer | A pattern that would recur, or a systemic issue | Per-task review comments |
| Workflow (COMPLETE) | Always — one Feature enrichment per completed run | `.bob/state/` files |

### After Writing a Concept

1. Append one bullet to the relevant subdirectory `index.md`:

   ```markdown
   * [Title](filename.md) - description
   ```

2. Append one entry to root `log.md`:

   ```markdown
   ## YYYY-MM-DD
   * **Creation**: Added [Title](/knowledge/subdir/filename.md).
   ```

### Naming Conventions

- Dated concepts (Features, Decisions tied to a workflow): `YYYY-MM-DD-<slug>.md`
- Evergreen concepts (Packages, Patterns, Playbooks): `<slug>.md`
- Slugs: lowercase, hyphens, no special characters

---

## Agent Reading Rules

Agents read `.knowledge/` as structured prior art before starting work.

### Progressive Disclosure Pattern

```bash
# 1. Start at the index — don't read everything
cat .knowledge/index.md

# 2. Identify relevant subdirectories from the index
cat .knowledge/packages/index.md
cat .knowledge/decisions/index.md

# 3. Traverse only relevant concepts
cat .knowledge/packages/ratelimit.md
cat .knowledge/decisions/no-shared-cache.md

# 4. Follow links to spec docs
cat internal/ratelimit/SPECS.md
cat internal/ratelimit/NOTES.md
```

Never read the entire `.knowledge/` tree upfront. Use `index.md` for progressive disclosure,
then traverse only what is relevant to the current task.

### When a Feature Concept Is Active

If a Feature concept is the source of the workflow prompt:

- Extract the `# Prompt` section as the task description
- Pre-load all linked packages and decisions as prior art
- Follow links from packages to their SPECS.md and NOTES.md
- Treat `# Acceptance Criteria` as hard constraints

---

## Workspace Detection

Bob detects the workspace mode at INIT and writes `.bob/state/workspace.md`.
All downstream agents read this file to know what context is available.

```bash
# OKF workspace
OKF=false
[ -d .knowledge ] && [ -f .knowledge/index.md ] && OKF=true

# Spec-driven workspace
SPEC_DRIVEN=false
find . -name "SPECS.md" -o -name "NOTES.md" -o -name "TESTS.md" -o -name "BENCHMARKS.md" \
  2>/dev/null | grep -q . && SPEC_DRIVEN=true
grep -rq "NOTE: Any changes to this file must be reflected" \
  --include="*.go" . 2>/dev/null && SPEC_DRIVEN=true

# Mode
if $OKF && $SPEC_DRIVEN; then MODE=full
elif $OKF;               then MODE=okf
elif $SPEC_DRIVEN;        then MODE=spec-driven
else                           MODE=none; fi
```

### `.bob/state/workspace.md` format

```markdown
---
okf: true
spec-driven: true
mode: full
feature: .knowledge/features/2026-06-28-add-rate-limiting.md
---

# Workspace Context

OKF bundle: .knowledge/ (index at .knowledge/index.md)
Spec-driven modules: [list detected directories]
Active feature: Add token bucket rate limiter
```

### Behavior Per Mode

| Mode | Brainstormer reads | Enrichment writes |
|------|-------------------|-------------------|
| `none` | Cold discovery from code | Nothing |
| `spec-driven` | SPECS.md + NOTES.md for in-scope modules | SPECS.md + NOTES.md |
| `okf` | `.knowledge/` index → package + decision concepts | OKF concepts + log.md |
| `full` | `.knowledge/` index → follows links to SPECS.md + NOTES.md | Both |

---

## Initializing a `.knowledge/` Bundle

To start a new OKF bundle in a project:

```bash
mkdir -p .knowledge/{features,packages,decisions,playbooks,patterns}
```

Create `.knowledge/index.md`:

```markdown
# [Project Name] — Knowledge

* [Features](features/index.md) - Planned and completed work
* [Packages](packages/index.md) - Go package catalog
* [Decisions](decisions/index.md) - Cross-cutting architectural decisions
* [Playbooks](playbooks/index.md) - How to handle recurring situations
* [Patterns](patterns/index.md) - Reusable code patterns
```

Create `.knowledge/log.md`:

```markdown
# Knowledge Log

## YYYY-MM-DD
* **Initialization**: Created .knowledge/ bundle.
```

Create each subdirectory `index.md` with an appropriate heading and empty list.

The `bob-design` skill creates package concepts when scaffolding new modules.
