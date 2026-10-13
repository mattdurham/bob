# Bob skill complexity review

Reviewed 2026-10-13 at commit `66d3ba7`.

The review uses Matt's stated criteria: skills should be a few paragraphs long, have few branching paths, and make it easy to understand what invoking them will do.

The main source of complexity is workflow machinery embedded in prompts: phase routing, fixed agent rosters, duplicated instructions, retry counters, persistent state, and publication recovery. Many of these skills need a smaller behavioral scope. Editing their wording alone would leave the ceremony intact.

## Scope and measurements

The inventory includes every skill family under `skills/`, including `talk-to-codex` and the version template. The review compares the main, simple, and Goose versions and follows relevant delegation and installation code. This is a static review; it does not measure actual agent cost or success rates, and it is not a comprehensive audit of every agent prompt.

| Measure | Value |
|---|---:|
| Skill families, including the version template | 24 |
| Public families / internal families | 21 / 3 |
| Main definitions | 23 `SKILL.md` files + 1 template |
| Simple / Goose variants | 9 / 15 |
| All skill definitions and variants | 48 files, 12,492 lines |
| Main definitions alone | 11,727 lines, 58,046 whitespace-delimited words |
| Median main definition length | 561 lines |
| Main definitions over 500 lines | 14 |

Line counts include templates and examples. They are a maintenance signal, not a complexity score. Only relevant skills load for an invocation; the totals are not a claim about context consumed by one run.

## Findings, in priority order

### 1. Review does much more than its name suggests

