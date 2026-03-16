---
name: talk-to-codex
description: Have a conversation with Codex about code — ask questions or instruct it to write code, looping until consensus.
user-invocable: true
category: workflow
---

# talk-to-codex — Conversational Code Collaboration with Codex

Start a multi-turn conversation with OpenAI Codex. Claude drives the conversation autonomously,
going back and forth until consensus is reached, then summarizes the outcome.

Two modes:

- **ask** (default) — Ask Codex questions about code. Read-only. Claude challenges and explores
  until both agree on an answer.
- **code** — Instruct Codex to write or modify code. Claude reviews changes and loops until
  both agents are satisfied with the result.

## Usage

Invoked as `/talk-to-codex [mode] <prompt>`.

**Arguments from the user are passed after the skill name.** Parse ARGUMENTS to extract:
- `[mode]` — optional, either `ask` or `code`. If omitted, infer from intent:
  - Questions, "how does X work", "why does Y" → **ask**
  - "implement X", "refactor Y", "add Z", "fix W" → **code**
- `<prompt>` — the question or instructions for Codex

If no prompt is provided, ask the user what they want to discuss.

## Resolving the Codebase

Determine which codebase Codex should have access to:

1. **Default** — use Claude's current working directory. This is the common case.
2. **Different codebase** — if the user references a project by name that doesn't match the
   current working directory, use the `/ref` skill's `projects.json` registry at
   `~/.claude/skills/ref/projects.json` to resolve the path. Match using the same strategy
   as the ref skill (exact, substring, semantic).

The resolved path is passed as the `cwd` parameter to Codex.

## Running the Conversation

**This entire conversation MUST run in a subagent** to avoid polluting the main context window.
Launch an Agent with `subagent_type: "general-purpose"` and pass it the full instructions below.

---

## Ask Mode

### Subagent Instructions (ask)

#### 1. Start the Codex session

Call `mcp__codex__codex` with:
- `prompt`: The user's question, enriched with any relevant context Claude has gathered
- `cwd`: The resolved codebase path
- `sandbox`: `"read-only"`

Save the `threadId` from the response — it's needed for follow-up messages.

#### 2. Analyze and respond

Read Codex's response. Consider:
- Does the answer fully address the question?
- Are there gaps, ambiguities, or areas that need clarification?
- Does Claude's own knowledge of the codebase suggest follow-up questions?
- Are there alternative approaches worth exploring?

If the answer is complete and Claude agrees, skip to step 4.

#### 3. Continue the conversation

Call `mcp__codex__codex-reply` with:
- `threadId`: The thread ID from step 1
- `prompt`: Claude's follow-up question, counterpoint, or request for clarification

Repeat steps 2-3 until one of these conditions is met:
- **Consensus**: Both Claude and Codex agree on the answer
- **Sufficient coverage**: The topic has been thoroughly explored from multiple angles
- **Diminishing returns**: Further back-and-forth isn't adding new information

#### 4. Summarize

Produce a structured summary:

```
## Codex Conversation: <topic>

### Question
<original question>

### Consensus
<the agreed-upon answer or conclusion>

### Key Points
- <important point 1>
- <important point 2>
- ...

### Disagreements (if any)
- <area where Claude and Codex diverged, and how it was resolved>

### Rounds
<number of back-and-forth exchanges>
```

---

## Code Mode

### Subagent Instructions (code)

#### 1. Start the Codex session

Call `mcp__codex__codex` with:
- `prompt`: Clear instructions for what Codex should implement. Include any spec, plan, or
  design context that Claude has from the current conversation. Be specific about files,
  functions, expected behavior, and constraints.
- `cwd`: The resolved codebase path
- `sandbox`: `"workspace-write"`

Save the `threadId` from the response — it's needed for follow-up messages.

#### 2. Review the changes

After Codex completes its work, review what it did:
- Read the files Codex modified or created (use Read/Glob/Grep tools on the cwd)
- Check that the implementation matches the spec/instructions
- Look for correctness issues, edge cases, style problems, or missed requirements
- Verify the changes don't break existing patterns or conventions in the codebase

If the changes look good and match the spec, skip to step 4.

#### 3. Request corrections

Call `mcp__codex__codex-reply` with:
- `threadId`: The thread ID from step 1
- `prompt`: Specific feedback on what needs to change. Be precise — reference files, line
  numbers, and expected behavior. One focused correction per message is better than a list
  of ten things.

After Codex responds, go back to step 2.

Repeat until:
- **Satisfied**: Claude is happy with the implementation
- **Consensus**: Both agents agree the work is complete

#### 4. Summarize

Produce a structured summary:

```
## Codex Conversation: <topic>

### Task
<original instructions>

### Changes Made
- <file>: <what changed>
- <file>: <what changed>
- ...

### Review Status
<satisfied / concerns remaining>

### Key Decisions
- <decision 1>
- <decision 2>
- ...

### Rounds
<number of review cycles>
```

---

## Behavior Notes

- Claude should bring its own knowledge of the codebase into the conversation — don't just
  passively accept Codex's answers or code. Challenge them when appropriate.
- In **ask** mode, if Codex suggests code changes, note them in the summary but do NOT apply
  them. All changes are routed through Claude in the main conversation.
- In **code** mode, Codex writes directly to the workspace. Claude reviews but does not
  duplicate the work — the point is to let Codex do the writing.
- If Codex seems confused or stuck in a loop, stop early and report what happened.
- Keep individual messages to Codex focused — one question or point per message works better
  than walls of text.
