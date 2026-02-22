# Belayin' Pin Bob - Captain of Your Agents
# Makefile for installing Bob workflow skills and subagents

.PHONY: help install install-skills install-agents install-lsp install-mcp install-crush-skills install-crush-agents install-guidance install-statusline install-worktree install-personality allow hooks enable-agent-teams resolve-copilot ci clean

help:
	@echo "🏴‍☠️ Belayin' Pin Bob - Captain of Your Agents"
	@echo ""
	@echo "Bob is a workflow orchestration system implemented through Claude skills and subagents."
	@echo "No MCP servers needed - just intelligent workflow coordination!"
	@echo ""
	@echo "Available targets:"
	@echo "  make install                  - Install everything (skills + agents + LSP) [RECOMMENDED]"
	@echo "  make install-skills           - Install workflow skills to Claude Code"
	@echo "  make install-agents           - Install specialized subagents to Claude Code"
	@echo "  make install-crush-skills     - Install workflow skills to Crush"
	@echo "  make install-crush-agents     - Install specialized subagents to Crush"
	@echo "  make install-lsp              - Install Go LSP plugin"
	@echo "  make install-mcp [DIRS=...]   - Install filesystem MCP server (required for Bob)"
	@echo "                                  DIRS: comma-delimited paths (default: \$$HOME/source,/tmp)"
	@echo "  make install-guidance PATH=/path - Copy AGENTS.md & CLAUDE.md to repo"
	@echo "  make install-statusline       - Install statusline script and configure Claude Code"
	@echo "  make install-worktree         - Install create-worktree script to ~/.local/bin"
	@echo "  make install-personality [PERSONALITY=...] - Install Bob personality"
	@echo "                                  Options: default, pirate, cartoon_pirate (default: no override)"
	@echo "  make allow                    - Apply permissions from config/claude-permissions.json"
	@echo "  make enable-agent-teams       - Enable experimental agent teams feature"
	@echo "  make hooks                    - [OPTIONAL] Install pre-commit hooks (tests, linting, formatting)"
	@echo "  make ci                       - Run full CI pipeline locally (tests, lint, fmt, race, GHA)"
	@echo "  make resolve-copilot PR=<url> - Resolve Copilot review comments and re-request review"
	@echo "  make clean                    - Clean temporary files"
	@echo ""
	@echo "Quick start:"
	@echo "  make install                  - Install everything (skills + agents + LSP)"
	@echo "  make install-mcp              - Install filesystem MCP server (required)"
	@echo "  make enable-agent-teams       - Enable experimental agent teams (for /bob:team-work)"
	@echo "  make hooks                    - [OPTIONAL] Install pre-commit hooks"
	@echo "  make allow                    - Apply permissions"
	@echo "  /work \"feature description\" - Start a workflow"
	@echo ""
	@echo "Examples:"
	@echo "  make install PERSONALITY=pirate"
	@echo "  make install PERSONALITY=cartoon_pirate"
	@echo "  make install-mcp DIRS=\"/home/matt/projects,/tmp\""
	@echo "  make install-guidance PATH=/home/matt/myproject"
	@echo "  make install-statusline"

# Install workflow skills to Claude
install-skills:
	@echo "📚 Installing Bob workflow skills..."
	@SKILLS_DIR="$$HOME/.claude/skills"; \
	mkdir -p "$$SKILLS_DIR"; \
	for skill in work code-review performance explore brainstorming writing-plans project team-work; do \
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
	@echo "  /bob:project     - Project initialization (inspired by GSD)"
	@echo "  /bob:work        - Full development workflow"
	@echo "  /bob:code-review - Code review workflow"
	@echo "  /bob:performance - Performance optimization"
	@echo "  /bob:explore     - Codebase exploration"
	@echo "  /bob:team-work   - Team-based workflow (requires enable-agent-teams)"
	@echo "  /brainstorming   - Creative ideation"
	@echo "  /writing-plans   - Implementation planning"
	@echo "  /bob:version     - Show Bob version info"

