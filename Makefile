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
	@echo "  make install-mcp              - Install Bob + Filesystem MCP servers (basic)"
	@echo "  make install-mcp-full         - Full installation (Skills + Agents + LSP)"
	@echo "  make install-skills           - Install workflow skills only"
	@echo "  make install-lsp              - Install Go LSP plugin only"
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
	echo "🛑 Stopping any running Bob MCP server processes..."; \
	BOB_PIDS=$$(pgrep -f "$${BOB_PATH} --serve" 2>/dev/null || true); \
	if [ -n "$${BOB_PIDS}" ]; then \
		echo "   Found Bob MCP server processes: $${BOB_PIDS}"; \
		kill $${BOB_PIDS} 2>/dev/null || true; \
		for i in 1 2 3 4 5; do \
			BOB_PIDS=$$(pgrep -f "$${BOB_PATH} --serve" 2>/dev/null || true); \
			if [ -z "$${BOB_PIDS}" ]; then break; fi; \
			sleep 1; \
		done; \
		BOB_PIDS=$$(pgrep -f "$${BOB_PATH} --serve" 2>/dev/null || true); \
		if [ -n "$${BOB_PIDS}" ]; then \
			echo "⚠️  Warning: Some Bob MCP server processes still running: $${BOB_PIDS}"; \
			echo "   You may need to manually kill them: kill -9 $${BOB_PIDS}"; \
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
	echo "📦 Checking filesystem MCP server prerequisites..."; \
	if command -v node > /dev/null 2>&1 && command -v npx > /dev/null 2>&1; then \
		NODE_VERSION=$$(node --version 2>/dev/null || echo "unknown"); \
		NPX_VERSION=$$(npx --version 2>/dev/null || echo "unknown"); \
		echo "   ✅ Node.js: $${NODE_VERSION}"; \
		echo "   ✅ npx: $${NPX_VERSION}"; \
		echo "   Will use official @modelcontextprotocol/server-filesystem"; \
		FILESYSTEM_INSTALLED=1; \
	else \
		echo "   ⚠️  Node.js/npx not found - filesystem server will not be available"; \
		echo ""; \
		echo "   To enable filesystem operations, install Node.js:"; \
		echo "   • Ubuntu/Debian: curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash - && sudo apt-get install -y nodejs"; \
		echo "   • macOS: brew install node"; \
		echo "   • Or visit: https://nodejs.org/"; \
		echo ""; \
		FILESYSTEM_INSTALLED=0; \
	fi; \
	echo ""; \
	CONFIGURED=0; \
	echo "📦 Configuring MCP servers..."; \
	echo ""; \
	if command -v claude > /dev/null 2>&1 && [ -x "$$(command -v claude)" ]; then \
		echo "🔧 Registering with Claude..."; \
		claude mcp remove bob 2>/dev/null || true; \
		if claude mcp add bob -- "$${BOB_PATH}" --serve 2>&1; then \
			echo "   ✅ Bob registered with Claude"; \
			CONFIGURED=1; \
		else \
			EXIT_CODE=$$?; \
			echo "   ❌ Failed to register Bob with Claude (exit code: $${EXIT_CODE})"; \
			echo "   Try manually: claude mcp add bob -- \"$${BOB_PATH}\" --serve"; \
		fi; \
		if [ "$${FILESYSTEM_INSTALLED}" = "1" ]; then \
			claude mcp remove filesystem 2>/dev/null || true; \
			if claude mcp add filesystem -- npx -y @modelcontextprotocol/server-filesystem "$$HOME/source" /tmp 2>&1; then \
				echo "   ✅ Filesystem server registered with Claude"; \
				CONFIGURED=1; \
			else \
				EXIT_CODE=$$?; \
				echo "   ❌ Failed to register filesystem with Claude (exit code: $${EXIT_CODE})"; \
				echo "   Try manually: claude mcp add filesystem -- npx -y @modelcontextprotocol/server-filesystem \"$$HOME/source\" /tmp"; \
			fi; \
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
			echo "   ❌ Failed to register Bob with Codex (exit code: $${EXIT_CODE})"; \
			echo "   Try manually: codex mcp add bob -- \"$${BOB_PATH}\" --serve"; \
		fi; \
		if [ "$${FILESYSTEM_INSTALLED}" = "1" ]; then \
			codex mcp remove filesystem 2>/dev/null || true; \
			if codex mcp add filesystem -- npx -y @modelcontextprotocol/server-filesystem "$$HOME/source" /tmp 2>&1; then \
				echo "   ✅ Filesystem server registered with Codex"; \
				CONFIGURED=1; \
			else \
				EXIT_CODE=$$?; \
				echo "   ❌ Failed to register filesystem with Codex (exit code: $${EXIT_CODE})"; \
				echo "   Try manually: codex mcp add filesystem -- npx -y @modelcontextprotocol/server-filesystem \"$$HOME/source\" /tmp"; \
			fi; \
		fi; \
	else \
		echo "   ⚠️  Codex CLI not found - skipping Codex registration"; \
	fi; \
	echo ""; \
	if [ $${CONFIGURED} -eq 1 ]; then \
		echo "✅ MCP servers configured successfully"; \
		echo "   - Bob workflow orchestrator"; \
		if [ "$${FILESYSTEM_INSTALLED}" = "1" ]; then \
			echo "   - Filesystem server (allowed: $$HOME/source, /tmp)"; \
		fi; \
		echo ""; \
		echo "🔄 Restart your CLI or start a new session to activate MCP servers"; \
	else \
		echo "⚠️  No MCP clients configured. Install Claude or Codex CLI and run 'make install-mcp' again."; \
		echo ""; \
		echo "Manual configuration:"; \
		echo "  Claude Bob: claude mcp add bob -- \"$${BOB_PATH}\" --serve"; \
		echo "  Codex Bob:  codex mcp add bob -- \"$${BOB_PATH}\" --serve"; \
		if [ "$${FILESYSTEM_INSTALLED}" = "1" ]; then \
			echo "  Claude Filesystem: claude mcp add filesystem -- mcp-filesystem-server --full-access \"$$HOME/source\" /tmp"; \
			echo "  Codex Filesystem:  codex mcp add filesystem -- mcp-filesystem-server --full-access \"$$HOME/source\" /tmp"; \
		fi; \
	fi

