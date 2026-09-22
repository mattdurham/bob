---
name: bob-audit
description: Goose-native read-only audit of documented invariants and codebase health
---

# Bob Audit for Goose

This workflow is strictly read-only: do not fix code or documentation. Discover
spec-driven modules, then audit documented invariants against implementation and
tests. A proposed invariant is not authoritative until the user accepts it.

For each bounded module/domain, use:

```text
audit = delegate(source: "Explore", instructions: "Read-only invariant audit; write the named .bob/state artifact", async: true)
load(source: "<audit task id>")
```

Run independent module audits asynchronously, retain all task ids, and validate
every artifact after loading. Check `SPECS.md` first, then `NOTES.md`, `TESTS.md`,
`BENCHMARKS.md`, code, and tests. Separate documented invariant violations from
undocumented risks. If structural Go analysis is relevant, delegate a bounded
call-graph/complexity/coupling scan and load its result.

Score only verified evidence. Write the final audit report with satisfied,
drifted, unverified, and proposed invariants; include `file:line`, impact, and
recommended next action. Present proposed-invariant review to the user rather
than silently changing the source of truth.
