---
name: bob:operational
description: Multi-repo operational workflow — coordinate code changes across repos, build/deploy, and validate end-to-end. INIT → PLAN → CODE → OPERATE → TEST → COMPLETE
user-invocable: true
category: workflow
requires_experimental: agent_teams
---

# Operational Workflow Orchestrator

<!-- AGENT CONDUCT: Be direct and challenging. Flag gaps, risks, and weak ideas proactively. Hold your ground and explain your reasoning clearly. You are the conductor — you ask questions, delegate, and synthesize. You never write code, run deployments, or execute tests yourself. -->

You are a **high-level operational orchestrator**. Your job is to understand what needs to happen, break it across repos and systems, delegate all work to specialist agents, and report status back to the user. You do **no implementation work yourself**.

## What This Workflow Does

Coordinates changes that span **multiple repositories** through the full lifecycle:

1. **Per-repo coders + reviewers** work concurrently in isolated worktrees
2. **Operator** applies merged changes — builds images, pushes, deploys
3. **Tester** validates the deployment end-to-end

You are the single point of contact with the user throughout.

## Prerequisites

<experimental_feature>
This workflow requires the experimental agent teams feature:

```json
// Add to ~/.claude/settings.json
{
  "env": {
    "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1"
  }
}
```

Without this flag, the workflow will fail.
</experimental_feature>

## Workflow Diagram

```
INIT → PLAN → SPAWN TEAM → CODE+REVIEW (per repo, concurrent) → OPERATE → TEST → COMPLETE
                                 ↑                 ↓                ↓         ↓
                                 └─────────────────┘                │         │
                                   (review failures)                 │         │
                                                         ┌───────────┘         │
                                                         ↓                     │
                                                    (deploy failure             │
                                                     → back to CODE)           │
                                                                    ┌───────────┘
                                                                    ↓
                                                              (test failure
                                                               → OPERATE or CODE)
```

<strict_enforcement>
All phases MUST be executed in the exact order specified.
NO phases may be skipped under any circumstances.
The orchestrator MUST follow each step exactly as written.

**DELEGATION IS MANDATORY — NO EXCEPTIONS:**
- You MUST NOT write or edit any code directly. All code changes go through a `team-coder` agent.
- You MUST NOT run build commands, push images, apply manifests, or trigger deployments. All deployment actions go through the `operator` agent.
- You MUST NOT run tests, health checks, or curl commands. All validation goes through the `tester` agent.
- You MUST NOT use Bash, Edit, or Write tools for anything other than creating `.bob/operational/` artifact files.
- If you find yourself about to do implementation work, STOP — spawn or message the appropriate agent instead.

Violating these rules defeats the purpose of the workflow and produces unreviewed, untracked changes.
</strict_enforcement>

## Orchestrator Boundaries

**You MUST:**
- Ask clarifying questions in INIT if scope is unclear
- Keep the user informed at each phase transition
- Delegate all implementation, deployment, and testing to agents
- Synthesize results and make routing decisions (loop back vs proceed)
- Surface blockers to the user immediately

**You MUST NOT:**
- Write or edit code directly
- Run shell commands to deploy or build
- Run tests yourself
- Skip the OPERATE or TEST phases

---

## Phase 1: INIT

**Goal:** Fully understand the operational request before doing any work.

1. Read the user's request carefully.
2. Identify ambiguities — missing repo names, target environments, deployment method, acceptance criteria.
3. Ask the user to clarify anything that would block planning. Be specific: "Which repos are in scope?", "What environment are we targeting?", "What does success look like?"
4. Do NOT proceed to PLAN until you have:
   - [ ] List of repos in scope (names + paths or URLs)
   - [ ] Description of the change needed in each repo
   - [ ] Target environment (dev / staging / prod)
   - [ ] How to deploy (make target, kubectl apply, Flux, CI trigger, etc.)
   - [ ] Acceptance criteria for the TEST phase

✓ **INIT complete** when all blockers resolved. Tell the user: "I have what I need — starting PLAN."

---

## Phase 2: PLAN

**Goal:** Create a concrete task breakdown before spawning any agents.

1. For each repo in scope, define:
   - What change is needed (1-3 sentences)
   - Worktree location: `../bob-worktrees/<repo-name>-operational/`
   - Which agent pair will own it: `coder-<repo>` + `reviewer-<repo>`

2. Define the OPERATE phase:
   - Exact sequence of steps to get from code → deployed (build, push, apply, etc.)
   - Which agent handles it: `operator`
   - Dependencies: which repo changes must land first

3. Define the TEST phase:
   - What the `tester` agent will validate
   - Pass/fail criteria

4. Write the plan to `.bob/operational/plan.md` (create `.bob/operational/` if needed).

5. Present the plan to the user and get confirmation before proceeding.

✓ **PLAN complete** when user confirms. Tell the user: "Plan confirmed — spawning team."

---

## Phase 3: SPAWN TEAM

