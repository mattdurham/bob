---
name: bob-ship
description: Publish an already verified and reviewed changeset, honor configured push controls, and report CI and PR feedback.
---

# Publish a Reviewed Change

Confirm the requested changeset has current passing verification in `.bob/state/test-results.md` and a clean final review in `.bob/state/review.md`. Compare the reviewed scope and revision with live state; missing, stale, failed, or unresolved results block publication. Preserve unrelated user changes. Do not repair code in this skill: return failures or any hook changes to the calling workflow for verification and review before publishing.

Before committing, pushing, or creating/updating a PR, read and follow [the publication procedure](references/publication-procedure.md). The parent owns publication. Run the bounded worker jobs — the commit/prepare pass and the CI/PR monitoring pass — as `task` tool items (`agent: "task"`) carrying the commit/prepare duties and the CI/PR monitoring duties self-contained. Launch both passes from the parent's single `task` tool call where they are independent; the parent then `wait`s for all delivered results and consolidates. Workers never re-delegate; the parent performs any further dispatch. Respect a caller's single-agent boundary (perform the job directly when required). Read `BOB_CONFIRM_BEFORE_PUSH` from the environment; when it prints exactly `1`, prepare locally, show every unpushed commit and the exact proposed PR title and body, then wait for the specified `push` reply. Revalidate the approved state before publication. Keep the existing publication gate and recovery controls on every push; never bypass a failed gate. Collect and validate `.bob/state/commit.md`. Preserve failed publication artifacts and report what actually reached the remote. A declined publication ends this invocation without retrying.

After successful publication, inspect CI and PR feedback through a bounded monitor `task` item or directly. Repeat checks after a pause while CI is pending; report operational failures rather than assuming readiness. Write `.bob/state/ship-status.md` with branch, commit, PR URL, CI state, and reason: `COMPLETE` only after publication succeeds and monitoring is `READY`; `NEEDS_BRAINSTORM` for failed CI or requested changes; `FAILED` for operational failure or declined publication. Return needed code changes to the caller, keep the review skill's status intact, and report the result. Merge only with the caller's final user approval.

---

## Entry Gate: Verified, Reviewed, and Current

Before any commit, push, or PR mutation you MUST pass the entry gate:

1. Read `.bob/state/test-results.md` and `.bob/state/review.md` with `read`.
   - Verification must be current and passing; review must be a clean final review.
2. Compare the reviewed scope and revision with live state. Confirm the reviewed
   commit/revision matches `git rev-parse HEAD` (and, when recorded, the branch
   matches `git branch --show-current`) for the verified scope. Confirm the code
   on disk at that revision is what was tested and reviewed.
3. Block publication (exit `FAILED`, return to caller) if any of:
   - test results are missing, stale, or failed;
   - the review is missing, stale, or has unresolved/requested changes;
   - the live revision no longer matches the reviewed revision.
4. Never repair code in this skill. Any failure — including a hook that fails
   with a block or requires edits — is returned to the calling workflow for
   verification and review before publishing. Preserve unrelated user changes.

Passing the gate authorizes the publication procedure, not a bypass of it.

---

## Read the Publication Procedure

Read `references/publication-procedure.md` with `read` before committing,
pushing, or creating/updating a PR. It is authoritative on the control
procedure: the confirm flag, the two confirm-mode passes, the resume arms, the
preview contract, the publication gate, and recovery. Its worker examples are
written in the Claude `subagent()` API and must be translated to omp's `task`
tool exactly as the sections below render them. Where a literal worker example
differs from the omp `task` renderings below, the renderings below govern; the
control procedure itself is unchanged.

---

## Check the Confirm Flag (decision input, never from memory)

Check the flag with a literal command — the printed value is the decision
input:

```bash
echo "confirm-before-push: ${BOB_CONFIRM_BEFORE_PUSH:-unset}"
```

Confirm mode is ON only when the command prints `confirm-before-push: 1`.
Decide from the printed line, never from environment memory.

---

## COMMIT Phase

Write `.bob/state/commit-prompt.md` before any commit pass:

```markdown
# Commit Instructions

Commit only the changes within the verified and reviewed workflow scope.
Preserve unrelated user changes. Do not repair code; return any hook failure
requiring edits to the calling workflow for verification and review.

Context:
- Review findings: .bob/state/review.md
- Implementation details: .bob/state/implementation-status.md (if exists)
- Original plan: .bob/state/plan.md (if exists)

Commit message guidance:
- Summarize what was fixed based on the review findings
- If this was a new feature + review fixes, lead with the feature
- Include brief note on issues addressed
```

### Confirm mode OFF

