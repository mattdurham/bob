---
name: bob-explore
description: Goose-native codebase exploration with parallel analysis and adversarial challenge
---

# Bob Explore for Goose

This workflow is read-only. The parent creates `.bob/state/`, delegates bounded
work, owns barriers and retries, and writes the final synthesis.

1. **DISCOVER:** Delegate to `Explore`; call `load(source: "<discovery task id>")`
   and require `.bob/state/discovery.md`.
2. **ANALYZE:** Launch four `team-analyst` jobs with
   `delegate(..., async: true)` for structure, data/control flow, patterns, and
   dependencies. Give each a unique `.bob/state/analysis-*.md` path. Load all
   four and reject missing artifacts.
3. **CHALLENGE:** Launch one `team-challenger` per analysis asynchronously.
   Each receives one analysis artifact and writes a unique challenge result.
   Call `load(source: "<challenge task id>")` for all jobs.
4. **LOOP:** A FAIL creates one bounded re-analysis job for that domain, followed
   by one rechallenge. Stop after max 2 challenge loops and document remaining
   uncertainty.
5. **DOCUMENT:** Synthesize verified findings into the requested document with
   `file:line` evidence, spec contracts, architecture, flow, dependencies,
   uncertainties, and challenger corrections.

Never modify source files, create commits, or let a delegate coordinate peers.
