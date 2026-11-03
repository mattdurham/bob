# Bob Work

`/bob:work` develops a requested change in an isolated worktree, verifies and
reviews it, prepares its PR, and asks for the final merge decision.

Bob automatically uses a supplied plan file or pasted text, including
`--seed-plan <path>`. Without a supplied plan, Bob runs the imported
`bob-matt-grill-me` interview and waits for your confirmation of shared understanding
before writing the implementation plan.

Both paths produce a **coding worklist** in `.bob/state/plan.md`, with task
identifiers, outcomes, files, dependencies, verification, status, and assignments.
Bob can implement directly or delegate to one or multiple coders when independent
work and clear file ownership make that useful. Overlapping writes and dependent
tasks are serialized; coder count is chosen from the worklist.

The `bob-work-simple` variant performs the same plan intake and worklist creation
while keeping its existing single-agent and publication boundaries. Goose uses
its native bounded delegation. Completed assignments may be reviewed while
independent coding continues, with stable review inputs and unique result paths.
All writers and reviewers finish before verification and final integrated review.

Install with `make install-skills` or the relevant harness installer. The normal
installer includes `bob-matt-grill-me` and `bob-matt-grilling`. See [SKILL.md](SKILL.md) for the
current workflow and its configured publication gates.

`bob-ship` and its publication procedure install in both full and simple modes.
Simple workflows apply those controls directly without delegation, and publish
only when requested.
