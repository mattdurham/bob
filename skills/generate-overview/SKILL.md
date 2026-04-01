---
name: bob:generate-overview
description: Generate a stylized, self-contained HTML report for any Go codebase element — functions, packages, interfaces, structs, handlers, CLI commands, and more
user-invocable: true
category: documentation
---

# Generate Overview — HTML Report

You produce a **single self-contained HTML file** describing a Go codebase element. The report
is modern, clean, and interactive — no external dependencies. You detect what the user is
pointing at, choose the right report template, research thoroughly, then generate.

---

## Phase 1: IDENTIFY TARGET

Ask the user what they want documented:

> What should I generate a report for? Point me at something:
> a function, package, file, struct, interface, directory, handler, CLI command, etc.

The user may give you a function name, file path, package path, or description. Use your
tools to resolve it to a concrete code element.

**Detect the target type** by examining what the user pointed at:

| Target Type | How to Detect |
|---|---|
| **Function** | User names a function; or you resolve to a `func` declaration |
| **Interface** | User names an interface; resolves to `type X interface` |
| **Struct** | User names a struct; resolves to `type X struct` |
| **Package** | User names a package or directory containing `.go` files with `package` declarations |
| **File** | User gives a specific `.go` file path |
| **Directory** | Path contains mixed files (not a single Go package) |
| **HTTP Handler** | Function signature matches `(w http.ResponseWriter, r *http.Request)` or returns `http.Handler`; or is registered with a router |
| **CLI Command** | Uses `cobra.Command`, `flag.FlagSet`, `os.Args`, or similar CLI framework patterns |
| **Middleware** | Function that wraps/returns `http.Handler` or `http.HandlerFunc`; or follows `func(next http.Handler) http.Handler` pattern |
| **Config/Constants** | User points at a `const` or `var` block, or a config struct |
| **Protobuf/gRPC** | `.proto` file, or Go file with generated gRPC service stubs |
| **Makefile/Build** | `Makefile`, `Taskfile.yml`, `.goreleaser.yml`, or similar |

Tell the user what you detected:

> I see this is a **[target type]**: `[name]` in `[location]`.
> Generating report...

If ambiguous, ask. Do NOT ask for a "depth level" — the target type determines the report format.

---

## Phase 2: RESEARCH

Gather everything needed for the report. Run these in parallel where possible.

### 2a. Read the target

- Read the target code thoroughly
- For functions: read the full function body and its doc comment
- For packages: read all exported types, functions, and their doc comments
- For files: read the entire file
- For interfaces: find all implementations via grep

### 2b. Find context

- Who calls this? (`grep` for usage sites)
- What does this call? (read function bodies for outbound calls)
- Are there tests? Read relevant test files
- Any spec documents? (SPECS.md, NOTES.md, CLAUDE.md in the package)

### 2c. Collect metadata

- File path and line numbers for all code elements
- Package import path
- Git blame for last-modified info if relevant

---

## Phase 3: GENERATE HTML

Generate a **single self-contained HTML file**. All CSS and JS must be inline. No external
resources, CDNs, or dependencies. The file must look great when opened in any browser.

### Global HTML Design System

Every report shares these design properties:

