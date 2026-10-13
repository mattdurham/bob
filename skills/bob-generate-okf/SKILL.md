---
name: bob-generate-okf
description: Create or update an OKF knowledge catalog linking the project’s packages, decisions, patterns, and confirmed planned work to authoritative sources.
---

# Maintain the Knowledge Catalog

Create or update the repository’s .knowledge/ catalog using [the installed OKF
format reference](references/okf-format.md) before authoring concepts. Start
with existing indexes, then inspect relevant code, tests, specifications,
decisions, and confirmed planned work. Identify useful package, feature,
decision, playbook, and pattern concepts with concrete source pointers. Ask only
for material intent or facts that cannot be established from the available
sources. Distinguish documented decisions, observed behavior, inferred
rationale, and proposals; never fabricate history, performance claims, or
roadmap commitments.

Extend an existing catalog by default, preserving user-authored material,
concept identities, links, and history. Replace or remove existing content only
when the request clearly calls for it. Write concise concepts that link to
authoritative code and specifications, then maintain the root and section
indexes and append meaningful updates to log.md using the actual date. Create
the standard indexes for a new catalog, leaving unsupported sections empty. Keep
the catalog as a navigation layer rather than duplicating its sources; avoid
changes to implementation or module specifications.

Check concept frontmatter, naming, resources, cross-references, anchors, and
index coverage against the installed format and the actual repository. Verify
that assertions and feature status have evidence, label unresolved assumptions,
and preserve unknown concept types during updates. Report the catalog path,
meaningful additions or corrections, important entry points, and any gaps that
remain. Finish when the requested catalog is written and checked; further
questions or user corrections can be handled directly without a mandatory
approval loop.