First check for a prepared commit left by an earlier confirmation pause
(including one abandoned or declined — disabling confirmation supersedes a
prior stop): if `.bob/state/commit.md` reads `STATUS: AWAITING_CONFIRMATION`,
its BRANCH and HEAD match live `git branch --show-current` and
`git rev-parse HEAD`, and the tree is clean apart from `.bob/state` (same
root-scoped status command as the resume check below), append this line to
`.bob/state/commit-prompt.md`:

> An earlier prepare pass already created commit [HEAD] on this branch; do
> not create a new commit — publish it: push, create the PR, and delete
> .bob/state/pr-body.md and .bob/state/pr-title.txt once the PR exists.

Otherwise delete any stale `.bob/state/pr-body.md` and
`.bob/state/pr-title.txt` (a mismatched `commit.md` record is ignored and does
not activate this resume arm). A `FAILED` record — whether from a
`CONFIRM_MODE: PUBLISH` run or an earlier flag-off publication attempt — does
not activate this arm; recovery with confirmation off is manual.

Then run the COMMIT task item (below, `CommitPrepare` with a `CONFIRM_MODE`
absent). After it finishes, `wait`, then read `.bob/state/commit.md`:
- `STATUS: SUCCESS` → move to MONITOR.
- `STATUS: FAILED` → report the failure and exit `STATUS: FAILED`.

> **CONFIRM_MODE: PREPARE / PUBLISH**: the spawn task text carries a
> `CONFIRM_MODE` line (exactly `PREPARE` or `PUBLISH`); that line overrides
> anything stale in `.bob/state/commit-prompt.md` regarding commit/create-PR
> behavior. The task items below respect that.

### Confirm mode ON — two passes, with the user's decision between them

**a. Resume check.** If `.bob/state/commit.md` already reads
`STATUS: AWAITING_CONFIRMATION`, its BRANCH and HEAD match live
`git branch --show-current` and `git rev-parse HEAD`, the tree is clean apart
from `.bob/state` — the status command, literally excluding the state directory
where INIT created it (`--show-prefix` is captured before entering the root; it
is empty when already there):

```bash
(P=$(git rev-parse --show-prefix) && cd "$(git rev-parse --show-toplevel)" && git status --porcelain -- . ":(exclude,literal)${P}.bob/state")
```

— exits successfully and prints nothing, and both state files are readable
(`test -r .bob/state/pr-body.md && test -r .bob/state/pr-title.txt`),
**skip to (c)**: an earlier paused run is being resumed.
If commit.md instead reports `STATUS: FAILED` with `PUSHED: yes` and
`PR_CONFIRMED: no`, its BRANCH and HEAD still match live state, the tree is
clean apart from `.bob/state` (same root-scoped status command), and both state
files are readable (same `test -r` command), skip to (c) — on `push`, the
publish pass re-runs the push step (safe to repeat), then continues at Step 6
to create or update the PR. Likewise if commit.md reports a publish FAILURE
whose reason is a publication gate block, with BRANCH and HEAD still matching
live state, the tree clean apart from `.bob/state` (same status command), and
both state files readable (same `test -r` command): skip to (c) — on `push`,
the publish pass re-runs the gated push.
Otherwise delete any stale `.bob/state/pr-body.md` and
`.bob/state/pr-title.txt` before continuing — a tree with changes outside
`.bob/state` never resumes: it falls through to (b), so the fresh prepare pass
commits the new work and the preview covers it.

**b. Prepare pass.** Run the `CommitPrepare` task item with
`CONFIRM_MODE: PREPARE`. It creates the commit, writes the proposed PR body to
`.bob/state/pr-body.md` and the proposed PR title to `.bob/state/pr-title.txt`,
does NOT push and does NOT create a PR, and writes
`STATUS: AWAITING_CONFIRMATION` (or FAILED) to `.bob/state/commit.md`.
After it finishes, `wait`, then read `.bob/state/commit.md`:
`STATUS: FAILED` → report and exit FAILED; any status other than
`AWAITING_CONFIRMATION` → treat as FAILED and say the pause did not happen.

**c. Present the preview mechanically — never summarize or restate it.**

1. Show every commit the push will publish — run
   `git log --oneline HEAD --not --remotes=origin` and present its output
   verbatim (the push publishes the branch ref, so unpushed ancestors ship with
   it).
2. Show the commit details from `.bob/state/commit.md` (branch, SHA, message,
   files; for a publish-failure report, every field it carries, including its
   publication state), then the PR title and body:
   ```bash
   cat .bob/state/pr-body.md
   ```
3. Ask exactly:
   `Push this branch and create or update its PR with the title and body shown above? [push / stop]`

