# Belayin' Pin Bob - Captain of Your Agents
# Makefile for installing Bob workflow skills and subagents

.PHONY: help install install-skills install-agents install-lsp install-mcp install-guidance allow hooks clean

help:
	@echo "🏴‍☠️ Belayin' Pin Bob - Captain of Your Agents"
	@echo ""
	@echo "Bob is a workflow orchestration system implemented through Claude skills and subagents."
	@echo "No MCP servers needed - just intelligent workflow coordination!"
	@echo ""
	@echo "Available targets:"
	@echo "  make install                  - Install everything (skills + agents + LSP) [RECOMMENDED]"
	@echo "  make install-skills           - Install workflow skills only"
	@echo "  make install-agents           - Install specialized subagents"
	@echo "  make install-lsp              - Install Go LSP plugin"
	@echo "  make install-mcp [DIRS=...]   - Install filesystem MCP server (required for Bob)"
	@echo "                                  DIRS: comma-delimited paths (default: \$$HOME/source,/tmp)"
	@echo "  make install-guidance PATH=/path - Copy AGENTS.md & CLAUDE.md to repo"
	@echo "  make allow                    - Apply permissions from config/claude-permissions.json"
	@echo "  make hooks                    - [OPTIONAL] Install pre-commit hooks (tests, linting, formatting)"
	@echo "  make clean                    - Clean temporary files"
	@echo ""
	@echo "Quick start:"
	@echo "  make install                  - Install everything (skills + agents + LSP)"
	@echo "  make install-mcp              - Install filesystem MCP server (required)"
	@echo "  make hooks                    - [OPTIONAL] Install pre-commit hooks"
	@echo "  make allow                    - Apply permissions"
	@echo "  /work \"feature description\" - Start a workflow"
	@echo ""
	@echo "Examples:"
	@echo "  make install-mcp DIRS=\"/home/matt/projects,/tmp\""
	@echo "  make install-guidance PATH=/home/matt/myproject"

# Install workflow skills to Claude
install-skills:
	@echo "📚 Installing Bob workflow skills..."
	@SKILLS_DIR="$$HOME/.claude/skills"; \
	mkdir -p "$$SKILLS_DIR"; \
	for skill in work code-review performance explore brainstorming writing-plans; do \
		if [ -d "skills/$$skill" ]; then \
			echo "   Installing $$skill skill..."; \
			mkdir -p "$$SKILLS_DIR/$$skill"; \
			cp "skills/$$skill/SKILL.md" "$$SKILLS_DIR/$$skill/SKILL.md"; \
		else \
			echo "   ⚠️  Skill $$skill not found, skipping..."; \
		fi; \
	done
	@SKILLS_DIR="$$HOME/.claude/skills"; \
	echo "   Generating bob:version skill..."; \
	GIT_HASH=$$(git rev-parse HEAD); \
	GIT_SHORT=$$(git rev-parse --short HEAD); \
	GIT_DATE=$$(git log -1 --format=%cd --date=format:'%Y-%m-%d %H:%M:%S'); \
	GIT_BRANCH=$$(git rev-parse --abbrev-ref HEAD); \
	GIT_REMOTE=$$(git config --get remote.origin.url || echo "local"); \
	INSTALL_DATE=$$(date '+%Y-%m-%d %H:%M:%S'); \
	BOB_REPO_PATH=$$(pwd); \
	SKILL_COUNT=$$(find skills -name "SKILL.md" -o -name "SKILL.md.template" | wc -l); \
	AGENT_COUNT=$$(find agents -name "SKILL.md" 2>/dev/null | wc -l || echo "0"); \
	if [ -f "$$HOME/.claude/hooks-config.json" ] && [ -f "$$HOME/.claude/hooks/pre-commit-checks.sh" ]; then \
		HOOKS_STATUS="**Hooks:** ✓ Installed\n- Pre-commit quality checks (tests, linting, formatting)\n- Run \`make hooks\` to reinstall or update"; \
	else \
		HOOKS_STATUS="**Hooks:** ✗ Not installed\n- Run \`make hooks\` to install pre-commit quality checks"; \
	fi; \
	mkdir -p "$$SKILLS_DIR/bob-version"; \
	sed -e "s|{{GIT_HASH}}|$$GIT_HASH|g" \
	    -e "s|{{GIT_DATE}}|$$GIT_DATE|g" \
	    -e "s|{{GIT_BRANCH}}|$$GIT_BRANCH|g" \
	    -e "s|{{GIT_REMOTE}}|$$GIT_REMOTE|g" \
	    -e "s|{{INSTALL_DATE}}|$$INSTALL_DATE|g" \
	    -e "s|{{BOB_REPO_PATH}}|$$BOB_REPO_PATH|g" \
	    -e "s|{{SKILL_COUNT}}|$$SKILL_COUNT|g" \
	    -e "s|{{AGENT_COUNT}}|$$AGENT_COUNT|g" \
	    -e "s|{{HOOKS_STATUS}}|$$HOOKS_STATUS|g" \
	    skills/bob-version/SKILL.md.template > "$$SKILLS_DIR/bob-version/SKILL.md"
	@echo "✅ Skills installed to ~/.claude/skills/"
	@echo ""
	@echo "Available workflow commands:"
	@echo "  /bob:work        - Full development workflow"
	@echo "  /bob:code-review - Code review workflow"
	@echo "  /bob:performance - Performance optimization"
	@echo "  /bob:explore     - Codebase exploration"
	@echo "  /brainstorming   - Creative ideation"
	@echo "  /writing-plans   - Implementation planning"
	@echo "  /bob:version     - Show Bob version info"

