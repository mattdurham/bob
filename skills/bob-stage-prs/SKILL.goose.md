---
name: bob-stage-prs
description: Goose-native staging of a large changeset into ordered reviewable pull requests
---

# Bob Stage PRs for Goose

This workflow is idempotent and supports `RESUME → ADVANCE → CONFIRM → EXECUTE`.
Persist the stack, branch names, commit ranges, dependency/merge order, and
publication state under `.bob/state/` after every unit.

Use a bounded read-only analysis delegate to partition the changeset:

```text
analysis = delegate(source: "Explore", instructions: "Analyze the exact diff and write the staged PR plan artifact", async: true)
load(source: "<analysis task id>")
```

Validate that each proposed PR is buildable, reviewable, behaviorally coherent,
and ordered so dependencies merge first. On RESUME, read state and external PR
status before proposing the next action. On ADVANCE, skip already merged units
and recompute only affected descendants.

Before the first external publication, show the complete stack and ask for one
CONFIRM decision. On EXECUTE, delegate each ordered publication unit to
`commit-agent`, load it synchronously, and validate the updated state artifact
before continuing. Stop immediately on a failed branch, commit, push, or PR.
Report merge order and the next resumable command.
