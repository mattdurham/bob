#!/bin/bash

# Install Go LSP (gopls) plugin for Claude
# Uses Claude's plugin system for gopls integration

set -e

echo "🔧 Installing Go LSP plugin..."
echo ""

# Check if gopls is installed
if ! command -v gopls &> /dev/null; then
    echo "⚠️  Warning: gopls not found in PATH"
    echo "   Installing gopls..."
    
    if ! command -v go &> /dev/null; then
        echo "❌ Error: Go not installed. Install Go first:"
        echo "   https://go.dev/doc/install"
        exit 1
    fi
    
    go install golang.org/x/tools/gopls@latest
    
    if ! command -v gopls &> /dev/null; then
        echo "⚠️  gopls installed but not in PATH"
        echo "   Add \$HOME/go/bin to your PATH:"
        echo "   export PATH=\$PATH:\$HOME/go/bin"
        exit 1
    fi
    
    echo "✅ gopls installed: $(gopls version)"
else
    echo "✅ gopls already installed: $(gopls version)"
fi

# Install Claude plugin for gopls
echo ""
if command -v claude &> /dev/null; then
    echo "🔧 Installing gopls-lsp Claude plugin..."
    
    if claude plugin install gopls-lsp@claude-plugins-official 2>&1; then
        echo "✅ gopls-lsp plugin installed"
        echo "   Go LSP now available in Claude for code intelligence"
    else
        EXIT_CODE=$?
        echo "⚠️  Failed to install gopls-lsp plugin (exit code: $EXIT_CODE)"
        echo "   Try manually: claude plugin install gopls-lsp@claude-plugins-official"
    fi
else
    echo "⚠️  Claude CLI not found - skipping plugin installation"
    echo "   Install Claude CLI and run: make install-lsp"
fi

# Install for Codex if available
echo ""
if command -v codex &> /dev/null; then
    echo "🔧 Checking Codex LSP support..."
    
    # Codex may use different plugin system
    echo "   ⚠️  Codex LSP configuration not implemented yet"
    echo "   Check Codex documentation for Go LSP setup"
else
    echo "   ⚠️  Codex CLI not found - skipping"
fi

echo ""
echo "✅ Go LSP installation complete"
echo ""
echo "Features enabled:"
echo "  ✓ Code completion"
echo "  ✓ Hover documentation"
echo "  ✓ Go-to-definition"
echo "  ✓ Find references"
echo "  ✓ Code actions"

exit 0
