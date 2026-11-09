# Project book format

The Markdown sources are authoritative. Generated HTML is disposable output.
Keep the book current-state and narrative; Git retains its history.

## Repository layout

Use this layout at the repository root unless the user explicitly selects
another book root:

```text
book/
  .gitignore
  book.toml
  src/
    SUMMARY.md
    introduction.md
    guided-tour.md
    how-it-works.md
    components/
      index.md
      <component>.md
    reference.md
```

Use one stable, kebab-case filename per meaningful component. Preserve existing
chapter paths and identities during updates. Add local assets below `book/src/`
only when they improve the explanation.

Create `book/.gitignore` with:

```gitignore
/html/
```

Create `book/book.toml` with the real project title:

```toml
[book]
title = "<project title>"
language = "en"
src = "src"

[build]
build-dir = "html"
create-missing = false
```

`create-missing = false` makes an incomplete index an error instead of silently
creating empty chapters. Never edit files under `book/html/` or place source
material there.

## Ordered index

`book/src/SUMMARY.md` is the table of contents and reading contract. Link every
authored chapter exactly once, in this order:

```markdown
# Summary

- [Introduction](introduction.md)
- [Guided tour](guided-tour.md)
- [How it works](how-it-works.md)
- [Component workshop](components/index.md)
  - [Example component](components/example-component.md)
- [Reference](reference.md)
```

Replace the example component with the actual component chapters. Add nested
sections or split reference material when the project warrants it, while
preserving the global progression from operation to internals. An existing book
may already use equivalent descriptive titles; preserve those identities rather
than forcing a rename.

## Reading path and chapter contracts

### 1. Introduction — meet the system

Explain the problem, purpose, capabilities, limits, intended audience, and a
small mental model. Introduce essential vocabulary in plain language. Give the
reader enough context to understand why the system exists before showing its
machinery.

### 2. Guided tour — operate it

Lead through setup and one realistic end-to-end task, then the normal operations
a reader needs. Commands must come from the repository and be exercised before
being described as verified; otherwise identify the missing verification.
Explain inputs, observable results, and relevant cleanup or safety behavior.
Keep internals out except where they clarify an operating decision.

### 3. How it works — understand the whole machine

Explain architecture, ownership, major control/data flows, external dependencies,
and cross-cutting decisions. Show how the parts cooperate before dissecting
them. Include rationale with the relevant decision. Use diagrams when they make
relationships clearer; keep assets local and mdBook-compatible.

### 4. Component workshop — disassemble every part

Use `components/index.md` to introduce the workshop and orient the reader to the
component map. Give every meaningful component its own indexed chapter. Cover,
as applicable:

- responsibility and position in the system;
- public inputs, outputs, contracts, and invariants;
- dependencies, state ownership, and important internal flow;
- errors, edge cases, concurrency, security, and resource behavior;
- verification scenarios and the tests that exercise them;
- stable source pointers such as `path/to/file.go` plus the relevant symbol.

Include only evidenced performance statements and benchmarks. Trivial helpers
belong in their owning component's chapter. A short orientation is useful, but
do not repeat the guided tour or architecture chapter in each component.

### 5. Reference — look up exact details

Provide concise APIs, commands, configuration, file/component mapping, and a
glossary as applicable. Keep explanations in earlier chapters; reference entries
are for precise lookup and must agree with those explanations.

## Prose and evidence

Write paragraphs and headings that lead into one another, not collections of
notes, changelogs, or renamed specification tables. Explain why behavior exists
where that rationale is known. Prefer small worked examples and selective code
excerpts over full source dumps. Define acronyms and avoid unexplained project
jargon. Use inline repository paths and symbol names instead of line numbers,
which drift.

Preserve user-authored prose unless the requested change or new evidence makes
it inaccurate. Never fabricate intent, historical decisions, command output,
test results, benchmarks, or future capabilities. Ask about material unknowns;
leave no TODO-shaped substitute for missing content.

## Agent guidance installation

After validating the book, add this block to the root `AGENTS.md` and/or
`CLAUDE.md` that the repository already uses. If both exist, update both so
harness-specific agents receive the contract. If neither exists, create
`AGENTS.md` with the block. Replace only an existing marked block and preserve
all surrounding content:

```markdown
<!-- bob-book:start -->
## Project book

Read `book/src/SUMMARY.md` before planning changes. The book records intended
behavior; code records current behavior. Update relevant chapters before
implementation and reconcile them after verification. Surface code/book
conflicts with evidence and obtain explicit resolution before changing either
contract.
<!-- bob-book:end -->
```

For a non-default book root, adjust the index path in the installed guidance.
If existing guidance assigns conflicting authority to specs, OKF, or another
artifact, request an explicit ruling instead of installing contradictory
instructions.
