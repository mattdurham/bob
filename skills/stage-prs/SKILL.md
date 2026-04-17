---
name: bob:stage-prs
description: Break a large changeset into ordered, reviewable PRs — idempotent, call again to advance the stack after merges
user-invocable: true
category: workflow
---

# Stage PRs Workflow Orchestrator

You orchestrate splitting a large changeset into a series of **small, ordered, reviewable pull requests**. Each PR targets ≤200 lines changed (excluding tests and docs). PRs may depend on each other and are stacked (each PR's base is the previous PR's branch).

**This workflow is idempotent** — call it multiple times. On each invocation it checks for an existing plan and PR states first, advancing the stack if the frontier PR has merged, before falling through to plan/create new PRs if needed.

You are a **pure orchestrator** — you never write code or run git commands yourself beyond scoping queries.

## Workflow Diagram

```
RESUME ──── no plan found ──────────────────→ ANALYZE → PLAN → CONFIRM → EXECUTE → COMPLETE
   │
   ├── plan found, PRs pending/open ────────→ report status, exit
   │
   └── plan found, frontier PR merged ──────→ ADVANCE → report status, exit
                                                  │
                                                  └── more PRs to advance → loop ADVANCE
```

---

## Orchestrator Boundaries

**You ONLY:**
- ✅ Run `git` and `gh` read-only commands to scope the changeset and check PR status
- ✅ Spawn subagents via Task tool
- ✅ Write `.bob/state/stage-prs-*.md` files
- ✅ Present the plan and status to the user
- ✅ Route based on user response and subagent output

**You NEVER:**
- ❌ Write or edit source code files
- ❌ Run `git commit`, `git push`, `gh pr create`, `git rebase` directly
- ❌ Make grouping decisions without first consulting the Explore agent
- ❌ Skip the CONFIRM phase on a fresh plan — the user must approve before PRs are created

---

## Execution Rules

- All subagents MUST run with `run_in_background: true`
- After spawning, STOP — wait for completion notification before proceeding
- The CONFIRM phase is a hard pause: present the plan, then wait for user input

---

## Phase 0: RESUME

**Goal:** Detect an existing plan and check the current state of the PR stack. This phase runs on EVERY invocation before anything else.

**Actions:**

1. Check for an existing plan:
   ```bash
   ls .bob/state/stage-prs-plan.md 2>/dev/null
   ```

2. **If no plan exists** → fall through to Phase 1 (ANALYZE). Done with this phase.

3. **If a plan exists**, read `.bob/state/stage-prs-plan.md` and extract:
   - All PRs in the series (title, branch, base, PR URL if present)
   - Their order

4. Check the status of every PR that has a URL by running for each:
   ```bash
   gh pr view <url> --json number,title,state,mergedAt,baseRefName,headRefName
   ```
   States are: `OPEN`, `MERGED`, `CLOSED`.

5. Build a mental state table:

   | # | Title | Branch | State | Merged? |
   |---|-------|--------|-------|---------|
   | 1 | ... | ... | MERGED | ✓ |
   | 2 | ... | ... | OPEN | — |
   | 3 | ... | ... | not created | — |

6. **Routing:**

   | Condition | Action |
   |-----------|--------|
   | All PRs merged | → Report "Stack complete, all PRs merged." and exit |
   | All created PRs open, none merged | → Report current status table and exit |
   | Some PRs not yet created | → Report status + note which haven't been created, exit |
   | Latest merged PR's *successor* has not been rebased yet | → Go to ADVANCE phase |
   | No change since last run | → Report status and exit |

   **How to detect "needs advance":** The PR immediately after the last merged PR should have its branch rebased onto `main` (or the current merge target). Check:
   ```bash
   # Is the next PR's branch already based on current main?
   git fetch origin
   git merge-base --is-ancestor origin/main origin/<next-branch>
   # Exit 0 = main is ancestor (already up to date)
   # Exit 1 = needs rebase
   ```

---

## Phase 0a: ADVANCE

**Goal:** Rebase the next PR's branch onto main after the previous PR merged, producing a clean diff.

**Context:** GitHub auto-retargets stacked PRs to `main` when their base PR merges. But the branch itself still contains the previous PR's commits. A rebase removes those commits and makes the diff accurate.

**Actions:**

