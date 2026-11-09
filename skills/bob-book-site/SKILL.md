---
name: bob-book-site
description: Build and verify a static HTML site from a repository’s mdBook-compatible project book without changing its Markdown sources.
user-invocable: true
---

# Build the Project Book Site

Resolve the requested mdBook root, defaulting to the repository's `book/`
directory. Require `book.toml` and `src/SUMMARY.md`; a missing or incomplete
book must be completed with `bob-book`, not repaired with `mdbook init` or empty
chapters.

Require `mdbook` on `PATH` and check its version. If it is unavailable, stop and
report that prerequisite with the installation command `cargo install mdbook
--locked`; do not silently install global tooling or substitute a different
renderer.

Run `mdbook build <book-root>` from the repository root. Honor the book's
configured build directory, normally `<book-root>/html/`. Never edit Markdown
sources or hand-edit generated HTML. Treat build warnings about missing,
unindexed, or unreadable chapters as book defects to reconcile through
`bob-book`, not as site cosmetic issues.

Verify that the build exits successfully, `index.html` exists in the configured
output directory, and the generated navigation contains the ordered chapters
from `SUMMARY.md`. When browser tooling is available, open the generated site
and check the introduction, one workshop chapter, navigation, and a narrow
viewport. Report the output path, mdBook version, exercised checks, and any
remaining verification limits. Hosting and deployment are outside this skill.
