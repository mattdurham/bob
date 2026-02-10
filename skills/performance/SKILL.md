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

## Execution Rules

**CRITICAL: All subagents MUST run in background**

- ✅ **ALWAYS use `run_in_background: true`** for ALL Task calls
- ✅ **After spawning agents, STOP** - do not poll or check status
- ✅ **Wait for agent completion notification** - you'll be notified automatically
- ❌ **Never use foreground execution** - it blocks the workflow

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
     run_in_background: true,
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
     run_in_background: true,
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
     run_in_background: true,
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
     run_in_background: true,
     prompt: "Run tests and new benchmarks.
             Compare to baseline in bots/benchmark-before.txt.
             Write results to bots/perf-results.md.")
```

**Output:** `bots/perf-results.md`

**Decision:**
- Targets met + tests pass → REVIEW
- Targets not met → ANALYZE (deeper optimization)
- Tests fail → OPTIMIZE (fix broken code)

---

## Phase 6: REVIEW

**Goal:** Comprehensive code review of optimized code by 7 specialized agents

Spawn 7 reviewer agents in parallel (single message, 7 Task calls):
```
Task(subagent_type: "workflow-reviewer",
     description: "Code quality review",
     run_in_background: true,
     prompt: "Review optimized code for logic, bugs, and best practices.
             Write findings to bots/review-code.md with severity levels.")

Task(subagent_type: "security-reviewer",
     description: "Security vulnerability review",
     run_in_background: true,
     prompt: "Scan for security vulnerabilities introduced by optimizations.
             Write findings to bots/review-security.md with severity levels.")

Task(subagent_type: "performance-analyzer",
     description: "Performance bottleneck review",
     run_in_background: true,
     prompt: "Verify optimizations didn't introduce new bottlenecks.
             Write findings to bots/review-performance.md with severity levels.")

Task(subagent_type: "docs-reviewer",
     description: "Documentation accuracy review",
     run_in_background: true,
     prompt: "Ensure performance changes are documented correctly.
             Write findings to bots/review-docs.md with severity levels.")

Task(subagent_type: "architect-reviewer",
     description: "Architecture and design review",
     run_in_background: true,
     prompt: "Evaluate if optimizations maintain good architecture.
             Write findings to bots/review-architecture.md with severity levels.")

Task(subagent_type: "code-reviewer",
     description: "Comprehensive code quality review",
     run_in_background: true,
     prompt: "Deep review of optimization quality and maintainability.
             Write findings to bots/review-code-quality.md with severity levels.")

Task(subagent_type: "golang-pro",
     description: "Go-specific code review",
     run_in_background: true,
     prompt: "Review Go optimizations for idiomatic patterns.
             Write findings to bots/review-go.md with severity levels.")
```

After all 7 agents complete, consolidate findings into `bots/review.md`.

**Output:** `bots/review.md` (consolidated report)

**Decision:**
- No issues → COMMIT
- Issues found → OPTIMIZE (fix issues before commit)

---

## Phase 7: COMMIT

Create commit with metrics:
```bash
git commit -m "perf: optimize [component]

Improvements:
- Speed: [before] → [after] ([X]% faster)
- Memory: [before] → [after] ([X]% reduction)

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Phase 8: MONITOR

Monitor CI performance tests.

**If failures:** Loop to ANALYZE

---

## Phase 9: COMPLETE

Optimization complete! Performance targets met.
