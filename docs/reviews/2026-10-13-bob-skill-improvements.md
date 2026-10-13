# Bob skill improvements

Reviewed using [writing-for-agents](../../skills/bob-matt-writing-for-agents/SKILL.md) and
[codebase-design](../../skills/bob-matt-codebase-design/SKILL.md), with Matt's constraints:
a few paragraphs per skill, few branching paths, and an obvious purpose.

This covers all 23 Bob skill families and the repository's `talk-to-codex`
integration, including their simple and Goose variants. It builds on the
[complexity review](2026-10-13-skill-complexity-review.md) and accounts for the
27 newly imported skills. These are recommendations; the skill definitions
have not been rewritten.

## How the two skills inform this review

**Writing-for-agents:** Remove repeated rules, generic advice, stale examples,
and mandatory scripts of conversation. Keep a clear task and completion
criterion. Put substantial templates and conditional reference material in
supporting files that are actually installed.

**Codebase-design:** Apply the deletion test to every skill and coordinator.
If removing it removes only coordination, merge or retire it. If removing it
would scatter a useful capability across callers, keep that capability behind
a small interface. The interface includes side effects, configuration, and
failure behavior as well as the command name.

Shortening an entrypoint while preserving the same mandatory agents, stages,
artifacts, and routing rules would preserve most of the complexity. Reduce the
behavioral machinery first.

## Recommendations for every skill

### 1. `bob-work` — simplified

[Source](../../skills/bob-work/SKILL.md). Three paragraphs preserve existing-plan
intake or a confirmed one-question-at-a-time interview, an isolated worktree,
and a coding worklist. Implementation and bounded delegation follow ready tasks
and file ownership; optional early reviews use stable inputs and supplement the
final integrated review. Conflicting inputs require a recorded decision before
verification. Current checks and a clean read-only review precede `bob-ship`,
which retains publication approval, recovery, CI, and the final merge decision.
Mandatory phase agents, fixed loops, copied shell procedures, and routing tables
were removed. The simple variant stays direct in the current workspace and
commits or publishes only when requested; Goose uses bounded parent-owned jobs.

### 2. `bob-work-agents` — removed

The adaptive `bob-work` already provides isolated development, plan intake,
a coding worklist, and task-based delegation. Removed the duplicate mandatory
phase-agent protocol, installer entries, and active comparisons to it. Installation
retires known unchanged copies after a replacement installs, preserving custom
copies, links, and companion files.

### 3. `bob-work-teams` — merged into `bob-work` and removed

The coding worklist now supports optional reviews of completed assignments while
independent coding continues. Reviewers use stable inputs and return separate
reports; the parent routes fixes and waits for all workers before verification
and final integrated review. Changed inputs invalidate earlier reviews. Roster
size and native team use follow the work rather than a fixed team protocol.
Removed the separate skill and installer entries; known unchanged installations
retire after their replacement is present, preserving custom copies and resources.

### 4–5. `bob-explore` and `bob-explore-teams` — removed

Their useful exploration guidance now lives in `bob:matt-research`: trace relevant
structure and behavior, check contracts and assumptions against source evidence,
and report uncertainties. Research chooses direct or bounded delegated work and
writes one Markdown report. The fixed teams and phase machinery were removed.

### 6. `bob-cleanup` — simplified

[Source](../../skills/bob-cleanup/SKILL.md). Three paragraphs retain scoped,
behavior-preserving cleanup, contract and user-work preservation, relevant
verification, and review. Delegation depends on scope and file ownership rather
than a fixed team. Bug fixes that change behavior and disputed contracts require
a separate decision. Unresolved verification failures block completion. Commits
and publication require an explicit request; publication keeps `bob-ship` gates.
The simple variant performs the same work without delegation.

### 7. `bob-cleanup-teams` — removed

The simplified `bob-cleanup` covers the same behavior-preserving cleanup and
chooses delegation when useful. Removed the duplicated team protocol and all
installer entries. Installation retires known unchanged copies only after the
replacement is present; customized copies and companion files are preserved.

### 8. `bob-code-review` and `bob-ship` — separated and simplified

`bob-code-review` is read-only and produces findings, plans, or draft comments.
`bob-ship` publishes a changeset only after current passing verification and a
clean final review. Its three-paragraph entrypoint covers the prerequisites,
approval, publication, monitoring, and return status. Detailed preview, resume,
partial-failure, recovery, and approval procedures are preserved in
`references/publication-procedure.md`, required before publication mutations.
Direct execution or bounded workers respect the calling workflow's boundaries;
failed gates and declined publication cannot be bypassed. The separate
publication gate scripts and commit-agent controls are unchanged. The canonical
ship package and its reference also install in simple mode so direct callers
can apply the same controls without delegation.

