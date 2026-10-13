---
name: bob-postmortem
description: Write a blameless incident postmortem grounded in evidence, with impact, timeline, contributing causes, and concrete follow-up actions.
---

# Write an Incident Postmortem

Write a blameless postmortem for the specified incident using the available
incident notes, logs, metrics, alerts, deployment history, and relevant code or
runbooks. Gather accessible evidence before asking the user for material facts
that remain missing. Use the organization’s incident definitions, severity
rubric, and document format when available. Preserve timestamps and timezones,
label estimates, and leave missing or contradictory information explicit rather
than fabricating a complete story.

Explain the observed impact, chronological timeline, initiating event, failure
mechanism, contributing conditions, detection, mitigation, and resolution.
Separate confirmed causes from hypotheses and open questions; do not force a
single root cause or equate a missing test with the technical failure. Focus on
system behavior and the conditions responders faced, with source links for
material claims. Capture what helped, what impeded response, and supported
circumstances that limited impact. Calculate timing metrics only when their
endpoints and definitions are known.

Produce a concise report in the established format, using [the optional
template](references/postmortem-template.md) when useful. Include specific
follow-up actions tied to demonstrated risks, with known owners, priorities, and
agreed dates; label suggestions and leave unassigned owners or dates explicit.
Save to the requested location, preserving existing authored content, or return
the draft when no file was requested. Report unresolved questions and supporting
evidence without requiring another intake or approval round simply to finish the
draft.
