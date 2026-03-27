---
name: bob:explore
description: Team-based codebase exploration with adversarial challenge - DISCOVER → ANALYZE → CHALLENGE → DOCUMENT
user-invocable: true
category: workflow
requires_experimental: agent_teams
---

# Team Exploration Workflow (Agent Teams)

<!-- AGENT CONDUCT: Be direct and challenging. Flag gaps, risks, and weak ideas proactively. -->

You orchestrate **read-only exploration** with an adversarial **CHALLENGE** phase that stress-tests the analysis using concurrent specialist agents coordinated through a **shared task list**.

## Prerequisites

<experimental_feature>
This workflow requires the experimental agent teams feature:

```json
// Add to ~/.claude/settings.json
{
  "env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  }
}
```

Or set environment variable:
```bash
export CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1
```

Without this flag, the workflow will fail.
</experimental_feature>

## Workflow Diagram

```
INIT → DISCOVER → SPAWN TEAM → CREATE TASKS → EXECUTE (ANALYZE ↔ CHALLENGE) → DOCUMENT → COMPLETE
                                                          ↑            ↓
                                                          └────────────┘
                                                       (challenger FAILs →
                                                        re-analysis tasks)
```

**Read-only:** No code changes, no commits.

Analysts and challengers are persistent teammates that claim tasks from a shared list. Loop-back is natural — challengers create re-analysis tasks, analysts pick them up.

---

<strict_enforcement>
All phases MUST be executed in the exact order specified.
NO phases may be skipped under any circumstances.
The orchestrator MUST follow each step exactly as written.
Each phase has specific prerequisites that MUST be satisfied before proceeding.
</strict_enforcement>

## Orchestrator Boundaries

**The team lead coordinates. It never analyzes.**

**Team Lead CAN:**
- Create and manage the agent team
- Spawn teammates with specific prompts
- Create tasks using TaskCreate
- Monitor task list with TaskList
- Read files to make routing decisions
- Merge analysis files into unified documents
- Display brief status updates to the user between phases
- Clean up team when workflow complete