1. Identify the "next" PR — the first PR in the series whose state is `OPEN` or `not created` and whose predecessor is `MERGED`.

2. Write `.bob/state/stage-prs-advance-N.md`:
   ```markdown
   # Advance PR N: <title>

   ## Situation
   PR <N-1> ("<previous title>") has merged into `main`.
   PR <N> ("<this title>") branch `<branch>` needs to be rebased onto `main`
   so that its diff is clean (no carry-over commits from PR <N-1>).

   ## Steps

   1. Fetch latest main:
      ```bash
      git fetch origin
      ```

   2. Rebase this PR's branch onto main:
      ```bash
      git rebase origin/main origin/<branch> --onto origin/main
      # or if already checked out:
      git checkout <branch>
      git rebase origin/main
      ```

   3. If rebase conflicts:
      - Note which files conflict
      - If conflicts are simple (the merged PR's changes no longer exist), resolve by accepting current changes
      - If conflicts are complex, abort (`git rebase --abort`) and report back with details

   4. Force-push the rebased branch:
      ```bash
      git push --force-with-lease origin <branch>
      ```

   5. Verify the PR base is now `main` (GitHub should have auto-retargeted):
      ```bash
      gh pr view <pr-url> --json baseRefName
      ```
      If base is still the old branch (not yet auto-retargeted by GitHub):
      ```bash
      gh pr edit <pr-url> --base main
      ```

   6. Write status back to this file:
      ```markdown
      STATUS: COMPLETE
      Branch rebased: <branch>
      PR URL: <url>
      New base: main
      ```
      Or on failure:
      ```markdown
      STATUS: CONFLICT
      Conflicting files: [list]
      Rebase aborted. Manual resolution required.
      ```

   ## PR URL
   <url>

   ## Feature branch (full stack origin, for reference)
   <original-feature-branch>
   ```

3. Spawn commit-agent to perform the rebase and force-push:
   ```
   Task(subagent_type: "commit-agent",
        description: "Advance stack: rebase PR N onto main",
        run_in_background: true,
        prompt: "Read `.bob/state/stage-prs-advance-N.md` for full instructions.
                Rebase the branch onto main and force-push. Update PR base if needed.
                Write status back to that file.")
   ```

4. After completion, read `.bob/state/stage-prs-advance-N.md`:
   - `STATUS: COMPLETE` → report to user: "✓ PR N (<title>) rebased onto main. PR: <url>"
   - `STATUS: CONFLICT` → report conflict details to user, exit. Do not attempt further advances.

