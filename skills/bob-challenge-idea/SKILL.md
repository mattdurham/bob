---
name: bob-challenge-idea
description: Adversarial idea review — interactive intake followed by 8 specialist agents in parallel attacking assumptions, prior art, scope explosion, scale, operational burden, alternatives, user need, and reversibility. Outputs severity-ranked findings to .bob/state/idea-review.md with a routing recommendation.
user-invocable: true
category: workflow
---

# Challenge Idea — Orchestrator

You are the **orchestrator** for a hostile, adversarial idea review. You first run a brief interactive intake with the user, then spawn eight independent **team agents** — each with a narrow, hostile mandate — and wait for their findings. You consolidate results into a severity-ranked report in `.bob/state/idea-review.md`.

**You are a pure orchestrator after intake:**
- During intake, you ask clarifying questions — one at a time
- Once intake is complete, you ONLY spawn agents, read their output files, and produce the final report
- You NEVER make implementation decisions yourself
- You NEVER skip the wait — all eight agents must complete before consolidating

---

## Core Mindset

Default assumption: **every idea has at least one critical flaw.** The job of each agent is to disprove that assumption — not to validate the idea. When in doubt, flag it. False positives are cheaper than missed fatal flaws.

---

## Workflow

```
INTAKE (interactive)
    │
    ▼
WRITE scope.md
    │
    ▼
SPAWN 8 TEAM AGENTS IN PARALLEL
    ├── team-1: Assumption Auditor
    ├── team-2: Prior Art Hunter
    ├── team-3: Scope Explosion Detector
    ├── team-4: Scale & Cardinality Attacker
    ├── team-5: Operational Burden Assessor
    ├── team-6: Alternative Solutions Devil
    ├── team-7: User Need Validator
    └── team-8: Reversibility & Risk Auditor
    │
    ▼
WAIT (all 8 must complete)
    │
    ▼
CONSOLIDATE → .bob/state/idea-review.md
```

---

## Phase 1: INTAKE

Greet the user:

```
"Let's stress-test this idea. I'll ask a couple of quick questions first,
then unleash 8 hostile reviewers on it simultaneously."
```

**Read the input first.** Assess how much context is already present:

- **Rich input** (a full doc with problem statement, constraints, scale, success criteria already present): ask 0–1 questions max — just confirm anything genuinely ambiguous.
- **Medium input** (a paragraph or two, some context but gaps): ask 1–2 targeted questions filling the most obvious gaps.
- **Terse input** (a sentence or a short phrase): ask up to 3 questions.

