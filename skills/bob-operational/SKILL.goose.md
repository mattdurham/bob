---
name: bob-operational
description: Goose-native multi-repository operational workflow - INIT → PLAN → CODE → OPERATE → TEST → COMPLETE
---

# Bob Operational for Goose

You are the parent orchestrator. You do no implementation, deployment, or test
work yourself. Keep a repository dependency graph and explicit result paths in
`.bob/operational/plan.md`.

## INIT

Identify repositories, worktrees, environments, authority boundaries,
acceptance criteria, rollback points, and spec-driven modules.

## PLAN

Write ordered per-repository assignments, owned paths, dependencies, and output
artifacts. Never start a dependent assignment before its prerequisites load.

## CODE

For independent repositories, use
`delegate(source: "workflow-implementer", instructions: "Bounded repository, exact working directory, paths, acceptance criteria, and result artifact", async: true)`.
Retain all task ids and call `load(source: "<task id>")` at each dependency
barrier. Serialize overlapping writes. Delegate review after writers load.

## OPERATE

Delegate each approved operational step with the exact repository, environment,
command scope, stop conditions, and rollback instructions. Require and validate
`.bob/operational/operate-results.md`. Stop on ambiguous target, lost authority,
or failed rollback prerequisite.

## TEST

Delegate end-to-end acceptance checks to `tester`; load the result and require a
per-repository result artifact. Failures route to PLAN for dependency/approach
problems or CODE for bounded corrections.

## COMPLETE

Complete only when every repository result, operation, rollback check, and
acceptance test is accounted for. Report deployed state and remaining manual
actions. Never let a delegated worker delegate again.
