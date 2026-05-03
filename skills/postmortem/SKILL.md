---
name: bob:postmortem
description: Write a blameless incident postmortem — gather timeline, root cause, impact, action items
user-invocable: true
category: workflow
---

# Postmortem Skill

You are writing a **blameless incident postmortem**. The goal is to understand what happened, why it happened, and what concrete actions will prevent recurrence — not to assign blame.

## Principles (Google SRE / industry standard)

- **Blameless**: focus on systems and processes, not individuals
- **Actionable**: every lesson must produce a concrete action item with an owner
- **Honest**: include what went wrong AND what went well AND where we got lucky
- **Timely**: written while the incident is fresh

## What to gather

Ask the user for the following if not provided. Ask one question at a time, not all at once:

1. **What happened?** — brief description of the incident
2. **When?** — start time, detection time, mitigation time, resolution time (with timezone)
3. **Who was involved?** — responders, on-call, leads
4. **Impact** — who was affected, how many users/services, severity (P1/P2/P3/P4)
5. **Timeline** — key events in chronological order (what happened, who did what, when)
6. **Root cause** — the underlying technical reason (not the trigger)
7. **Trigger** — the immediate cause that started the incident
8. **Detection** — how was it detected (alert, user report, etc.)
9. **Resolution** — what fixed it
10. **What went well** — monitoring caught it fast, runbooks worked, etc.
11. **What went wrong** — gaps in alerting, process failures, etc.
12. **Where we got lucky** — things that could have been worse
13. **Action items** — concrete follow-up tasks

## Output format

Produce a postmortem in this structure:

```markdown
# Postmortem: [Title]

**Date:** [incident date]  
**Status:** [Draft / In Review / Complete]  
**Severity:** [P1 / P2 / P3 / P4]  
**Authors:** [names]

---

## Summary

[2–3 sentence description of what happened, impact, and how it was resolved]

## Impact

| Dimension           | Detail                                       |
| ------------------- | -------------------------------------------- |
| Duration            | [start] → [resolution] ([X hours Y minutes]) |
| Detected            | [detection time] — [TTD: X min]              |
| Mitigated           | [mitigation time] — [TTM: X min]             |
| Resolved            | [resolution time] — [TTR: X min]             |
| Users affected      | [count or percentage]                        |
| Services affected   | [list]                                       |
| Error budget impact | [if known]                                   |

## Timeline

| Time (TZ) | Event            |
| --------- | ---------------- |
| HH:MM     | [event]          |
| HH:MM     | [action by whom] |
| ...       | ...              |

## Root Cause

[The underlying technical reason. Not the trigger — the systemic issue that allowed this to happen.]

## Trigger

[The specific event that initiated the incident]

## Detection

[How the incident was detected. Include alert name if applicable.]

## Resolution

[What was done to resolve it. Step by step if relevant.]

## What Went Well

- [thing that worked — monitoring, runbook, communication, etc.]
- ...

## What Went Wrong

- [gap or failure — missing alert, unclear runbook, toil, etc.]
- ...

## Where We Got Lucky

- [things that could have been much worse]
- ...

## Action Items

| Action                        | Owner       | Priority   | Due    |
| ----------------------------- | ----------- | ---------- | ------ |
| [specific, measurable action] | [name/team] | [P1/P2/P3] | [date] |
| ...                           | ...         | ...        | ...    |

## Supporting Information

[Links to dashboards, logs, alert history, Slack threads, runbooks]
```

## Guidance

**Root cause vs trigger**: The trigger is "the deploy at 14:32 introduced a nil pointer". The root cause is "we have no integration tests for nil config values" or "the deploy pipeline doesn't run smoke tests". Dig deeper than the trigger.

**Action items must be**:

- Specific and measurable (not "improve monitoring" — "add alert for X metric exceeding Y threshold")
- Owned by a named person or team
- Time-bounded

**Severity guide**:

- P1: complete outage, all users affected
- P2: major degradation, most users affected or data loss risk
- P3: partial degradation, subset of users affected
- P4: minor issue, minimal user impact

**TTD / TTM / TTR**:

- TTD (Time to Detect): incident start → detection
- TTM (Time to Mitigate): incident start → user impact stopped
- TTR (Time to Resolve): incident start → full resolution + root cause confirmed

## Instructions

1. Greet the user briefly, explain you'll ask a few questions to write the postmortem
2. Ask for information you don't have, one question at a time
3. If the user provides everything upfront, skip to writing
4. Write the complete postmortem document
5. Ask if they want to adjust anything (tone, add detail, change action items)
6. Offer to save it to a file if they name one
