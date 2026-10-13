# Publication, Approval, and Recovery

Read this before any commit, push, or PR mutation. It preserves the existing
publication control procedure. Translate the worker examples to the harness's
native tools; they describe job inputs and result contracts. When a calling
workflow requires single-agent execution, perform the jobs directly using the
installed commit-agent and monitor-agent guidance without spawning workers.
Every direct or delegated push uses the commit-agent publication gate, including
retries. Missing required guidance or gate tooling blocks publication.

## COMMIT

**Goal:** Commit the reviewed and fixed code; when confirm-before-push is
enabled, pause for the user's approval before anything is published.

**Actions:**

1. Check the confirm flag with a literal command — the printed value is the
   decision input; never assert the flag's state from memory:
   ```bash
   echo "confirm-before-push: ${BOB_CONFIRM_BEFORE_PUSH:-unset}"
   ```
   Confirm mode is ON only when the command prints `confirm-before-push: 1`.

2. Write `.bob/state/commit-prompt.md`:
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

3. **Confirm mode OFF** — first check for a prepared commit left by an
   earlier confirmation pause (including one abandoned or declined —
   disabling confirmation supersedes a prior stop): if
   `.bob/state/commit.md` reads `STATUS: AWAITING_CONFIRMATION`, its BRANCH
   and HEAD match live `git branch --show-current` and `git rev-parse HEAD`,
   and the tree is clean apart from `.bob/state` (same root-scoped status
   command as 4a), append this line to `.bob/state/commit-prompt.md`:
   "An earlier prepare pass already created commit [HEAD] on this branch; do
   not create a new commit — publish it: push, create the PR, and delete
   .bob/state/pr-body.md and .bob/state/pr-title.txt once the PR exists."
   Otherwise delete any stale `.bob/state/pr-body.md` and
   `.bob/state/pr-title.txt` (a mismatched `commit.md` record is ignored and
   does not activate this resume arm). A `FAILED` record — whether from a
   `CONFIRM_MODE: PUBLISH` run or an earlier flag-off publication attempt —
   does not activate this arm; recovery with confirmation off is manual.
   Then spawn commit-agent as before:
   ```
   subagent({
  agent: "commit-agent",
  task: "Read .bob/state/commit-prompt.md for instructions.
                Create commit, push branch, create PR.
                Write status to .bob/state/commit.md.",
  context: "fresh"
})
   ```

   After completion, read `.bob/state/commit.md`:
   - STATUS: SUCCESS → move to MONITOR
   - STATUS: FAILED → report failure and exit with STATUS: FAILED

