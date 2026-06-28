---
name: bob:internal:brainstorming
description: "You MUST use this before any creative work - creating features, building components, adding functionality, or modifying behavior. Explores user intent, requirements and design before implementation."
user-invocable: false
category: internal
---

# Brainstorming Ideas Into Designs

## Overview

Help turn ideas into fully formed designs and specs through natural collaborative dialogue.

**Greeting:**
```
"Hey! Let's think through this idea together.
I'll ask you some questions to make sure we're on the right track..."
```

Start by understanding the current project context, then ask questions one at a time to refine the idea. Once you understand what you're building, present the design in small sections (200-300 words), checking after each section whether it looks right so far.

## CLAUDE.md Module Awareness

Before brainstorming, check if any directories in scope contain a `CLAUDE.md` file.

If found, these are **documented modules**. Read `CLAUDE.md` to understand the module's
numbered invariants, axioms, assumptions, and non-obvious constraints before proposing any
changes. Factor CLAUDE.md updates into any proposed approach.

**CLAUDE.md is the source of truth for invariants.** If the user's request contradicts an
existing numbered invariant in CLAUDE.md, you MUST flag the conflict and ask the user to
confirm they want to change it before proceeding. Do not silently comply — invariants can be
changed, but only deliberately.

## The Process

**Understanding the idea:**
- Check for CLAUDE.md modules in scope, then current project state (files, docs, recent commits)
- Ask questions one at a time to refine the idea
- Prefer multiple choice questions when possible, but open-ended is fine too
- Only one question per message - if a topic needs more exploration, break it into multiple questions
- Focus on understanding: purpose, constraints, success criteria

**Exploring approaches:**
- Propose 2-3 different approaches with trade-offs
- Present options conversationally with your recommendation and reasoning
- Lead with your recommended option and explain why

**Presenting the design:**
- Once you believe you understand what you're building, present the design
- Break it into sections of 200-300 words
- Ask after each section whether it looks right so far
- Cover: architecture, components, data flow, error handling, testing
- Be ready to go back and clarify if something doesn't make sense

## After the Design

**Documentation:**
- Write the validated design to `docs/plans/YYYY-MM-DD-<topic>-design.md`
- Use elements-of-style:writing-clearly-and-concisely skill if available
- Commit the design document to git

**Implementation (if continuing):**
- Use superpowers:using-git-worktrees to create isolated workspace
- Use superpowers:writing-plans to create detailed implementation plan

## Key Principles

- **One question at a time** - Don't overwhelm with multiple questions
- **Multiple choice preferred** - Easier to answer than open-ended when possible
- **YAGNI ruthlessly** - Remove unnecessary features from all designs
- **Explore alternatives** - Always propose 2-3 approaches before settling
- **Incremental validation** - Present design in sections, validate each
- **Be flexible** - Go back and clarify when something doesn't make sense
- **ALWAYS ask questions** - This skill is interactive by design. "Don't-ask mode" is a permission setting for tool approvals — it does NOT mean skip design questions. Never short-circuit brainstorming by dumping a "full recommendation" without asking the user first. The user wants to be consulted on design decisions.
