#!/bin/bash

# Resolve all GitHub Copilot review comments on a PR and re-request review
# Usage: resolve-copilot-comments.sh [PR_NUMBER]

set -e

# --- Dependency checks ---

if ! command -v gh &> /dev/null; then
    echo "❌ Error: gh CLI not found"
    echo "   Install: https://cli.github.com/"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo "❌ Error: jq not found"
    echo "   Install: sudo apt install jq (or brew install jq)"
    exit 1
fi

# --- Help ---

if [[ "$1" == "--help" || "$1" == "-h" ]]; then
    echo "Usage: resolve-copilot-comments.sh <PR_URL_OR_NUMBER>"
    echo ""
    echo "Resolves all unresolved GitHub Copilot review comments on a PR"
    echo "and re-requests a Copilot review."
    echo ""
    echo "Arguments:"
    echo "  PR_URL_OR_NUMBER  GitHub PR URL or number (required)"
    echo "                    e.g. https://github.com/owner/repo/pull/123"
    echo "                    e.g. 123"
    echo ""
    echo "Options:"
    echo "  -h, --help    Show this help message"
    exit 0
fi

# --- Determine PR number ---

INPUT="$1"

if [[ -z "$INPUT" ]]; then
    echo "❌ Error: PR URL or number is required"
    echo "   Usage: resolve-copilot-comments.sh <PR_URL_OR_NUMBER>"
    echo "   Example: resolve-copilot-comments.sh https://github.com/owner/repo/pull/123"
    exit 1
fi

# Extract PR number from URL or use as-is
if [[ "$INPUT" =~ /pull/([0-9]+) ]]; then
    PR_NUMBER="${BASH_REMATCH[1]}"
    echo "✅ Extracted PR #${PR_NUMBER} from URL"
elif [[ "$INPUT" =~ ^[0-9]+$ ]]; then
    PR_NUMBER="$INPUT"
else
    echo "❌ Error: Could not parse PR number from: $INPUT"
    echo "   Expected a PR URL (https://github.com/owner/repo/pull/123) or number"
    exit 1
fi

# --- Get repo owner/name ---

REPO=$(gh repo view --json nameWithOwner --jq '.nameWithOwner')
if [[ -z "$REPO" ]]; then
    echo "❌ Error: Could not determine repository"
    exit 1
fi

OWNER=$(echo "$REPO" | cut -d'/' -f1)
REPO_NAME=$(echo "$REPO" | cut -d'/' -f2)

echo ""
echo "📋 Fetching review threads for ${REPO}#${PR_NUMBER}..."

# --- Query all review threads ---

THREADS_JSON=$(gh api graphql -f query='
query($owner: String!, $repo: String!, $pr: Int!) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $pr) {
      reviewThreads(first: 100) {
        nodes {
          id
          isResolved
          comments(first: 1) {
            nodes {
              author {
                login
              }
            }
          }
        }
      }
    }
  }
}' -f owner="$OWNER" -f repo="$REPO_NAME" -F pr="$PR_NUMBER")

# --- Filter unresolved Copilot threads ---

COPILOT_THREADS=$(echo "$THREADS_JSON" | jq -r '
  .data.repository.pullRequest.reviewThreads.nodes[]
  | select(.isResolved == false)
  | select(.comments.nodes[0].author.login == "copilot-pull-request-reviewer[bot]")
  | .id
')

if [[ -z "$COPILOT_THREADS" ]]; then
    echo "✅ No unresolved Copilot review threads found"
else
    THREAD_COUNT=$(echo "$COPILOT_THREADS" | wc -l | tr -d ' ')
    echo "🔧 Found ${THREAD_COUNT} unresolved Copilot thread(s)"
    echo ""

    RESOLVED=0
    FAILED=0

    while IFS= read -r THREAD_ID; do
        if gh api graphql -f query='
mutation($threadId: ID!) {
  resolveReviewThread(input: {threadId: $threadId}) {
    thread {
      isResolved
    }
  }
}' -f threadId="$THREAD_ID" > /dev/null 2>&1; then
            RESOLVED=$((RESOLVED + 1))
        else
            FAILED=$((FAILED + 1))
            echo "⚠️  Failed to resolve thread: ${THREAD_ID}"
        fi
    done <<< "$COPILOT_THREADS"

    echo "✅ Resolved ${RESOLVED}/${THREAD_COUNT} thread(s)"
    if [[ "$FAILED" -gt 0 ]]; then
        echo "⚠️  ${FAILED} thread(s) failed to resolve"
    fi
fi

# --- Re-request Copilot review ---

echo ""
echo "🔄 Re-requesting Copilot review..."

if gh extension list 2>/dev/null | grep -q "copilot-review"; then
    gh copilot-review "$PR_NUMBER"
    echo "✅ Copilot review re-requested via gh-copilot-review extension"
else
    echo "⚠️  gh-copilot-review extension not installed"
    echo "   Install: gh extension install ChrisCarini/gh-copilot-review"
    echo ""
    echo "   Falling back to --add-reviewer..."
    gh pr edit "$PR_NUMBER" --add-reviewer "copilot" 2>/dev/null || true
    echo "   ⚠️  Fallback may not trigger a new review if Copilot already reviewed."
    echo "   For reliable re-requests, install the extension above."
fi

echo ""
echo "🎉 Done! PR #${PR_NUMBER} Copilot comments resolved."