# Install workflow skills to Crush
install-crush-skills:
	@echo "📚 Installing Bob workflow skills to Crush..."
	@CRUSH_SKILLS_DIR=$${CRUSH_SKILLS_DIR:-$$HOME/.config/crush/skills}; \
	mkdir -p "$$CRUSH_SKILLS_DIR"; \
	for skill in work code-review performance explore brainstorming writing-plans project; do \
		if [ -d "skills/$$skill" ]; then \
			echo "   Installing $$skill skill..."; \
			mkdir -p "$$CRUSH_SKILLS_DIR/$$skill"; \
			cp "skills/$$skill/SKILL.md" "$$CRUSH_SKILLS_DIR/$$skill/SKILL.md"; \
		else \
			echo "   ⚠️  Skill $$skill not found, skipping..."; \
		fi; \
	done
	@CRUSH_SKILLS_DIR=$${CRUSH_SKILLS_DIR:-$$HOME/.config/crush/skills}; \
	echo "   Generating bob-version skill..."; \
	GIT_HASH=$$(git rev-parse HEAD); \
	GIT_SHORT=$$(git rev-parse --short HEAD); \
	GIT_DATE=$$(git log -1 --format=%cd --date=format:'%Y-%m-%d %H:%M:%S'); \
	GIT_BRANCH=$$(git rev-parse --abbrev-ref HEAD); \
	GIT_REMOTE=$$(git config --get remote.origin.url || echo "local"); \
	INSTALL_DATE=$$(date '+%Y-%m-%d %H:%M:%S'); \
	BOB_REPO_PATH=$$(pwd); \
	SKILL_COUNT=$$(find skills -name "SKILL.md" -o -name "SKILL.md.template" | wc -l); \
	AGENT_COUNT=$$(find agents -name "SKILL.md" 2>/dev/null | wc -l || echo "0"); \
	mkdir -p "$$CRUSH_SKILLS_DIR/bob-version"; \
	sed -e "s|{{GIT_HASH}}|$$GIT_HASH|g" \
	    -e "s|{{GIT_DATE}}|$$GIT_DATE|g" \
	    -e "s|{{GIT_BRANCH}}|$$GIT_BRANCH|g" \
	    -e "s|{{GIT_REMOTE}}|$$GIT_REMOTE|g" \
	    -e "s|{{INSTALL_DATE}}|$$INSTALL_DATE|g" \
	    -e "s|{{BOB_REPO_PATH}}|$$BOB_REPO_PATH|g" \
	    -e "s|{{SKILL_COUNT}}|$$SKILL_COUNT|g" \
	    -e "s|{{AGENT_COUNT}}|$$AGENT_COUNT|g" \
	    skills/bob-version/SKILL.md.template > "$$CRUSH_SKILLS_DIR/bob-version/SKILL.md"
	@CRUSH_SKILLS_DIR=$${CRUSH_SKILLS_DIR:-$$HOME/.config/crush/skills}; \
	echo "✅ Skills installed to $$CRUSH_SKILLS_DIR"
	@echo ""
	@echo "Set CRUSH_SKILLS_DIR environment variable to use a custom directory:"
	@echo "  export CRUSH_SKILLS_DIR=/path/to/crush/skills"
	@echo "  make install-crush-skills"

