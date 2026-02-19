#!/bin/bash
# Status line command for Claude Code
# Based on user's shell PS1 configuration

# Read JSON input from stdin
input=$(cat)

# Extract data from JSON
user=$(whoami)
host=$(hostname -s)
cwd=$(echo "$input" | jq -r '.workspace.current_dir')
remaining=$(echo "$input" | jq -r '.context_window.remaining_percentage // empty')

# Get git branch and worktree info if in a git repo (skip optional locks for performance)
git_branch=""
worktree_info=""
git_changes=""
if git -C "$cwd" rev-parse --git-dir >/dev/null 2>&1; then
    git_branch=$(git -C "$cwd" --no-optional-locks branch --show-current 2>/dev/null)

    # Check if this is a worktree (not the main repo)
    git_dir=$(git -C "$cwd" rev-parse --git-dir 2>/dev/null)
    git_common_dir=$(git -C "$cwd" rev-parse --git-common-dir 2>/dev/null)

    # If git-dir ends with .git/worktrees/*, it's a worktree
    if [[ "$git_dir" == *".git/worktrees/"* ]]; then
        # Extract worktree name (task) from current directory
        task_name=$(basename "$cwd")

        # Extract repo name from parent directory pattern: <repo-name>-worktrees
        parent_dir=$(basename "$(dirname "$cwd")")
        if [[ "$parent_dir" == *"-worktrees" ]]; then
            repo_name="${parent_dir%-worktrees}"
            worktree_info=$(printf " \033[35m[worktree:%s/%s]\033[00m" "$repo_name" "$task_name")
        else
            # Fallback if pattern doesn't match
            worktree_info=$(printf " \033[35m[worktree:%s]\033[00m" "$task_name")
        fi
    fi

    # Get lines changed (staged + unstaged)
    added=0
    removed=0

    # Count unstaged changes
    while IFS=$'\t' read -r add rem _; do
        [[ "$add" != "-" ]] && added=$((added + add))
        [[ "$rem" != "-" ]] && removed=$((removed + rem))
    done < <(git -C "$cwd" --no-optional-locks diff --numstat 2>/dev/null)

    # Count staged changes
    while IFS=$'\t' read -r add rem _; do
        [[ "$add" != "-" ]] && added=$((added + add))
        [[ "$rem" != "-" ]] && removed=$((removed + rem))
    done < <(git -C "$cwd" --no-optional-locks diff --cached --numstat 2>/dev/null)

    # Show changes if any exist
    if [ "$added" -gt 0 ] || [ "$removed" -gt 0 ]; then
        git_changes=$(printf " \033[32m+%d\033[00m/\033[31m-%d\033[00m" "$added" "$removed")
    fi

    if [ -n "$git_branch" ]; then
        git_branch=" (git:$git_branch)"
    fi
fi

# Build status line with colors (matching your PS1 style)
# Green for user@host, blue for path, magenta for worktree indicator, green/red for changes
status_line=$(printf "\033[01;32m%s@%s\033[00m:\033[01;34m%s\033[00m%s%s%s" "$user" "$host" "$cwd" "$git_branch" "$worktree_info" "$git_changes")

# Add context remaining if available
if [ -n "$remaining" ]; then
    status_line="$status_line $(printf "\033[33m[ctx:%s%%]\033[0m" "$remaining")"
fi

echo "$status_line"
