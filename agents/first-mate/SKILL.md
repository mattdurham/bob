---
name: first-mate
description: First-mate CLI reference guide for spec lookup and code graph analysis in spec-driven Go projects
user-invocable: false
category: reference
---

# First-Mate CLI Reference

`first-mate` is a code graph and spec analysis CLI. Use it when working in a spec-driven project (one with SPECS.md, NOTES.md, TESTS.md, or BENCHMARKS.md files).

## When to Use First-Mate

Detect spec-driven modules:
```bash
find . -name "SPECS.md" -o -name "NOTES.md" -o -name "TESTS.md" -o -name "BENCHMARKS.md" | head -5
```

If any are found, use first-mate for spec lookup and structural analysis. It is faster and more accurate than manual grep for these tasks.

---

## Setup: Load the Code Graph

Run once per session before any other commands:
```bash
first-mate parse_tree
```

This parses all Go files under the working directory into an in-memory graph. Required before `call_graph`, `query_nodes`, `find_deadcode`, etc.

---

## Spec Lookup

Read spec documents from the project:
```bash
first-mate read_docs kind="SPECS"       # all SPECS.md files
first-mate read_docs kind="NOTES"       # all NOTES.md design decisions
first-mate read_docs kind="TESTS"       # all TESTS.md test scenarios
first-mate read_docs kind="BENCHMARKS"  # all BENCHMARKS.md metric targets
first-mate read_docs pattern="SPEC-001" # search for a specific spec ID or keyword
```

Find which code nodes a spec covers:
```bash
first-mate find_spec query="FuncName"         # find spec coverage for a symbol
first-mate find_spec query="authentication"   # keyword search across specs
first-mate list_specs                          # list all known specs
first-mate get_spec id="SPEC-001"             # get a specific spec by ID
```

---

## Code Structure

Get call graphs:
```bash
first-mate call_graph function_id="pkg.FuncName" direction="callees"  # what this calls (default)
first-mate call_graph function_id="pkg.FuncName" direction="callers"  # who calls this
first-mate call_graph function_id="pkg.FuncName" direction="both"     # both directions
first-mate call_graph function_id="pkg.FuncName" depth=5              # deeper traversal
```

Find the path between two functions:
```bash
first-mate call_path from="pkg.FuncA" to="pkg.FuncB"
```

Query nodes by property (CEL expressions):
```bash
first-mate query_nodes expr='kind=="function" && cyclomatic > 15'          # complex functions
first-mate query_nodes expr='kind=="function" && cyclomatic > 40'          # over threshold
first-mate query_nodes expr='kind=="function"' sort_by="cyclomatic" top_n=10  # top 10 most complex
first-mate query_nodes expr='kind=="function" && len(caller_ids)==0 && !external'  # uncalled functions
first-mate query_nodes expr='kind=="interface"'                            # all interfaces
```

Find all implementations of an interface:
```bash
first-mate find_implementations interface_id="pkg.InterfaceName"
```

Find dead code (exported symbols never referenced):
```bash
first-mate find_deadcode
```

Find all TODO/FIXME/HACK comments:
```bash
first-mate find_todos
```

Full structural overview (compact format):
```bash
first-mate encode_ccgf
```

---

## Static Analysis

Run all analysis in one call (recommended — annotates nodes for follow-up queries):
```bash
first-mate run_analysis
```

After `run_analysis`, query annotated nodes:
```bash
first-mate query_nodes expr='lint_count > 0'    # nodes with lint issues
first-mate query_nodes expr='race_count > 0'    # nodes with race issues
first-mate query_nodes expr='heap_allocs > 5'   # heavy heap allocators
```

Run individual checks:
```bash
first-mate find_races      # statically detect race patterns (heuristic — also run go test -race)
first-mate run_vet         # go vet with node annotation
first-mate run_lint        # golangci-lint with node annotation
first-mate run_escape      # escape analysis — annotates heap_allocs on nodes
first-mate run_tests       # go test with coverage — annotates test_status and coverage on nodes
```