```
DESIGN TOKENS:
- Font stack: system-ui, -apple-system, "Segoe UI", sans-serif
- Mono font: "SF Mono", "Cascadia Code", "JetBrains Mono", "Fira Code", Consolas, monospace
- Base size: 15px body, 13px code
- Max width: 1200px centered, with sidebar where applicable
- Border radius: 8px cards, 6px code blocks, 4px inline code
- Shadows: subtle, using rgba(0,0,0,0.06) for light, rgba(0,0,0,0.3) for dark

COLOR PALETTE (light mode):
- Background: #ffffff page, #f8f9fb cards/sidebar, #f1f3f5 code blocks
- Text: #1a1a2e primary, #495057 secondary, #868e96 muted
- Accent: #4263eb primary actions/links, #748ffc hover
- Borders: #e9ecef default, #dee2e6 heavy
- Syntax: #d73a49 keywords, #6f42c1 types, #005cc5 functions, #032f62 strings, #6a737d comments

COLOR PALETTE (dark mode):
- Background: #1a1b26 page, #24283b cards/sidebar, #1e2030 code blocks
- Text: #c0caf5 primary, #9aa5ce secondary, #565f89 muted
- Accent: #7aa2f7 primary, #89b4fa hover
- Borders: #3b4261 default, #414868 heavy
- Syntax: #f7768e keywords, #bb9af7 types, #7aa2f7 functions, #9ece6a strings, #565f89 comments

INTERACTIVE FEATURES (include in every report):
- Dark/light mode toggle (top-right, respects prefers-color-scheme)
- Collapsible sections via <details>/<summary> with smooth animation
- Smooth scroll to anchors
- Copy button on all code blocks
- Keyboard shortcut: 't' to toggle theme, '/' to focus search (if search exists)
```

### Report Templates by Target Type

**IMPORTANT — Source links everywhere:** Every file path, function name, type name, line
number reference, and declaration in the report MUST be a clickable `vscode://file/` link
(see the **Source Links** section below for format and styling). When a template says
"file:line link" or "linked" or "with line number", it means a `vscode://file/` source link.
Do not render any file:line reference as plain text.

---

#### FUNCTION Report

A line-by-line annotated walkthrough of a single function.

**Layout:** Single column, no sidebar.

**Sections:**
1. **Header** — Function signature in a large code block. Doc comment rendered below.
2. **At a Glance** — Card grid:
   - Package & file location (with line number)
   - Receiver type (if method)
   - Parameters table (name, type, purpose)
   - Return values table (type, purpose, error conditions)
3. **Annotated Source** — The core of this report:
   - Show the full function source with line numbers
   - Each logical block gets a **margin annotation** — a short callout on the right side
     explaining what that block does and why
   - Group lines into logical blocks (setup, validation, core logic, error handling, cleanup)
   - Use colored left-border strips to distinguish block types:
     - Blue: setup/initialization
     - Green: core logic / happy path
     - Yellow: validation / guards
     - Red: error handling
     - Gray: cleanup / defer
   - For complex expressions, add inline tooltips or expandable explanations
4. **Call Graph** — What this function calls (as a simple list with file:line links).
   What calls this function (usage sites found via grep).
5. **Error Paths** — If the function returns errors, list each error return with:
   - The condition that triggers it
   - The error value/message
   - Line number link
6. **Related** — Links to: test functions, related functions in same package, interface
   it implements (if method).

---

#### INTERFACE Report

**Layout:** Two-column — main content + right sidebar listing all implementors.

**Sections:**
1. **Header** — Interface name, package, doc comment.
2. **Contract** — Full interface definition in a code block.
3. **Method Catalog** — For each method:
   - Signature
   - Purpose (from doc comments or inferred)
   - Parameters and return values
   - Behavioral contract / expectations
4. **Implementors** — Each known implementation:
   - Type name and package (linked)
   - Which methods it implements (checkmarks in a matrix if >3 methods)
   - Brief description of how this implementation differs
   - Key implementation detail or trade-off
5. **Usage Patterns** — How callers typically use this interface. Code snippets from real usage.
6. **Sidebar** — Sticky list of implementors with jump links.

---

#### STRUCT Report

**Layout:** Single column with collapsible sections.

**Sections:**
1. **Header** — Struct name, package, doc comment.
2. **Field Reference** — Table:
   - Field name | Type | JSON/YAML tag | Purpose
   - Group by concern if >8 fields (use sub-headers)
   - Highlight unexported fields with a subtle indicator
3. **Constructors** — Functions that return this struct (`NewX`, `MakeX`, etc.):
   - Signature, doc comment, which fields they set
