#!/usr/bin/env bash
set -euo pipefail

# Usage: create-worktree <branch-name>
# Creates a git worktree in ../<source-directory>-worktrees/<branch-name>

if [ $# -eq 0 ]; then
    echo "Error: Branch name required"
    echo "Usage: create-worktree <branch-name>"
    exit 1
fi

BRANCH_NAME="$1"

# Ensure we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "Error: Not in a git repository"
    exit 1
fi

# Check if we're already in a worktree
COMMON_DIR=$(git rev-parse --git-common-dir 2>/dev/null || echo "")
GIT_DIR=$(git rev-parse --git-dir 2>/dev/null || echo "")

if [ "$COMMON_DIR" != "$GIT_DIR" ] && [ "$COMMON_DIR" != ".git" ]; then
    echo "Already in a worktree: $(git rev-parse --show-toplevel)"
    echo "Cannot create a worktree from within another worktree."
    echo "Navigate to the main repository first."
    exit 1
fi

# Get the git root directory (where .git is)
GIT_ROOT=$(git rev-parse --show-toplevel)
cd "$GIT_ROOT"

# Get the source directory name (e.g., "bob" from "/home/user/source/bob")
SOURCE_DIR=$(basename "$GIT_ROOT")

# Create worktree parent directory (e.g., "../bob-worktrees")
WORKTREE_PARENT="$(dirname "$GIT_ROOT")/${SOURCE_DIR}-worktrees"
mkdir -p "$WORKTREE_PARENT"

# Full path to the new worktree
WORKTREE_PATH="${WORKTREE_PARENT}/${BRANCH_NAME}"

# Check if worktree already exists
if [ -d "$WORKTREE_PATH" ]; then
    echo "Error: Worktree already exists at $WORKTREE_PATH"
    exit 1
fi

# Create the git worktree
echo "Creating worktree: $WORKTREE_PATH"
git worktree add -b "$BRANCH_NAME" "$WORKTREE_PATH"

# Output the cd command for the user to run
echo ""
echo "Worktree created successfully!"
echo "To switch to it, run:"
echo "  cd $WORKTREE_PATH"
