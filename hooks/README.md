# Bob Pre-Commit Hooks

Pre-commit hooks for Go projects that enforce quality standards before allowing commits.

## Installation

From the Bob repository:

```bash
make hooks
```

This installs:
- `pre-commit-checks.sh` → `~/.claude/hooks/`
- Merges hooks config into `~/.claude/hooks-config.json`
- Creates automatic backup
- Deduplicates existing hooks

**Safe to run multiple times** - intelligent merge preserves existing hooks.

## Pre-Commit Hook

### `pre-commit-checks.sh`

**Trigger:** Before any `git commit` command
**Purpose:** Ensures code quality before allowing commits

**Checks Performed:**

1. **Go Formatting** (`gofmt`)
   - Automatically formats code if needed
   - Ensures consistent style

2. **Tests** (`go test ./...`)
   - Runs all tests in the project
   - **Blocks commit if any tests fail** ❌
   - Timeout: 30 seconds

3. **Linting** (`golangci-lint`)
   - Comprehensive linting checks
   - **Blocks commit if linting fails** ❌
   - Timeout: 2 minutes
   - Install: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`

4. **Complexity** (`gocyclo`)
   - Checks cyclomatic complexity
   - **Warns but doesn't block** ⚠️
   - Flags functions with complexity > 15
   - Install: `go install github.com/fzipp/gocyclo/cmd/gocyclo@latest`

### How It Works

When Claude attempts to run `git commit`:

```
1. PreToolUse hook intercepts the command
2. Runs pre-commit-checks.sh
3. If checks pass (exit 0) → Commit proceeds
4. If checks fail (exit 1) → Commit is blocked
```

### Configuration

Located in `~/.claude/hooks-config.json`:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash(git commit*)",
        "hooks": [
          {
            "type": "command",
            "command": "~/.claude/hooks/pre-commit-checks.sh",
            "timeout": 120,
            "blocking": true
          }
        ]
      }
    ]
  }
}
```

### Bypass (Emergency Only)

If you need to bypass checks temporarily:

```bash
# Option 1: Use git directly in terminal (not through Claude)
git commit -m "message"

# Option 2: Disable hook temporarily
mv ~/.claude/hooks-config.json ~/.claude/hooks-config.json.disabled
# ... make commit ...
mv ~/.claude/hooks-config.json.disabled ~/.claude/hooks-config.json
```

### Best Practices

- ✅ **Let the hook run** - It catches issues early
- ✅ **Fix issues immediately** - Don't bypass checks
- ✅ **Run checks manually** before committing:
  ```bash
  go test ./...
  golangci-lint run
  gocyclo -over 15 .
  ```
- ❌ **Don't bypass** for convenience - Technical debt accumulates

## Other Hooks

### PostToolUse Hooks

- **Skill execution** - Reloads CLAUDE.md after skill runs
- **Edit/Write** - Analyzes Go code structure after changes
- **Bash** - Monitors PR creation and status

### Stop Hook

Self-check before finishing conversation:
- Verifies tests were run
- Checks for incomplete work
- Ensures quality standards met

---

*Part of Bob's workflow orchestration system*