[`bob-code-review`](../../skills/bob-code-review/SKILL.md#L8) reviews, fixes, tests, commits, pushes, creates a PR, and monitors CI. Publication is the normal successful path. Its [COMMIT phase](../../skills/bob-code-review/SKILL.md#L342) alone spans 154 lines, including confirmation on/off paths, prepared-commit recovery, failed-push recovery, stale-preview handling, and exact reply matching.

This makes both command selection and execution harder to understand. A user looking for review must distinguish normal review, simple review, adversarial review, and simple adversarial review, with different side effects.

**Recommendation:** Give `review` a focused contract: inspect the requested scope and report verified findings. Keep fixing and publication as clearly named actions. Preserve publication guarantees in one implementation rather than reproducing recovery logic in several prompts.

### 2. Several public choices have almost no behavioral distinction

`bob-explore` and `bob-explore-teams` are each 625 lines. Their differences are the name, a claimed distinction already present in both, tool argument spelling, and a cross-reference. `bob-cleanup` and `bob-cleanup-teams` are each 776 lines; their differences are the name and tool argument spelling. Neither pair gives the user a meaningful choice of outcome.

The three development entry points also make the user choose orchestration architecture: `work`, `work-agents`, and `work-teams`. Their main definitions total 2,167 lines before loading their workers or `code-review`.

**Recommendation:** Keep one public command per user task. Retain existing names as aliases during migration. Agent execution strategy should be an implementation choice unless the user explicitly wants to control it.

### 3. Fixed teams turn small requests into large workflows

[`work`](../../skills/bob-work/SKILL.md#L163) always splits the plan between two coders. [`explore`](../../skills/bob-explore/SKILL.md#L170) creates two analysts and two challengers after separate discovery, then covers four prescribed dimensions. [`adversarial-review`](../../skills/bob-adversarial-review/SKILL.md#L92) and [`challenge-idea`](../../skills/bob-challenge-idea/SKILL.md#L96) each require eight reviewers and all eight results.

Cleanup first invokes audit, launches specialist discovery, creates another persistent team, tracks per-iteration artifacts, runs final review, and then invokes `code-review`, which reviews again. The workflow includes a useful no-findings shortcut, but that shortcut occurs after substantial discovery machinery.

**Recommendation:** Make direct work the default. Delegate only when there is a concrete independent task or an independent check worth paying for. Remove fixed headcounts, required kickoff messages, team ceremonies, and per-agent status artifacts from the skill contract.

### 4. Branching rules conflict with each other

[`work`](../../skills/bob-work/SKILL.md#L38) forbids the orchestrator from running git commands, then instructs it to create a worktree directly. It forbids implementation decisions, then makes it responsible for resolving plan/code disagreements before tests.

[`work-agents`](../../skills/bob-work-agents/SKILL.md#L92) forbids all file writes, then asks the orchestrator to write a brainstorm prompt and record conflict rulings. [`work-teams`](../../skills/bob-work-teams/SKILL.md#L169) prohibits teammates from pass/fail judgments, then its reviewer prompt instructs them to APPROVE or NEEDS_FIXES.

[`cleanup`](../../skills/bob-cleanup/SKILL.md#L70) forbids phase skipping, but later skips to final review when there are no tasks. It promises a clean final review, allows remaining MEDIUM+ findings after five iterations, and gives new functionality a separate unlimited loop. Its roster descriptions alternate between three, four, and five discovery/reviewer counts.

These contradictions require the model to invent precedence while following supposedly exact rules.

**Recommendation:** Replace repeated CAN/CANNOT sections and routing matrices with a few outcome constraints. Use one failure rule: fix a bounded problem and verify the affected work; surface a blocker or a needed decision when progress is no longer useful. Revisit the approach when evidence shows the approach is wrong.

### 5. “Simple” has several incompatible meanings

[`CLAUDE.md`](../../CLAUDE.md#L216) describes simple as a specification format choice, with [only spec content changing](../../CLAUDE.md#L255). [README](../../README.md#L49) says simple workflows may still use agents, identifying adversarial review as the exception. All nine current simple skill files prohibit delegation. They also change workspace handling, publication behavior, output paths, and sometimes the task itself.

For example, normal `stage-prs` manages branches and a PR stack; simple `stage-prs` stages commits and requires separate authorization for PRs. Normal explore writes `exploration-report.md`; simple explore writes `exploration.md`. The default installer publishes both normal and available simple siblings. `SPEC=simple` skips families without simple files rather than providing a uniformly simpler version of the full catalog.

Goose files are much shorter, which is useful evidence that long prose is removable, but their semantics also differ. Normal `code-review` fixes CRITICAL/HIGH findings up to its loop cap; Goose `code-review` immediately returns `NEEDS_BRAINSTORM` for them.

**Recommendation:** Establish one short behavioral definition for each task. Keep runtime tool differences in a small adapter. Remove the public full/simple taxonomy once the shorter definitions become the default. Do not preserve every historical behavior as a mode.

### 6. Templates and specialist references dominate several skills

`generate-overview` contains eleven report layouts plus visual tokens, linking rules, and highlighting rules. `generate-okf` includes most of a knowledge-catalog schema and repeated templates. `design` embeds the four-document scaffolding format. `internal-go-coding` is a reference manual rather than a workflow.

These details have more justification than team greetings or routing recaps, but they obscure the action the skill performs. `generate-feature-page` already references overview's visual system instead of duplicating it, although that reference creates an extra skill dependency.

**Recommendation:** Keep the skill to the requested outcome and essential constraints. Ship actual templates, shared styling, and narrowly relevant reference material as supporting resources. Load only the resource needed for the current output. Avoid relocating the same phase engine into another long markdown file that every invocation must read.

### 7. Review premises manufacture work

[`adversarial-review`](../../skills/bob-adversarial-review/SKILL.md#L22) starts from “every file contains at least one issue.” [`challenge-idea`](../../skills/bob-challenge-idea/SKILL.md#L22) starts from “every idea has at least one critical flaw.” Both favor flagging uncertainty; repeated reports are treated as higher confidence. These premises can generate findings, reconciliation, and follow-up work without a demonstrated defect. This is a risk inferred from the prompts, not measured behavior.

The simple adversarial review already supplies the better contract: actionable, demonstrable findings; meaningful impact; explicitly allow no verified findings.

**Recommendation:** Keep skeptical inspection and require evidence. Separate an unanswered question from a verified defect. Severity should follow impact, not the amount of ceremony or the number of agents repeating a claim.

## Review of every skill family

Lengths below refer to the main definition, including its examples and templates. “High” means substantial excess ceremony or branching against the requested few-paragraph standard; it is not a defect severity.

| Skill | Lines | What it actually does | Complexity and recommendation |
|---|---:|---|---|
| `bob-work` | 294 | Develops in a worktree with planning, two coders, tests, and a nested review/publication lifecycle | **High.** Keep one short development contract; remove fixed team size, seeded-phase machinery, and nested lifecycle ownership. |
| `bob-work-agents` | 696 | Develops through sequential phase agents and a coder that delegates again | **High.** Retire as a separate public choice; preserve an alias if needed. |
| `bob-work-teams` | 1,177 | Develops through persistent coders/reviewers and a shared task list | **High.** Retire as a public architecture choice; use teams only for suitable work. |
| `bob-explore` | 625 | Maps code, coordinates four analysis dimensions and challenges, writes reports | **High.** Keep a short read-only exploration contract tied to the user's question. |
| `bob-explore-teams` | 625 | Almost the same exploration workflow | **High.** Merge into `explore`; alias the old name. |
| `bob-cleanup` | 776 | Audits, discovers issues, coordinates fixes/reviews, tests, reviews again, publishes | **High.** Keep a short behavior-preserving cleanup contract; remove the audit and team prerequisites. |
| `bob-cleanup-teams` | 776 | Almost the same cleanup workflow | **High.** Merge into `cleanup`; alias the old name. |
| `bob-code-review` | 632 | Reviews, fixes, tests, publishes, and monitors | **High.** Make `review` report findings; move publication to a clearly named action. |
| `bob-adversarial-review` | 750 | Runs eight specialist reviews; optionally writes reproduction tests | **High.** Fold useful checks into `review`; use focused extra review only when justified. |
| `bob-audit` | 770 | Checks invariants, automatically adds Go health analysis and scoring, proposes invariants | **High.** Focus on checking stated invariants. Make structural analysis a separate requested task. |
| `bob-challenge-idea` | 648 | Interviews, runs eight adversarial lenses, produces an idea review | **High.** Use a few paragraphs of evidence-based idea critique. Consider combining overlapping risk work with `premortem`. |
| `bob-specs` | 592 | Creates or applies four spec documents and Go note headers | **High in prompt size.** Rename to `specs` or similarly explicit wording; move scaffolding into templates. |
| `bob-generate-overview` | 530 | Generates one HTML report using one of eleven target layouts | **High in prompt size.** Combine explanation entry points; extract shared style and templates. |
| `bob-generate-feature-page` | 197 | Generates a linked HTML feature explanation with animations | **Removed after individual review.** Overlap with learning and explanation skills did not justify retaining this entrypoint. |
| `bob-generate-okf` | 635 | Interviews and generates a `.knowledge/` catalog | **High in prompt size.** Keep as an optional knowledge task with schema/templates as resources. Remove fixed question progression and repeated completion prompts. |
| `bob-stage-prs` | 523 | Plans, publishes, resumes, rebases, and advances a PR stack | **High, with genuine state complexity.** Keep as an advanced task; put resumable git mechanics in tested code and make the skill a short contract. |
| `bob-operational` | 321 | Coordinates multiple repos, deployment, and deployed acceptance checks | **High, with genuine scope complexity.** Keep optional; state targets, dependencies, stop conditions, and evidence. Drop mandatory persistent role rosters. |
| `bob-premortem` | 197 | Imagines failure and produces prioritized risks/mitigations | **Low branching, moderate size.** Keep; move the large output template out and ask only for missing facts. |
| `bob-postmortem` | 148 | Produces an incident timeline, causes, impact, and actions | **Low branching, moderate size.** Keep; a few paragraphs plus an optional template are sufficient. |
| `bob-internal-brainstorming` | 79 | Interactive design discussion, design-document commit, external skill handoffs | **Moderate despite its length.** Remove universal mandatory invocation, section-by-section confirmations, automatic design commits, and undeclared skill dependencies. Fold needed clarification into `work`. |
| `bob-internal-writing-plans` | 133 | Spawns a planner to write complete code and frequent-commit steps | **Moderate.** Fold concise planning into `work`; plan acceptance criteria and change boundaries instead of a second implementation. |
| `bob-internal-go-coding` | 349 | Go coding examples and a recurring review checklist | **Moderate reference burden.** Keep relevant guidance as an optional reference, remove copies from orchestration prompts, and independently verify technical claims. |
| `bob-version` | 64 | Displays installation metadata and a manually listed catalog | **Low.** Keep as a small utility; derive the catalog from installed metadata so it stays accurate. |
| `talk-to-codex` | 190 | Runs an ask/code conversation through a subagent and Codex until agreement | **Moderate.** Keep optional integration; one short contract with explicit mode and a bounded exchange. Treat agreement as review input, then verify the result. |

## Proposed standard for the rewrite

Use three or four paragraphs, roughly 150–250 words as a working target rather than a hard limit. Each skill should make these things obvious: what task it performs, what it can change, what evidence counts as success, and what it reports when blocked.

Keep one main path. Remove phase names, routing tables, fixed headcounts, repeated prohibitions, scripted greetings, and implementation-level tool examples. Input that changes the user's intended task can warrant a separate command; ordinary problem solving should use the agent's judgment.

Preserve useful constraints: respect existing user changes and contracts, verify changed behavior, review the final diff, ground findings in evidence, honor configured publication approval, and state unresolved decisions clearly. Required review does not require multiple review layers.

Persist a plan or results when needed for actual handoff or resumption. Create artifacts for a consumer or a useful durable record, rather than because a phase exists. Use scripts for deterministic publication, stack recovery, or other mechanics that truly require exact state transitions.

Make the shorter task definitions canonical. Keeping full/simple/runtime versions with different behavior would leave the command-selection problem unresolved. Supporting resources should provide templates or specific reference material; they should not recreate the same long orchestrator under a different filename.

## Example shape for `work`

This is an illustration of the desired contract, not an installed replacement. The final publication endpoint should be decided once for the development task and reflected in its name and description.

> Implement the requested change. Read the relevant repository guidance, code, tests, and contracts first. Establish the intended behavior and a short plan; use a supplied plan as context. Resolve material contradictions and ask only for missing decisions that prevent useful progress.
>
> Work in an isolated worktree when needed. Preserve unrelated user changes and follow existing patterns. Delegate independent work when it provides a clear benefit, with explicit file ownership. Update documentation when the change affects a documented contract or decision.
>
> Run the relevant verification and review the resulting diff. Fix demonstrated problems and recheck affected behavior. Reconsider the approach when the evidence warrants it; surface a persistent blocker with its evidence. Report what changed, what was verified, and what remains unresolved. Follow the agreed publication scope and configured approval requirement.

## Suggested order

1. Define the few-paragraph contract and publication endpoint for `work` and `review`.
2. Rewrite `work`, `explore`, `review`, `cleanup`, and `audit` as the core tasks; eliminate their team/full/simple public forks.
3. Consolidate duplicate names as aliases and update the catalog/documentation from the actual definitions.
4. Extract documentation templates and the few deterministic git operations that need real state machines.
5. Trim the focused optional skills to the same standard and reconcile runtime behavior.

This review adds recommendations only. No skill, installer, or agent behavior was changed.