4. **Confirm mode ON** — two passes, with the user's decision between them:

   a. Resume check: if `.bob/state/commit.md` already reads
      `STATUS: AWAITING_CONFIRMATION`, its BRANCH and HEAD match live
      `git branch --show-current` and `git rev-parse HEAD`, the tree is clean
      apart from `.bob/state` — the status command, literally excluding the
      state directory where INIT created it (`--show-prefix` is captured
      before entering the root; it is empty when already there):
      `(P=$(git rev-parse --show-prefix) && cd "$(git rev-parse --show-toplevel)" && git status --porcelain -- . ":(exclude,literal)${P}.bob/state")`
      exits successfully and prints nothing — and both state files are
      readable
      (`test -r .bob/state/pr-body.md && test -r .bob/state/pr-title.txt`
      exits successfully), skip to (c): an earlier paused run is being
      resumed. If commit.md instead reports `STATUS: FAILED` with
      `PUSHED: yes` and `PR_CONFIRMED: no`, its BRANCH and HEAD still match
      live state, the tree is clean apart from `.bob/state` (same root-scoped
      status command as above), and both state files are readable (same
      `test -r` command), skip to (c) — on `push`, the publish pass re-runs
      the push step (safe to repeat), then continues at Step 6 to create or
      update the PR. Likewise if commit.md reports a publish FAILURE whose
      reason is a publication gate block, with BRANCH and HEAD still matching
      live state, the tree clean apart from `.bob/state` (same root-scoped
      status command as above), and both state files readable (same
      `test -r` command): skip to (c) — on `push`, the publish pass re-runs
      the gated push. Otherwise delete any stale `.bob/state/pr-body.md` and
      `.bob/state/pr-title.txt` before continuing — a tree with changes
      outside `.bob/state` never resumes: it falls through to (b), so the
      fresh prepare pass commits the new work and the preview covers it.

   b. Spawn the prepare pass:
      ```
      subagent({
  agent: "commit-agent",
  task: "CONFIRM_MODE: PREPARE
                Read .bob/state/commit-prompt.md for instructions.
                Create the commit, write the proposed PR body to
                .bob/state/pr-body.md and the proposed PR title to
                .bob/state/pr-title.txt, but do NOT push and do NOT create
                a PR. Write status to .bob/state/commit.md.",
  context: "fresh"
})
      ```
      Read `.bob/state/commit.md`: STATUS: FAILED → report and exit FAILED;
      any status other than AWAITING_CONFIRMATION → treat as FAILED and say the
      pause did not happen.

   c. Present the preview mechanically — never summarize or restate it.
      First show every commit the push will publish — run
      `git log --oneline HEAD --not --remotes=origin` and present its output
      verbatim (the push publishes the branch ref, so unpushed ancestors ship
      with it; the list makes that visible). Then show the commit details from
      `.bob/state/commit.md` (branch, SHA, message, files; for a
      publish-failure report, show every field it carries, including its
      publication state), the PR title, and the body via:
      ```bash
      cat .bob/state/pr-body.md
      ```
      Then ask exactly:
      `Push this branch and create or update its PR with the title and body shown above? [push / stop]`

   d. Route on the reply:
      - Exactly `push` → spawn the publish pass:
        ```
        subagent({
  agent: "commit-agent",
  task: "CONFIRM_MODE: PUBLISH
                Branch: [BRANCH from commit.md]
                Commit SHA: [HEAD from commit.md]
                PR title: [TITLE from commit.md]
                Verify the branch and SHA match live state, then push and
                create the PR with the approved title and body files. Do NOT
                create a commit. Write status to .bob/state/commit.md.",
  context: "fresh"
})
        ```
        Read `.bob/state/commit.md`: SUCCESS → move to MONITOR; FAILED → relay
        the reported failure reason in one sentence and exit FAILED (a moved
        HEAD means: re-run /bob:ship to attempt a fresh preview; the
        fresh prepare pass re-presents only work a confirm-mode pass recorded,
        and a clean live commit with no such record reports "nothing to
        commit").
      - Exactly `stop` → delete `.bob/state/pr-body.md` and
        `.bob/state/pr-title.txt`, then exit with
        STATUS: FAILED, reason "user declined publication; branch retained at
        [sha]; approved PR creation or update was not confirmed" (add "the
        branch was already pushed by an earlier attempt" when commit.md's For
        Orchestrator PUSHED field is yes — never claim nothing was pushed in
        that case). Never continue to MONITOR or COMPLETE.
      - Anything else → ask the question again (repeat only the question, not
        the preview).

---

## MONITOR

**Goal:** Watch CI and PR feedback.

**Actions:**

1. Write `.bob/state/monitor-prompt.md`:
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

2. Spawn monitor-agent:
   ```
   subagent({
  agent: "monitor-agent",
  task: "Read .bob/state/monitor-prompt.md for instructions.
                Check PR status, CI checks, review feedback.
                Write full status report to .bob/state/monitor.md.",
  context: "fresh"
})
   ```

3. After completion, read `.bob/state/monitor.md` and route:

   | Monitor Status | Action |
   |----------------|--------|
   | READY (all green) | → COMPLETE |
   | WAITING (CI in progress) | → Re-spawn monitor-agent after a pause |
   | NEEDS_WORK (CI failed or changes requested) | → Exit with NEEDS_BRAINSTORM |

---

## Return to the Calling Workflow

Write `.bob/state/ship-status.md` with status, reason, branch, commit, PR URL,
and CI state. `COMPLETE` means publication succeeded and monitoring is READY.
`NEEDS_BRAINSTORM` reports CI failures or requested changes to the caller.
`FAILED` reports an operational failure or declined publication. A decline is
terminal for this invocation: surface it and stop without retrying or merging.
Never write `.bob/state/code-review-status.md`; that belongs to the review skill.
