# agent-loader Tests

No TypeScript test runner is configured in this project. These are the manual
verification cases and expected behaviours for `buildBuiltinTools`.

## buildBuiltinTools

### Case 1: undefined input → all defaults
```
buildBuiltinTools(undefined)
// → ["read", "write", "edit", "bash", "find", "grep", "ls"]
```

### Case 2: CC PascalCase names → pi lowercase names
```
buildBuiltinTools(["Read", "Write", "Bash"])
// → ["read", "write", "bash"]
```

### Case 3: Glob maps to find
```
buildBuiltinTools(["Glob", "Grep"])
// → ["find", "grep"]
```

### Case 4: Coordination tools filtered out (not pi built-ins)
```
buildBuiltinTools(["Read", "Task", "TaskCreate", "Write"])
// → ["read", "write"]
```

### Case 5: All coordination tools → fallback to ["read"]
```
buildBuiltinTools(["Task", "TaskList", "TaskGet", "TaskUpdate"])
// → ["read"]
```

### Case 6: Deduplication
```
buildBuiltinTools(["Read", "Read", "Bash"])
// → ["read", "bash"]
```

### Case 7: Empty array → all defaults
```
buildBuiltinTools([])
// → ["read", "write", "edit", "bash", "find", "grep", "ls"]
```

### Case 8: Mixed valid + coordination
```
buildBuiltinTools(["Read", "Glob", "Grep", "Task", "Write", "Bash"])
// (workflow-brainstormer's tool list)
// → ["read", "find", "grep", "write", "bash"]
```

## Integration verification (manual)

After deploying the fixed extension, spawn an agent and confirm it has tools:

```
subagent({ agent: "workflow-brainstormer", task: "List the files in the current directory using your ls tool and report them." })
```

Expected: agent lists files successfully (not "I have no tools available").