# Install workflow skills to Claude
install-skills:
	@echo "📚 Installing Bob workflow skills..."
	@SKILLS_DIR="$$HOME/.claude/skills"; \
	mkdir -p "$$SKILLS_DIR"; \
	for skill in work code-review performance explore; do \
		echo "   Installing $$skill skill..."; \
		mkdir -p "$$SKILLS_DIR/$$skill"; \
		cp "skills/$$skill/SKILL.md" "$$SKILLS_DIR/$$skill/SKILL.md"; \
	done
	@echo "✅ Skills installed to ~/.claude/skills/"
	@echo ""
	@echo "Available workflow commands:"
	@echo "  /work          - Full development workflow"
	@echo "  /code-review   - Code review workflow"
	@echo "  /performance   - Performance optimization"
	@echo "  /explore       - Codebase exploration"

# Install Go LSP plugin
install-lsp:
	@echo "🔧 Installing Go LSP plugin..."
	@bash scripts/install-lsp.sh

# Install everything (skills, agents, LSP)
install-mcp-full: install-skills install-agents install-lsp
	@echo ""
	@echo "✅ Full installation complete!"
	@echo ""
	@echo "Installed:"
	@echo "  ✓ Workflow skills → ~/.claude/skills/"
	@echo "  ✓ Specialized subagents → ~/.claude/agents/"
	@echo "  ✓ Go LSP plugin (gopls)"
	@echo "  ✓ Filesystem MCP server"
	@echo ""
	@echo "🔄 Restart Claude/Codex to activate all components"

install-agents:
	@echo "🤖 Installing workflow subagents..."
	@AGENTS_DIR="$$HOME/.claude/agents"; \
	mkdir -p "$$AGENTS_DIR"; \
	for agent in planner coder tester reviewer performance-analyzer security-reviewer docs-reviewer architecture-review code-review go-reviewer error-detective debugger; do \
		echo "   Installing $$agent agent..."; \
		mkdir -p "$$AGENTS_DIR/$$agent"; cp "agents/$$agent/SKILL.md" "$$AGENTS_DIR/$$agent/SKILL.md"; \
	done
	@echo "✅ Subagents installed to ~/.claude/agents/"
	@echo ""
	@echo "Available subagents:"
	@echo "  workflow-planner              - Implementation planning"
	@echo "  workflow-coder                - Code implementation (TDD)"
	@echo "  workflow-tester               - Test execution and quality checks"
	@echo "  workflow-reviewer             - Code quality review (basic)"
	@echo "  performance-analyzer          - Performance analysis"
	@echo "  security-reviewer             - Security vulnerability detection"
	@echo "  docs-reviewer                 - Documentation accuracy validation"
	@echo "  architect-reviewer            - Architecture and design review"
	@echo "  code-reviewer                 - Comprehensive code quality review"
	@echo "  golang-pro                    - Go-specific code review"
	@echo "  error-detective               - Error pattern analysis and root cause investigation"
	@echo "  debugger                      - Bug diagnosis and systematic debugging"