4. **Methods** — Grouped by concern:
   - For each: signature, one-line purpose, receiver type (pointer vs value)
   - Expandable: full doc comment and parameter details
5. **Used By** — Where this struct is instantiated or referenced.
   Top 10 usage sites with file:line links.
6. **Implements** — Interfaces this struct satisfies (if any).

---

#### PACKAGE Report

**Layout:** Two-column — main content + left sidebar with navigation.

**Sections:**
1. **Header** — Package name, import path, package doc comment.
2. **Overview** — 3-5 sentence summary of what this package does and why it exists.
3. **Sidebar Navigation** — Sticky nav listing all sections with counts:
   - Types (N) / Functions (N) / Constants (N) / Variables (N)
4. **Architecture** — ASCII or text description of how the major types relate.
   Show dependency flow between internal components.
5. **Public API** — Grouped catalog:
   - **Interfaces** — name, one-line purpose, method count
   - **Structs** — name, one-line purpose, field count
   - **Functions** — name, signature, one-line purpose
   - **Constants & Variables** — grouped by block, with values
   Each item is collapsible to show full details.
6. **Internal Design** — Key unexported types and helpers that a contributor should know about.
7. **Dependencies** — What this package imports (grouped: stdlib, internal, external).
   What imports this package.
8. **Testing** — Test file overview. How to run tests. Notable test helpers.

---

#### FILE Report

**Layout:** Two-column — main content + right sidebar with structure outline.

**Sections:**
1. **Header** — File path, package, last modified info.
2. **Sidebar** — Sticky outline of every declaration in the file:
   - type, func, const, var — with line numbers
   - Click to scroll to that section
   - Search/filter box at top of sidebar
3. **File Overview** — What this file is responsible for. Inferred from its contents and
   the package context.
4. **Declarations** — Walk through the file top-to-bottom:
   - Each declaration gets a card:
     - Name and kind (type/func/const/var)
     - Line range
     - Doc comment
     - Syntax-highlighted source (collapsible for long items)
     - Annotations for non-obvious logic
5. **Imports** — What this file imports, grouped and annotated with why each is needed.
6. **Cross-References** — Other files in the package that reference declarations in this file.

---

#### DIRECTORY Report (non-package)

**Layout:** Single column with file tree.

**Sections:**
1. **Header** — Directory path, total file count, language breakdown.
2. **File Tree** — Interactive tree view:
   - Each file gets a one-line purpose annotation
   - Directories are collapsible
   - Entry points highlighted with a marker
   - Test files visually distinguished
3. **Package Map** — If directory contains multiple Go packages, show each with:
   - Package name, import path, purpose
   - Key exported symbols
4. **Dependency Flow** — How packages in this directory depend on each other.
5. **Entry Points** — Main functions, init functions, handler registrations.
6. **Configuration** — Config files present (go.mod, Makefile, Dockerfile, etc.)
   with brief description of each.

---

#### HTTP HANDLER Report

**Layout:** Single column, visually structured like API documentation.

**Sections:**
1. **Header** — Handler function name, route pattern (if discoverable), HTTP method.
2. **Endpoint Card** — Prominent card:
   - `METHOD /path/to/endpoint`
   - Brief description
3. **Middleware Chain** — Ordered list of middleware applied to this handler:
   - Each middleware: name, what it does, what it adds to context
   - Visual pipeline: `Request → Auth → RateLimit → Logger → [Handler] → Response`
4. **Request** — What this handler expects:
   - URL parameters
   - Query parameters
   - Request body schema (struct fields if JSON-decoded)
   - Required headers
5. **Response** — What this handler returns:
   - Success response (status code, body shape)
   - Error responses (each error status code, condition, body)
6. **Annotated Source** — Same as Function report's annotated source section.
7. **Auth & Permissions** — What auth is required (inferred from middleware or code).

---

#### CLI COMMAND Report

**Layout:** Single column, styled like a modern man page.

