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

Create bots/:
```bash
mkdir -p bots
```

---

## Phase 2: DISCOVER

**Goal:** Find relevant code

Spawn Explore agent:
```
Task(subagent_type: "Explore",
     description: "Discover codebase structure",
     run_in_background: true,
     prompt: "Find code related to [exploration goal].
             Map file structure, key components, relationships.
             Write findings to bots/discovery.md.")
```

**Output:** `bots/discovery.md`

---

## Phase 3: ANALYZE

**Goal:** Understand how code works

Spawn researcher:
```
Task(subagent_type: "researcher",
     description: "Analyze codebase",
     run_in_background: true,
     prompt: "Read files in bots/discovery.md.
             Understand logic, patterns, architecture.
             Write analysis to bots/analysis.md.")
```

**Input:** `bots/discovery.md`
**Output:** `bots/analysis.md`

---

## Phase 4: DOCUMENT

**Goal:** Create clear documentation

Create comprehensive report in `bots/exploration-report.md`:
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
