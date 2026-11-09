---
name: bob-book
description: Create, revise, or reconcile a repository’s authoritative Markdown book for book-first development, progressive human reading, and coding-agent guidance.
user-invocable: true
---

# Maintain the Project Book

Read [the book format](references/book-format.md) before creating or changing
files. Infer the requested operation: bootstrap a missing book, revise intent
before a code change, or reconcile an existing book with verified code. Perform
only the requested documentation work; never implement the described software
change or generate HTML in this skill.

Start with the existing `book/src/SUMMARY.md`, relevant chapters, and project
guidance. Then inspect the implementation, tests, examples, and existing
specifications or knowledge catalogs needed for the requested scope. Existing
specs and OKF concepts are evidence, not outputs to update. Find repository
facts yourself; ask only about intent or conflicts that the sources cannot
resolve.

## Bootstrap

Inventory the repository's meaningful components and their source files before
writing. A component owns a coherent responsibility and public behavior; trivial
helpers remain with their owner rather than becoming separate chapters. Create
the complete reading path and one indexed workshop chapter per component.
Ground behavioral claims in code and tests, and execute setup or tour commands
before presenting them as verified. Distinguish observed behavior, documented
intent, and inference. Ask about material intended behavior that remains
unknown instead of turning guesses into authoritative prose.

## Revise intent

Translate the requested change into the relevant narrative, contracts, examples,
failure behavior, and verification guidance before implementation. Preserve
unaffected chapters and user-authored prose. The request authorizes the
described book revision; it does not authorize unrelated contract changes or
code edits. Report the resulting intentional code/book difference so a separate
coding workflow can implement it.

## Reconcile

Compare the book's claims with the verified implementation and tests. Correct
stale explanations, examples, source pointers, and component coverage when the
underlying contract remains unchanged. Treat the book as intended behavior and
code as current behavior. For every behavioral mismatch, cite the chapter and
implementation evidence, explain the competing behaviors, and obtain explicit
resolution before changing either contract. A reconciliation run must not
silently rewrite the book to match code or bless code because it disagrees with
the book.

## Write and verify

Follow the prescribed file layout, global progression, and chapter contracts.
Write current-state, connected prose that a newcomer can read from beginning to
end. Define terms before relying on them, use real examples, and put design
rationale beside the behavior it explains. Keep technical detail in the
architecture, workshop, and reference chapters rather than turning the opening
into a specification. Use stable repository paths and symbols for traceability;
avoid copied source listings and drifting line-number citations. Leave no empty
chapters, placeholders, unsupported guarantees, invented history, or unverified
performance claims.

Install or refresh the marked project-book guidance in the repository's existing
root agent-guidance files as described in the format reference. Preserve all
text outside the marked block. If existing guidance gives another artifact
conflicting authority, surface that conflict and obtain a ruling before
declaring the book authoritative.

Check that every discovered component is indexed, every indexed file exists,
every chapter is reachable in the intended reading order, source paths and
symbols resolve, examples have the claimed verification, and generated HTML has
not been edited. Report the book path, operation performed, meaningful changes,
verified examples, unresolved conflicts, and any evidence gaps. Finish only
when the requested scope is complete or explicitly blocked on a decision.
