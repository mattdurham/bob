#!/bin/bash
# Pre-commit checks for Go projects
# Runs tests, linting, and formatting checks before allowing git commit

set -e

echo "🔍 Running pre-commit checks..."

# Only run checks if we're in a Go project
if [ ! -f "go.mod" ]; then
    echo "⚠️  Not a Go project (no go.mod), skipping checks"
    exit 0
fi

# Track if any checks fail
FAILED=0

# 1. Format check
echo "📝 Checking Go formatting..."
UNFORMATTED=$(gofmt -l . 2>/dev/null | grep -v vendor || true)
if [ -n "$UNFORMATTED" ]; then
    echo "❌ Code is not formatted. Running go fmt..."
    go fmt ./...
    echo "✅ Code formatted"
fi

# 2. Run tests
echo "🧪 Running tests (go test ./...)..."
if ! go test ./... -timeout=30s; then
    echo "❌ Tests failed!"
    FAILED=1
else
    echo "✅ All tests passed"
fi

# 3. Run golangci-lint if available
if command -v golangci-lint >/dev/null 2>&1; then
    echo "🔎 Running golangci-lint..."
    if ! golangci-lint run --timeout=2m; then
        echo "❌ Linting failed!"
        FAILED=1
    else
        echo "✅ Linting passed"
    fi
else
    echo "⚠️  golangci-lint not installed, skipping lint checks"
    echo "   Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
fi

# 4. Check cyclomatic complexity if gocyclo is available
if command -v gocyclo >/dev/null 2>&1; then
    echo "📊 Checking cyclomatic complexity..."
    COMPLEX=$(gocyclo -over 15 . 2>/dev/null | grep -v vendor || true)
    if [ -n "$COMPLEX" ]; then
        echo "⚠️  High complexity functions found:"
        echo "$COMPLEX"
        echo "   Consider refactoring functions with complexity > 15"
        # Don't fail on complexity, just warn
    else
        echo "✅ Complexity checks passed"
    fi
fi

# Final result
echo ""
if [ $FAILED -eq 0 ]; then
    echo "✅ All pre-commit checks passed - commit allowed"
    exit 0
else
    echo "❌ Pre-commit checks failed - commit blocked"
    echo ""
    echo "Fix the issues above and try again."
    exit 1
fi
