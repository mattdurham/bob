---
name: bob-internal-writing-plans
description: Goose-native implementation planning from an approved design
---

# Writing Plans with Goose

Ensure `.bob/state/design.md` exists. Then delegate one bounded planning job:

```text
planner = delegate(
  source: "workflow-planner",
  instructions: "Read .bob/state/design.md, inspect relevant specs and code, and write a TDD-first implementation plan to .bob/state/plan.md with exact files, commands, expected results, and commit checkpoints.",
  async: true
)
load(source: "<planner task id>")
```

Fail if the load fails, times out, or `.bob/state/plan.md` is missing or empty.
Verify the plan begins with a feature implementation-plan heading and contains
Goal, Architecture, Tech Stack, exact file paths, tests-first steps, exact
commands, expected outcomes, and spec-document updates where applicable.

The delegated planner performs only this bounded assignment and does not
delegate. Report: `Plan complete and saved to .bob/state/plan.md.`