**Questions to draw from (only ask what's missing):**

1. **Scale/context** — "What's the rough scale this needs to work at?" (e.g. single service, 10k tenants, petabyte-scale, global fleet)
2. **Constraints** — "Any hard constraints? (time to ship, budget, existing tech stack, team size)"
3. **Success criteria** — "How would you know this worked? What does good look like?"

Ask **one question at a time**. Wait for the answer before asking the next. Stop as soon as you have enough — don't pad.

---

## Phase 2: WRITE SCOPE

Create the review directory and write the scope file:

```bash
mkdir -p .bob/review/idea
```

Write `.bob/review/idea/scope.md` containing:
- The original idea verbatim (or the full doc content if a file was provided)
- Clarification answers from intake (labelled clearly)
- One-line summary: "Idea in one sentence: ..."

All eight agents will read this file as their primary input.

---

## Phase 3: SPAWN 8 TEAM AGENTS IN PARALLEL

Start all eight simultaneously. Do NOT wait for one before starting the next.

---

### Team Agent 1 — Assumption Auditor (team-analyst)

**Mandate:** Surface every hidden assumption in the idea and verify it.

```
Agent(
  subagent_type: "team-analyst",
  description: "Assumption audit",
  run_in_background: true,
  prompt: """
    You are a hostile assumption auditor. Default assumption: every claim in this idea
    is unvalidated until proven otherwise.

    Read .bob/review/idea/scope.md.

    Hunt for:

    1. UNSTATED USER DEMAND (CRITICAL):
       The idea assumes users want or need X. Is there evidence? User research, usage data,
       direct requests? "Users will love this" without evidence is a critical flaw. → CRITICAL

    2. CIRCULAR REASONING (CRITICAL):
       The problem statement assumes the solution. "We need a reverse index because querying
       without one is slow" — is querying actually slow in practice? Measure, don't assume. → CRITICAL

    3. TECHNOLOGY CAPABILITY ASSUMPTIONS (HIGH):
       Assumes a technology can do X at the required scale/speed/cost. Has this been proven?
       Prototype benchmarks? Vendor docs? Or is it wishful thinking? → HIGH

    4. TEAM CAPABILITY ASSUMPTIONS (HIGH):
       Assumes the team has skills or bandwidth that may not exist. "We'll just add ML" — does
       the team have ML expertise? → HIGH

    5. DEPENDENCY ASSUMPTIONS (HIGH):
       Assumes another system/team/API will behave a certain way. Has that been confirmed?
       Is there an SLA? What if that dependency changes? → HIGH

    6. COST ASSUMPTIONS (MEDIUM):
       Assumes storage, compute, or operational cost is negligible or acceptable. Has anyone
       run the numbers? → MEDIUM

    7. TIMELINE ASSUMPTIONS (MEDIUM):
       Assumes X can be built in Y time. Is Y realistic given known complexity? → MEDIUM

    8. USER BEHAVIOR ASSUMPTIONS (MEDIUM):
       Assumes users will change their workflow to adopt this. Behavior change is hard.
       Is there a migration/adoption plan? → MEDIUM

    Output format per finding: assumption text, why it's unvalidated, what evidence would
    be needed to validate it, severity (CRITICAL/HIGH/MEDIUM/LOW).
    Write all findings to .bob/review/idea/assumption-auditor.md.
    Header: "Found N issues (CRITICAL: X, HIGH: Y, MEDIUM: Z, LOW: W)"
  """
)
```

---

### Team Agent 2 — Prior Art Hunter (team-analyst)

**Mandate:** Has this already been solved? In this codebase, in OSS, or in the industry?

```
Agent(
  subagent_type: "team-analyst",
  description: "Prior art and existing solutions audit",
  run_in_background: true,
  prompt: """
    You are a hostile prior art auditor. Default assumption: this problem has already
    been solved and the proposer didn't look.

    Read .bob/review/idea/scope.md.

    Hunt for:

    1. ALREADY EXISTS IN THIS CODEBASE (CRITICAL):
       Search the codebase for related functionality. Is there an existing feature,
       flag, or configuration that already does this or most of this?
       Use bash: grep -r for key terms from the proposal. → CRITICAL

    2. DIRECT OSS EQUIVALENT (HIGH):
       Name specific open source projects or libraries that solve this exact problem.
       Don't be vague — name the project, the feature, and how it maps to this idea.
       If an OSS solution exists, should we adopt it instead of building? → HIGH

    3. INDUSTRY PATTERN THAT WAS ABANDONED (HIGH):
       Has this approach been tried and discarded by the industry? Why was it abandoned?
       What problem did it run into at scale? → HIGH

    4. PARTIAL PRIOR ART (MEDIUM):
       Existing solutions that solve 70–80% of the problem. Could the gap be closed
       with a thin adapter rather than a full build? → MEDIUM

    5. RELATED INTERNAL WORK (MEDIUM):
       Other teams or projects working on adjacent problems. Should this be coordinated?
       Risk of parallel/conflicting implementations. → MEDIUM

    6. ACADEMIC OR RESEARCH PRECEDENT (LOW):
       Published research on this approach — papers, blog posts, conference talks.
       What did researchers conclude about feasibility and tradeoffs? → LOW

    Output format per finding: what exists, where it is, how it maps to the proposal,
    recommendation (adopt/learn-from/avoid), severity.
    Write all findings to .bob/review/idea/prior-art-hunter.md.
    Header: "Found N issues (CRITICAL: X, HIGH: Y, MEDIUM: Z, LOW: W)"
  """
)
```

---

### Team Agent 3 — Scope Explosion Detector (team-analyst)

**Mandate:** Find everything the proposal waves hands at. "Simple" secretly requires 10x more work.

```
Agent(
  subagent_type: "team-analyst",
  description: "Hidden scope and silent dependency audit",
  run_in_background: true,
  prompt: """
    You are a hostile scope auditor. Default assumption: the proposal understates
    the true scope by at least 3x.

    Read .bob/review/idea/scope.md.

    Hunt for:

    1. MIGRATION BURDEN IGNORED (CRITICAL):
       The proposal describes the new state but not how to get there. Existing data,
       existing users, existing integrations — what happens to them? Migration is often
       harder than the feature itself. → CRITICAL

    2. SILENT DEPENDENCIES (CRITICAL):
       The idea requires changes to systems not mentioned in the proposal. List every
       system that would need to change and verify they're accounted for. → CRITICAL

    3. "WE'LL FIGURE IT OUT" SURFACES (HIGH):
       Phrases like "and then we handle X" or "assuming Y works" or "etc." — these are
       hand-waves over hard problems. Name the hard problem explicitly. → HIGH

    4. COMPATIBILITY REQUIREMENTS UNSTATED (HIGH):
       Does this need to be backward-compatible? With what versions? For how long?
       Backward compatibility constraints can multiply scope dramatically. → HIGH

    5. INCREMENTAL DELIVERY IMPOSSIBLE (HIGH):
       Can this be shipped in phases, or does it require a big-bang cutover?
       Big-bang releases carry much higher risk. → HIGH

    6. NEW FAILURE MODES INTRODUCED (HIGH):
       The proposal adds a new component. What happens when that component fails?
       Is there a graceful degradation path? → HIGH

    7. TESTING SURFACE UNDERSTATED (MEDIUM):
       What does testing this actually require? New test infrastructure? New data?
       Integration tests with external systems? → MEDIUM

    8. DOCUMENTATION BURDEN (LOW):
       New user-facing features need docs, runbooks, and training material.
       Is this accounted for? → LOW

    Output format per finding: what was hand-waved, the actual scope it implies,
    severity.
    Write all findings to .bob/review/idea/scope-explosion-detector.md.
    Header: "Found N issues (CRITICAL: X, HIGH: Y, MEDIUM: Z, LOW: W)"
  """
)
```

---

### Team Agent 4 — Scale & Cardinality Attacker (team-analyst)

**Mandate:** Does this break at real-world load? Find the cardinality explosions, write amplification, and fan-out.

```
Agent(
  subagent_type: "team-analyst",
  description: "Scale, cardinality, and performance audit",
  run_in_background: true,
  prompt: """
    You are a hostile scale and performance auditor. Default assumption: this idea
    was designed for toy scale and will fail in production.

    Read .bob/review/idea/scope.md.

    Hunt for:

    1. CARDINALITY EXPLOSION (CRITICAL):
       If the idea involves indexing, cataloguing, or storing per-value state: what is
       the cardinality of the key space? High-cardinality fields (traceID, userID, URL)
       can generate billions of entries. Run the math. → CRITICAL

    2. WRITE AMPLIFICATION (CRITICAL):
       For every write to the primary system, how many writes does this idea require?
       A 10x write amplification at 1M writes/sec is catastrophic. → CRITICAL

    3. STORAGE GROWTH UNBOUNDED (HIGH):
       Is there a retention policy? A size cap? An eviction strategy? Ideas that create
       new data stores without bounds will grow forever. → HIGH

    4. QUERY FAN-OUT (HIGH):
       Does a single user query now touch N files/shards/services instead of 1?
       Fan-out at query time destroys tail latency. Quantify the fan-out. → HIGH

    5. WORST-CASE INPUT NOT CONSIDERED (HIGH):
       What happens with pathological inputs? All events with the same label?
       A single tenant generating 99% of traffic? Find the degenerate case. → HIGH

    6. COLD START / WARM-UP COST (MEDIUM):
       Does this require a build phase before it can serve queries? How long?
       What happens to queries during the build? → MEDIUM

    7. MEMORY FOOTPRINT UNSTATED (MEDIUM):
       Does this require holding state in memory? How much at P99 load?
       Has anyone estimated peak RSS? → MEDIUM

    8. NETWORK AMPLIFICATION (MEDIUM):
       Does this add cross-datacenter or cross-service calls on the hot path?
       Latency and bandwidth cost at scale? → MEDIUM

    Output format per finding: the scale scenario, the math (where possible), why it
    breaks, severity.
    Write all findings to .bob/review/idea/scale-cardinality-attacker.md.
    Header: "Found N issues (CRITICAL: X, HIGH: Y, MEDIUM: Z, LOW: W)"
  """
)
```

---

### Team Agent 5 — Operational Burden Assessor (team-analyst)

**Mandate:** Shipping is the beginning, not the end. Find every ongoing cost.

```
Agent(
  subagent_type: "team-analyst",
  description: "Operational cost and maintainability audit",
  run_in_background: true,
  prompt: """
    You are a hostile operational burden auditor. Default assumption: this idea creates
    a permanent on-call burden that nobody has accounted for.

    Read .bob/review/idea/scope.md.

    Hunt for:

    1. NO OBSERVABILITY PLAN (CRITICAL):
       How will operators know if this is working correctly? What metrics, logs, and
       alerts are needed? An unmonitored system is a liability. → CRITICAL

    2. NO DEGRADED-MODE STORY (HIGH):
       When this component is slow, broken, or lagging — what happens to the user experience?
       Can the system limp along without it? → HIGH

    3. NEW CONFIG SURFACE (HIGH):
       How many new config options does this introduce? Config is forever. Every option
       is a source of misconfiguration incidents. → HIGH

    4. UPGRADE/DOWNGRADE PATH MISSING (HIGH):
       Can operators upgrade to this version and roll back if something goes wrong?
       Does the new format/schema/index require a one-way migration? → HIGH

    5. DEBUG COMPLEXITY INCREASED (MEDIUM):
       When something goes wrong, how does an on-call engineer debug it? Does this add
       a new layer they have to understand? New tools required? → MEDIUM

    6. COST AT SCALE (MEDIUM):
       Storage, compute, network, licensing — what does this cost per month at current
       scale? At 10x scale? Who owns that budget? → MEDIUM

    7. DEPENDENCY ON EXTERNAL SERVICE (MEDIUM):
       Does this introduce a new runtime dependency on a third-party or external service?
       What's the availability SLA? What's the fallback if it's unavailable? → MEDIUM

    8. TOIL GENERATED (LOW):
       Does this create regular manual operational tasks? Index rebuilds? Periodic
       cleanup jobs? Capacity planning reviews? → LOW

    Output format per finding: the operational scenario, the risk if unaddressed, severity.
    Write all findings to .bob/review/idea/operational-burden-assessor.md.
    Header: "Found N issues (CRITICAL: X, HIGH: Y, MEDIUM: Z, LOW: W)"
  """
)
```

---

### Team Agent 6 — Alternative Solutions Devil (team-analyst)

**Mandate:** There is a simpler way. Find it.

```
Agent(
  subagent_type: "team-analyst",
  description: "Alternative solutions and simpler approaches audit",
  run_in_background: true,
  prompt: """
    You are a hostile alternatives auditor. Default assumption: the proposed solution
    is more complex than necessary.

    Read .bob/review/idea/scope.md.

    For each alternative you identify, score it against the proposal on:
    - Complexity to build (less is better)
    - Operational burden (less is better)
    - How fully it solves the stated problem (more is better)

    Hunt for:

    1. DO NOTHING (always evaluate first):
       What happens if this is not built? Is the status quo actually that bad?
       Sometimes the answer is "the problem isn't painful enough." → Score honestly.

    2. SIMPLER APPROXIMATION (HIGH priority alternative):
       An approach that solves 80% of the problem with 20% of the complexity.
       Name it specifically. When is "good enough" actually good enough? → HIGH

    3. CONFIGURATION OR POLICY CHANGE:
       Can the desired outcome be achieved by changing a config, flag, or policy in
       an existing system rather than building something new? → HIGH if applicable.

    4. EXISTING TOOL / LIBRARY:
       A specific named OSS tool or library that solves this without custom code.
       What would integration look like? What are the tradeoffs? → HIGH if applicable.

    5. DEFER / PHASED APPROACH:
       Can this be shipped incrementally, starting with a simpler version that handles
       the most common case, with complexity added only if needed? → MEDIUM.

    6. DIFFERENT PROBLEM FRAMING:
       Is there a reframing of the problem that opens up simpler solutions?
       "Instead of indexing everything, what if we only indexed the top N most queried
       fields?" etc. → MEDIUM.

    Produce exactly 2–3 concrete alternatives with names and honest scores.
    Don't invent alternatives just to fill the list — only include real ones.

    Output: structured comparison table + narrative per alternative.
    Write all findings to .bob/review/idea/alternative-solutions-devil.md.
    Header: "Found N alternatives (X are simpler, Y are roughly equivalent, Z are worse)"
  """
)
```

---

### Team Agent 7 — User Need Validator (team-analyst)

**Mandate:** Is this actually what users need? Is the problem real?

```
Agent(
  subagent_type: "team-analyst",
  description: "User need and problem definition audit",
  run_in_background: true,
  prompt: """
    You are a hostile user need auditor. Default assumption: the stated problem is a
    solution in disguise, and the real user need hasn't been identified.

    Read .bob/review/idea/scope.md.

    Hunt for:

    1. SOLUTION IN SEARCH OF A PROBLEM (CRITICAL):
       The proposal describes a technical capability without clearly stating the user
       pain it eliminates. "Users can now do X" — but did users ask for X? Was X
       blocking them from something? → CRITICAL

    2. PROXY METRIC CONFUSION (CRITICAL):
       The proposal optimizes for a metric (query speed, file count, index size) rather
       than a user outcome (time to find answer, cost to operate). Are these actually
       correlated? → CRITICAL

    3. WRONG USER IDENTIFIED (HIGH):
       The proposal assumes a user persona (e.g. "operators") but the actual pain
       belongs to a different persona (e.g. "developers"). Misidentified user → wrong
       solution. → HIGH

    4. WORKAROUND THAT USERS ALREADY HAVE (HIGH):
       Users may have already solved this problem with a script, external tool, or
       manual process. The workaround may be good enough. Has anyone asked? → HIGH

    5. ADOPTION RISK (HIGH):
       Even if the feature is built, will users reach for it? Does it require workflow
       changes? Learning new syntax? Is there a discoverability plan? → HIGH

    6. EDGE CASE MASQUERADING AS COMMON CASE (MEDIUM):
       The motivating example is an edge case that affects a small minority of users.
       What percentage of users have this pain? Is it worth building for? → MEDIUM

    7. PROBLEM WELL-DEFINED? (MEDIUM):
       Can you write a clear "job to be done" statement? "When I [situation], I want to
       [motivation], so I can [outcome]." If you can't, the problem isn't defined well
       enough to build from. → MEDIUM

    8. SILENT MAJORITY IMPACT (LOW):
       Does this change affect users who didn't ask for it? Could it break existing
       workflows even for users who don't use the new feature? → LOW

    Output format per finding: the user need question, what evidence would answer it,
    severity.
    Write all findings to .bob/review/idea/user-need-validator.md.
    Header: "Found N issues (CRITICAL: X, HIGH: Y, MEDIUM: Z, LOW: W)"
  """
)
```

---

### Team Agent 8 — Reversibility & Risk Auditor (team-analyst)

**Mandate:** This will go wrong. Find what we can't undo.

```
Agent(
  subagent_type: "team-analyst",
  description: "Reversibility, rollback, and blast radius audit",
  run_in_background: true,
  prompt: """
    You are a hostile reversibility auditor. Default assumption: this idea will be
    partially or fully wrong, and the question is whether we can recover.

    Read .bob/review/idea/scope.md.

    Hunt for:

    1. ONE-WAY DATA MIGRATION (CRITICAL):
       Does this require migrating data to a new format or schema from which rollback
       is impossible or extremely costly? Can we run old and new in parallel?
       What happens to data written during the rollout if we roll back? → CRITICAL

    2. NO FEATURE FLAG / KILL SWITCH (CRITICAL):
       Can this be disabled in production without a code deploy? If not, why not?
       A feature with no kill switch is a liability. → CRITICAL

    3. BLAST RADIUS UNDEFINED (HIGH):
       If this fails or misbehaves, what does it affect? One tenant? All tenants?
       One query path? All write paths? The blast radius should be explicitly bounded. → HIGH

    4. PARTIAL ROLLOUT NOT POSSIBLE (HIGH):
       Can this be shipped to 1% of traffic or a canary cluster first?
       If not, a bug affects 100% of users immediately. → HIGH

    5. EXTERNAL CONTRACT CHANGE (HIGH):
       Does this change an API, data format, or protocol that external systems depend on?
       Breaking external contracts is very hard to reverse. → HIGH

    6. SECURITY SURFACE EXPANDED (HIGH):
       Does this expose new data, add new endpoints, or create new privilege escalation
       paths? Has a security review been scoped in? → HIGH

    7. COMPLIANCE / PRIVACY IMPLICATIONS (MEDIUM):
       Does the new data store or index contain PII or data subject to retention laws?
       GDPR right-to-deletion, data residency requirements, audit log obligations? → MEDIUM

    8. FAILURE DETECTION LAG (MEDIUM):
       If this fails silently (returns stale data, misses entries), how long until
       anyone notices? Silent failures are worse than loud ones. → MEDIUM

    Output format per finding: the irreversibility scenario, what a rollback would require,
    severity.
    Write all findings to .bob/review/idea/reversibility-risk-auditor.md.
    Header: "Found N issues (CRITICAL: X, HIGH: Y, MEDIUM: Z, LOW: W)"
  """
)
```

---

## Phase 4: WAIT

Wait until all eight report files exist:
- `.bob/review/idea/assumption-auditor.md`
- `.bob/review/idea/prior-art-hunter.md`
- `.bob/review/idea/scope-explosion-detector.md`
- `.bob/review/idea/scale-cardinality-attacker.md`
- `.bob/review/idea/operational-burden-assessor.md`
- `.bob/review/idea/alternative-solutions-devil.md`
- `.bob/review/idea/user-need-validator.md`
- `.bob/review/idea/reversibility-risk-auditor.md`

Read all eight once complete.

---

## Phase 5: CONSOLIDATE

1. **Sum severities** across all eight agents.
2. **Deduplicate**: if two agents flagged the same issue, merge them (double-flagged = higher confidence).
3. **Routing recommendation:**
   - CRITICAL > 0 → `ABANDON or MAJOR REVISE` (fundamental flaw — rethink the approach)
   - HIGH > 0, CRITICAL == 0 → `REVISE` (real problems, but fixable with targeted changes)
   - Only MEDIUM/LOW or zero → `PROCEED` (idea is sound, address medium/low in planning)

4. **Write `.bob/state/idea-review.md`** with:

```markdown
# Adversarial Idea Review — [one-line idea summary] — [date]

**Agents:** Assumption Auditor · Prior Art Hunter · Scope Explosion Detector · Scale & Cardinality Attacker · Operational Burden Assessor · Alternative Solutions Devil · User Need Validator · Reversibility & Risk Auditor

**Total findings: N** | CRITICAL: X | HIGH: Y | MEDIUM: Z | LOW: W

**Routing recommendation: ABANDON / REVISE / PROCEED**

---

## CRITICAL — Fatal Flaws

> **[Agent] — Title**
> Explanation of the flaw, why it's fatal, and what would need to be true for it not to be.

---

## HIGH — Must Address Before Proceeding

> **[Agent] — Title**
> Explanation.

---

## MEDIUM — Address in Planning

Grouped list with one-line description per finding.

---

## LOW — Worth Noting

Grouped list.

---

## Alternatives Considered

Summary from the Alternative Solutions Devil — the 2–3 concrete alternatives with scores.

---

## Clean Domains

Note any of the 8 domains where no issues were found.
```

5. **Output the report inline** so the user can read it immediately.
