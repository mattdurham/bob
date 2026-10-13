# Optional Incident Postmortem Template

Follow the organization's existing format when available. Adapt these sections
to the incident; omit irrelevant fields and keep unknown facts explicit.
Replace placeholders with supported facts. A document's draft/review status is
separate from the incident's operational resolution status.

```markdown
# Postmortem: [incident title]

Date: [incident date and timezone]
Document status: [draft / in review / approved, if established]
Incident status: [ongoing / mitigated / resolved / unknown]
Severity: [organization's classification or unknown]
Authors/responders: [known names or teams]

## Summary

[What happened, observed impact, and current resolution state.]

## Impact

[Known affected services/users, scope, duration, and any supported estimates.
Label estimates and cite their sources.]

| Milestone | Timestamp and timezone | Evidence |
|-----------|------------------------|----------|
| Impact began | [known timestamp or unknown] | [source] |
| Detected | [known timestamp or unknown] | [source] |
| Mitigated | [known timestamp or unknown] | [source] |
| Resolved | [known timestamp or unknown] | [source] |

## Timeline

| Time and timezone | Event | Evidence |
|-------------------|-------|----------|
| [timestamp] | [observed event or response action] | [source] |

## Causes and Contributing Factors

[Initiating event, failure mechanism, and contributing system/process conditions.
Distinguish demonstrated causality, hypotheses, and unanswered questions.]

## Detection and Response

[How detection, mitigation, and resolution happened; what helped, what hindered,
and any supported circumstances that limited the impact.]

## Follow-up Actions

| Action and expected outcome | Owner | Priority | Due | Status |
|-----------------------------|-------|----------|-----|--------|
| [specific preventive or detection improvement] | [known owner or unassigned] | [known/proposed priority] | [agreed/proposed date or TBD] | [known status] |

## Open Questions and Supporting Evidence

[Unresolved facts, contradictory sources, and links to logs, dashboards, changes,
incident notes, alerts, or runbooks.]
```

Use one consistent timezone or show conversions clearly. Calculate timing
metrics only from known endpoints and name the interval used. Resolution time
does not require a confirmed cause unless the organization's definition says
so. Do not infer severity, owners, deadlines, or approval merely to fill fields.
