# DISCOVER Phase

You are currently in the **DISCOVER** phase of the exploration workflow.

## Your Goal
Discover and understand the codebase structure without making any changes.

## What To Do

### 1. Understand the Request
- Clarify what the user wants to explore
- What questions need answering?
- What patterns or features to investigate?

### 2. Search the Codebase
Use read-only tools:
```bash
# Find files
find . -name "*.go" -o -name "*.ts" -o -name "*.js"

# Search for patterns
grep -r "pattern" --include="*.go"

# List directory structure
tree -L 3
```

Or use Claude Code tools:
- **Glob** - Find files by pattern
- **Grep** - Search file contents
- **Read** - Read files
- **Bash** (read-only commands only)

### 3. Document Findings
Write discoveries to `bots/explore.md`:

```markdown
# Exploration: <Topic>

## Question
[What we're trying to understand]

## Discoveries

### File Structure
- [Directory organization]
- [Key files found]

### Patterns Found
- [Pattern 1: Description and locations]
- [Pattern 2: Description and locations]

### Key Components
- [Component 1: What it does]
- [Component 2: What it does]

### Dependencies
- [External packages used]
- [Internal modules]
```

## CRITICAL RULES
- ❌ **NO FILE CHANGES** - This is read-only exploration
- ❌ **NO EDITS** - Do not modify any files
- ❌ **NO WRITES** - Do not create new files (except bots/explore.md)
- ❌ **NO WORKTREE** - Work in current directory
- ❌ **NO COMMITS** - This is exploration only

## When You're Done
Once you've gathered information:

1. Share findings with user
2. Ask if they need deeper exploration
3. Report progress:
   ```
   workflow_report_progress(
       worktreePath: "<current-directory>",
       currentStep: "ANALYZE",
       metadata: {
           "filesExplored": 10,
           "patternsFound": 3
       }
   )
   ```

## Next Phase
After reporting, move to **ANALYZE** to deeply understand what you found.