**d. Route on the reply.**

- Exactly `push` → run the `CommitPublish` task item with the branch, commit
  SHA, and PR title from `.bob/state/commit.md` and `CONFIRM_MODE: PUBLISH`.
  It verifies live state matches, then pushes through the gate and creates or
  updates the PR from the title/body files, without creating a commit. After
  it finishes, `wait`, read `.bob/state/commit.md`:
  - `SUCCESS` → move to MONITOR.
  - `FAILED` → relay the reported failure reason in one sentence and exit
    FAILED. A moved HEAD means: re-run /bob:ship to attempt a fresh preview (a
    fresh prepare pass re-presents only work a confirm-mode pass recorded; a
    clean live commit with no such record reports "nothing to commit").
- Exactly `stop` → delete `.bob/state/pr-body.md` and
  `.bob/state/pr-title.txt`, then exit with `STATUS: FAILED`, reason "user
  declined publication; branch retained at [sha]; approved PR creation or
  update was not confirmed" (add "the branch was already pushed by an earlier
  attempt" when commit.md's For Orchestrator PUSHED field is yes — never claim
  nothing was pushed in that case). Never continue to MONITOR or COMPLETE.
- Anything else → ask the question again (repeat only the question, not the
  preview).

Revalidate the approved state (entry gate) immediately before publication; a
real push or PR mutation never proceeds on stale approval.

---

## COMMIT Workspace Task Items

Run from the repository root. Items are self-contained, `agent: "task"`,
read-only except for writing their `OUTPUT_*` files under `.bob/state`. Launch
each as the procedure requires (the confirm gate serializes them); use the
harness `task` tool with a single `tasks[]` entry, then `wait` for the result.
Workers never re-delegate.

### `CommitPrepare` — commit the reviewed changes (create commit; push/PR only per mode)

```text
name: CommitPrepare
agent: task
task: |
  Work in the bob repo root. Read `.bob/state/commit-prompt.md` for context
  (review findings, implementation, plan) and honor the CONFIRM_MODE line if
  present. Do not repair code; return any hook failure requiring edits to the
  calling workflow. Preserve unrelated user changes. Never stage or commit
  `.bob/state`.

  Step 1 — Review state (run in parallel): `git status`, `git diff HEAD`,
  `git log --oneline -10`, `git branch --show-current`.

  Step 2 — Draft the commit message from the review findings: a type-prefixed
  subject under 70 chars (feat:/fix:/refactor:/test:/docs:/chore:/perf:/style:),
  then a body explaining what changed and why, then a `Co-Authored-By:` line.

  Step 3 — Stage specific files by name (never `git add -A` or `git add .`).
  Check each file with `git diff <file>` first; skip files containing secrets
  (.env, credentials, API keys, private keys).

  Step 4 — Commit with a heredoc:
  `git commit -m "$(cat <<'EOF' ... EOF)"`. Create a NEW commit — never amend,
  never `--no-verify`, never force. If CONFIRM_MODE is `PREPARE`, STOP here:
  do not push, do not create a PR.

  CONFIRM_MODE: PREPARE only — Draft the PR now with the `write` tool:
  - Write the proposed PR body to `.bob/state/pr-body.md`.
  - Write the proposed PR title (one non-empty line, no CR/LF bytes) to
    `.bob/state/pr-title.txt`.
  Never interpolate either draft into Bash command text. Then write
  `.bob/state/commit.md` with `write`, status `STATUS: AWAITING_CONFIRMATION`,
  plus the proposal as described at the end. If an earlier paused run is being
  resumed (see resume arms), re-report the matching existing commit instead of
  failing with "nothing to commit". A clean tree with no confirm-mode record is
  a plain "nothing to commit" failure — never offer a commit no confirm-mode
  pass recorded, and never re-offer one whose publication already succeeded.

  If any step fails, write `.bob/state/commit.md` with `STATUS: FAILED`, the
  failed step, the error, an explanation, and a suggested action; then stop.

  Output — `.bob/state/commit.md`, status `SUCCESS | FAILED` (or
  `AWAITING_CONFIRMATION` for a PREPARE run), committing branch/SHA/message/
  files. A PREPARE run's For Orchestrator block:
  `STATUS: AWAITING_CONFIRMATION`, `BRANCH`, `HEAD`, `TITLE`, `PUSHED: no`,
  `NEXT_PHASE: CONFIRM`.
solutionSpace: |
  Straight port of the commit-agent instructions; commit/PR mechanics are fixed.
  Only judgment (message wording, which files) and the confirm-mode branch are
  open. Do not redesign the git workflow.
```