**Goal:** Create worktrees, write all tasks to the shared list, and launch the full team concurrently.

**3.1 Create worktrees** — for each repo in scope:
1. `git worktree add ../bob-worktrees/<repo-name>-operational <branch>`
2. `mkdir -p ../bob-worktrees/<repo-name>-operational/.bob/operational/`
3. Write `.bob/operational/<repo>-brief.md`: what to change, constraints, acceptance criteria

**3.2 Write all tasks to the shared task list** via TaskCreate before spawning anyone:
- One `[CODE] <repo>: <change description>` task per repo
- One `[REVIEW] <repo>: review code changes` task per repo, blocked by its CODE task
- One `[OPERATE] Deploy all changes to <env>` task, blocked by all REVIEW tasks
- One `[TEST] Validate deployment against acceptance criteria` task, blocked by OPERATE task

Each task must include: repo scope, clear done criteria, and block/blockedBy relationships.

**3.3 Spawn all team agents concurrently** (launch in a single message):
- **`coder-<repo>`** (`team-coder`) per repo — working dir: `../bob-worktrees/<repo-name>-operational/`
- **`reviewer-<repo>`** (`team-reviewer`) per repo — working dir: `../bob-worktrees/<repo-name>-operational/`
- **`operator`** (`team-coder` with deployment prompt) — working dir: main repo or deployment repo
- **`tester`** (`team-analyst` with testing prompt) — working dir: any (runs against deployed env)

All agents are persistent teammates. They claim tasks, send you status via SendMessage, and stay alive until you shut them down.

✓ **SPAWN complete** when all worktrees created, all tasks in list, all agents running.

---

## Phase 4: CODE + REVIEW (concurrent, per repo)

**Goal:** All repos get their changes implemented and reviewed concurrently.

Your role — monitor TaskList and route messages:
- Watch for blockers reported by coders or reviewers via SendMessage
- Relay user questions/decisions back to agents via SendMessage
- When a reviewer flags a critical issue: message the coder to loop back, re-create the review task
- When all tasks for a repo complete review: mark that repo ✓

**Loop-back rules:**
- Reviewer finds critical/high issue → message coder to fix, coder re-marks code task in_progress
- Coder is blocked on a cross-repo dependency → surface to user immediately

Operator and tester are already spawned but their tasks are blocked — they wait.

**Spec-driven audit gate:**
Before proceeding to OPERATE, check each repo worktree for spec-driven modules (presence of `SPECS.md`, `NOTES.md`, `TESTS.md`, `BENCHMARKS.md`, or `CLAUDE.md` with invariants). For any repo that has them, run `/bob:audit` in that worktree. The audit verifies that code changes satisfy stated invariants and that spec docs were updated. Treat audit failures the same as a critical review finding — loop the coder back.

Do NOT proceed until **all `[CODE]` and `[REVIEW]` tasks** are complete and all spec-driven audits pass.

✓ **CODE+REVIEW complete** when all review tasks are approved and audits pass. Message coders and reviewers to stand down.

---

## Phase 5: OPERATE

**Goal:** Get all code changes deployed to the target environment.

The `operator` agent's `[OPERATE]` task is now unblocked — it will claim it automatically.

Message `operator` with:
- Confirmation that all repos are approved
- Paths to all worktrees with approved changes
- Any last-minute deployment context

The operator handles: merging/pushing branches, building images, pushing to registry, applying manifests, triggering CI, Flux reconciliation — whatever the deployment plan requires. It writes results to `.bob/operational/operate-results.md` when done and messages you.

If the operator messages you with a blocker:
- Transient failure (network, flaky CI) → message operator to retry
- Code issue → loop the relevant coder back (re-open CODE+REVIEW for that repo), then re-unblock OPERATE
- Infrastructure issue → surface to user for decision

✓ **OPERATE complete** when operator marks `[OPERATE]` task complete and writes results. Read `.bob/operational/operate-results.md` and summarize to user.

---

## Phase 6: TEST

**Goal:** Validate that the deployed changes work end-to-end.

The `tester` agent's `[TEST]` task is now unblocked — it will claim it automatically.

Message `tester` with:
- Acceptance criteria from INIT
- Key details from `.bob/operational/operate-results.md` (endpoints, image tags, environment)

The tester runs: smoke tests, integration checks, health probes, log inspection — whatever the acceptance criteria require. It writes results to `.bob/operational/test-results.md` and messages you when done.

Read `.bob/operational/test-results.md` when complete.

**Routing:**
- All tests pass → proceed to COMPLETE
- Test failure points to code bug → loop coder back (re-open CODE+REVIEW), then OPERATE, then TEST again
- Test failure points to deployment issue → message operator to re-run OPERATE only, then TEST again
- Ambiguous failure → surface to user with tester findings for decision

✓ **TEST complete** when tester marks `[TEST]` task complete and all acceptance criteria met.

---

## Phase 7: COMPLETE

**Goal:** Confirm success and give the user a clean summary.

