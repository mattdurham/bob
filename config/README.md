# Bob Configuration

This directory contains configuration files for Bob workflows.

## claude-permissions.json

Defines the default permission allow list for Claude Code.

**Usage:**
```bash
# Apply permissions to ~/.claude/settings.json
make allow
```

The `make allow` command **intelligently merges** these permissions into your existing Claude settings:

- ✅ **Union merge**: Combines permissions from config + existing settings (no overwrites)
- ✅ **Preserves custom permissions**: Any permissions you added manually are kept
- ✅ **Deduplicates**: Removes duplicate entries automatically
- ✅ **Automatic backup**: Creates `settings.json.backup` before applying
- ✅ **Preserves other settings**: StatusLine, plugins, etc. remain untouched
- ✅ **Idempotent**: Safe to run multiple times

**Example:**
```bash
# Your settings.json has: ["Bash", "Read", "CustomTool"]
# Config has: ["Bash", "Read", "Write", "Edit"]
# After make allow: ["Bash", "CustomTool", "Edit", "Read", "Write"]
#                    ↑ All permissions combined, sorted, deduplicated
```

**Managing Permissions:**

1. **Edit permissions:** Modify `config/claude-permissions.json`
2. **Apply to Claude:** Run `make allow`
3. **Share across machines:** Commit this file to version control

**Current Permissions:**
- `Bash` - Shell command execution
- `Read` - File reading
- `Write` - File creation
- `Edit` - File editing
- `Skill` - Skill invocation
- `WebFetch` - Web content fetching
- `mcp__*` - All MCP server tools (wildcard)

## Permission Modes

Claude Code supports three permission modes via `defaultMode`:

### `dontAsk` (Recommended)
- Tools in the `allow` list are automatically permitted **without prompting**
- Security boundary is **still enforced** - only listed tools are allowed
- Best for: Trusted workflows where you want speed without constant prompts
- Current setting: ✅ **Active**

### `ask`
- Prompts user for approval on **every tool use**, even for allowed tools
- Most interactive but safest mode
- Best for: Learning what Claude does, auditing workflows, or untrusted code

### `bypassPermissions` (Dangerous)
- **Completely bypasses** the permission system
- **ALL tools allowed** without any checks
- ⚠️ **Use only in fully trusted, sandboxed environments**
- Best for: Never (use `dontAsk` with comprehensive allow list instead)

**Current Mode:** `dontAsk` - Fast execution with security boundaries maintained.

## Why This Matters

Having permissions in version control means:
- ✅ Consistent permissions across all machines
- ✅ Easy onboarding for new team members
- ✅ Permissions can be reviewed in PRs
- ✅ Rollback capability if needed
- ✅ Documentation of required tool access
