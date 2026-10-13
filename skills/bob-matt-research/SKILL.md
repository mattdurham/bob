---
name: bob-matt-research
description: Investigate a topic or explore a codebase using primary sources, then save grounded findings in a Markdown report.
metadata:
  author: Matt Pocock
  source: https://github.com/mattpocock/skills/tree/49dd158d1076134a641b33efb035946536778336/skills/engineering/research
  upstream-commit: "49dd158d1076134a641b33efb035946536778336"
license: MIT
---

Establish the question and scope from the request; ask only when a missing detail would materially change the investigation. Investigate against primary sources: official documentation, source code, specifications, and first-party APIs. Follow claims back to the source that owns them. Work directly or delegate bounded, independent questions when that will help; collect delegated results before drawing conclusions.

For codebase exploration, trace the relevant entry points, component relationships, data flow, dependencies, and tests. Read documented contracts and design decisions, then check them against the implementation. Challenge the initial explanation by looking for counterexamples, stale documentation, missing assumptions, and edge cases. Distinguish demonstrated behavior from intended behavior, inference, and unanswered questions. Keep the investigation read-only except for its report; do not change source code, specifications, or external systems.

Write the findings to one Markdown report with source citations or file locations, the answer to the question, and remaining uncertainties. Save it at the requested path or where the repository keeps research notes; otherwise choose a sensible location and say where. Summarize the useful findings for the user without treating speculation or unverified claims as facts.