**Sections:**
1. **Header** — Command name, one-line description.
2. **Synopsis** — Usage string in a code block.
3. **Description** — Full description from the command's `Long` field or doc comment.
4. **Flags** — Table:
   - Flag | Short | Type | Default | Env Var | Description
   - Required flags highlighted
5. **Subcommands** — If this is a parent command:
   - Each subcommand: name, one-line description
   - Collapsible to show that subcommand's flags
6. **Environment Variables** — Any env vars read by this command.
7. **Examples** — Usage examples from the command definition or tests.
8. **Source** — Link to the file:line where the command is defined.

---

#### MIDDLEWARE Report

**Layout:** Single column with a visual flow diagram.

**Sections:**
1. **Header** — Middleware function name, package.
2. **Flow Diagram** — Visual representation:
   ```
   ┌─────────────┐
   │   Request    │
   └──────┬──────┘
          ▼
   ┌─────────────┐
   │  [Before]   │  ← what this middleware does before calling next
   │  Auth check │
   └──────┬──────┘
          ▼
   ┌─────────────┐
   │    next()   │  ← inner handler
   └──────┬──────┘
          ▼
   ┌─────────────┐
   │  [After]    │  ← what this middleware does after (if anything)
   │  Log access │
   └──────┬──────┘
          ▼
   ┌─────────────┐
   │  Response   │
   └─────────────┘
   ```
   Rendered as inline SVG.
3. **Context Injection** — What this middleware adds to `context.Context`:
   - Key, value type, how to retrieve it downstream
4. **Short-Circuit Conditions** — When does this middleware stop the chain:
   - Condition, HTTP status returned, error body
5. **Ordering Dependencies** — Must come before/after other middleware. Why.
6. **Annotated Source** — Same style as Function report.

---

#### CONFIG / CONSTANTS Report

**Layout:** Single column, styled as a reference card.

**Sections:**
1. **Header** — Block location, purpose.
2. **Reference Table** — For each constant/variable:
   - Name | Type | Value | Used By | Description
   - "Used By" links to call sites
   - Sortable columns (JS)
3. **Grouping** — If constants use `iota`, show the enumeration with meaning of each value.
4. **Configuration Map** — For config structs:
   - Field | Type | Default | Env Var | CLI Flag | Description
   - Mark required vs optional
   - Show validation rules if present
5. **Impact** — What changes when you modify each value. Brief notes per item.

---

#### PROTOBUF / gRPC SERVICE Report

**Layout:** Two-column — methods list in sidebar, details in main.

**Sections:**
1. **Header** — Service name, proto file, package.
2. **Sidebar** — List of all RPC methods with streaming indicators:
   - `→` unary
   - `→→` server streaming
   - `←→` bidirectional
3. **Service Overview** — What this service does, who calls it.
4. **Methods** — For each RPC:
   - Method name, streaming type
   - Request message (fields table)
   - Response message (fields table)
   - Error codes returned
5. **Messages** — All message types used by this service:
   - Field number, name, type, description
   - Nested messages shown indented
6. **Generated Code** — Where the generated Go code lives, key files.

---

#### MAKEFILE / BUILD CONFIG Report

**Layout:** Single column with visual dependency graph.

**Sections:**
1. **Header** — File name, build system type.
2. **Target Graph** — Visual DAG of target dependencies (inline SVG):
   - Commonly-used targets highlighted
   - Default target marked
3. **Target Reference** — For each target:
   - Name, dependencies, description (from comments)
   - Commands it runs (collapsible)
   - Variables that affect it
4. **Variables** — All variables defined:
   - Name, default value, description, overridable?
5. **Environment** — Env vars that affect the build.
6. **Quick Reference** — Top 5 most useful commands in a copy-friendly card.

---

## Source Links

All "jump to source" references in the report must be clickable links that open the file
in the user's editor. Use **absolute file paths** resolved from the working directory.

**Link format:** Use `vscode://file/{absolute_path}:{line}:{column}` URIs.

