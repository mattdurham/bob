---
name: bob:premortem
description: Run a premortem — imagine a project has failed and work backwards to identify risks before they happen
user-invocable: true
category: workflow
---

# Premortem Skill

You are facilitating a **premortem** for a project or change. A premortem is the opposite of a postmortem: instead of analyzing what went wrong after the fact, you imagine it is some point in the future and the project has already failed — then work backwards to identify the causes before they happen.

## Why premortems work

From Gary Klein (HBR, 2007): prospective hindsight — imagining an event has already occurred — increases the ability to identify reasons for future outcomes by 30%. Teams that run premortems catch risks that normal planning misses because people feel safe saying "this will fail because..." rather than arguing against a plan mid-meeting.

## What to gather

Ask the user for the following if not provided. Ask one at a time.

1. **What is the project or change?** — description, scope, timeline
2. **Who is involved?** — team, stakeholders
3. **What does success look like?** — the goal / definition of done
4. **What is the timeline?** — when does it start, when is it due
5. **Any known risks already identified?** — so we don't duplicate

## Premortem process

### Step 1: Set the scene

Frame it for the team (include this in output):

> "It is [future date — end of the project]. The project has failed. Not just a little — it failed badly. Take a moment to picture it. Now: what happened?"

### Step 2: Brainstorm failure causes

Generate a comprehensive list of potential failure causes across these categories:

**Technical risks**

- Architecture or design flaws
- Dependencies that could break
- Performance, scalability, or reliability issues
- Security vulnerabilities
- Integration failures

**Execution risks**

- Timeline too aggressive
- Scope creep
- Key person dependency / bus factor
- Skills or knowledge gaps
- Testing gaps

**Process risks**

- Unclear requirements or changing specs
- Poor communication between teams
- Missing sign-offs or approvals
- Deployment or rollback complexity

**External risks**

- Third-party service failures
- Regulatory or compliance issues
- Vendor delays
- User adoption problems

**What could go right (but might not)**

- Assumptions we're relying on that could be wrong
- Things we're lucky to have that might not last

### Step 3: Prioritize

For each risk, assess:

- **Likelihood**: Low / Medium / High
- **Impact**: Low / Medium / High
- **Priority** = likelihood × impact

Focus discussion on High/High and High/Medium items.

### Step 4: Mitigations

For each high-priority risk, define:

- A concrete mitigation action
- An owner
- A due date (before the project starts or early in execution)

## Output format

```markdown
# Premortem: [Project Name]

**Date:** [today's date]  
**Project timeline:** [start] → [end]  
**Participants:** [names if provided]

---

## The Scenario

It is [end date]. The project has failed. Here's what went wrong...

---

## Risk Register

| #   | Risk               | Category  | Likelihood | Impact | Priority    |
| --- | ------------------ | --------- | ---------- | ------ | ----------- |
| 1   | [risk description] | Technical | High       | High   | 🔴 Critical |
| 2   | [risk description] | Execution | Medium     | High   | 🟠 High     |
| 3   | [risk description] | Process   | Low        | Medium | 🟡 Medium   |
| ... |                    |           |            |        |             |

**Priority key**: 🔴 Critical (H/H) · 🟠 High (H/M or M/H) · 🟡 Medium (M/M or L/H) · 🟢 Low

---

## Critical Risks — Mitigations

### 🔴 [Risk 1]

**Why this matters**: [explain the failure scenario]

**Mitigation**: [specific action to reduce likelihood or impact]  
**Owner**: [name/team]  
**Due**: [date — should be before project start or milestone]

### 🔴 [Risk 2]

...

---

## High Risks — Mitigations

### 🟠 [Risk N]

...

---

## Assumptions We're Relying On

These are things we're assuming will go right. If any of them break, the project is in trouble:

- [assumption] — _mitigation if wrong: [action]_
- ...

---

## What Could Go Right (That Might Not)

- [positive factor we're counting on] — _how to protect it: [action]_
- ...

---

## Action Items

| Action                       | Owner  | Priority   | Due    |
| ---------------------------- | ------ | ---------- | ------ |
| [specific preventive action] | [name] | [P1/P2/P3] | [date] |
| ...                          |        |            |        |

---

## Revisit Schedule

- [ ] [date before project start] — confirm mitigations are in place
- [ ] [mid-project milestone] — review risks, add new ones
- [ ] [go-live minus 1 week] — final risk review
```

## Guidance

**Encourage honesty**: The premortem works because people feel safe stating concerns as "this failed because..." rather than arguing against the plan. Don't debate risks during brainstorm — capture everything first.

**Dig into assumptions**: The most dangerous risks are the ones the team isn't discussing. Ask "what are we assuming will just work?"

**Focus on prevention, not prediction**: The goal isn't to predict the future — it's to take actions now that reduce risk.

**Distinguish likelihood from impact**: A low-likelihood catastrophic risk (data loss, security breach) may warrant more mitigation than a high-likelihood minor risk.

**Set a revisit date**: A premortem isn't done once. Re-run it at major milestones or when scope changes significantly.

## Instructions

1. Briefly explain what a premortem is and why it's valuable
2. Gather project info (ask one question at a time if not provided upfront)
3. Generate a comprehensive risk list across all categories
4. Prioritize by likelihood × impact
5. Write the full premortem document
6. Ask if they want to adjust priorities, add risks, or assign owners
7. Offer to save to a file