### 9. `bob-adversarial-review` — keep as a focused deep review

[Source](../../skills/bob-adversarial-review/SKILL.md). A correctness investigation
of contracts, ownership, concurrency, and edge cases can add value beyond the
imported Standards/Spec review. Keep that purpose. Cut the compulsory eight
reviewers, assumption that every file contains a defect, severity inflation,
and artifact-existence completion rule. Use the simple variant's requirement
for demonstrable findings. Put domain checklists in references. Reproduction
test writing should be a separately requested follow-up to the read-only review.

**Proposed description:** Investigate a specified change or module for concrete
correctness defects and report verified failure modes with evidence.

### 10. `bob-audit` — keep invariant verification, remove automatic health scoring

[Source](../../skills/bob-audit/SKILL.md). Checking documented guarantees across
modules is a useful capability beyond reviewing a diff. Keep a record of each
invariant's evidence and whether it is satisfied, violated, or unverified.
Cut automatic Go call-graph analysis, invented composite scores, tool-detection
branches, and the separate scoring/consolidation agents. Structural analysis
can be requested independently. Proposed invariants should remain suggestions
in the report; applying them is a later authoring task.

**Proposed description:** Check a module's documented invariants against its
implementation and tests, reporting violations and unverified guarantees.

### 11. `bob-challenge-idea` — fold into risk review and imported grilling

[Source](../../skills/bob-challenge-idea/SKILL.md). The eight mandatory lenses
create substantial machinery around overlapping idea and risk assessment.
Use imported `grilling` for clarification and `grill-me` when explicitly requested
by the user. Preserve useful written risk questions in `premortem` reference
material: assumptions, alternatives, operating costs, demand, and reversibility.
Remove the assumption that every idea has a critical flaw. A risk register is
not the same output as the current severity-ranked idea report, so migration
must disclose the changed output rather than silently redirect the old command.

### 12. `bob-specs` — keep the capability, clarify its name

[Source](../../skills/bob-specs/SKILL.md). It scaffolds living module specs;
it does not perform general software design. Rename it to `bob-specs` or make
that purpose unmistakable in its description. Keep extracting real contracts
and responsibility boundaries. Move document skeletons into installed templates,
follow the repository's chosen spec convention, and cut the separate new/apply
phase engines. Separate documented facts from inferred guarantees when working
from existing code. Imported `codebase-design` supplies design vocabulary and
`to-spec` synthesizes a project spec; neither replaces this local scaffolding.

**Proposed description:** Create or update a module's living specifications,
following its repository conventions and grounding guarantees in evidence.

### 13. `bob-generate-overview` — keep a short HTML reference contract

[Source](../../skills/bob-generate-overview/SKILL.md). A single self-contained
HTML reference is a distinct output. Keep accurate source links, readable
content, and usable offline output. Cut the eleven inline layout prescriptions,
palette tables, repeated link rules, and token-by-token highlighting instructions.
Ship shared styles and useful templates as assets. Infer the target from the
request and ask only when ambiguous. Completion includes opening the report
and checking its links and controls.

**Proposed description:** Generate a self-contained HTML reference for a code
element or package, grounded in source and tests.

### 14. `bob-generate-feature-page` — removed

Removed after individual review at the user's request. Its teaching and
interactive explanation overlap did not justify another Bob entrypoint.
`bob-generate-overview` remains available for offline reports of existing code;
Matt's `teach` and `prototype` cover learning and interactive design exploration.

### 15. `bob-generate-okf` — keep, extract the schema and remove the interview script

[Source](../../skills/bob-generate-okf/SKILL.md). A linked `.knowledge/` catalog
is a useful maintained artifact. Keep navigation, source pointers, preservation
of user-authored content, and honest labeling of inferred decisions. Move the
schema and repeated concept/index templates into installed references and assets.
Ask for facts the code cannot supply, with no mandatory six-question interview
or loop that requires the user to say `done`. Replace the hardcoded
`2026-06-30` date with the actual date. Extending an existing catalog is the
normal path; replacing it is a separate destructive action requiring clear intent.

**Proposed description:** Create or update an OKF knowledge catalog that links
the project's packages, decisions, patterns, and confirmed planned work.

### 16. `bob-stage-prs` — keep; implement the git state machine in code

