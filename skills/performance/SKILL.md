---
name: performance
description: Performance optimization workflow - BENCHMARK → ANALYZE → OPTIMIZE → VERIFY
user-invocable: true
category: workflow
---

# Performance Optimization Workflow

You orchestrate **performance optimization** through benchmarking, analysis, optimization, and verification.

## Workflow Diagram

```
INIT → BENCHMARK → ANALYZE → OPTIMIZE → VERIFY → COMMIT → MONITOR → COMPLETE
                      ↑                      ↓
                      └──────────────────────┘
                    (loop if targets not met)
```

---

## Phase 1: INIT

Understand performance goals:
- What needs optimization?
- Current vs target metrics?
- Acceptable trade-offs?

Create bots/:
```bash
mkdir -p bots
```

Write targets to `bots/perf-targets.md`

---

## Phase 2: BENCHMARK

**Goal:** Establish baseline

Spawn workflow-tester for benchmarking:
```
Task(subagent_type: "workflow-tester",
     description: "Run benchmarks",
     prompt: "Run performance benchmarks.
             Save results to bots/benchmark-before.txt.")
```

**Output:** `bots/benchmark-before.txt`

---

## Phase 3: ANALYZE

**Goal:** Identify bottlenecks

Spawn workflow-performance-analyzer:
```
Task(subagent_type: "workflow-performance-analyzer",
     description: "Analyze performance",
     prompt: "Analyze benchmarks in bots/benchmark-before.txt.
             Identify bottlenecks and recommend optimizations.
             Write findings to bots/perf-analysis.md.")
```

**Input:** `bots/benchmark-before.txt`, `bots/perf-targets.md`
**Output:** `bots/perf-analysis.md`

---

## Phase 4: OPTIMIZE

**Goal:** Implement optimizations

Spawn workflow-coder:
```
Task(subagent_type: "workflow-coder",
     description: "Implement optimizations",
     prompt: "Implement optimizations from bots/perf-analysis.md.
             Keep code readable, add comments for complex changes.")
```

**Input:** `bots/perf-analysis.md`
**Output:** Optimized code

---

## Phase 5: VERIFY

**Goal:** Verify improvements

Spawn workflow-tester:
```
Task(subagent_type: "workflow-tester",
     description: "Verify optimizations",
     prompt: "Run tests and new benchmarks.
             Compare to baseline in bots/benchmark-before.txt.
             Write results to bots/perf-results.md.")
```

**Output:** `bots/perf-results.md`

**Decision:**
- Targets met + tests pass → COMMIT
- Targets not met → ANALYZE (deeper optimization)
- Tests fail → OPTIMIZE (fix broken code)

---

## Phase 6: COMMIT

Create commit with metrics:
```bash
git commit -m "perf: optimize [component]

Improvements:
- Speed: [before] → [after] ([X]% faster)
- Memory: [before] → [after] ([X]% reduction)

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Phase 7: MONITOR

Monitor CI performance tests.

**If failures:** Loop to ANALYZE

---

## Phase 8: COMPLETE

Optimization complete! Performance targets met.
