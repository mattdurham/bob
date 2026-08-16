---
name: bob:cleanup
description: Single-agent code cleanup that removes complexity without adding functionality.
user-invocable: true
category: workflow
---

# Bob Cleanup — Single Agent

Perform cleanup directly in the current workspace. Never call `Task`, `Agent`,
`subagent`, create teammates, or use agent teams.

Inspect the code and its tests, identify safe simplifications, and record a brief
plan in `.bob/state/plan.md`. Make only behavior-preserving changes: remove dead
code, reduce needless abstraction, clarify names and comments, and update stale
documentation. Preserve public contracts and user changes. Run formatting, tests,
and relevant static checks, then review the final diff yourself. Report changes,
verification, and any cleanup intentionally left for later. Commit or push only
when the user requested it. If you push and open or update a PR, and
`BOB_CONFIRM_BEFORE_PUSH` is exactly `1` (check with `echo
"confirm-before-push: ${BOB_CONFIRM_BEFORE_PUSH:-unset}"`), commit locally
first, then show the user every commit the push will publish (`git log
--oneline HEAD --not --remotes=origin`) and the proposed PR title and body
verbatim, and ask `Push this branch and create or update its PR with the title
and body shown above? [push / stop]` before running `git push` or `gh pr
create`/`gh pr edit`. Only proceed on an exact `push` reply.
