
## Workflow Mode: Local

<workflow_override>
This installation uses LOCAL workflow mode. The COMMIT, MONITOR, and COMPLETE phases
are overridden as described below. These overrides take precedence over the default
phase instructions.

**CRITICAL RULE — You MUST ask the user what to do after committing.**
Do NOT autonomously push, create PRs, or merge branches. The user decides.
</workflow_override>

### COMMIT Phase Override (LOCAL mode)

When you reach the COMMIT phase, do NOT automatically push or create a PR.

Instead:

1. Verify `.bob/state/review.md` exists (same prerequisite as default)
2. Spawn commit-agent to stage and commit locally ONLY:
   ```
   Task(subagent_type: "commit-agent",
        description: "Commit changes locally",
        run_in_background: true,
        prompt: "1. Verify .bob/state/review.md exists
                 2. Run git status and git diff
                 3. Stage relevant files (never git add -A)
                 4. Create commit with descriptive message
                 5. DO NOT push. DO NOT create a PR.
                 Working directory: [worktree-path]")
   ```

**⚠️ MANDATORY — Do NOT proceed without asking the user.**

3. After the commit completes, you **MUST** use `AskUserQuestion` to ask the user
   what they want to do next. Do NOT skip this step. Do NOT assume an answer.
   Present these options:
   - **Submit a PR** — Push the branch and create a PR via `gh pr create`, then proceed to MONITOR as normal
   - **Merge into parent branch** — Merge the feature branch into the base branch locally, skip MONITOR, proceed to COMPLETE

   Wait for the user's response before taking any further action.

### MONITOR Phase Override (LOCAL mode)

If the user chose "Merge into parent branch" in the COMMIT phase, SKIP the MONITOR
phase entirely (there is no PR to monitor). Proceed directly to COMPLETE.

If the user chose "Submit a PR", run the MONITOR phase as normal.

### COMPLETE Phase Override (LOCAL mode)

If the user chose "Merge into parent branch":

1. Spawn a Bash agent to merge the worktree branch into the parent branch:
   ```
   Task(subagent_type: "Bash",
        description: "Merge into parent branch",
        run_in_background: true,
        prompt: "Merge the current feature branch into its parent branch:
                 1. FEATURE_BRANCH=$(git branch --show-current)
                 2. Determine the parent/base branch (usually main or master)
                 3. cd to the main repo (git worktree list to find it)
                 4. git merge $FEATURE_BRANCH
                 5. Report success or merge conflicts")
   ```
2. If merge succeeds, celebrate.
3. If merge conflicts, inform the user and let them resolve manually.

If the user chose "Submit a PR", run the COMPLETE phase as normal (offer to merge the PR).
