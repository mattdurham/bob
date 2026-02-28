---
name: bob:explore
description: Read-only codebase exploration - DISCOVER → ANALYZE → DOCUMENT
user-invocable: true
category: workflow
---

# Codebase Exploration Workflow

You orchestrate **read-only exploration** to understand codebase structure and functionality.

## Workflow Diagram

```
INIT → DISCOVER → ANALYZE → DOCUMENT → COMPLETE
```

**Read-only:** No code changes, no commits.

---

## Execution Rules

**CRITICAL: All subagents MUST run in background**

- ✅ **ALWAYS use `run_in_background: true`** for ALL Task calls
- ✅ **After spawning agents, STOP** - do not poll or check status
- ✅ **Wait for agent completion notification** - you'll be notified automatically
- ❌ **Never use foreground execution** - it blocks the workflow

---

## Phase 1: INIT

Understand exploration goal:
- What to explore?
- Specific feature/component?
- Architecture overview?

Create .bob/:
```bash
mkdir -p .bob/state
```

---

## Phase 2: DISCOVER

**Goal:** Find relevant code and understand its contracts

Spawn Explore agent:
```
Task(subagent_type: "Explore",
     description: "Discover codebase structure",
     run_in_background: true,
     prompt: "Find code related to [exploration goal].
             Map file structure, key components, relationships.

             SPEC-DRIVEN MODULES: For every directory you encounter, check for
             SPECS.md, NOTES.md, TESTS.md, BENCHMARKS.md, or .go files containing:
               // NOTE: Any changes to this file must be reflected in the corresponding SPECS.md or NOTES.md.
             If found, this is a spec-driven module. Read SPECS.md FIRST — it is the
             authoritative contract for the module's behavior, invariants, and public API.
             Read NOTES.md for design decisions and rationale. These documents take
             priority over reading implementation code for understanding what the module
             does and why.

             Write findings to .bob/state/discovery.md.
             For spec-driven modules, include a section summarizing the contracts and
             key design decisions from the spec docs.")
```

**Output:** `.bob/state/discovery.md`

---

## Phase 3: ANALYZE

**Goal:** Understand how code works — specs first, then implementation

Spawn researcher:
```
Task(subagent_type: "researcher",
     description: "Analyze codebase",
     run_in_background: true,
     prompt: "Read files in .bob/state/discovery.md.
             Understand logic, patterns, architecture.

             For any spec-driven modules identified in discovery, analyze the
             implementation THROUGH the lens of the specs: does the code match
             its contracts? Are invariants maintained? Note any drift between
             SPECS.md and the actual implementation.

             Write analysis to .bob/state/analysis.md.")
```

**Input:** `.bob/state/discovery.md`
**Output:** `.bob/state/analysis.md`

---

## Phase 4: DOCUMENT

**Goal:** Create clear documentation

Create comprehensive report in `.bob/state/exploration-report.md`:
- Overview of what was explored
- Architecture and structure
- Key components explained
- Flow diagrams (ASCII)
- Code examples
- Patterns observed
- Important files
- Questions/TODOs

---

## Phase 5: COMPLETE

Present findings to user:
- Summarize learnings
- Show key insights
- Point to detailed docs

**Next steps:**
- Explore deeper?
- Related areas?
- Start implementation? (switch to /work)