### `CommitPublish` — push through the gate and create/update the PR

```text
name: CommitPublish
agent: task
task: |
  Work in the bob repo root. Run with `CONFIRM_MODE: PUBLISH`; a matching
  BRANCH, commit SHA, and approved PR title were named by the orchestrator.

  Verify live state first: `git branch --show-current` and `git rev-parse
  HEAD` must equal the named BRANCH and SHA byte-for-byte. On ANY mismatch STOP
  with a FAILED report saying "HEAD moved since the preview; re-run /bob:ship
  to attempt a fresh preview". Never create a commit in this mode. Read
  `.bob/state/pr-title.txt` with the `read` tool and compare it in memory to the
  task-provided title; NEVER place the task-provided title in shell command
  text while verifying it. On mismatch, STOP with a FAILED report saying the
  title state is stale — re-run /bob:ship for a fresh preview.

  Note any pre-Step-3 movement-block re-run guard from `.bob/state/commit.md`:
  when a recorded gate block names branch/HEAD that moved, require live state to
  equal exactly the values that block line names before any staging/commit/push;
  if either differs, stop without staging, committing, or pushing — the
  hook-made state must be restored to the named pre-hook branch and HEAD, never
  published or used as a repair base.

  Step 5 — Push through the publication gate (allow the call at least 150s —
  the repo's hook gets up to 90):
  `bash -- "[agent-directory]/scripts/push-with-gate.sh"`
  (the script runs the repo's `.bob/hooks/pre-publish` if present, then pushes,
  in one process).
  - Exit 0 → note the branch and continue to Step 6.
  - Nonzero and LAST line starts with `BOB_GATE_BLOCK:` → the gate blocked
    publication; final for this run. Do not push another way, do not use
    `--no-verify`, do not continue to Step 6. Write the FAILED report quoting
    the block line and any hook output above it (the script surfaces the last
    8 KiB), record branch and SHA, suggested action: "fix what the block line
    reports, then re-run — the commit exists; a re-run with a clean tree
    retries publication without recommitting".
  - The script is missing or won't start (call errors without a
    `BOB_GATE_BLOCK:` line — e.g. "No such file or directory") → install error;
    write the FAILED report saying so; never fall back to plain `git push`.
  - Nonzero without that final block line (script itself ran) → ordinary push
    failure; retry only by re-running this same wrapper call, never plain
    `git push`.
  - The call died without printing a verdict → re-run the push call; it is safe
    to repeat (the gate runs again; re-pushing an already-pushed commit is a
    no-op). Never skip to Step 6 on a dead call: a matching remote ref does not
    prove this run's gate ran.

  Step 6 — Create or update the PR with the approved files (never a fresh
  heredoc, never the title substituted then re-parsed; reading both from their
  files passes them literally, so shell syntax in a title is never evaluated):
  - If an OPEN PR for this branch already exists:
    `gh pr edit [PR number] --title "$(cat .bob/state/pr-title.txt)" --body-file .bob/state/pr-body.md`.
    If the edit fails, STOP and report FAILED with `PR_CONFIRMED: no`,
    retaining both state files — only a successful edit followed by Step 7
    verification of that same PR reports that PR's URL as success.
  - Otherwise: `gh pr create --title "$(cat .bob/state/pr-title.txt)" --body-file .bob/state/pr-body.md`.
  Capture the PR URL.

  Step 7 — Verify: `git log -1 --oneline`, `git status`; verify the exact PR
  created/updated with
  `gh pr view [PR number or URL] --json number,url,title`.
  Only after Step 7 verifies the PR, delete `.bob/state/pr-body.md` and
  `.bob/state/pr-title.txt` — never on a failure, so a re-run can finish
  publication. If the push succeeded but PR creation/update failed, report
  FAILED noting the branch is pushed.

  Step 8 — Write `.bob/state/commit.md` with `write`; `STATUS: SUCCESS` on
  publication, listing branch, SHA, message, files, PR number/URL/title, and
  `NEXT_PHASE: MONITOR`. On failure write `STATUS: FAILED`; a PUBLISH failure
  must carry `BRANCH`, `HEAD`, `TITLE`, `PUSHED: yes/no`, `PR_CONFIRMED:
  yes/no`, `RETRY`. `PR_CONFIRMED: yes` means Step 7 verified the exact PR
  after successful creation or a successful existing-PR edit.
solutionSpace: |
  Direct port of the commit-agent PUBLISH flow. The gate wrapper, resume arms,
  and step order are fixed; only reading/writing state and choosing the
  create-vs-edit path (given live PR presence) are open.
```

---

## MONITOR Phase

