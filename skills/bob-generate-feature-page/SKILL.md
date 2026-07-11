---
name: bob-generate-feature-page
description: Generate a self-contained, navigable explainer bundle for a feature of the current codebase — a sidebar of sub-concepts, detail pages, and linked animated diagrams grounded in the real code
user-invocable: true
category: documentation
---

# Generate Feature Page — Animated Explainer Bundle

You produce a **self-contained explainer bundle** for a feature of the current codebase — a
capability implemented across one or more files/packages, not a single function/struct/package.
(For a single declaration, use `bob-generate-overview` instead.)

The bundle is a small directory: one `index.html` with a sidebar of sub-concepts and a detail
pane, plus one animated diagram page per sub-concept under `animations/`. Every sub-concept and
every animation must be grounded in code/docs you actually read — this is not a generic
explainer for abstract outside concepts (e.g. "explain OAuth2" in the abstract is out of scope;
"explain how our OAuth2 middleware works" is in scope).

---

## Phase 1: DISCOVER

Ask the user what feature to explain:

> What feature should I explain? Name it (e.g. "labelstore", "the OKF bundle system", "the
> audit workflow") and I'll find where it lives in this repo.

Resolve the name to a bounded slice of the codebase:

- `grep`/`glob` for the feature name across code, package names, and docs
- Identify entry points, key packages/files, and related docs (README, SPECS.md, NOTES.md,
  CLAUDE.md)
- If the boundary is ambiguous (e.g. does "labelstore" include its callers, or just the
  package itself?), ask the user to confirm scope before proceeding

Tell the user what you found:

> This feature spans: `[files/packages]`. Researching...

---

## Phase 2: RESEARCH

Gather everything needed to explain the feature accurately. For anything beyond a couple of
files, delegate this to an `Explore` subagent rather than reading everything inline — scope can
span many files.

- Read all in-scope files: exported and key unexported types/functions, doc comments
- Follow callers/callees across the scope boundary to understand how the feature connects to
  the rest of the system
- Read tests in scope for behavioral evidence (what's actually guaranteed vs incidental)
- Read any SPECS.md, NOTES.md, CLAUDE.md, or README in scope
- For each candidate sub-concept, collect **concrete evidence**: specific function/type names,
  real call order, actual state transitions or data flow. This evidence is the only thing later
  animations may be built from — never invent a flow that isn't backed by something you read.

---

## Phase 3: DECOMPOSE

Break the feature into **sub-concepts** — its natural components or sub-topics. Not a fixed
count: a small feature might yield 3, a complex one 6-8. Each sub-concept must correspond to
something a reader would want to look up independently (e.g. for "labelstore": *Actors*,
*Storage model*, *Dedup logic*, *Config & lifecycle*).

For each sub-concept, note:
- One-line purpose
- Key types/functions/files (with line numbers)
- The specific evidence (from Phase 2) its animation will depict
- **Animation style** — classify the sub-concept as one of:
  - **Relationship** (static structure: who talks to whom, what depends on what, ownership) →
    animated SVG
  - **Process** (an ordered sequence: a call chain, a state machine, a request lifecycle) →
    canvas step animation
  When a sub-concept has both a structure and a sequence worth showing, pick whichever the
  evidence emphasizes more; don't force both into one animation.

---

## Phase 4: GENERATE

### Design tokens

Reuse `bob-generate-overview`'s design system exactly — same font stack, color palette (light
and dark), border radii, shadows, dark/light toggle behavior, and source-link styling. All of
Bob's generated HTML should look like one family. See `bob-generate-overview`'s SKILL.md for
the full token values; do not invent a new palette.

### `index.html`

**Layout:** Left sidebar (sub-concept list) + right detail pane.

- **Header** — feature name, one-paragraph summary of what it does and why it exists
- **Sidebar** — one entry per sub-concept; clicking one shows that sub-concept's detail block
  in the right pane (in-page JS, no navigation)
- **Detail pane**, per sub-concept:
  - Purpose (1-3 sentences)
  - Key types/functions/files, each a clickable `vscode://file/{absolute_path}:{line}:{column}`
    source link (relative path as link text, absolute path in the href — same convention as
    generate-overview)
  - A real `<a href="animations/<concept>.html">` link, styled as a prominent button:
    "Watch: `<Concept Name>` →". This is an actual page navigation, not a popup
    (`window.open`) and not an in-page hash-route swap.
- Dark/light toggle, keyboard shortcut `t`, respects `prefers-color-scheme`

### `animations/<concept>.html`

One standalone, self-contained page per sub-concept. Every animation, regardless of style,
carries:

- Title + one-paragraph caption stating what it shows and citing the specific code evidence
  it's based on (e.g. "Based on `Put()` in `pkg/labelstore/writer.go:42`")
- `← Back to <feature>` link to `../index.html`
- Same dark/light toggle as `index.html`

Build one of two ways, per the **animation style** chosen in Phase 3:

**Relationship → animated SVG**

- Inline `<svg>` diagram: nodes for the real components/types involved, edges for their actual
  relationships (calls, ownership, data dependency) — labeled with real names, not generic
  placeholders
- Motion via CSS keyframes/transitions: e.g. pulsing/highlighting an edge on hover, a subtle
  looping pulse along a dependency line, fade-in on scroll into view
- No step controls needed — this is a structural diagram, not a sequence. Auto-plays its
  ambient motion (loop), interactive on hover
- Do not draw an edge or relationship you don't have concrete evidence for

**Process → canvas step animation**

- `<canvas>` element, sized to fill most of the viewport
- Inline JS driving a **discrete-step state machine** — not continuous physics. Each step
  redraws the canvas to reflect one real transition drawn from Phase 2 evidence (e.g.
  "`labelWriter.Put()` called" → highlight that box, draw an arrow to storage). Do not animate
  a transition you don't have concrete evidence for.
- Controls bar: Play, Pause, Step Forward, Step Back, Reset — same visual tokens as the rest of
  the bundle

---

## Output Structure

Flat directory named after the feature (kebab-case), written to the current working directory:

```
<feature>-explainer/
  index.html
  animations/
    <concept-1>.html
    <concept-2>.html
    ...
```

Example: `labelstore-explainer/index.html`, `labelstore-explainer/animations/dedup-logic.html`.

Every individual HTML file is self-contained — no CDNs, no external stylesheets/scripts, all
CSS/JS inline. The bundle as a whole is multiple linked files; each file on its own has zero
external dependencies.

---

## Phase 5: OUTPUT

1. Before writing, verify every `<a href>` between `index.html` and `animations/*.html` (and
   back) resolves to a file you're actually about to write — no broken internal links.
2. Write the directory and all files.
3. Open the bundle's `index.html` in the default browser:
   ```bash
   open <feature>-explainer/index.html     # macOS
   xdg-open <feature>-explainer/index.html # Linux
   ```
4. Tell the user:
   > Explainer bundle written to `<feature>-explainer/` and opened in your browser.
5. Do NOT print the HTML to the conversation. It's too long and not useful in terminal.

---

## Quality Checklist

Before writing the files, verify the bundle meets these standards:

- [ ] `index.html` and every `animations/*.html` file are independently self-contained (no
  external deps)
- [ ] Every sub-concept and every animation step/edge is traceable to specific code/doc
  evidence read in Phase 2 — cited in the animation's caption
- [ ] Each animation used the right style for its sub-concept: relationships as animated SVG
  (no step controls), processes as canvas step animations (with Play/Pause/Step/Reset controls)
- [ ] Dark/light toggle and shared design tokens are consistent across `index.html` and all
  animation pages
- [ ] All internal links (`index.html` ↔ `animations/*.html`) verified non-broken
- [ ] No fixed cap on sub-concept count — scale to the feature's actual complexity
- [ ] Every file path, function reference, and declaration in `index.html` has a clickable
  `vscode://file/` source link
- [ ] Valid HTML5, semantic elements (`<nav>`, `<main>`, `<article>`, `<aside>`)
- [ ] Responsive — readable on narrow viewports
- [ ] `index.html` opened in browser automatically at the end