[Source](../../skills/bob-stage-prs/SKILL.md). PR stacking and resuming after
merges have genuine state complexity. Keep coherent, dependency-ordered PRs,
the approved publication plan, and resumable progress. Move branch operations,
merge tracking, rebase/lease checks, and partial-publication recovery into a
tested helper. Make the 200-line target a preference rather than a reason to
produce incoherent splits. Verify that every intermediate branch works; copying
final versions of whole files is insufficient by itself to establish that.
Completion is the actual stack, merge order, and next resumable action.

**Proposed description:** Split a changeset into an approved stack of coherent
pull requests and safely advance it as earlier requests merge.

### 17. `bob-operational` — keep a small contract around real dependencies

[Source](../../skills/bob-operational/SKILL.md). Cross-repository implementation
and deployment is a distinct task. Keep explicit repos, target environments,
dependencies, deployment authority, rollback prerequisites, review, and deployed
acceptance checks. Cut compulsory persistent coder/reviewer pairs and idle
operator/tester roles. Execute independent work concurrently when useful;
serialize genuine prerequisites. Replace ambiguous shared output paths with
named per-repository results. Completion must describe deployed state and
acceptance evidence, including any partial failure.

**Proposed description:** Coordinate a requested change across repositories,
deploy it to the specified environment, and verify its acceptance criteria.

### 18. `bob-premortem` — keep and trim

[Source](../../skills/bob-premortem/SKILL.md). A prioritized risk register with
preventive actions is a clear output. Keep failure scenarios, assumptions,
impact, mitigations, and known owners. Cut the explanatory research claim,
exhaustive intake script, repeated risk sections, and obligatory revisit schedule.
Use an optional output template and include only relevant categories. Mark
unknown owners, dates, and likelihoods as unknown instead of fabricating them.

**Proposed description:** Identify plausible project failure scenarios and
produce prioritized risks with practical mitigations.

### 19. `bob-postmortem` — keep and trim

[Source](../../skills/bob-postmortem/SKILL.md). Incident reconstruction is a clear
standalone task. Keep evidence, impact, timeline, contributing causes, and owned
follow-up actions. Move the report skeleton into a template. Cut the 13-question
intake sequence and universal severity definitions; follow the organization's
terms. Gather missing material in a compact request and distinguish established
facts from hypotheses. Completion is a usable draft that identifies remaining gaps.

**Proposed description:** Write an evidence-based incident postmortem with its
timeline, impact, contributing causes, and concrete follow-up actions.

### 20. `bob-internal-brainstorming` — removed

Removed both the normal and simple wrappers after individual review. Calling
workflows now use `bob-matt-grilling` for missing material decisions, preserving
existing specifications, confirmed decisions, and the local design artifact.
Automatic design commits and external `superpowers` dependencies are gone.

### 21. `bob-internal-writing-plans` — removed

Removed the normal, simple, and Goose wrappers after individual review. Work
consumers retain local planning with change boundaries, ownership, dependencies,
acceptance criteria, specification constraints, and meaningful verification.
`bob-matt-to-spec` and `bob-matt-to-tickets` remain separately requested tracker workflows.

### 22. `bob-internal-go-coding` — convert to a focused reference and correct it

[Source](../../skills/bob-internal-go-coding/SKILL.md). Keep non-obvious guidance
on ownership, bounded fan-out, cancellation, and error propagation in references
reached when relevant. Remove copies from every coder/fix prompt and avoid
unconditional style mandates. Correct the claim that `make` lengths and slice
indices require `int`: a local compiler check accepted both `int64` and `uint64`
arguments. Bounds and resource-budget validation remain valuable. A short
reference pointer can replace this separately installed internal skill.

### 23. `bob-version` — removed

Removed the generated diagnostic utility after individual review at the user's
request. Installers no longer generate a version skill or advertise its command.
Existing generated entrypoints are retired during installation.

### 24. `talk-to-codex` — removed

Removed after individual review. Its review and implementation overlap did not
justify retaining the extra cross-harness conversation workflow. Installers no
longer test for Codex merely to install this entrypoint, and known installed
copies are retired without removing customized files or companion resources.

## The variants require a semantic migration

Team variants mainly choose execution machinery and can merge into their task.
Several simple variants choose a different task outcome. They cannot safely
become aliases to a command that publishes automatically.

