# COMMIT Phase

You are currently in the **COMMIT** phase of the performance workflow.

## Your Goal
Commit performance improvements.

## Continuation Behavior

**IMPORTANT:** Do NOT ask continuation questions like:
- "Should I proceed?"
- "Ready to continue?"
- "Shall I move to the next step?"
- "Done. Continue?"

**AUTOMATICALLY PROCEED** after completing your tasks.

**ONLY ASK THE USER** when:
- Choosing between multiple approaches/solutions
- Clarifying unclear requirements
- Confirming potentially risky/destructive actions (deletes, force pushes, etc.)
- Making architectural or design decisions

## What To Do

### 1. Review Changes
```bash
git status
git diff
```

### 2. Stage Files
```bash
git add path/to/optimized/file.go
git add path/to/test/file_test.go
```

### 3. Commit with Performance Data
```bash
git commit -m "$(cat <<'EOF'
perf: optimize XYZ for 50% performance improvement

Performance improvements:
- Reduced time/op by 50% (1234ns → 617ns)
- Reduced memory/op by 50% (512B → 256B)
- Reduced allocations/op by 50% (10 → 5)

Changes:
- Used sync.Pool for buffer reuse
- Replaced O(n²) algorithm with O(n)

Benchmarks: See bots/performance.md

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
EOF
)"
```

### 4. Verify
```bash
git log -1 --stat
```

## DO NOT
- ❌ Do not commit without benchmark data
- ❌ Do not use `git add .`
- ❌ Do not skip Co-Authored-By

## When You're Done
After committing:

1. Tell user: "Performance improvements committed: <hash>"
2. Ask: "Ready to push and create PR?"
3. Wait for approval
4. Report progress:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "MONITOR",
       metadata: {
           "committed": true,
           "improvementPct": 50
       }
   )
   ```

