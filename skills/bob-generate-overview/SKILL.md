---
name: bob-generate-overview
description: Explain an existing codebase element in a self-contained offline HTML report grounded in code, tests, and specifications.
---

# Explain Existing Code

Explain the requested codebase element in one self-contained HTML report.
Resolve the target from the request and repository; ask only when its identity
or scope is unclear. Read its implementation, relevant callers and dependencies,
tests, and applicable specifications before describing behavior. Explain its
purpose, important contracts, control or data flow, edge cases, and connections
to the surrounding system as relevant. Ground claims in the code you read and
label uncertainty; never invent guarantees or performance results.

Choose a readable, responsive layout suited to the subject and audience. Use
navigation, diagrams, code excerpts, or interactive controls when they improve
understanding. Keep CSS, JavaScript, fonts, and visual assets inline or
browser-native so the single file works offline without CDNs or external
dependencies. Include accurate jump-to-source links using escaped absolute paths
and 1-based line numbers, with relative paths as readable labels; use the user’s
editor scheme when known, otherwise vscode://file links. Escape code and
repository text as display content. Follow relevant repository presentation
conventions without imposing a fixed palette, report template, or set of
interactions.

Write the report to the requested location, or <target-name>-report.html in the
current directory, preserving unrelated files. Check that its explanations agree
with the inspected code, source locations are accurate, internal links and any
controls work, and it remains usable offline and on narrow screens. Open it in
the default browser when the environment supports that, and report the output
path and any verification limits. Keep implementation files unchanged and
provide the artifact link instead of pasting the HTML into the conversation.