| Simple variant | Improvement and compatibility issue |
|---|---|
| `bob-work-simple` | Use its concise implementation structure as a starting point. Its current local-diff/optional-commit endpoint differs from `work`'s PR endpoint; preserve or explicitly migrate that meaning. |
| `bob-explore-simple` | Good short baseline. Keep evidence and uncertainty, remove the categorical ban on delegation, and normalize the report contract. |
| `bob-cleanup-simple` | Good short baseline. Keep behavior preservation and verification; clearly decide its local-diff versus publishing endpoint before merging names. |
| `bob-code-review-simple` | Retain review-only semantics. A migration toward imported `code-review` must also cover uncommitted diffs, which that import's fixed-point/HEAD instructions do not explicitly include. Never alias this to shipping. |
| `bob-adversarial-review-simple` | Best evidence standard for the deep review. Shorten the inline checklist into a relevant reference. Preserve read-only behavior and its distinction between proposed and generated tests. |
| `bob-audit-simple` | Short baseline, but narrow generic code-health scanning to the stated invariants and report what could not be verified. |
| `bob-stage-prs-simple` | It stages commits rather than publishing a PR stack. Preserve that distinction during migration; an alias to normal staging would expand side effects. |
| `bob-internal-brainstorming-simple` | Retire alongside the full wrapper; retain only useful clarification behavior in the consuming task. |
| `bob-internal-writing-plans-simple` | Fold into its consumer; remove the obligation to reproduce complete implementation details. |

The 15 Goose variants show that much shorter instructions are possible, but
they are independently maintained behaviors. For example, normal code review
attempts fixes before returning `NEEDS_BRAINSTORM`, while Goose returns that
status immediately for CRITICAL/HIGH findings. Establish one task contract per
retained skill, then generate narrow harness adaptations. Agent count, source
editing authority, publication behavior, and completion criteria should agree
across adapters.

## Shared improvements needed for the rewrite

- **Publication:** Keep the pre-publish hook, configured push preview, explicit
  commit/branch identity, and truthful recovery after a partial push. Implement
  these once in tested code; a short skill must still honor them.
- **Verification:** Keep repository verification configuration. Its `verify:`
  parsing and execution grammar belongs in one helper, with results consumed by
  the task. Verify and review the final changeset, including files altered by
  verification commands.
- **Resources:** Bob's normal skill installers mostly copy entrypoints. Extend
  them to ship templates/assets/references before shortening skills to depend
  on those files. The imported-package installer already copies complete packages.
- **Artifacts:** Keep records with a real consumer, such as a PR-stack ledger
  used on resume. Remove artifacts whose sole reason to exist is a phase label.
- **Invocations:** Write descriptions around the actual task and side effects.
  Preserve current invocation policies unless deliberately changed; a short
  description should avoid attracting unrelated edits into a publishing workflow.
- **Conflicts:** Keep explicit handling of contradictory inputs. Decide from
  evidence and surface unresolved decisions; compiling code alone cannot prove
  that it implements the intended contract.
- **Migration:** Update callers, agents, installers, tests, and the catalog
  together. Remove obsolete Bob-owned installed names deliberately while
  preserving unrelated skills and imported attribution.

## Example of the intended size: `bob-work`

> Implement the requested change in an isolated worktree, using the repository's
> guidance and documented contracts. Establish the acceptance criteria and a
> short plan; use an existing plan when provided. Clarify material contradictions
> and continue work that does not depend on an unresolved decision.
>
> Implement the smallest coherent change and update affected contracts or design
> records. Delegate independent assignments when useful, with clear file ownership.
> Run the repository's verification, review the final diff, and fix demonstrated
> problems. Reconsider the approach when the evidence shows it is wrong.
>
> Publish the verified, reviewed change through Bob's shared publication helper,
> honoring the pre-publish hook and configured push approval. Report the PR and
> CI results, and obtain the final merge decision. If blocked, report the precise
> blocker and retained state instead of repeatedly restarting the workflow.

This is a proposed contract, not a claim that the shared helper already exists
or covers the full lifecycle. The helper must be implemented and verified before
the replacement skill relies on it.

## Recommended implementation order

1. Settle and implement the shared verification/publication contracts. Keep the
   distinction between review, local implementation, and shipping visible.
2. Rewrite `work`, `explore`, `cleanup`, and `audit` to a few paragraphs. Replace
   the misleading review/publication entrypoint with the proposed shipping task.
3. Merge the four team/agent duplicates and retire the brainstorming/planning
   wrappers. Migrate simple variants according to their actual side effects.
4. Ship resources and trim the spec, documentation, risk, and incident skills.
5. Verify retained task behavior across harnesses with realistic tasks: small
   edits, read-only reviews, no-op cleanup, missing audit evidence, a blocked
   publication, and a resumed PR stack. Test outcomes and authority boundaries
   rather than required headings or phase names.
