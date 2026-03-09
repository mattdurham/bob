---
name: bob:explore-teams
description: Team-based codebase exploration with adversarial challenge - DISCOVER → ANALYZE → CHALLENGE → DOCUMENT
user-invocable: true
category: workflow
---

# Team Exploration Workflow

You orchestrate **read-only exploration** with an adversarial **CHALLENGE** phase that stress-tests the analysis using concurrent specialist agents.

## Workflow Diagram

```
INIT → DISCOVER → ANALYZE → CHALLENGE → DOCUMENT → COMPLETE
                     ↑           ↓
                     └───────────┘
                   (challenge fails)
```

**Read-only:** No code changes, no commits.

---

## Execution Rules

**CRITICAL: All subagents MUST run in background**

- ALWAYS use `run_in_background: true` for ALL Agent calls
- After spawning agents, STOP - do not poll or check status
- Wait for agent completion notification - you'll be notified automatically
- Never use foreground execution - it blocks the workflow

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
Agent(subagent_type: "Explore",
     description: "Discover codebase structure",
     run_in_background: true,
     prompt: "Find code related to [exploration goal].
             Map file structure, key components, relationships.

             CLAUDE.md MODULES: For every directory you encounter, check for a CLAUDE.md
             file. If found, this is a documented module. Read CLAUDE.md FIRST — it
             contains the authoritative numbered invariants, axioms, assumptions, and
             non-obvious constraints for the module. These documents take priority over
             reading implementation code for understanding what the module does and why.

             Write findings to .bob/state/discovery.md.
             For CLAUDE.md modules, include a section summarizing the invariants and
             key constraints from CLAUDE.md.")
```

**Output:** `.bob/state/discovery.md`

---

## Phase 3: ANALYZE

**Goal:** Understand how code works from multiple angles using concurrent specialist agents

**This phase uses a team approach.** Instead of a single researcher, multiple agents analyze the codebase concurrently, each focusing on a different dimension. Their findings are then merged into a unified analysis.

**Actions:**

Spawn ALL analyst agents concurrently (in a single message with multiple Agent calls):

**Analyst 1 — Structure & Components**
```
Agent(subagent_type: "Explore",
     name: "analyst-structure",
     description: "Analyze structure and components",
     run_in_background: true,
     prompt: "Read .bob/state/discovery.md for context on what was found.
             Then read the identified source files.

             Your focus: STRUCTURE and COMPONENTS.
             - What are the key types, interfaces, and structs?
             - What is each component's responsibility?
             - How are components organized (packages, modules, layers)?
             - What are the public APIs vs internal details?
             - What are the key abstractions?

             For any CLAUDE.md modules identified in discovery, read CLAUDE.md
             FIRST to understand the documented invariants before reading code.

             Write findings to .bob/state/analyze-structure.md.")
```

**Analyst 2 — Data Flow & Control Flow**
```
Agent(subagent_type: "Explore",
     name: "analyst-flow",
     description: "Analyze data and control flow",
     run_in_background: true,
     prompt: "Read .bob/state/discovery.md for context on what was found.
             Then read the identified source files.

             Your focus: DATA FLOW and CONTROL FLOW.
             - How does data enter the system?
             - What transformations happen along the way?
             - What are the key code paths (happy path, error paths)?
             - How is state managed and passed between components?
             - What are the entry points and exit points?
             - Are there async/concurrent flows? How do they coordinate?

             Trace actual execution paths through the code. Be concrete —
             reference specific functions, methods, and call chains.

             Write findings to .bob/state/analyze-flow.md.")
```

**Analyst 3 — Patterns & Conventions**
```
Agent(subagent_type: "Explore",
     name: "analyst-patterns",
     description: "Analyze patterns and conventions",
     run_in_background: true,
     prompt: "Read .bob/state/discovery.md for context on what was found.
             Then read the identified source files.

             Your focus: PATTERNS and CONVENTIONS.
             - What design patterns are used (factory, strategy, observer, etc.)?
             - What are the error handling conventions?
             - What testing patterns are used?
             - What naming conventions are followed?
             - Are there recurring idioms or helper patterns?
             - How is configuration handled?
             - How are dependencies injected or managed?

             For any CLAUDE.md modules, read CLAUDE.md for documented
             invariants and constraints.

             Write findings to .bob/state/analyze-patterns.md.")
```

**Analyst 4 — Dependencies & Integration**
```
Agent(subagent_type: "Explore",
     name: "analyst-dependencies",
     description: "Analyze dependencies and integration",
     run_in_background: true,
     prompt: "Read .bob/state/discovery.md for context on what was found.
             Then read the identified source files.

             Your focus: DEPENDENCIES and INTEGRATION.
             - What external dependencies are used and why?
             - How do components depend on each other?
             - Are there circular dependencies?
             - What are the integration points (APIs, databases, file systems, networks)?
             - How is the system configured and initialized?
             - What are the deployment or build concerns?

             Map the dependency graph. Identify coupling hotspots.

             Write findings to .bob/state/analyze-dependencies.md.")
```

**After ALL analysts complete:**

Read all four analysis files and merge them into a unified `.bob/state/analysis.md`:
- Architecture overview (from structure + dependencies)
- Key components and responsibilities (from structure)
- Data flow and control flow (from flow)
- Patterns and conventions (from patterns)
- Dependencies and integration points (from dependencies)
- Invariant compliance (from any analyst that found CLAUDE.md modules)
- Assumptions and open questions (collected from all analysts)

If looping back from CHALLENGE, include in each analyst's prompt:
```
"Previous analysis had issues identified by challengers.
 Read .bob/state/challenge-accuracy.md, .bob/state/challenge-completeness.md,
 .bob/state/challenge-architecture.md, .bob/state/challenge-operational.md,
 and .bob/state/challenge-fresh-eyes.md for specific issues to address
 within your focus area."
```

**Input:** `.bob/state/discovery.md`
**Output:** `.bob/state/analyze-structure.md`, `.bob/state/analyze-flow.md`, `.bob/state/analyze-patterns.md`, `.bob/state/analyze-dependencies.md` → merged into `.bob/state/analysis.md`

---

## Phase 4: CHALLENGE

**Goal:** Stress-test the analysis using concurrent specialist agents that each focus on a different aspect. If challengers find significant gaps or errors, loop back to ANALYZE.

**This is the key differentiator from bob:explore.** Multiple agents run concurrently, each probing the analysis from a different angle. They read the analysis and the source code independently to find mistakes, gaps, and unsupported claims.

**Actions:**

Spawn ALL challenger agents concurrently (in a single message with multiple Agent calls):

**Challenger 1 — Accuracy**
```
Agent(subagent_type: "Explore",
     name: "challenger-accuracy",
     description: "Challenge analysis accuracy",
     run_in_background: true,
     prompt: "You are an adversarial reviewer checking the ACCURACY of an analysis.

             Read .bob/state/analysis.md, then independently read the source code
             referenced in .bob/state/discovery.md.

             Your job: find factual errors in the analysis.
             - Are component descriptions correct?
             - Are function behaviors described accurately?
             - Are data types and signatures right?
             - Are claimed relationships between components real?
             - Does the code actually do what the analysis says it does?

             Be skeptical. Verify claims against the actual code.

             Write your findings to .bob/state/challenge-accuracy.md:
             - VERDICT: PASS or FAIL
             - List of errors found (with file:line evidence)
             - List of unverified claims
             - Confidence level (high/medium/low)")
```

**Challenger 2 — Completeness**
```
Agent(subagent_type: "Explore",
     name: "challenger-completeness",
     description: "Challenge analysis completeness",
     run_in_background: true,
     prompt: "You are an adversarial reviewer checking the COMPLETENESS of an analysis.

             Read .bob/state/analysis.md, then independently explore the codebase
             using .bob/state/discovery.md as a starting point.

             Your job: find what the analysis MISSED.
             - Are there important components not mentioned?
             - Are there key code paths not covered?
             - Are error handling patterns described?
             - Are edge cases and failure modes documented?
             - Are there important dependencies or integrations missed?
             - Are there configuration or initialization flows omitted?

             Look beyond what the analysis covered.

             Write your findings to .bob/state/challenge-completeness.md:
             - VERDICT: PASS or FAIL
             - List of significant omissions
             - List of minor gaps
             - Suggested areas for deeper exploration")
```

**Challenger 3 — Architecture**
```
Agent(subagent_type: "Explore",
     name: "challenger-architecture",
     description: "Challenge architecture claims",
     run_in_background: true,
     prompt: "You are an adversarial reviewer checking the ARCHITECTURE claims in an analysis.

             Read .bob/state/analysis.md, then independently read the source code
             referenced in .bob/state/discovery.md.

             Your job: challenge the architectural understanding.
             - Are the described patterns actually used consistently?
             - Are dependency directions correct?
             - Are layer boundaries real or assumed?
             - Does the data flow description match reality?
             - Are there hidden coupling or circular dependencies?
             - Are concurrency/threading models described correctly?
             - Is the module boundary analysis accurate?

             Think structurally. Challenge assumptions about how pieces fit together.

             Write your findings to .bob/state/challenge-architecture.md:
             - VERDICT: PASS or FAIL
             - List of architectural mischaracterizations
             - List of missed architectural patterns
             - Corrected architectural understanding (if needed)")
```

**Challenger 4 — Operational / SRE**
```
Agent(subagent_type: "Explore",
     name: "challenger-operational",
     description: "Challenge from SRE/ops perspective",
     run_in_background: true,
     prompt: "You are an adversarial reviewer with an SRE/OPERATIONAL mindset.

             Read .bob/state/analysis.md, then independently read the source code
             referenced in .bob/state/discovery.md.

             Your job: evaluate the analysis through the lens of running this
             code in production.
             - What happens when things fail? Are failure modes documented?
             - Is observability addressed (logging, metrics, tracing)?
             - Are there resource leaks (goroutines, file handles, connections)?
             - What are the scaling bottlenecks?
             - Are timeouts, retries, and circuit breakers present where needed?
             - How does the system degrade under load or partial failure?
             - Are there operational concerns the analysis ignores (deployment,
               configuration, secrets management, health checks)?
             - What would wake you up at 3am?

             Think like someone who has to keep this running in production.

             Write your findings to .bob/state/challenge-operational.md:
             - VERDICT: PASS or FAIL
             - List of operational risks not covered in the analysis
             - List of missing failure modes
             - Recommendations for operational concerns to document")
```

**Challenger 5 — Fresh Eyes**
```
Agent(subagent_type: "Explore",
     name: "challenger-fresh-eyes",
     description: "Challenge with unbiased fresh perspective",
     run_in_background: true,
     prompt: "You are an adversarial reviewer providing a FRESH, UNBIASED perspective.

             IMPORTANT: Read the source code in .bob/state/discovery.md FIRST.
             Form your OWN understanding of what this code does and how it works
             BEFORE reading the analysis. Then read .bob/state/analysis.md and
             compare it against your independent understanding.

             Your job: catch groupthink and blind spots by coming at this cold.
             - Does your independent reading match the analysis narrative?
             - Are there simpler explanations the analysis overcomplicates?
             - Are there complexities the analysis glosses over?
             - Does the analysis make assumptions that aren't in the code?
             - Is the analysis telling a coherent story, or papering over
               inconsistencies?
             - What surprised you about the code that the analysis didn't mention?
             - What questions would a newcomer to this codebase ask that
               the analysis doesn't answer?

             Be the outsider. No prior context, no shared assumptions.
             If something doesn't make sense to you, it's a gap.

             Write your findings to .bob/state/challenge-fresh-eyes.md:
             - VERDICT: PASS or FAIL
             - List of discrepancies between your reading and the analysis
             - List of blind spots or groupthink patterns detected
             - Questions a newcomer would still have after reading the analysis")
```

**After ALL challengers complete:**

Read all five challenge files and make a routing decision:

**Routing rules:**
- **Any FAIL verdict** → Loop back to ANALYZE with challenger feedback
  - Append challenger findings to `.bob/state/discovery.md` as additional context
  - The re-analysis must address the specific issues raised
  - Maximum 2 challenge loops (after 2 failures, proceed to DOCUMENT with caveats noted)
- **All PASS** → Proceed to DOCUMENT

When looping back, re-run the full ANALYZE team. Each analyst's prompt already includes
instructions to read challenger feedback when looping back (see Phase 3).

**Output:** `.bob/state/challenge-accuracy.md`, `.bob/state/challenge-completeness.md`, `.bob/state/challenge-architecture.md`, `.bob/state/challenge-operational.md`, `.bob/state/challenge-fresh-eyes.md`

---

## Phase 5: DOCUMENT

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

## Phase 6: COMPLETE

Present findings to user:
- Summarize learnings
- Show key insights
- Note confidence level (based on challenge results)
- Point to detailed docs

**Next steps:**
- Explore deeper?
- Related areas?
- Start implementation? (switch to /work)