# Install specialized subagents to Crush
install-crush-agents:
	@echo "🤖 Installing workflow subagents to Crush..."
	@CRUSH_SKILLS_DIR=$${CRUSH_SKILLS_DIR:-$$HOME/.config/crush/skills}; \
	mkdir -p "$$CRUSH_SKILLS_DIR"; \
	AGENT_COUNT=0; \
	if [ -d "agents" ]; then \
		for agent_dir in agents/*; do \
			if [ -d "$$agent_dir" ] && [ -f "$$agent_dir/SKILL.md" ]; then \
				agent=$$(basename "$$agent_dir"); \
				echo "   Installing $$agent agent..."; \
				mkdir -p "$$CRUSH_SKILLS_DIR/$$agent"; \
				cp "$$agent_dir/SKILL.md" "$$CRUSH_SKILLS_DIR/$$agent/SKILL.md"; \
				if [ -f "$$agent_dir/style.md" ]; then \
					cp "$$agent_dir/style.md" "$$CRUSH_SKILLS_DIR/$$agent/style.md"; \
				fi; \
				if [ -f "$$agent_dir/golang-pro.md" ]; then \
					cp "$$agent_dir/golang-pro.md" "$$CRUSH_SKILLS_DIR/$$agent/golang-pro.md"; \
				fi; \
				AGENT_COUNT=$$((AGENT_COUNT + 1)); \
			fi; \
		done; \
	else \
		echo "   ⚠️  No agents directory found"; \
	fi; \
	echo "✅ $$AGENT_COUNT subagents installed to $$CRUSH_SKILLS_DIR"
	@echo ""
	@echo "Set CRUSH_SKILLS_DIR environment variable to use a custom directory:"
	@echo "  export CRUSH_SKILLS_DIR=/path/to/crush/skills"
	@echo "  make install-crush-agents"

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
				if [ -f "$$agent_dir/style.md" ]; then \
					cp "$$agent_dir/style.md" "$$AGENTS_DIR/$$agent/style.md"; \
				fi; \
				if [ -f "$$agent_dir/golang-pro.md" ]; then \
					cp "$$agent_dir/golang-pro.md" "$$AGENTS_DIR/$$agent/golang-pro.md"; \
				fi; \
				AGENT_COUNT=$$((AGENT_COUNT + 1)); \
			fi; \
		done; \
	else \
		echo "   ⚠️  No agents directory found"; \
	fi; \
	echo "✅ $$AGENT_COUNT subagents installed to ~/.claude/agents/"
	@echo ""
	@echo "Specialized subagents available:"
	@echo ""
	@echo "Level 1 Orchestrators:"
	@echo "  workflow-coder                - EXECUTE phase coordinator (spawns 3 Level 2 agents)"
	@echo "  review-consolidator           - Merges 9 review findings into single report"
	@echo "  review-router                 - Makes routing decisions based on severity"
	@echo ""
	@echo "Level 2 Workers - Implementation:"
	@echo "  workflow-brainstormer         - Research & creative ideation"
	@echo "  workflow-planner              - Implementation planning"
	@echo "  workflow-implementer          - Code implementation (TDD, golang-pro guide)"
	@echo "  workflow-task-reviewer        - Task completion validation"
	@echo "  workflow-code-quality         - Go idioms & best practices (Uber Style Guide)"
	@echo "  workflow-tester               - Test execution and quality checks"
	@echo ""
	@echo "Level 2 Workers - Review (9 specialized reviewers):"
	@echo "  workflow-reviewer             - Multi-pass code quality review"
	@echo "  security-reviewer             - OWASP Top 10, vulnerability detection"
	@echo "  performance-analyzer          - Performance bottlenecks & optimization"
	@echo "  docs-reviewer                 - Documentation accuracy validation"
	@echo "  architect-reviewer            - Architecture & design review"
	@echo "  code-reviewer                 - Deep code quality analysis"
	@echo "  go-reviewer                   - Go-specific code review"
	@echo "  debugger                      - Bug diagnosis and debugging"
	@echo "  error-detective               - Error pattern analysis"
	@echo ""
	@echo "Level 2 Workers - Operations:"
	@echo "  commit-agent                  - Git operations & PR creation"
	@echo "  monitor-agent                 - CI/CD & PR monitoring"

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

# Install everything (skills, agents, LSP, personality) - PRIMARY COMMAND
# Usage: make install [PERSONALITY=pirate|cartoon_pirate]
install: install-skills install-agents install-crush-skills install-crush-agents install-lsp
	@if [ -n "$(PERSONALITY)" ] && [ "$(PERSONALITY)" != "default" ]; then \
		echo ""; \
		echo "🎭 Installing personality: $(PERSONALITY)..."; \
		$(MAKE) install-personality PERSONALITY=$(PERSONALITY); \
	fi
	@echo ""
	@echo "✅ Full installation complete!"
	@echo ""
	@echo "Installed to Claude Code:"
	@echo "  ✓ Workflow skills → ~/.claude/skills/"
	@echo "  ✓ Specialized subagents → ~/.claude/agents/"
	@if [ -f "$$HOME/.claude/bob-personality.md" ]; then \
		ACTIVE=$$(head -1 "$$HOME/.claude/bob-personality.md" | sed 's/# Bob Personality: //'); \
		echo "  ✓ Personality → $$ACTIVE"; \
	else \
		echo "  ✓ Personality → Default (built-in)"; \
	fi
	@echo ""
	@echo "Installed to Crush:"
	@echo "  ✓ Workflow skills → ~/.config/crush/skills/"
	@echo "  ✓ Specialized subagents → ~/.config/crush/skills/"
	@echo ""
	@echo "Installed:"
	@echo "  ✓ Go LSP plugin (if available)"
	@echo ""
	@echo "Optional (not installed by default):"
	@echo "  - Pre-commit hooks → Run 'make hooks' to install"
	@echo "  - Personality → Run 'make install PERSONALITY=pirate' or 'make install PERSONALITY=cartoon_pirate'"
	@echo ""
	@echo "🔄 Restart Claude/Crush to activate all components"
	@echo ""
	@echo "Quick start:"
	@echo "  /bob:work \"Add new feature\" - Start full development workflow"
	@echo "  /bob:team-work \"feature\"    - Team-based workflow (run 'make enable-agent-teams' first)"
	@echo "  /bob:code-review             - Review existing code"
	@echo "  /bob:performance             - Optimize performance"

# Install Bob personality
# Usage: make install-personality PERSONALITY=pirate|cartoon_pirate|default
# If PERSONALITY is not set or empty, removes any installed personality (uses built-in default)
# If PERSONALITY=default, also removes installed personality (same as built-in)
# When a personality is set, also injects a personality override into installed skills
install-personality:
	@PERSONALITY_FILE="$$HOME/.claude/bob-personality.md"; \
	SKILLS_DIR="$$HOME/.claude/skills"; \
	INJECT_LINE="## Personality Override"; \
	if [ -z "$(PERSONALITY)" ] || [ "$(PERSONALITY)" = "default" ]; then \
		if [ -f "$$PERSONALITY_FILE" ]; then \
			rm "$$PERSONALITY_FILE"; \
			echo "✅ Personality reset to default (removed override file)"; \
		else \
			echo "✅ Already using default personality (no override file)"; \
		fi; \
		for skill in work team-work brainstorming code-review explore performance project writing-plans; do \
			SKILL_FILE="$$SKILLS_DIR/$$skill/SKILL.md"; \
			if [ -f "$$SKILL_FILE" ] && grep -q "$$INJECT_LINE" "$$SKILL_FILE" 2>/dev/null; then \
				if [ -d "skills/$$skill" ] && [ -f "skills/$$skill/SKILL.md" ]; then \
					cp "skills/$$skill/SKILL.md" "$$SKILL_FILE"; \
					echo "   Restored $$skill skill to default"; \
				fi; \
			fi; \
		done; \
	elif [ -f "personalities/$(PERSONALITY).md" ]; then \
		cp "personalities/$(PERSONALITY).md" "$$PERSONALITY_FILE"; \
		echo "✅ Personality set to: $(PERSONALITY)"; \
		echo "   Installed to: $$PERSONALITY_FILE"; \
		for skill in work team-work brainstorming code-review explore performance project writing-plans; do \
			SKILL_FILE="$$SKILLS_DIR/$$skill/SKILL.md"; \
			if [ -f "$$SKILL_FILE" ]; then \
				if ! grep -q "$$INJECT_LINE" "$$SKILL_FILE" 2>/dev/null; then \
					TMP=$$(mktemp); \
					awk 'NR==1 && /^---$$/{front=1; print; next} front && /^---$$/{front=0; print; print ""; print "## Personality Override"; print ""; print "**Read `~/.claude/bob-personality.md` and adopt that personality for ALL user-facing messages.** The personality file'"'"'s voice, greetings, status updates, completions, errors, and vocabulary override all hardcoded messages in this document."; print ""; next} {print}' "$$SKILL_FILE" > "$$TMP" && mv "$$TMP" "$$SKILL_FILE"; \
					echo "   Injected personality override into $$skill skill"; \
				fi; \
			fi; \
		done; \
	else \
		echo "❌ Unknown personality: $(PERSONALITY)"; \
		echo "   Available personalities:"; \
		for p in personalities/*.md; do \
			name=$$(basename "$$p" .md); \
			echo "     - $$name"; \
		done; \
		exit 1; \
	fi
	@echo ""
	@echo "Available personalities:"
	@for p in personalities/*.md; do \
		name=$$(basename "$$p" .md); \
		if [ "$$name" = "$(PERSONALITY)" ]; then \
			echo "  → $$name (active)"; \
		else \
			echo "    $$name"; \
		fi; \
	done
	@echo ""
	@echo "🔄 Restart Claude Code for personality changes to take effect"

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

# Install statusline script and configure Claude Code to use it
install-statusline:
	@echo "📊 Installing Claude Code statusline..."
	@if [ ! -f "scripts/statusline-command.sh" ]; then \
		echo "❌ Error: scripts/statusline-command.sh not found"; \
		exit 1; \
	fi
	@cp scripts/statusline-command.sh "$$HOME/.claude/statusline-command.sh"
	@chmod +x "$$HOME/.claude/statusline-command.sh"
	@echo "✅ Installed statusline script to ~/.claude/statusline-command.sh"
	@if ! command -v jq >/dev/null 2>&1; then \
		echo "⚠️  jq not found - skipping settings.json update"; \
		echo "   Add this to ~/.claude/settings.json manually:"; \
		echo '   "statusLine": {"type": "command", "command": "$$HOME/.claude/statusline-command.sh", "padding": 0}'; \
		exit 0; \
	fi
	@SETTINGS_FILE="$$HOME/.claude/settings.json"; \
	if [ ! -f "$$SETTINGS_FILE" ]; then \
		echo '{}' > "$$SETTINGS_FILE"; \
	fi; \
	echo "Configuring statusLine in ~/.claude/settings.json..."; \
	cp "$$SETTINGS_FILE" "$$SETTINGS_FILE.backup"; \
	TMP_FILE=$$(mktemp); \
	SCRIPT_PATH="$$HOME/.claude/statusline-command.sh"; \
	jq --arg cmd "$$SCRIPT_PATH" '.statusLine = {"type": "command", "command": $$cmd, "padding": 0}' "$$SETTINGS_FILE" > "$$TMP_FILE"; \
	if [ $$? -eq 0 ]; then \
		mv "$$TMP_FILE" "$$SETTINGS_FILE"; \
		echo "✅ Configured statusLine in ~/.claude/settings.json"; \
		echo "✅ Backup saved to ~/.claude/settings.json.backup"; \
	else \
		echo "❌ Failed to update settings.json"; \
		rm -f "$$TMP_FILE"; \
		exit 1; \
	fi
	@echo ""
	@echo "Statusline shows:"
	@echo "  user@host:path (git:branch) [worktree:repo/task] +added/-removed [ctx:XX%]"
	@echo ""
	@echo "🔄 Restart Claude Code for the statusline to take effect"

# Install create-worktree script to ~/.local/bin
install-worktree:
	@echo "🌳 Installing create-worktree script..."
	@if [ ! -f "create-worktree.sh" ]; then \
		echo "❌ Error: create-worktree.sh not found"; \
		exit 1; \
	fi
	@mkdir -p "$$HOME/.local/bin"
	@cp create-worktree.sh "$$HOME/.local/bin/create-worktree"
	@chmod +x "$$HOME/.local/bin/create-worktree"
	@echo "✅ Installed to ~/.local/bin/create-worktree"
	@echo ""
	@SHELL_RC=""; \
	if [ -n "$$ZSH_VERSION" ] || [ -f "$$HOME/.zshrc" ]; then \
		SHELL_RC="$$HOME/.zshrc"; \
	elif [ -n "$$BASH_VERSION" ] || [ -f "$$HOME/.bashrc" ]; then \
		SHELL_RC="$$HOME/.bashrc"; \
	fi; \
	if [ -n "$$SHELL_RC" ]; then \
		if grep -q "^worktree()" "$$SHELL_RC" 2>/dev/null; then \
			echo "✅ Shell function already exists in $$SHELL_RC"; \
		else \
			echo "Adding worktree() shell function to $$SHELL_RC..."; \
			echo "" >> "$$SHELL_RC"; \
			echo "# Git worktree helper function - creates worktree in ../<repo>-worktrees/<branch> and cd's to it" >> "$$SHELL_RC"; \
			echo "worktree() {" >> "$$SHELL_RC"; \
			echo "    local branch=\"\$$1\"" >> "$$SHELL_RC"; \
			echo "    create-worktree \"\$$branch\" && cd \"\$$(git rev-parse --show-toplevel)/../\$$(basename \$$(git rev-parse --show-toplevel))-worktrees/\$$branch\"" >> "$$SHELL_RC"; \
			echo "}" >> "$$SHELL_RC"; \
			echo "✅ Added worktree() function to $$SHELL_RC"; \
		fi; \
	else \
		echo "⚠️  Could not detect shell RC file"; \
		echo "Add this to your ~/.bashrc or ~/.zshrc:"; \
		echo ""; \
		echo "  worktree() {"; \
		echo "      local branch=\"\$$1\""; \
		echo "      create-worktree \"\$$branch\" && cd \"\$$(git rev-parse --show-toplevel)/../\$$(basename \$$(git rev-parse --show-toplevel))-worktrees/\$$branch\""; \
		echo "  }"; \
	fi
	@echo ""
	@echo "Usage:"
	@echo "  worktree <branch-name>    - Create worktree and switch to it"
	@echo ""
	@echo "🔄 Reload your shell to use the worktree command:"
	@if [ -f "$$HOME/.zshrc" ]; then \
		echo "  source ~/.zshrc"; \
	elif [ -f "$$HOME/.bashrc" ]; then \
		echo "  source ~/.bashrc"; \
	fi
	@if ! echo "$$PATH" | grep -q "$$HOME/.local/bin"; then \
		echo ""; \
		echo "⚠️  Warning: ~/.local/bin is not in your PATH"; \
		echo "Add this to your ~/.bashrc or ~/.zshrc:"; \
		echo "  export PATH=\"\$$HOME/.local/bin:\$$PATH\""; \
	fi

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

# Enable experimental agent teams feature
enable-agent-teams:
	@echo "🧪 Enabling experimental agent teams feature..."
	@if ! command -v jq >/dev/null 2>&1; then \
		echo "❌ Error: jq is required but not installed"; \
		echo "Install with: sudo apt-get install jq  (or your package manager)"; \
		exit 1; \
	fi
	@SETTINGS_FILE="$$HOME/.claude/settings.json"; \
	if [ ! -f "$$SETTINGS_FILE" ]; then \
		echo "Creating new settings file..."; \
		echo '{}' > "$$SETTINGS_FILE"; \
	fi
	@echo "Backing up existing settings..."
	@cp "$$HOME/.claude/settings.json" "$$HOME/.claude/settings.json.backup"
	@TMP_FILE=$$(mktemp); \
	jq '.env.CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS = "1" | .teammateMode = "auto"' "$$HOME/.claude/settings.json" > "$$TMP_FILE"; \
	if [ $$? -eq 0 ]; then \
		mv "$$TMP_FILE" "$$HOME/.claude/settings.json"; \
		echo "✅ Experimental agent teams enabled"; \
		echo "✅ Backup saved to ~/.claude/settings.json.backup"; \
	else \
		echo "❌ Failed to update settings.json"; \
		rm -f "$$TMP_FILE"; \
		exit 1; \
	fi
	@echo ""
	@echo "Agent teams configuration:"
	@echo "  ✓ CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1"
	@echo "  ✓ teammateMode=auto (split panes if in tmux, otherwise in-process)"
	@echo ""
	@echo "Optional: Install tmux for split pane display"
	@if ! command -v tmux >/dev/null 2>&1; then \
		echo "  ⚠️  tmux not installed (split panes not available)"; \
		echo "  Install with: brew install tmux (macOS) or apt-get install tmux (Linux)"; \
	else \
		echo "  ✓ tmux is installed (split panes available)"; \
	fi
	@echo ""
	@echo "Usage:"
	@echo "  /bob:team-work \"Add new feature\" - Start team-based workflow"
	@echo ""
	@echo "🔄 Restart Claude Code for changes to take effect"

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

# Resolve Copilot review comments on a PR
# Usage: make resolve-copilot PR=https://github.com/owner/repo/pull/123
resolve-copilot:
	@if [ -z "$(PR)" ]; then \
		echo "❌ Error: PR is required"; \
		echo "Usage: make resolve-copilot PR=https://github.com/owner/repo/pull/123"; \
		exit 1; \
	fi
	@bash scripts/resolve-copilot-comments.sh "$(PR)"

# Run full CI pipeline locally (mirrors what GitHub Actions would run)
# This is the single command that must pass before committing.
ci:
	@echo "🔄 Running full CI pipeline locally..."
	@echo ""
	@PASS=0; FAIL=0; SKIP=0; \
	HAS_GO=$$(find . -name '*.go' -not -path './vendor/*' 2>/dev/null | head -1); \
	if [ -n "$$HAS_GO" ]; then \
		echo "── go test ./..."; \
		if go test ./... > /tmp/bob-ci.log 2>&1; then \
			echo "   ✅ PASS"; PASS=$$((PASS + 1)); \
		else \
			echo "   ❌ FAIL"; tail -20 /tmp/bob-ci.log | sed 's/^/   /'; FAIL=$$((FAIL + 1)); \
		fi; \
		echo "── go test -race ./..."; \
		if go test -race ./... > /tmp/bob-ci.log 2>&1; then \
			echo "   ✅ PASS"; PASS=$$((PASS + 1)); \
		else \
			echo "   ❌ FAIL"; tail -20 /tmp/bob-ci.log | sed 's/^/   /'; FAIL=$$((FAIL + 1)); \
		fi; \
		echo "── go test -cover ./..."; \
		if go test -cover ./... > /tmp/bob-ci.log 2>&1; then \
			echo "   ✅ PASS"; PASS=$$((PASS + 1)); \
		else \
			echo "   ❌ FAIL"; tail -20 /tmp/bob-ci.log | sed 's/^/   /'; FAIL=$$((FAIL + 1)); \
		fi; \
		echo "── go fmt"; \
		if test -z "$$(gofmt -l . 2>/dev/null)"; then \
			echo "   ✅ PASS"; PASS=$$((PASS + 1)); \
		else \
			echo "   ❌ FAIL"; gofmt -l . 2>/dev/null | sed 's/^/   /'; FAIL=$$((FAIL + 1)); \
		fi; \
		if command -v golangci-lint > /dev/null 2>&1; then \
			echo "── golangci-lint"; \
			if golangci-lint run > /tmp/bob-ci.log 2>&1; then \
				echo "   ✅ PASS"; PASS=$$((PASS + 1)); \
			else \
				echo "   ❌ FAIL"; tail -20 /tmp/bob-ci.log | sed 's/^/   /'; FAIL=$$((FAIL + 1)); \
			fi; \
		else \
			echo "── golangci-lint"; echo "   ⏭️  SKIP (not installed)"; SKIP=$$((SKIP + 1)); \
		fi; \
		if command -v gocyclo > /dev/null 2>&1; then \
			echo "── gocyclo (threshold: 40)"; \
			if ! gocyclo -over 40 . 2>/dev/null | grep -q .; then \
				echo "   ✅ PASS"; PASS=$$((PASS + 1)); \
			else \
				echo "   ❌ FAIL"; gocyclo -over 40 . 2>/dev/null | sed 's/^/   /'; FAIL=$$((FAIL + 1)); \
			fi; \
		else \
			echo "── gocyclo"; echo "   ⏭️  SKIP (not installed)"; SKIP=$$((SKIP + 1)); \
		fi; \
	else \
		echo "── go tests"; echo "   ⏭️  SKIP (no .go files found)"; SKIP=$$((SKIP + 1)); \
	fi; \
	if [ -d ".github/workflows" ]; then \
		for wf in .github/workflows/*.yml .github/workflows/*.yaml; do \
			[ -f "$$wf" ] || continue; \
			WF_NAME=$$(basename "$$wf"); \
			echo "── GHA: $$WF_NAME"; \
			grep -E '^\s+run:\s' "$$wf" 2>/dev/null | sed 's/.*run:\s*//' | while read -r cmd; do \
				[ -z "$$cmd" ] && continue; \
				echo "   → $$cmd"; \
				if eval "$$cmd" > /tmp/bob-ci.log 2>&1; then \
					echo "     ✅ PASS"; \
				else \
					echo "     ❌ FAIL"; tail -10 /tmp/bob-ci.log | sed 's/^/     /'; \
				fi; \
			done; \
		done; \
	else \
		echo "── GitHub Actions"; echo "   ⏭️  SKIP (no .github/workflows/ directory)"; SKIP=$$((SKIP + 1)); \
	fi; \
	echo ""; \
	echo "── Summary: $$PASS passed, $$FAIL failed, $$SKIP skipped"; \
	rm -f /tmp/bob-ci.log; \
	if [ "$$FAIL" -gt 0 ]; then \
		echo "❌ CI pipeline FAILED"; exit 1; \
	else \
		echo "✅ CI pipeline PASSED"; \
	fi

# Clean temporary files
clean:
	@echo "🧹 Cleaning temporary files..."
	@find . -name "*.tmp" -delete 2>/dev/null || true
	@find . -name ".DS_Store" -delete 2>/dev/null || true
	@echo "✅ Clean complete"