# Install specialized subagents
install-agents:
	@echo "🤖 Installing workflow subagents..."
	@AGENTS_DIR="$$HOME/.claude/agents"; \
	mkdir -p "$$AGENTS_DIR"; \
	AGENT_COUNT=0; \
	if [ -d "agents" ]; then \
		for agent_dir in agents/*; do \
			if [ -d "$$agent_dir" ] && [ -f "$$agent_dir/SKILL.md" ]; then \
				agent=$$(basename "$$agent_dir"); \
				echo "   Installing $$agent agent..."; \
				mkdir -p "$$AGENTS_DIR/$$agent"; \
				cp "$$agent_dir/SKILL.md" "$$AGENTS_DIR/$$agent/SKILL.md"; \
				AGENT_COUNT=$$((AGENT_COUNT + 1)); \
			fi; \
		done; \
	else \
		echo "   ⚠️  No agents directory found"; \
	fi; \
	echo "✅ $$AGENT_COUNT subagents installed to ~/.claude/agents/"
	@echo ""
	@echo "Specialized subagents available:"
	@echo "  workflow-planner              - Implementation planning"
	@echo "  workflow-coder                - Code implementation (TDD)"
	@echo "  workflow-tester               - Test execution and quality checks"
	@echo "  workflow-reviewer             - Code quality review"
	@echo "  performance-analyzer          - Performance analysis"
	@echo "  security-reviewer             - Security vulnerability detection"
	@echo "  docs-reviewer                 - Documentation accuracy validation"
	@echo "  architect-reviewer            - Architecture and design review"
	@echo "  code-reviewer                 - Comprehensive code quality review"
	@echo "  golang-pro                    - Go-specific code review"
	@echo "  error-detective               - Error pattern analysis"
	@echo "  debugger                      - Bug diagnosis and debugging"

# Install Go LSP plugin
install-lsp:
	@echo "🔧 Installing Go LSP plugin..."
	@if [ -f "scripts/install-lsp.sh" ]; then \
		bash scripts/install-lsp.sh; \
	else \
		echo "   ⚠️  LSP installation script not found, skipping..."; \
	fi

# Install filesystem MCP server (required for Bob workflows)
# Usage: make install-mcp [DIRS=/path1,/path2,/path3]
# If DIRS not specified, defaults to $HOME/source and /tmp
install-mcp:
	@echo "📁 Installing filesystem MCP server..."
	@if ! command -v claude >/dev/null 2>&1; then \
		echo "❌ Error: claude command not found"; \
		echo "   Please install Claude Code first"; \
		exit 1; \
	fi
	@if ! command -v npm >/dev/null 2>&1; then \
		echo "❌ Error: npm not found"; \
		echo "   Please install Node.js and npm first:"; \
		echo "   - Ubuntu/Debian: sudo apt-get install nodejs npm"; \
		echo "   - macOS: brew install node"; \
		echo "   - Or visit: https://nodejs.org/"; \
		exit 1; \
	fi
	@if ! command -v npx >/dev/null 2>&1; then \
		echo "❌ Error: npx not found"; \
		echo "   Please install Node.js (npx comes with npm 5.2+):"; \
		echo "   - Ubuntu/Debian: sudo apt-get install nodejs npm"; \
		echo "   - macOS: brew install node"; \
		echo "   - Or visit: https://nodejs.org/"; \
		exit 1; \
	fi
	@if [ -n "$(DIRS)" ]; then \
		MCP_DIRS=$$(echo "$(DIRS)" | tr ',' ' '); \
	else \
		MCP_DIRS="$$HOME/source /tmp"; \
	fi; \
	if claude mcp list | grep -q "filesystem:"; then \
		echo "   ⚠️  Filesystem MCP server already installed"; \
		echo "   Remove it first with: claude mcp remove filesystem"; \
	else \
		echo "   Installing filesystem MCP server..."; \
		claude mcp add filesystem -- npx -y @modelcontextprotocol/server-filesystem $$MCP_DIRS; \
		echo "   ✅ Filesystem MCP server installed"; \
		echo ""; \
		echo "Configured directories:"; \
		for dir in $$MCP_DIRS; do \
			echo "  ✓ $$dir"; \
		done; \
	fi

# Install everything (skills, agents, LSP) - PRIMARY COMMAND
install: install-skills install-agents install-lsp
	@echo ""
	@echo "✅ Full installation complete!"
	@echo ""
	@echo "Installed:"
	@echo "  ✓ Workflow skills → ~/.claude/skills/"
	@echo "  ✓ Specialized subagents → ~/.claude/agents/"
	@echo "  ✓ Go LSP plugin (if available)"
	@echo ""
	@echo "Optional (not installed by default):"
	@echo "  - Pre-commit hooks → Run 'make hooks' to install"
	@echo ""
	@echo "🔄 Restart Claude to activate all components"
	@echo ""
	@echo "Quick start:"
	@echo "  /work \"Add new feature\"     - Start full development workflow"
	@echo "  /code-review                 - Review existing code"
	@echo "  /performance                 - Optimize performance"

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
	@cp CLAUDE.md "$(PATH)/CLAUDE.md"
	@if [ -f "AGENTS.md" ]; then \
		cp AGENTS.md "$(PATH)/AGENTS.md"; \
		echo "✅ Installed: $(PATH)/AGENTS.md"; \
	fi
	@echo "✅ Installed: $(PATH)/CLAUDE.md"
	@echo ""
	@echo "These files configure the repo to use Bob workflow skills."
	@echo "Commit them to your repo so Claude knows about Bob workflows!"

# Apply permissions from config to ~/.claude/settings.json
allow:
	@echo "🔐 Applying Claude permissions..."
	@if [ ! -f "config/claude-permissions.json" ]; then \
		echo "❌ Error: config/claude-permissions.json not found"; \
		exit 1; \
	fi
	@if ! command -v jq >/dev/null 2>&1; then \
		echo "❌ Error: jq is required but not installed"; \
		echo "Install with: sudo apt-get install jq  (or your package manager)"; \
		exit 1; \
	fi
	@SETTINGS_FILE="$$HOME/.claude/settings.json"; \
	if [ ! -f "$$SETTINGS_FILE" ]; then \
		echo "Creating new settings file..."; \
		cp config/claude-permissions.json "$$SETTINGS_FILE"; \
	else \
		echo "Backing up existing settings..."; \
		cp "$$SETTINGS_FILE" "$$SETTINGS_FILE.backup"; \
		echo "Intelligently merging permissions (union of allow lists)..."; \
		TMP_FILE=$$(mktemp); \
		jq -s '.[0] as $$existing | .[1] as $$config | $$existing * $$config | .permissions.allow = (($$existing.permissions.allow // []) + ($$config.permissions.allow // []) | unique)' "$$SETTINGS_FILE" config/claude-permissions.json > "$$TMP_FILE"; \
		if [ $$? -eq 0 ]; then \
			mv "$$TMP_FILE" "$$SETTINGS_FILE"; \
			echo "✅ Backup saved to: $$SETTINGS_FILE.backup"; \
		else \
			echo "❌ Merge failed, restoring from backup"; \
			rm -f "$$TMP_FILE"; \
			exit 1; \
		fi; \
	fi
	@echo "✅ Permissions applied to ~/.claude/settings.json"
	@echo ""
	@echo "Active permissions:"
	@jq -r '.permissions.allow[]' "$$HOME/.claude/settings.json" | sed 's/^/  ✓ /'
	@echo ""
	@echo "Default mode: $$(jq -r '.permissions.defaultMode' "$$HOME/.claude/settings.json")"

# Install pre-commit hooks
hooks:
	@echo "🪝 Installing pre-commit hooks..."
	@if [ ! -d "hooks" ]; then \
		echo "❌ Error: hooks/ directory not found"; \
		exit 1; \
	fi
	@if ! command -v jq >/dev/null 2>&1; then \
		echo "❌ Error: jq is required but not installed"; \
		echo "Install with: sudo apt-get install jq  (or your package manager)"; \
		exit 1; \
	fi
	@echo "Installing hook scripts..."
	@mkdir -p "$$HOME/.claude/hooks"
	@cp hooks/pre-commit-checks.sh "$$HOME/.claude/hooks/"
	@chmod +x "$$HOME/.claude/hooks/pre-commit-checks.sh"
	@if [ -f "hooks/README.md" ]; then \
		cp hooks/README.md "$$HOME/.claude/hooks/"; \
	fi
	@echo "✅ Hook scripts installed"
	@echo ""
	@HOOKS_CONFIG="$$HOME/.claude/hooks-config.json"; \
	if [ ! -f "$$HOOKS_CONFIG" ]; then \
		echo "Creating new hooks configuration..."; \
		cp hooks/hooks-config.json "$$HOOKS_CONFIG"; \
	else \
		echo "Backing up existing hooks configuration..."; \
		cp "$$HOOKS_CONFIG" "$$HOOKS_CONFIG.backup"; \
		echo "Merging hooks configuration..."; \
		TMP_FILE=$$(mktemp); \
		jq -s '.[0] as $$existing | .[1] as $$new | $$existing * $$new | .hooks.PreToolUse = (($$existing.hooks.PreToolUse // []) + ($$new.hooks.PreToolUse // []) | unique_by(.matcher))' "$$HOOKS_CONFIG" hooks/hooks-config.json > "$$TMP_FILE"; \
		if [ $$? -eq 0 ]; then \
			mv "$$TMP_FILE" "$$HOOKS_CONFIG"; \
			echo "✅ Backup saved to: $$HOOKS_CONFIG.backup"; \
		else \
			echo "❌ Merge failed, restoring from backup"; \
			rm -f "$$TMP_FILE"; \
			exit 1; \
		fi; \
	fi
	@echo "✅ Hooks configuration merged"
	@echo ""
	@echo "Enabling hookify plugin..."
	@SETTINGS_FILE="$$HOME/.claude/settings.json"; \
	if [ -f "$$SETTINGS_FILE" ]; then \
		TMP_FILE=$$(mktemp); \
		jq '.enabledPlugins."hookify@claude-plugins-official" = true' "$$SETTINGS_FILE" > "$$TMP_FILE" && mv "$$TMP_FILE" "$$SETTINGS_FILE"; \
		echo "✅ Hookify plugin enabled"; \
	fi
	@echo ""
	@echo "📋 Installed hooks:"
	@echo "  ✓ pre-commit-checks.sh - Runs tests, linting, formatting before commits"
	@echo "  ✓ hookify plugin enabled"
	@echo ""
	@echo "🔍 Hook will run automatically before 'git commit' commands"
	@echo "   Blocks commits if:"
	@echo "   - Tests fail (go test ./...)"
	@echo "   - Linting fails (golangci-lint)"
	@echo "   - Code not formatted (go fmt)"
	@echo ""
	@echo "🔄 Restart Claude Code for hooks to take effect"
	@echo "📚 See ~/.claude/hooks/README.md for details"

# Clean temporary files
clean:
	@echo "🧹 Cleaning temporary files..."
	@find . -name "*.tmp" -delete 2>/dev/null || true
	@find . -name ".DS_Store" -delete 2>/dev/null || true
	@echo "✅ Clean complete"