- Resolve all paths to absolute (e.g. `/Users/joe/dev/project/pkg/auth/handler.go:42`)
- Line numbers are 1-based
- Column defaults to 1 if not relevant
- Example: `vscode://file///Users/joe/dev/project/pkg/auth/handler.go:42:1`

**Where to add source links:**

- Every file path reference (e.g. `pkg/auth/handler.go:42` → clickable)
- Function names in call graphs and "Used By" sections
- Struct/interface names in "Implements" and "Implementors" sections
- Import paths that resolve to local packages
- Sidebar entries in File and Package reports
- Every declaration card's line range
- "Used By" and "Callers" entries
- Error path line references in Function reports

**Styling for source links:**

```css
a.src-link {
  font-family: var(--mono);
  font-size: 0.85em;
  color: var(--accent);
  text-decoration: none;
  border-bottom: 1px dashed var(--accent);
  opacity: 0.8;
  transition: opacity 0.15s;
}
a.src-link:hover {
  opacity: 1;
  border-bottom-style: solid;
}
```

Display the **relative path** as the link text (for readability) but use the **absolute
path** in the `href` URI. Add a small external-link icon (inline SVG, 12px) after each
source link to signal it opens an editor.

---

## Phase 4: OUTPUT

1. Write the HTML file to the current directory:
   - Filename: `<target-name>-report.html` (kebab-case)
   - Example: `rate-limiter-report.html`, `auth-middleware-report.html`

2. Open the report in the default browser:
   ```bash
   open <filename>    # macOS
   xdg-open <filename>  # Linux
   ```

3. Tell the user:
   > Report written to `<filename>` and opened in your browser.

4. Do NOT print the HTML to the conversation. It's too long and not useful in terminal.

---

## HTML Quality Checklist

Before writing the file, verify your HTML meets these standards:

- [ ] Valid HTML5 doctype and lang attribute
- [ ] All CSS inline in a `<style>` tag (no external stylesheets)
- [ ] All JS inline in a `<script>` tag (no external scripts)
- [ ] Dark/light toggle works and respects `prefers-color-scheme`
- [ ] All code blocks have syntax highlighting via CSS classes
- [ ] Copy buttons on code blocks work
- [ ] Collapsible sections use `<details>`/`<summary>`
- [ ] Sidebar navigation scrolls smoothly to anchors
- [ ] All source links use `vscode://file/` URIs with absolute paths
- [ ] Source links show relative path as display text, absolute in href
- [ ] Every file path, function reference, and declaration has a clickable source link
- [ ] No broken internal links
- [ ] Responsive — readable on narrow viewports
- [ ] Semantic HTML: `<nav>`, `<main>`, `<article>`, `<section>`, `<aside>`
- [ ] Page title set to the target name
- [ ] Total file size reasonable (< 500KB for most reports)

---

## Syntax Highlighting Rules

Apply these CSS classes to Go code tokens. Do NOT use an external library — apply classes
during generation based on Go syntax:

| Token Type | CSS Class | Light Color | Dark Color |
|---|---|---|---|
| Keyword (`func`, `type`, `if`, `return`, etc.) | `.kw` | `#d73a49` | `#f7768e` |
| Type name (`string`, `int`, `error`, custom) | `.typ` | `#6f42c1` | `#bb9af7` |
| Function name (in calls and declarations) | `.fn` | `#005cc5` | `#7aa2f7` |
| String literal | `.str` | `#032f62` | `#9ece6a` |
| Number literal | `.num` | `#005cc5` | `#ff9e64` |
| Comment | `.cmt` | `#6a737d` | `#565f89` |
| Operator / punctuation | `.op` | `#d73a49` | `#89ddff` |
| Package / import path | `.pkg` | `#e36209` | `#ff9e64` |

Apply highlighting by wrapping tokens in `<span class="XX">`. For the annotated source
sections, combine with line-number gutters and annotation margins.
