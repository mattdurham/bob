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
	@echo "  make install-mcp              - Install Bob as MCP server in Claude CLI"
	@echo "  make install-guidance PATH=/path - Copy AGENTS.md & CLAUDE.md to repo"
	@echo "  make clean                    - Clean build artifacts"
	@echo "  make test                     - Run tests"

# Run Bob MCP server
run:
	@echo "🏴‍☠️ Starting Bob MCP server..."
	@cd cmd/bob && go run . --serve

# Build Bob binary
build: install-deps
	@echo "🔨 Building Bob from: $$(pwd)/cmd/bob"
	@cd cmd/bob && go build -o bob
	@echo "✅ Bob built: $$(pwd)/cmd/bob/bob"
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

# Install Bob as MCP server in Claude and Codex
install-mcp: build
	@echo "🏴‍☠️ Installing Bob as MCP server..."
	@if [ -z "$$HOME" ]; then \
		echo "❌ Error: HOME environment variable not set"; \
		exit 1; \
	fi; \
	BOB_INSTALL_DIR="$${HOME}/.bob"; \
	BOB_PATH="$${BOB_INSTALL_DIR}/bob"; \
	mkdir -p "$${BOB_INSTALL_DIR}"; \
	echo "🛑 Stopping any running Bob processes..."; \
	if pgrep -x bob > /dev/null 2>&1; then \
		killall bob 2>/dev/null || true; \
		for i in 1 2 3 4 5; do \
			if ! pgrep -x bob > /dev/null 2>&1; then break; fi; \
			sleep 1; \
		done; \
		if pgrep -x bob > /dev/null 2>&1; then \
			echo "⚠️  Warning: Some Bob processes still running"; \
			echo "   You may need to manually kill them: killall -9 bob"; \
		fi; \
	fi; \
	if ! cp cmd/bob/bob "$${BOB_PATH}"; then \
		echo "❌ Error: Failed to copy Bob binary to $${BOB_PATH}"; \
		echo "   Check disk space and permissions"; \
		exit 1; \
	fi; \
	if ! chmod +x "$${BOB_PATH}"; then \
		echo "❌ Error: Failed to make Bob executable"; \
		exit 1; \
	fi; \
	echo "✅ Installed Bob to $${BOB_PATH}"; \
	echo ""; \
	CONFIGURED=0; \
	echo "📦 Configuring Bob as MCP server..."; \
	echo ""; \
	if command -v claude > /dev/null 2>&1 && [ -x "$$(command -v claude)" ]; then \
		echo "🔧 Registering with Claude..."; \
		claude mcp remove bob 2>/dev/null || true; \
		if claude mcp add bob -- "$${BOB_PATH}" --serve 2>&1; then \
			echo "   ✅ Bob registered with Claude"; \
			CONFIGURED=1; \
		else \
			EXIT_CODE=$$?; \
			echo "   ❌ Failed to register with Claude (exit code: $${EXIT_CODE})"; \
			echo "   Try manually: claude mcp add bob -- $${BOB_PATH} --serve"; \
		fi; \
	else \
		echo "   ⚠️  Claude CLI not found - skipping Claude registration"; \
	fi; \
	echo ""; \
	if command -v codex > /dev/null 2>&1 && [ -x "$$(command -v codex)" ]; then \
		echo "🔧 Registering with Codex..."; \
		codex mcp remove bob 2>/dev/null || true; \
		if codex mcp add bob -- "$${BOB_PATH}" --serve 2>&1; then \
			echo "   ✅ Bob registered with Codex"; \
			CONFIGURED=1; \
		else \
			EXIT_CODE=$$?; \
			echo "   ❌ Failed to register with Codex (exit code: $${EXIT_CODE})"; \
			echo "   Try manually: codex mcp add bob -- $${BOB_PATH} --serve"; \
		fi; \
	else \
		echo "   ⚠️  Codex CLI not found - skipping Codex registration"; \
	fi; \
	echo ""; \
	if [ $${CONFIGURED} -eq 1 ]; then \
		echo "✅ Bob configured successfully"; \
		echo ""; \
		echo "🔄 Restart your CLI or start a new session to activate Bob"; \
	else \
		echo "⚠️  No MCP clients configured. Install Claude or Codex CLI and run 'make install-mcp' again."; \
		echo ""; \
		echo "Manual configuration:"; \
		echo "  Claude: claude mcp add bob -- $${BOB_PATH} --serve"; \
		echo "  Codex:  codex mcp add bob -- $${BOB_PATH} --serve"; \
	fi
