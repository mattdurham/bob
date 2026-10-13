---
name: bob-challenge-idea
description: Goose-native idea challenge by eight parallel adversarial delegates
---

# Bob Challenge Idea for Goose

Perform interactive intake one question at a time, then write the agreed scope
to `.bob/state/idea-review/scope.md`. Launch eight bounded `team-analyst` jobs
with `delegate(..., async: true)`, retain all ids, and collect every result with
`load(source: "<task id>")`. Each job writes a unique artifact. Retry a missing
domain once and refuse to consolidate incomplete coverage.

### Team Agent 1 — Assumption Auditor
### Team Agent 2 — Prior Art Hunter
### Team Agent 3 — Scope Explosion Detector
### Team Agent 4 — Scale & Cardinality Attacker
### Team Agent 5 — Operational Burden Assessor
### Team Agent 6 — Alternative Solutions Devil
### Team Agent 7 — User Need Validator
### Team Agent 8 — Reversibility & Risk Auditor

Each delegate receives the scope, one named domain, evidence requirements, and
its exact artifact path. Challenge claims with evidence; no idea is assumed to
have a critical flaw. Each worker distinguishes demonstrated problems from
unverified assumptions, assigns severity from actual impact, and describes
possible remedies with cost and constraints. Missing evidence is uncertainty,
not proof of failure. After the load barrier, verify and deduplicate findings
without averaging away disagreement; repeated unsupported claims do not gain
confidence. Write `.bob/state/idea-review.md` with CRITICAL severe-impact findings,
HIGH issues that must be addressed, MEDIUM planning risks, LOW notes,
alternatives considered, clean domains, and a proceed/revise/stop routing
recommendation. Base it on feasibility, remedy cost, and the user's constraints:
revise for repairable flaws or important validation gaps; stop only when evidence
shows no feasible remedy at acceptable cost; proceed with explicit residual
risks when no blocker remains. A CRITICAL finding alone never implies stop.
Report unknown feasibility or cost as unknown. This workflow does not implement
the idea.