After `run_tests`:
```bash
first-mate query_nodes expr='kind=="function" && coverage < 50'              # low coverage
first-mate query_nodes expr='test_status == "untested" && cyclomatic > 10'  # risky untested functions
```

After `run_bench`:
```bash
first-mate query_nodes expr='bench_ns_op > 1000'   # slow benchmarks
```

---

## Graph Diff (Before/After Comparison)

Save a snapshot before making changes:
```bash
first-mate graph_snapshot    # returns snapshot name
```

After changes, compare:
```bash
first-mate graph_diff        # shows added, deleted, changed nodes
```

---

## Help Commands

```bash
first-mate help                    # list all available tools
first-mate help <tool>             # describe a specific tool and its parameters
first-mate query_help              # full CEL query language docs with examples
first-mate query_examples          # ready-to-use query examples grouped by use case
first-mate tools_help              # complete tool reference for all tools
first-mate ccgf_grammar            # CCGF format definition (call before encode_ccgf)
```

---

## query_nodes Field Reference

All fields available in CEL expressions for `query_nodes`:

### Always Available
| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique node ID (e.g. `"pkg.FuncName"`) |
| `kind` | string | Node type: `"function"`, `"type"`, `"interface"`, `"var"`, `"const"`, `"file"`, `"package"` |
| `name` | string | Symbol name |
| `file` | string | Source file path |
| `line` | int | Line number |
| `text` | string | Full source text of the node |
| `external` | bool | True if from an external package (not in this module) |
| `cyclomatic` | int | Cyclomatic complexity (functions only) |
| `cognitive` | int | Cognitive complexity (functions only) |
| `receiver` | string | Method receiver type (methods only) |
| `params` | string | Parameter list as string |
| `returns` | string | Return types as string |
| `parent_id` | string | ID of parent node (e.g. file containing this function) |
| `callee_ids` | list(string) | IDs of functions this node calls |
| `caller_ids` | list(string) | IDs of functions that call this node |
| `child_ids` | list(string) | IDs of child nodes (e.g. methods of a type) |

### Annotated by `run_lint`
| Field | Type | Description |
|-------|------|-------------|
| `lint_count` | int | Number of lint issues on this node |
| `lint_issues` | string | Semicolon-separated `"linter: message"` list |

### Annotated by `find_races` / `run_analysis`
| Field | Type | Description |
|-------|------|-------------|
| `race_count` | int | Number of race patterns detected |
| `race_issues` | string | Semicolon-separated race descriptions |

### Annotated by `run_escape` / `run_analysis`
| Field | Type | Description |
|-------|------|-------------|
| `heap_allocs` | int | Number of allocations that escape to heap |
| `stack_allocs` | int | Number of allocations that stay on stack |

### Annotated by `run_tests`
| Field | Type | Description |
|-------|------|-------------|
| `coverage` | float | Test coverage percentage (0–100) |
| `test_status` | string | `"pass"`, `"fail"`, `"covered"`, or `"untested"` |

### Annotated by `run_bench`
| Field | Type | Description |
|-------|------|-------------|
| `bench_ns_op` | float | Nanoseconds per operation |
| `bench_b_op` | float | Bytes allocated per operation |
| `bench_allocs_op` | float | Allocations per operation |

### Annotated by `run_profile`
| Field | Type | Description |
|-------|------|-------------|
| `pprof_flat_pct` | float | % of CPU time spent in this function |
| `pprof_cum_pct` | float | % of CPU time spent in this function + callees |

### Annotated by `run_vet`
| Field | Type | Description |
|-------|------|-------------|
| `vet_issues` | string | Semicolon-separated vet messages |

---

## query_edges Field Reference

Fields available in CEL expressions for `query_edges`:

| Field | Type | Description |
|-------|------|-------------|
| `from` | string | Source node ID |
| `to` | string | Target node ID |
| `kind` | string | Edge type: `"call"`, `"implements"`, `"contains"`, `"imports"` |

---

## Args Format

Args can be passed as `key=value` pairs or as JSON:
```bash
first-mate query_nodes expr='kind=="function"' top_n=10
first-mate query_nodes '{"expr":"kind==\"function\"","top_n":10}'
```