With `.bob/state/commit.md` showing `STATUS: SUCCESS` (PR URL present), write
`.bob/state/monitor-prompt.md`:

```markdown
# Monitor Instructions

Monitor the PR that was just created or updated.

PR details: see .bob/state/commit.md for the PR URL.

Check:
- CI/CD check status
- Review comments or change requests
- Whether PR is ready to merge

Write status to .bob/state/monitor.md with routing decision.
```

Run the `MonitorPR` task item below, or inspect directly. After it finishes,
`wait`, then read `.bob/state/monitor.md` and route:

| Monitor Status | Action |
|----------------|--------|
| READY (all green) | → COMPLETE |
| WAITING (CI in progress) | → after a pause, run `MonitorPR` again (repeat as long as CI is pending) |
| NEEDS_WORK (CI failed or changes requested) | → exit with `NEEDS_BRAINSTORM` |

Repeat the check after a pause while CI is pending. Report operational failures
(gh auth/network/rate-limit errors) as failures rather than assuming readiness.

---

## MONITOR Workspace Task Item

Run from the repository root. Self-contained, `agent: "task"`, read-only
except for writing `.bob/state/monitor.md`.

### `MonitorPR` — inspect CI and PR feedback

```text
name: MonitorPR
agent: task
task: |
  Work in the bob repo root. Read `.bob/state/monitor-prompt.md` with `read`
  for the PR to monitor (see `.bob/state/commit.md` for the URL/number).

  Step 1 — PR info: `gh pr view --json number,url,title,state,isDraft` (or by
  number if provided). Extract number, URL, title, state, draft.

  Step 2 — CI/CD: `gh pr checks --json name,status,conclusion,detailsUrl` and
  `gh run list --branch $(git branch --show-current) --limit 5 --json status,conclusion,name,databaseId`.
  Classify: all completed success = passing; any in_progress/queued = running;
  any failure/cancelled/timed_out/action_required = failing.

  Step 3 — Reviews: `gh pr view --json reviews`, plus
  `gh api repos/{owner}/{repo}/pulls/{pr_number}/comments` and
  `gh pr view --json reviewDecision`. Note APPROVED / CHANGES_REQUESTED /
  REVIEW_REQUIRED / null.

  Step 4 — Comments/threads: `gh pr view --json comments` and
  `gh api repos/{owner}/{repo}/pulls/{pr_number}/reviews` and
  `gh pr view --json reviewThreads`. Look for unresolved/blocking threads.

  Step 5 — Status checks detail: `gh api repos/{owner}/{repo}/commits/$(git rev-parse HEAD)/status` and
  `.../commits/$(git rev-parse HEAD)/check-runs`.

  Step 6 — Analyze health per this routing:
  if state == MERGED → COMPLETE;
  elif (ci_failures OR changes_requested OR blocking_comments) → BRAINSTORM
    (always re-brainstorm; never loop to EXECUTE/REVIEW);
  elif (all checks passed AND approved AND no unresolved threads) → COMPLETE;
  elif checks in progress → WAIT;
  else → WAIT (awaiting reviews).

  Step 7 — Write `.bob/state/monitor.md` with `write`: ISO timestamp, status
  READY | WAITING | NEEDS_WORK (or BLOCKED), PR details, a CI checks table
  (name/status/conclusion/details), review status, comments, a routing decision
  with reasoning, and a For Orchestrator block:
  `STATUS: [READY|NEEDS_WORK|WAITING]`, `NEXT_PHASE: [COMPLETE|BRAINSTORM|WAIT]`,
  `WAIT_TIME:` (when WAIT), `ISSUES_COUNT:`.
  On gh command failure, report the operational error (auth/repo/rate-limit) as
  NEEDS_WORK-or-FAILED detail rather than assuming readiness.
solutionSpace: |
  Straight port of the monitor-agent flow. Classification rules and routing are
  fixed; only reading live gh data and filling the report are open.
```

---

## Return to the Calling Workflow

Write `.bob/state/ship-status.md` with `write`, including branch, commit, PR
URL, CI state, and reason:

- `COMPLETE` — publication succeeded and monitoring is `READY`.
- `NEEDS_BRAINSTORM` — CI failed or changes requested (report the specific
  failures/feedback so the caller can address them).
- `FAILED` — operational failure or declined publication (report what actually
  reached the remote; for a decline, this invocation is terminal — surface it
  and stop without retrying or merging).

Never write `.bob/state/code-review-status.md`; that belongs to the review
skill. Return any needed code changes and hook failures to the caller for
verification and review. Keep the review skill's status intact. Report the
result. Merge only with the caller's final user approval.