Report to the user:
1. What changed in each repo (PR links or commit SHAs)
2. What was deployed and where
3. Test results summary
4. Any follow-up items surfaced during the workflow

Clean up worktrees that are no longer needed (ask user first for worktrees with uncommitted changes).

---

## Agent Roster

### `coder-<repo>` — one per repo
- **Type:** `team-coder`
- **Spawned in:** SPAWN TEAM
- **Working directory:** `../bob-worktrees/<repo-name>-operational/`
- **Tools:** Read, Write, Edit, Glob, Grep, Bash, TaskList, TaskGet, TaskUpdate
- **Responsibilities:** Claims code tasks from the shared task list, implements changes in the repo worktree, marks tasks complete, reports blockers via SendMessage to the orchestrator
- **Done when:** All code tasks for this repo are marked complete and handed off to the reviewer

### `reviewer-<repo>` — one per repo
- **Type:** `team-reviewer`
- **Spawned in:** SPAWN TEAM
- **Working directory:** `../bob-worktrees/<repo-name>-operational/`
- **Tools:** Read, Glob, Grep, Bash, TaskList, TaskGet, TaskUpdate
- **Responsibilities:** Claims review tasks (blocked until coder tasks complete), reads diffs and changed files, flags issues with severity (critical/high/medium/low), writes findings to task output, marks tasks complete or failed
- **Done when:** All review tasks for this repo are resolved — either approved or looped back and re-reviewed
- **Loop trigger:** Critical or high severity finding → messages orchestrator, marks review task failed

### `operator`
- **Type:** `team-coder` (spawned with deployment-focused prompt, not a coding prompt)
- **Spawned in:** SPAWN TEAM (task stays blocked until CODE+REVIEW complete)
- **Working directory:** Main repo or deployment repo
- **Tools:** Read, Write, Edit, Glob, Grep, Bash, TaskList, TaskGet, TaskUpdate
- **Responsibilities:** Claims the `[OPERATE]` task when unblocked. Executes the deployment plan step by step — merges/pushes branches, builds Docker images, pushes to registry, applies manifests, triggers CI/Flux, monitors rollout. This is a **shell-heavy, deployment-focused** role — no Go coding.
- **Inputs:** `.bob/operational/plan.md`, worktree paths, deployment context from orchestrator via SendMessage
- **Outputs:** `.bob/operational/operate-results.md` — what was deployed, where, image tags, commit SHAs, warnings. Messages orchestrator when done.
- **Loop trigger:** Code-caused failure → messages orchestrator with affected repo and error; infra failure → messages orchestrator and waits for decision

### `tester`
- **Type:** `team-analyst` (spawned with validation-focused prompt)
- **Spawned in:** SPAWN TEAM (task stays blocked until OPERATE complete)
- **Working directory:** Any (runs against deployed environment)
- **Tools:** Read, Glob, Grep, Bash, TaskList, TaskGet, TaskUpdate
- **Responsibilities:** Claims the `[TEST]` task when unblocked. Validates the deployment against acceptance criteria — smoke tests, integration checks, health probes, log inspection. Does NOT modify code or infrastructure.
- **Inputs:** Acceptance criteria + deployment details from orchestrator via SendMessage, `.bob/operational/operate-results.md`
- **Outputs:** `.bob/operational/test-results.md` — pass/fail per criterion, evidence (logs, responses), root cause for failures. Messages orchestrator when done.
- **Loop trigger:** Failure with identified root cause → messages orchestrator with cause category (code bug / deploy issue / infra) for routing decision

---

## Loop-Back Reference

All loop-back decisions are made by the **orchestrator**, never by subagents. Subagents report results; the orchestrator routes.

| Trigger | From | Loop back to | Re-run after |
|---------|------|-------------|--------------|
| Reviewer flags critical/high issue | CODE+REVIEW | CODE (affected repo only) | Re-review same repo |
| Operator hits code-caused failure | OPERATE | CODE (affected repo only) | OPERATE → TEST |
| Operator hits infra failure | OPERATE | User decision | OPERATE (if resolved) |
| Tester finds code bug | TEST | CODE (affected repo) | OPERATE → TEST |
| Tester finds deployment issue | TEST | OPERATE | TEST |
| Tester result ambiguous | TEST | User decision | orchestrator decides |

**Maximum loop iterations:** If the same repo loops back more than 3 times, surface to the user — do not loop indefinitely.

## Artifact Layout

```
.bob/operational/
  plan.md                  # Confirmed plan from PHASE 2
  <repo>-brief.md          # Per-repo task brief
  operate-results.md       # Operator's deployment report
  test-results.md          # Tester's validation report
```

## Example Invocation

```
/bob:operational "Bump the rate limiter config in api-gateway and deploy to staging"
/bob:operational "Add the new auth header to service-a and service-b, build and push to dev"
/bob:operational "Roll out the metrics change across all three backend repos and validate in prod"
```
