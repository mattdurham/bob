# Belayin' Pin Bob - Captain of Your Agents
# Makefile for building and running Bob workflow orchestrator

.PHONY: help run build install-deps clean test install-guidance install-mcp

help:
	@echo "🏴‍☠️ Belayin' Pin Bob - Captain of Your Agents"
	@echo ""
	@echo "Available targets:"
	@echo "  make run                      - Run Bob as MCP server"
	@echo "  make build                    - Build Bob binary"
	@echo "  make install-deps             - Install Go dependencies"
	@echo "  make install-mcp              - Install Bob as MCP server in Claude Desktop"
	@echo "  make install-guidance PATH=/path - Copy AGENTS.md & CLAUDE.md to repo"
	@echo "  make clean                    - Clean build artifacts"
	@echo "  make test                     - Run tests"

# Run Bob MCP server
run:
	@echo "🏴‍☠️ Starting Bob MCP server..."
	@cd cmd/bob && go run . --serve

# Build Bob binary
build: install-deps
	@echo "🔨 Building Bob..."
	@cd cmd/bob && go build -o bob
	@echo "✅ Bob built: cmd/bob/bob"
	@echo ""
	@echo "Run: ./cmd/bob/bob --serve"

# Install dependencies
install-deps:
	@echo "📦 Installing Go dependencies..."
	@cd cmd/bob && go mod download
	@echo "✅ Dependencies ready"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -f cmd/bob/bob
	@echo "✅ Clean complete"

# Run tests
test:
	@echo "🧪 Running Go tests..."
	@cd cmd/bob && go test ./... || true

# Install guidance files to another repo
install-guidance:
	@if [ -z "$(PATH)" ]; then \
		echo "❌ Error: PATH not specified"; \
		echo "Usage: make install-guidance PATH=/path/to/repo"; \
		exit 1; \
	fi
	@if [ ! -d "$(PATH)" ]; then \
		echo "❌ Error: Directory $(PATH) does not exist"; \
		exit 1; \
	fi
	@echo "🏴‍☠️ Installing Bob guidance to $(PATH)"
	@cp AGENTS.md "$(PATH)/AGENTS.md"
	@cp CLAUDE.md "$(PATH)/CLAUDE.md"
	@echo "✅ Installed:"
	@echo "   $(PATH)/AGENTS.md"
	@echo "   $(PATH)/CLAUDE.md"
	@echo ""
	@echo "These files configure the repo to use Bob MCP server."
	@echo "Commit them to your repo so Claude knows to use Bob!"

# Install Bob as MCP server in Claude Desktop
install-mcp: build
	@echo "🏴‍☠️ Installing Bob as MCP server..."
	@if [ -z "$$GITHUB_TOKEN" ]; then \
		echo "❌ Error: GITHUB_TOKEN environment variable not set"; \
		echo "Please set it in your environment: export GITHUB_TOKEN=your_token"; \
		exit 1; \
	fi
	@BOB_INSTALL_DIR="$$HOME/.bob"; \
	BOB_PATH="$$BOB_INSTALL_DIR/bob"; \
	mkdir -p "$$BOB_INSTALL_DIR"; \
	cp cmd/bob/bob "$$BOB_PATH"; \
	chmod +x "$$BOB_PATH"; \
	echo "✅ Installed Bob to $$BOB_PATH"; \
	if [ "$$(uname)" = "Darwin" ]; then \
		CONFIG_DIR="$$HOME/Library/Application Support/Claude"; \
	elif [ "$$(uname)" = "Linux" ]; then \
		CONFIG_DIR="$$HOME/.config/Claude"; \
	else \
		echo "❌ Error: Unsupported OS (only macOS and Linux supported)"; \
		exit 1; \
	fi; \
	CONFIG_FILE="$$CONFIG_DIR/claude_desktop_config.json"; \
	echo "📂 Config file: $$CONFIG_FILE"; \
	mkdir -p "$$CONFIG_DIR"; \
	if [ ! -f "$$CONFIG_FILE" ]; then \
		echo '{"mcpServers":{}}' > "$$CONFIG_FILE"; \
		echo "✅ Created new config file"; \
	fi; \
	if ! command -v jq > /dev/null 2>&1; then \
		echo "❌ Error: jq is required but not installed"; \
		echo "Install with: sudo apt-get install jq  (Linux)"; \
		echo "          or: brew install jq          (macOS)"; \
		exit 1; \
	fi; \
	TMP_FILE="$$(mktemp)"; \
	jq --arg bob_path "$$BOB_PATH" --arg github_token "$$GITHUB_TOKEN" \
		'.mcpServers.bob = {command: $$bob_path, args: ["--serve"], env: {GITHUB_TOKEN: $$github_token}}' \
		"$$CONFIG_FILE" > "$$TMP_FILE" && mv "$$TMP_FILE" "$$CONFIG_FILE"; \
	echo "✅ Bob configured as MCP server"; \
	echo ""; \
	echo "🔄 Restart Claude Desktop to activate Bob"; \
	echo ""; \
	echo "Configuration:"; \
	jq '.mcpServers.bob' "$$CONFIG_FILE"