**Team Lead CANNOT:**
- Write or edit source code files
- Analyze code itself (that's what analyst teammates do)
- Challenge analysis itself (that's what challenger teammates do)
- Make analytical conclusions

**All analysis and challenge work MUST be performed by teammates.**

---

## Autonomous Progression Rules

**CRITICAL: The team lead drives forward relentlessly. It does NOT ask for permission.**

The workflow runs autonomously from INIT through DOCUMENT. The team lead's job is to keep the pipeline moving — spawn teammates, create tasks, monitor progress, route to next phase. No pauses, no confirmations, no "should I continue?" prompts.

**Auto-routing rules:**

| Situation | Action | Prompt user? |
|-----------|--------|--------------|
| Analysts complete tasks | Challengers pick them up automatically | No |
| Challengers PASS tasks | Monitor until all challenged | No |
| Challengers FAIL tasks | Re-analysis tasks created, analysts pick them up | No — just log what happened |
| All tasks challenged and PASSed | Proceed to DOCUMENT | No |
| Max challenge rounds hit (2) | Proceed to DOCUMENT with caveats | No |
| COMPLETE phase | Present findings | Yes — ask about next steps |

**The ONLY user interaction is at COMPLETE.**

---

## Phase 1: INIT

**Goal:** Understand exploration goal

**Actions:**
1. **Greet the user:**
   ```
   "Hey! Bob here, ready to coordinate the exploration team.

   Exploring: [topic/component]

   Let me rally the agents to dig into this."
   ```

2. **Verify experimental flag is enabled:**
   ```
   Check if CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 is set
   If not, STOP and say:
   "Agent teams are not enabled.
   Run this command to enable them:

   make enable-agent-teams

   Then restart Claude Code and try again!"
   ```

3. Create .bob/:
   ```bash
   mkdir -p .bob/state
   ```

4. Move to DISCOVER phase

---

## Phase 2: DISCOVER

**Goal:** Find relevant code and build a map

Spawn an Explore agent (regular subagent, not a teammate):
```
Agent(subagent_type: "Explore",
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

## Phase 3: SPAWN TEAM

**Goal:** Create agent team with analyst and challenger teammates

**Actions:**

**Step 1: Create the agent team**

```
"I need to create an agent team for this exploration task.

Team structure:
- 2 analyst teammates (team-analyst agents)
- 2 challenger teammates (team-challenger agents)

All teammates should use the Sonnet model for balanced quality and speed.

Please create this team now."
```

**Step 2: Spawn analyst teammates**

**Analyst 1:**
```
"Spawn a teammate named 'analyst-1' to analyze code from the shared task list.

Teammate prompt:
'You are analyst-1, a team-analyst agent working on codebase exploration.

Your job:
1. Check TaskList for available analysis tasks (pending, task_type: analysis or re-analysis)
2. Claim a task using TaskUpdate (set status: in_progress, owner: analyst-1)
3. Read task details with TaskGet
4. Read .bob/state/discovery.md for context
5. Research the codebase thoroughly for your assigned dimension
6. Write evidence-based findings to the output file specified in the task
7. Mark task completed
8. Repeat until no more tasks available

SPEC-DRIVEN MODULES: For any directory with SPECS.md, NOTES.md, TESTS.md, or
BENCHMARKS.md, read SPECS.md FIRST to understand documented contracts before
reading code. Read NOTES.md for design decisions and rationale.

Quality standards:
- Every claim must cite a specific file:line
- Verify relationships by reading actual code, not guessing
- Be thorough within your assigned dimension
- For re-analysis tasks, read ALL challenger feedback and address specific issues

Working directory: [current-working-directory]'"
```

**Analyst 2:**
```
"Spawn a teammate named 'analyst-2' with the same prompt as analyst-1
(with name changed to analyst-2)."
```

**Step 3: Spawn challenger teammates**

**Challenger 1:**
```
"Spawn a teammate named 'challenger-1' to challenge completed analysis.

Teammate prompt:
'You are challenger-1, a team-challenger agent stress-testing analysis quality.

Your job:
1. Check TaskList for completed analysis tasks (completed, task_type: analysis or re-analysis, not yet challenged)
2. Claim a task using TaskUpdate (set metadata.challenging: true, challenger: challenger-1)
3. Read the analysis output file
4. INDEPENDENTLY read the source code and verify every major claim
5. Either PASS (analysis accurate) or FAIL (significant issues)
6. If FAIL: create a re-analysis task with specific issues and evidence
7. Repeat until all completed analysis tasks are challenged

Be SKEPTICAL. Verify claims against actual code. Don't trust the analysis —
check it yourself.

PASS criteria: Core claims accurate, no major gaps, evidence verifiable.
FAIL criteria: Factual errors, major missing areas, unsupported claims.

Minor issues = PASS with notes. Major issues = FAIL with re-analysis task.

Working directory: [current-working-directory]'"
```

**Challenger 2:**
```
"Spawn a teammate named 'challenger-2' with the same prompt as challenger-1
(with name changed to challenger-2)."
```

**Step 4: Verify team creation**

Check that all teammates are spawned:
```
"Show me the current team members and their status."
```

You should see:
- analyst-1 (active)
- analyst-2 (active)
- challenger-1 (active)
- challenger-2 (active)

---

## Phase 4: CREATE TASKS

**Goal:** Create analysis tasks in the shared task list

Create one task per analysis dimension. Analysts will claim and work them autonomously.

**Task 1 — Structure & Components:**
```
TaskCreate(
  subject: "Analyze: Structure & Components",
  description: "Read .bob/state/discovery.md for context, then analyze the codebase.

  YOUR FOCUS: STRUCTURE and COMPONENTS.
  - What are the key types, interfaces, and structs?
  - What is each component's responsibility?
  - How are components organized (packages, modules, layers)?
  - What are the public APIs vs internal details?
  - What are the key abstractions?

  For any spec-driven modules, read SPECS.md FIRST.

  Write findings to .bob/state/analyze-structure.md",
  activeForm: "Analyzing structure and components",
  metadata: {
    task_type: "analysis",
    dimension: "structure",
    output_file: "analyze-structure.md",
    priority: "high"
  }
)
```

**Task 2 — Data Flow & Control Flow:**
```
TaskCreate(
  subject: "Analyze: Data Flow & Control Flow",
  description: "Read .bob/state/discovery.md for context, then analyze the codebase.

  YOUR FOCUS: DATA FLOW and CONTROL FLOW.
  - How does data enter the system?
  - What transformations happen along the way?
  - What are the key code paths (happy path, error paths)?
  - How is state managed and passed between components?
  - What are the entry points and exit points?
  - Are there async/concurrent flows? How do they coordinate?

  Trace actual execution paths. Be concrete — reference specific functions,
  methods, and call chains.

  Write findings to .bob/state/analyze-flow.md",
  activeForm: "Analyzing data and control flow",
  metadata: {
    task_type: "analysis",
    dimension: "flow",
    output_file: "analyze-flow.md",
    priority: "high"
  }
)
```

**Task 3 — Patterns & Conventions:**
```
TaskCreate(
  subject: "Analyze: Patterns & Conventions",
  description: "Read .bob/state/discovery.md for context, then analyze the codebase.

  YOUR FOCUS: PATTERNS and CONVENTIONS.
  - What design patterns are used (factory, strategy, observer, etc.)?
  - What are the error handling conventions?
  - What testing patterns are used?
  - What naming conventions are followed?
  - Are there recurring idioms or helper patterns?
  - How is configuration handled?
  - How are dependencies injected or managed?

  For any spec-driven modules, read NOTES.md for documented design decisions.

  Write findings to .bob/state/analyze-patterns.md",
  activeForm: "Analyzing patterns and conventions",
  metadata: {
    task_type: "analysis",
    dimension: "patterns",
    output_file: "analyze-patterns.md",
    priority: "high"
  }
)
```

**Task 4 — Dependencies & Integration:**
```
TaskCreate(
  subject: "Analyze: Dependencies & Integration",
  description: "Read .bob/state/discovery.md for context, then analyze the codebase.

  YOUR FOCUS: DEPENDENCIES and INTEGRATION.
  - What external dependencies are used and why?
  - How do components depend on each other?
  - Are there circular dependencies?
  - What are the integration points (APIs, databases, file systems, networks)?
  - How is the system configured and initialized?
  - What are the deployment or build concerns?

  Map the dependency graph. Identify coupling hotspots.

  Write findings to .bob/state/analyze-dependencies.md",
  activeForm: "Analyzing dependencies and integration",
  metadata: {
    task_type: "analysis",
    dimension: "dependencies",
    output_file: "analyze-dependencies.md",
    priority: "high"
  }
)
```

**After creating all tasks:**

Broadcast to all teammates:
```
"Broadcast to all team members:

Exploration tasks are ready! Here's how it works:

- Analysts: 4 analysis tasks are available. Claim and analyze.
- Challengers: As analysis tasks complete, challenge them. Verify claims against actual code.
- If a challenge FAILs, a re-analysis task will be created. Analysts pick those up.

Let's get a thorough, verified exploration done."
```

---

## Phase 5: EXECUTE (ANALYZE + CHALLENGE)

**Goal:** Analysts and challengers work concurrently through the task list

This is where the team pattern shines. Instead of sequential phases:
- **Analysts** claim analysis tasks, research, write findings, mark complete
- **Challengers** claim completed analysis tasks, verify against code, PASS or FAIL
- **If FAIL**: Challenger creates re-analysis task → analyst picks it up → challenger re-verifies
- This loops naturally through the task list until everything passes

**Your role as team lead:**
1. Monitor task list progress
2. Message teammates as needed
3. Track challenge rounds (max 2 per dimension)
4. Decide when to proceed to DOCUMENT

**Monitoring loop:**

Periodically check the task list:
```
TaskList()
```

Track:
- Analysis tasks pending
- Analysis tasks in progress
- Analysis tasks completed (waiting for challenge)
- Analysis tasks challenged and PASSed
- Analysis tasks challenged and FAILed
- Re-analysis tasks pending/in-progress/completed

**Example progression:**

```
[Initial state]
TaskList: 4 analysis tasks pending

Message from analyst-1: "Claimed Structure analysis"
Message from analyst-2: "Claimed Flow analysis"
TaskList: 2 pending, 2 in progress

Message from analyst-1: "Completed Structure analysis → analyze-structure.md"
TaskList: 2 pending, 1 in progress, 1 completed

Message from challenger-1: "Challenging Structure analysis"
Message from analyst-1: "Claimed Patterns analysis"
TaskList: 1 pending, 2 in progress, 1 being challenged

Message from challenger-1: "Structure analysis PASSED — accurate, well-evidenced"
Message from analyst-2: "Completed Flow analysis → analyze-flow.md"
TaskList: 1 pending, 1 in progress, 1 completed, 1 challenged+passed

Message from challenger-2: "Challenging Flow analysis"
Message from challenger-2: "Flow analysis FAILED — incorrect call chain in auth module, missing error recovery path. Created re-analysis task."
TaskList: 0 pending, 1 in progress, 1 re-analysis pending, 1 challenged+passed, 1 challenged+failed

Message from analyst-2: "Claimed re-analysis of Flow"
...

[Final state]
TaskList: All analysis tasks challenged and PASSed (including re-analyses)
→ Proceed to merge + DOCUMENT
```

**Completion criteria:**

Move to DOCUMENT when:
- All 4 original analysis dimensions have a PASSed analysis (original or re-analysis)
- OR max 2 challenge rounds per dimension reached (proceed with caveats)

**Handling stalls:**

If teammates go idle but work remains:
- Message analysts: "There are pending re-analysis tasks. Please claim them."
- Message challengers: "There are completed analysis tasks waiting for challenge."

**After all tasks complete:**

**Step 1: Read all analysis files**

Read all four PASSed analysis files:
```
Read(.bob/state/analyze-structure.md)
Read(.bob/state/analyze-flow.md)
Read(.bob/state/analyze-patterns.md)
Read(.bob/state/analyze-dependencies.md)
```

**Step 2: Merge into unified analysis**

Write a unified `.bob/state/analysis.md` that synthesizes all four dimensions:
- Architecture overview (from structure + dependencies)
- Key components and responsibilities (from structure)
- Data flow and control flow (from flow)
- Patterns and conventions (from patterns)
- Dependencies and integration points (from dependencies)
- Spec compliance (from any analyst that found spec-driven modules)
- Challenge results (what was corrected during challenge rounds)
- Remaining uncertainties (any issues from max-round caveats)

**Step 3: Shut down teammates**

```
"Ask analyst-1 teammate to shut down"
"Ask analyst-2 teammate to shut down"
"Ask challenger-1 teammate to shut down"
"Ask challenger-2 teammate to shut down"
```

Wait for each to confirm shutdown.

---

## Phase 6: DOCUMENT

**Goal:** Create clear documentation incorporating challenge results

Create comprehensive report in `.bob/state/exploration-report.md`:
- Overview of what was explored
- Architecture and structure
- Key components explained
- Flow diagrams (ASCII)
- Code examples
- Patterns observed
- Important files
- Challenge results summary (what was confirmed, what was corrected)
- Remaining uncertainties (unresolved challenger concerns)
- Questions/TODOs

---

## Phase 7: COMPLETE

**Goal:** Present findings to user

**Actions:**

1. **Clean up the team:**
   ```
   "Clean up the agent team"
   ```

2. **Present findings:**
   - Summarize learnings
   - Show key insights
   - Note confidence level (based on challenge results)
   - Point to detailed docs in `.bob/state/`

3. **Ask about next steps:**
   ```
   "Exploration complete!

   [Summary of key findings]

   Detailed report: .bob/state/exploration-report.md
   Unified analysis: .bob/state/analysis.md

   What next?
   - Explore deeper into a specific area?
   - Related components?
   - Start implementation? (switch to /work)"
   ```

---

## Team Architecture

```
Team Lead (You)
  ↓
  ├── Teammate: analyst-1 (team-analyst agent)
  ├── Teammate: analyst-2 (team-analyst agent)
  ├── Teammate: challenger-1 (team-challenger agent)
  └── Teammate: challenger-2 (team-challenger agent)

Coordination:
  - Shared task list (TaskCreate, TaskList, TaskGet, TaskUpdate)
  - Direct messaging between teammates
  - Team lead monitors and steers
  - Natural loop-back via re-analysis tasks
```

**The loop-back mechanism:**
```
Analyst completes analysis task
  → Challenger claims it, verifies against code
  → PASS: done
  → FAIL: Challenger creates re-analysis task
    → Analyst claims re-analysis task
    → Analyst writes corrected analysis
    → Challenger claims corrected analysis
    → PASS or FAIL (max 2 rounds)
```

This is the same pattern as team-coder/team-reviewer in the work workflow — the loop happens naturally through the task list without requiring the orchestrator to re-run entire phases.

---

## Troubleshooting

**Teammates not appearing:**
1. Check experimental flag is enabled
2. Verify team creation message worked
3. List teammates to see status

**Analysts going idle but challenges pending:**
1. Message analysts: "Check for re-analysis tasks"
2. Verify re-analysis tasks have correct metadata

**Challengers too strict:**
- If challengers FAIL everything, review their feedback
- Message challengers: "Only FAIL for significant factual errors, not style"

**Challenge loop stuck:**
- After 2 rounds per dimension, proceed to DOCUMENT
- Note unresolved issues in the report
