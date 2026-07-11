# bob-generate-feature-page — Design

*Date: 2026-07-11*

## Overview

`bob-generate-feature-page` generates a self-contained explainer bundle for a **feature of the
current codebase** — a capability implemented across one or more files/packages, optionally
backed by docs (README, SPECS.md, design docs). It is modeled on `bob-generate-overview`, but
targets a bounded *slice* of the codebase rather than a single declaration (function, struct,
package), and adds animated, code-grounded diagrams as first-class output.

Where `bob-generate-overview` answers "what does this function/struct/package do," this skill
answers "how does this feature work, end to end, and how do its parts relate" — with a
navigable sidebar of sub-concepts and a linked animation per sub-concept showing its real
control/data flow.

## Non-goals

- Not for explaining arbitrary external concepts unrelated to the repo (e.g. "explain OAuth2" in
  the abstract). Every sub-concept and animation must be grounded in code/docs actually read
  from the target repo.
- Not a replacement for `bob-generate-overview` — that skill remains the right tool for a single
  function/struct/interface/package walkthrough.

## Workflow

```
DISCOVER → RESEARCH → DECOMPOSE → GENERATE → OUTPUT
```

### Phase 1: DISCOVER

User names a feature (e.g. "labelstore", "the OKF bundle system", "the audit workflow"). The
skill:

- Searches the repo (grep/glob) for where that feature lives — entry points, key packages,
  related docs (README, SPECS.md, NOTES.md, CLAUDE.md)
- Confirms scope with the user if ambiguous (e.g. "labelstore" might mean just
  `pkg/labelstore`, or that package plus its callers)
- Produces a bounded map of in-scope files/packages/docs

### Phase 2: RESEARCH

Same rigor as `bob-generate-overview` Phase 2, but at feature granularity. For any
non-trivial feature this delegates to an `Explore` subagent rather than reading everything
inline, since scope can span many files:

- Read all in-scope files: exported and key unexported types/functions, doc comments
- Follow callers/callees across the scope boundary
- Read tests for behavioral evidence
- Read any SPECS.md/NOTES.md/CLAUDE.md/README in scope
- Collect concrete evidence per candidate sub-concept: specific function names, real call
  order, actual state transitions — this evidence is what animations will be built from

### Phase 3: DECOMPOSE

From the research, break the feature into **sub-concepts** — natural components/sub-topics
(not a fixed count; small features might yield 3, complex ones 6-8). Each sub-concept becomes:

- One sidebar entry in `index.html`
- One detail block (purpose, key types/functions, source links)
- One linked animation page depicting that sub-concept's real structure or flow
- An **animation style** classification: *relationship* (static structure — who talks to
  whom, what depends on what) or *process* (an ordered sequence — a call chain, a state
  machine, a request lifecycle). Pick whichever the evidence emphasizes; don't force both into
  one animation.

### Phase 4: GENERATE

Produces the output bundle (see Output Structure below). Reuses `bob-generate-overview`'s
design tokens (fonts, color palette, dark/light toggle, source-link styling) so all of Bob's
generated HTML looks like one family.

Animations are built one of two ways, per their Phase 3 style classification:

- **Relationship → animated SVG.** Inline `<svg>` with nodes for real components and edges for
  their actual relationships, animated via CSS keyframes/transitions (hover highlight, ambient
  pulse along dependency lines). No step controls — it's a structural diagram, not a sequence.
- **Process → canvas step animation.** A discrete-step state machine (not continuous physics)
  driven by inline JS: each step redraws the canvas to reflect one real transition drawn from
  Phase 2 evidence (e.g. "labelWriter.Put() called" → highlight box, draw arrow to storage).
  Ships with Play/Pause/Step Forward/Step Back/Reset controls.

Every animation, either style, includes a short caption citing the code evidence it's based on.

### Phase 5: OUTPUT

Write the bundle, open `index.html` in the browser (`open`/`xdg-open`), verify no broken
internal links first.

## Output Structure

Flat directory named after the feature (kebab-case), written to the current working directory:

```
labelstore-explainer/
  index.html                    # sidebar (sub-concepts) + detail pane
  animations/
    actors.html                 # standalone: animated SVG (relationship) or canvas+controls (process)
    storage-model.html
    dedup-logic.html
    config-lifecycle.html
```

- `index.html`: left sidebar lists sub-concepts. Clicking one shows its detail text + source
  links (`vscode://file/...`) in the right pane, plus a real `<a href>` link ("Watch: Dedup
  Logic →") that navigates to `animations/dedup-logic.html` — not a popup, not a hash-route
  view swap.
- Each `animations/<concept>.html` is independently self-contained: title, one-paragraph
  caption, and either an animated SVG relationship diagram or a canvas process animation with
  controls bar (Play/Pause/Step Forward/Step Back/Reset), plus a "← Back to `<feature>`" link
  to `index.html`.
- Every individual HTML file has no external dependencies. The *bundle* as a whole is multiple
  linked files — same "no CDNs, all inline" rule as generate-overview, applied per file.

## Quality Checklist

- All files self-contained (no CDNs/external deps)
- Every animation traceable to specific real code evidence cited in its caption
- Dark/light toggle + shared tokens consistent across `index.html` and all animation pages
- All internal links (`index.html` ↔ `animations/*.html`) verified non-broken before opening
- `index.html` opened in browser automatically at the end
- No fixed cap on sub-concept count — scale to feature complexity

## Open Questions / Future Work

- Whether to support re-running against an already-generated bundle to update it incrementally
  (out of scope for v1 — v1 always regenerates fresh).