5. After a successful advance, check if the *next* PR in the series also needs advancing (i.e., PR N+1's predecessor is now considered "merged + rebased"). If yes, loop back to ADVANCE for PR N+1.

6. Once no more advances are needed, report the full current status table and exit.

---

## Phase 1: ANALYZE

**Goal:** Understand the full scope of the changeset.

**Actions:**

1. Ensure working tree is clean (or warn user):
   ```bash
   git status --short
   ```
   If there are uncommitted changes, warn the user: staged PRs require committed changes. Ask if they want to continue (skipping uncommitted changes) or stop to commit first.

2. Determine base branch (default: `main` or `master`; accept as argument):
   ```bash
   git symbolic-ref refs/remotes/origin/HEAD --short 2>/dev/null || echo "origin/main"
   ```

3. Collect the changeset statistics (excluding tests and docs):
   ```bash
   # All changed files with line counts
   git diff <base>...HEAD --stat

   # File list only
   git diff <base>...HEAD --name-only

   # Line counts per file (additions + deletions), raw
   git diff <base>...HEAD --numstat
   ```

4. Identify test and doc files to exclude from the 200-line budget:
   - Test files: `*_test.go`, `**/*_test.go`, `**/testdata/**`, `**/fixtures/**`
   - Doc files: `*.md`, `*.rst`, `*.txt`, `docs/**`, `*.yaml` (config-only), `*.json` (config-only)
   - Test and doc files still go into PRs — they just don't count toward the 200-line limit

5. Compute net non-test/non-doc line changes per file.

6. Write `.bob/state/stage-prs-analysis.md`:
   ```markdown
   # Changeset Analysis

   Base branch: <base>
   Current branch: <branch>
   Total files changed: N
   Total lines changed (all): N
   Total lines changed (excl. tests/docs): N

   ## Files (sorted by logical area)

   | File | +lines | -lines | net | test/doc? |
   |------|--------|--------|-----|-----------|
   ...

   ## Commit history (base..HEAD)
   <git log --oneline output>
   ```

7. Move to PLAN phase.

---

## Phase 2: PLAN

**Goal:** Group files into ordered, logically coherent PRs under the 200-line limit.

**Actions:**

1. Write `.bob/state/stage-prs-plan-prompt.md`:
   ```markdown
   # PR Staging Plan Instructions

   Read `.bob/state/stage-prs-analysis.md` for the full changeset.

   ## Your Task

   Group the changed files into a series of ordered pull requests where:

   1. **Each PR is ≤200 lines changed** (count only non-test, non-doc files).
      Test and doc files travel with the code they test/document — they don't count toward the limit.
   2. **PRs are logically coherent** — each PR should represent a single, reviewable unit of work
      (e.g., "add config struct", "wire up HTTP handler", "add integration test + docs").
   3. **PRs are ordered by dependency** — if PR B depends on types or functions introduced in PR A,
      PR A must come first. A reviewer should be able to understand each PR in isolation.
   4. **Prefer smaller PRs** — if a file group is well under 200 lines, consider if another small
      logical unit could be added without exceeding the limit. But don't force unrelated things together.
   5. **Foundational changes first** — data types, interfaces, config structs should come before
      the code that uses them.

   ## Output Format

   Write your plan to `.bob/state/stage-prs-plan.md` using this format:

   ```markdown
   # Staged PR Plan

   Base branch: <base>
   Total PRs: N

   ## PR 1: <short title>

   **Branch:** `<suggested-branch-name>`
   **Base:** `<base branch or previous PR branch>`
   **Summary:** One sentence describing what this PR does and why it comes first.
   **Lines changed (excl. tests/docs):** N

   ### Files
   - `path/to/file.go` (+N/-N)
   - `path/to/file_test.go` (+N/-N) [test]
   - `docs/thing.md` (+N/-N) [doc]

   ## PR 2: <short title>
   ...
   ```

   If you cannot group the changeset under the 200-line limit per PR without creating too many
   tiny incoherent PRs (more than ~8), note this and suggest a slightly higher limit or an
   alternative grouping strategy.
   ```

2. Spawn Explore agent to produce the plan:
   ```
   Task(subagent_type: "Explore",
        description: "Plan staged PR groupings",
        run_in_background: true,
        prompt: "Read `.bob/state/stage-prs-plan-prompt.md` for full instructions.
                Read `.bob/state/stage-prs-analysis.md` for the changeset details.
                Read the actual file contents for any files where you need to understand
                logical relationships (imports, types used, function calls).
                Write the PR plan to `.bob/state/stage-prs-plan.md`.")
   ```

3. After completion, read `.bob/state/stage-prs-plan.md` and move to CONFIRM.

---

## Phase 3: CONFIRM

**Goal:** Present the plan to the user and get explicit approval before touching git.

**Actions:**

1. Read `.bob/state/stage-prs-plan.md` in full.

2. Present the plan to the user clearly:
   - Show each PR with its title, branch name, base, summary, line count, and file list
   - Highlight the stacking order (PR N bases on PR N-1)
   - Note any PRs near the 200-line limit

3. Ask the user:
   > "Does this PR breakdown look right? You can:
   > - **Approve** to proceed with creating branches and PRs
   > - **Adjust** to change groupings (describe what to move)
   > - **Cancel** to stop here (no git changes made)"

4. **WAIT for user response** before proceeding.

5. If **Approved**: move to EXECUTE.

6. If **Adjust requested**: update `.bob/state/stage-prs-plan.md` to reflect requested changes,
   then re-present the updated plan and ask again.

7. If **Cancelled**: write `.bob/state/stage-prs-status.md` with `STATUS: CANCELLED` and exit.

---

## Phase 4: EXECUTE

**Goal:** Create branches and PRs for each staged PR in order.

For each PR in the plan (in order):

### 4a. Create branch and apply changes

Write `.bob/state/stage-prs-execute-N.md` (where N is the PR number):
```markdown
# Execute PR N: <title>

## Branch
- Branch name: <branch>
- Base branch: <base or previous PR branch>

## Files to include
<list of files from plan>

## Instructions

1. From the base branch, create the PR branch:
   ```bash
   git checkout <base>
   git checkout -b <branch>
   ```

2. For each file in this PR, cherry-pick the final version from the feature branch:
   ```bash
   git checkout <feature-branch> -- <file1> <file2> ...
   ```

3. Stage and commit all files for this PR:
   ```bash
   git add <files>
   git commit -m "<conventional-commit message summarizing this PR>"
   ```

4. Push the branch:
   ```bash
   git push -u origin <branch>
   ```

5. Create PR targeting `<base-branch>`:
   ```bash
   gh pr create --base <base> --title "<title>" --body "..."
   ```
   PR body should:
   - Summarize what this PR does
   - If it's not the first PR, reference the PR it stacks on: "Stacks on #N"
   - Note that it's part of a series: "Part N of M in the <feature> staging series"

6. Write PR URL and status to `.bob/state/stage-prs-execute-N.md`:
   ```markdown
   STATUS: COMPLETE
   PR URL: <url>
   Branch: <branch>
   ```

## Feature branch (to cherry-pick from)
<feature-branch-name>
```

Spawn commit-agent:
```
Task(subagent_type: "commit-agent",
     description: "Create branch and PR N: <title>",
     run_in_background: true,
     prompt: "Read `.bob/state/stage-prs-execute-N.md` for full instructions.
             Follow the steps exactly. Write status and PR URL back to that file.")
```

After each PR completes:
- Read `.bob/state/stage-prs-execute-N.md` for the PR URL
- Report the PR URL to the user: "✓ PR N created: <url>"
- Proceed to the next PR

If a PR creation fails:
- Report the failure to the user immediately
- Ask whether to retry, skip, or abort the remaining PRs
- Do not proceed automatically on failure

---

## Phase 5: COMPLETE

**Goal:** Report the full series of PRs to the user.

Write `.bob/state/stage-prs-status.md`:
```markdown
# Stage PRs: Complete

STATUS: COMPLETE
Timestamp: <ISO timestamp>

## PR Series

| # | Title | Branch | Base | PR URL |
|---|-------|--------|------|--------|
| 1 | ... | ... | main | https://... |
| 2 | ... | ... | pr-1-branch | https://... |
...

## Merge Order

Merge in order: PR 1 → PR 2 → ... → PR N
Each PR should be merged before the next is reviewed/merged,
so that the base branch is correct when GitHub evaluates the diff.
```

Present the summary to the user with:
- The full PR series table
- A reminder to merge in order
- A note that as each stacked PR is merged, GitHub will automatically re-target the next PR to main

---

## State Files Reference

| File | Written By | Purpose |
|------|-----------|---------|
| `.bob/state/stage-prs-analysis.md` | Orchestrator | Changeset statistics and file list |
| `.bob/state/stage-prs-plan-prompt.md` | Orchestrator | Instructions for the Explore agent |
| `.bob/state/stage-prs-plan.md` | Explore agent | The PR grouping plan (persists across invocations) |
| `.bob/state/stage-prs-execute-N.md` | Orchestrator | Per-PR execution instructions + status |
| `.bob/state/stage-prs-advance-N.md` | Orchestrator | Per-PR advance (rebase) instructions + status |
| `.bob/state/stage-prs-status.md` | Orchestrator | Final completion status |

---

## Usage Examples

```
# Split current branch against main
/bob:stage-prs

# Split against a specific base
/bob:stage-prs --base origin/release-1.2
```

---

## Constraints and Edge Cases

**Merge commits / complex histories:**
If the feature branch has merge commits (not a linear history), note this in ANALYZE and warn
the user. The cherry-pick approach works best on linear histories; suggest `git rebase main`
before running this workflow if the history is tangled.

**Files that can't be split:**
If a single file has >200 non-test/doc lines changed, it must go in its own PR. Note this in the
plan and flag it as "oversized — consider breaking up the file if review feedback requires it."

**Already-stacked branches:**
If the current branch already stacks on another non-main branch, honor that as the base.

**Draft PRs:**
All PRs in the series should be created as **drafts** except the first one (or per user preference).
This prevents reviewers from starting on PR 2 before PR 1 is merged.
