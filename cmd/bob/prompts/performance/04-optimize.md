# OPTIMIZE Phase

You are currently in the **OPTIMIZE** phase of the performance workflow.

## Your Goal
Implement performance optimizations.

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

### 1. Implement Optimizations
For each bottleneck:
- Apply the optimization strategy
- Keep the code readable
- Add comments explaining the optimization
- Maintain correctness

### 2. Common Optimization Techniques
- **Reduce allocations**: Use sync.Pool, preallocate slices
- **Improve algorithms**: Use better data structures
- **Cache results**: Memoization, precomputation
- **Reduce copying**: Use pointers, avoid unnecessary copies
- **Parallelize**: Use goroutines where appropriate
- **Batch operations**: Group operations together

### 3. Write Tests
Ensure correctness:
```go
// Test that optimization didn't break functionality
func TestOptimizedFunction(t *testing.T) {
    // Same tests as before
    // Verify behavior unchanged
}
```

### 4. Document Changes
Update `bots/performance.md`:

```markdown
## Optimizations Implemented

### Optimization 1: Reduce Allocations in XYZ
**Change**: Used sync.Pool for buffer reuse
**File**: file.go:123
**Expected Impact**: 50% reduction in allocations

### Optimization 2: Improved Algorithm in ABC
**Change**: Replaced O(n²) loop with map lookup
**File**: file.go:456
**Expected Impact**: 30% faster for large inputs
```

## DO NOT
- ❌ Do not sacrifice correctness for speed
- ❌ Do not make code unreadable
- ❌ Do not skip testing
- ❌ Do not optimize without measuring

## When You're Done
After implementing optimizations:

1. Tell user: "Optimizations implemented"
2. Report progress:
   ```
   workflow_report_progress(
       worktreePath: "<worktree-path>",
       currentStep: "VERIFY",
       metadata: {
           "optimizationsApplied": X
       }
   )
   ```

