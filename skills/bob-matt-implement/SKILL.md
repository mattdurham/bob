---
name: bob-matt-implement
description: "Implement a piece of work based on a spec or set of tickets."
disable-model-invocation: true
metadata:
  author: Matt Pocock
  source: https://github.com/mattpocock/skills/tree/49dd158d1076134a641b33efb035946536778336/skills/engineering/implement
  upstream-commit: "49dd158d1076134a641b33efb035946536778336"
license: MIT
---

Implement the work described by the user in the spec or tickets.

If the user passes a ticket reference, fetch it from the issue tracker and state its title before starting. If the reference is ambiguous, ask.

Call the Skill tool with "bob-matt-tdd" where possible, at pre-agreed seams.

Run typechecking regularly, single test files regularly, and the full test suite once at the end.

Once done, call the Skill tool with "bob-matt-code-review" to review the work.

Commit your work to the current branch.
