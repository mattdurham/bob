---
name: bob:project
description: Project initialization - INIT → DISCOVER → QUESTION → RESEARCH → DEFINE → COMPLETE
user-invocable: true
category: workflow
---

# Project Initialization Workflow

You are orchestrating **project initialization**. You guide the user through structured discovery, questioning, and planning to create a `.bob/planning/` directory with persistent project context.

**Attribution:** This workflow is inspired by the [Get Shit Done (GSD)](https://github.com/gsd-build/get-shit-done) project's `.planning/` structure and project initialization approach. Credit to the GSD team for pioneering persistent planning artifacts for AI-assisted development.

## Workflow Diagram

```
INIT → DISCOVER → QUESTION → RESEARCH → DEFINE → COMPLETE
```

**Output:** A `.bob/planning/` directory containing PROJECT.md and REQUIREMENTS.md (plus CODEBASE.md and RESEARCH.md when applicable).

---

## Execution Rules

**CRITICAL: All subagents MUST run in background**

- ✅ **ALWAYS use `run_in_background: true`** for ALL Task calls
- ✅ **After spawning agents, STOP** - do not poll or check status
- ✅ **Wait for agent completion notification** - you'll be notified automatically
- ❌ **Never use foreground execution** - it blocks the workflow

**Interactive phases (QUESTION) are the exception** - those require user input.

---

## Phase 1: INIT

**Goal:** Understand what the user wants to build and detect project state.

1. Greet the user
2. Accept project description (from command argument or ask)
3. Check for existing `.bob/planning/` directory:
   - If exists: offer to continue/reset
   - If not: create it

```bash
mkdir -p .bob/state .bob/planning
```

4. Detect brownfield vs greenfield:
   - Check if source code already exists (src/, lib/, cmd/, internal/, etc.)
   - If brownfield: note that DISCOVER phase will map the codebase

Move to DISCOVER.

---

## Phase 2: DISCOVER

**Goal:** Understand the existing codebase (brownfield) or skip (greenfield).

### Greenfield (no existing code)

Skip to QUESTION.

### Brownfield (existing code)

Spawn Explore agent to map the codebase:

```
Task(subagent_type: "Explore",
     description: "Map existing codebase",
     run_in_background: true,
     prompt: "Thoroughly explore this codebase and document:
             1. Technology stack (languages, frameworks, dependencies)
             2. Architecture patterns (how code is organized)
             3. Directory structure (key directories and their purpose)
             4. Conventions (naming, testing patterns, code style)
             5. Integrations (external services, APIs, databases)
             6. Concerns (technical debt, missing tests, inconsistencies)

             Write findings to .bob/planning/CODEBASE.md with clear sections for each area.
             Be thorough - this document is the foundation for all planning.")
```

**Output:** `.bob/planning/CODEBASE.md`

Move to QUESTION.

---

## Phase 3: QUESTION

**Goal:** Deep structured questioning to capture project intent.

This is the **most important phase**. Ask questions ONE AT A TIME using AskUserQuestion. Do NOT dump a wall of questions.

### Questioning Sequence

**Round 1: Vision & Motivation**
- What is this project? (one-sentence description)
- What problem does it solve?
- Who is it for? (target users/audience)
- Why build it now?

**Round 2: Success Criteria**
- What does "done" look like for v1?
- How will you measure success?
- What's the minimum viable scope?

**Round 3: Technical Decisions**
- Preferred language/framework? (or detect from brownfield)
- Any architectural preferences? (monolith, microservices, serverless, etc.)
- What should NOT be built? (explicit exclusions)
- Any hard constraints? (time, budget, infrastructure, dependencies)

**Round 4: Scope Boundaries**
- What features are in scope for v1?
- What features are explicitly out of scope?
- Any non-negotiable requirements?

### Questioning Rules

- Ask **one question at a time** using AskUserQuestion
- Provide **sensible defaults** as options where possible
- **Challenge vague answers** — "fast" means what? "scalable" to what level?
- **Surface assumptions** — make implicit things explicit
- If the user says "you decide" on technical choices, make a recommendation and confirm
- Skip questions that are already answered by the project description or brownfield analysis

### After Questioning

Write `.bob/planning/PROJECT.md`:

```markdown
# Project: [Name]

## Vision
[One paragraph describing what this is and why it matters]

## Problem Statement
[What problem this solves and for whom]

## Target Users
[Who will use this and in what context]

## Success Criteria
[Measurable outcomes for v1]

## Scope
### In Scope
- [Feature/capability 1]
- [Feature/capability 2]

### Out of Scope
- [Explicitly excluded item 1]
- [Explicitly excluded item 2]

## Technical Decisions
- **Language/Framework:** [choice]
- **Architecture:** [choice]
- **Key Libraries:** [choices]

## Constraints
- [Constraint 1]
- [Constraint 2]

## Assumptions
- [Assumption 1]
- [Assumption 2]
```

Move to RESEARCH.

---

## Phase 4: RESEARCH

**Goal:** Research unfamiliar technologies or patterns before defining requirements.

Evaluate whether research is needed:
- User chose an unfamiliar stack → research it
- Complex domain (auth, payments, real-time) → research patterns
- Brownfield with unusual architecture → research integration approaches
- Simple/familiar stack → skip to DEFINE

If research is needed, spawn Explore agent:

```
Task(subagent_type: "Explore",
     description: "Research technology patterns",
     run_in_background: true,
     prompt: "Research the following for project planning:
             [specific technologies/patterns to research]

             Focus on:
             1. Established architecture patterns for this stack
             2. Standard library ecosystem (don't reinvent wheels)
             3. Common pitfalls and how to avoid them
             4. Best practices for [specific concern]

             Read .bob/planning/PROJECT.md for full context.
             Write findings to .bob/planning/RESEARCH.md.
             Be practical - focus on what helps implementation decisions.")
```

**Output:** `.bob/planning/RESEARCH.md` (optional)

Move to DEFINE.

---

## Phase 5: DEFINE

**Goal:** Convert project vision into traceable requirements.

Read `.bob/planning/PROJECT.md` (and `.bob/planning/RESEARCH.md` if it exists).

Write `.bob/planning/REQUIREMENTS.md`:

```markdown
# Requirements

## Functional Requirements

### REQ-001: [Short title]
**Description:** [What the system must do]
**Acceptance Criteria:**
- [ ] [Measurable criterion 1]
- [ ] [Measurable criterion 2]
**Priority:** Must-have | Should-have | Nice-to-have

### REQ-002: [Short title]
...

## Non-Functional Requirements

### NFR-001: [Short title]
**Description:** [Quality attribute]
**Acceptance Criteria:**
- [ ] [Measurable criterion]
**Priority:** Must-have | Should-have | Nice-to-have
```

### Requirements Rules

- Every requirement gets a **REQ-ID** (REQ-001, NFR-001, etc.)
- Every requirement has **acceptance criteria** (testable, measurable)
- Prioritize using **MoSCoW** (Must/Should/Could/Won't)
- Keep requirements **atomic** — one thing per requirement
- Must-have requirements define the **minimum viable scope**

Present requirements summary to user and confirm before moving on.

Move to COMPLETE.

---

## Phase 6: COMPLETE

**Goal:** Present the project plan and next steps.

Summarize what was created:
- `.bob/planning/PROJECT.md` — Project vision, scope, and decisions
- `.bob/planning/CODEBASE.md` — Existing codebase analysis (brownfield only)
- `.bob/planning/RESEARCH.md` — Technology research (if needed)
- `.bob/planning/REQUIREMENTS.md` — Traceable requirements with REQ-IDs

Suggest next steps:

```
Next steps:
  1. Review .bob/planning/ files and adjust if needed
  2. Start work with: /bob:work "description of what to build"
  3. The work workflow will automatically use .bob/planning/ context
```

---

## Integration with Bob Workflows

The `.bob/planning/` directory integrates with other Bob workflows:

- **`/bob:work`**: Reads PROJECT.md and REQUIREMENTS.md for context during brainstorm and planning
- **`/bob:explore`**: References CODEBASE.md for exploration starting points
- **`/bob:code-review`**: Checks implementations against REQUIREMENTS.md acceptance criteria
- **`/brainstorming`**: Uses PROJECT.md to skip already-answered questions and ground ideation

---

## File Summary

| File | Created In | Purpose |
|------|-----------|---------|
| `.bob/planning/PROJECT.md` | QUESTION | Vision, scope, decisions |
| `.bob/planning/CODEBASE.md` | DISCOVER | Existing code analysis (brownfield) |
| `.bob/planning/RESEARCH.md` | RESEARCH | Technology research (optional) |
| `.bob/planning/REQUIREMENTS.md` | DEFINE | Traceable requirements |
