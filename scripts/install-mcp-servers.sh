#!/bin/bash

# Install GitHub MCP server for Bob
# Uses official @github/mcp-server (made by GitHub)

set -e

echo "📦 Installing GitHub MCP server..."
echo ""

# Check npm/npx available
if ! command -v npx &> /dev/null; then
    echo "⚠️  Warning: npx not found. Skipping GitHub MCP server installation."
    echo "   Install Node.js/npm to use GitHub MCP server."
    exit 0
fi

# Install GitHub MCP Server
echo "🔧 Installing official GitHub MCP server..."
echo "   Package: @github/mcp-server"

# Test if package is available
if npx --yes @github/mcp-server --version &> /dev/null 2>&1; then
    echo "   ✅ GitHub MCP server available"
    GITHUB_INSTALLED=1
else
    echo "   ⚠️  GitHub MCP server package not found"
    echo "   Trying alternative: npx @github/mcp-server"
    GITHUB_INSTALLED=0
fi

# Register with Claude CLI
echo ""
if command -v claude &> /dev/null; then
    echo "🔧 Registering GitHub MCP server with Claude..."

    if [ "$GITHUB_INSTALLED" = "1" ]; then
        claude mcp remove github 2>/dev/null || true

        # Try to register (may need GitHub token)
        if claude mcp add github -- npx -y @github/mcp-server 2>&1; then
            echo "   ✅ GitHub MCP server registered with Claude"
            echo "   Note: You may need to configure GitHub token for authentication"
            echo "   See: https://github.com/github/github-mcp-server"
        else
            echo "   ⚠️  Failed to register GitHub MCP server"
            echo "   Try manually: claude mcp add github -- npx -y @github/mcp-server"
            echo "   Documentation: https://github.com/github/github-mcp-server"
        fi
    else
        echo "   ⚠️  Skipping registration - package not available"
    fi
else
    echo "   ⚠️  Claude CLI not found - skipping registration"
fi

# Register with Codex CLI
echo ""
if command -v codex &> /dev/null; then
    echo "🔧 Registering GitHub MCP server with Codex..."

    if [ "$GITHUB_INSTALLED" = "1" ]; then
        codex mcp remove github 2>/dev/null || true

        if codex mcp add github -- npx -y @github/mcp-server 2>&1; then
            echo "   ✅ GitHub MCP server registered with Codex"
        else
            echo "   ⚠️  Failed to register with Codex"
        fi
    fi
else
    echo "   ⚠️  Codex CLI not found - skipping"
fi

# Summary
echo ""
if [ "$GITHUB_INSTALLED" = "1" ]; then
    echo "✅ GitHub MCP server installation complete"
    echo ""
    echo "Server provides:"
    echo "  • GitHub API integration"
    echo "  • Repository operations"
    echo "  • Issue and PR management"
    echo "  • Code search capabilities"
    echo ""
    echo "Next steps:"
    echo "  1. Configure GitHub token (if needed)"
    echo "  2. Restart Claude CLI"
    echo "  3. Test with GitHub operations"
else
    echo "⚠️  GitHub MCP server not available"
    echo "   This is normal if the package hasn't been published yet"
    echo "   Bob will work fine without it"
    echo ""
    echo "   Check status: https://github.com/github/github-mcp-server"
fi

exit 0
