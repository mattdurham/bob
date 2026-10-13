# Optional Premortem Risk Register

Use an existing project risk format when available. Adapt this template to the
actual scope; unknown likelihoods, owners, or dates stay explicit. Scenario
analysis describes possibilities, not predictions or established incident facts.

```markdown
# Premortem: [project or change]

Date: [actual date]
Scope and success criteria: [confirmed scope and goal]
Time horizon: [known milestone or stated scenario horizon]
Sources: [plan, specifications, constraints, prior incidents, user context]

## Failure Scenario

[Imagine the project has failed to achieve its goal. Describe a plausible
outcome and the conditions that could have led to it.]

## Prioritized Risks

| Risk and failure mechanism | Impact | Likelihood and evidence | Priority and rationale | Mitigation or validation | Early warning | Owner / due |
|----------------------------|--------|-------------------------|------------------------|--------------------------|---------------|-------------|
| [specific scenario] | [consequence] | [estimate with evidence, or unknown] | [relative priority and reason] | [practical action] | [observable signal] | [known owner/date or unassigned/TBD] |

## Unresolved Assumptions

[What must be true for the plan to work, what remains uncertain, and which
experiment or decision could resolve each important uncertainty.]

## Recommended Next Actions

[The few preventive or validation steps that deserve attention first, with
cost/dependencies where known. Link to the corresponding risks rather than
repeating their full descriptions.]
```

Use qualitative likelihood only when there is a basis for it; do not multiply
labels such as High and Medium into an apparently precise score. Consider the
cost of prevention and severity of consequences as well as likelihood. Rare
catastrophic risks may merit attention. Separate proposed owners, dates, and
priorities from agreed commitments. Suggest a revisit point only when it helps
address a concrete milestone or unresolved risk.
