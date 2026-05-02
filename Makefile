# Belayin' Pin Bob - Captain of Your Agents
# Makefile for installing Bob workflow skills and subagents

SPEC ?= full

.PHONY: help all install install-skills install-agents install-lsp install-guidance install-statusline install-worktree install-personality install-plugins allow hooks enable-agent-teams resolve-copilot ci clean install-navigator install-no-python install-engram install-pi

all: install install-statusline install-worktree allow enable-agent-teams hooks install-engram
	@echo ""
	@echo "✅ Full system installation complete!"
	@echo "🔄 Restart Claude to activate all components"

help:
	@echo "🏴‍☠️ Belayin' Pin Bob - Captain of Your Agents"
	@echo ""
	@echo "Bob is a workflow orchestration system implemented through Claude skills and subagents."
	@echo ""
	@echo "Available targets:"
	@echo "  make all                      - Install + configure everything (the kitchen sink)"
	@echo "  make install                  - Install everything (skills + agents + LSP) [RECOMMENDED]"
	@echo "  make install SPEC=simple      - Install with simple spec mode (CLAUDE.md only per folder)"
	@echo "  make install-skills           - Install workflow skills to Claude Code"
	@echo "  make install-agents           - Install specialized subagents to Claude Code"
	@echo "  make install-lsp              - Install Go LSP plugin"
	@echo "  make install-plugins          - Install Claude plugins (grafana-engineering@grafana-ai-kit)"
	@echo "  make install-guidance PATH=/path - Copy AGENTS.md & CLAUDE.md to repo"
	@echo "  make install-no-python        - Add no-Python preference to ~/.claude/CLAUDE.md"
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
	@echo "  make install-navigator        - Build + install navigator HTTP MCP server to ~/.local/bin"
	@echo "  make install-engram           - Install engram persistent memory binary + Claude Code plugin"
	@echo "  make install-pi               - Install bob-agents pi extension to .pi/, skills to ~/.pi/agent/skills/"
	# @echo "  make install-bob-plugin       - Build + install bob Zellij plugin (requires Rust + zellij)"
	@echo ""
	@echo "Quick start:"
	@echo "  make install                  - Install everything (skills + agents + LSP)"
	@echo "  make enable-agent-teams       - Enable experimental agent teams (for /bob:work)"
	@echo "  make hooks                    - [OPTIONAL] Install pre-commit hooks"
	@echo "  make allow                    - Apply permissions"
	@echo "  /bob:work \"feature\" - Start a workflow"
	@echo ""
	@echo "Examples:"
	@echo "  make install PERSONALITY=pirate"
	@echo "  make install PERSONALITY=cartoon_pirate"
	@echo "  make install-guidance PATH=/home/matt/myproject"
	@echo "  make install-statusline"

# Install workflow skills to Claude
install-skills:
	@echo "📚 Installing Bob workflow skills..."
	@SKILLS_DIR="$$HOME/.claude/skills"; \
	mkdir -p "$$SKILLS_DIR"; \
	for skill in work explore brainstorming writing-plans audit code-review cleanup generate-overview stage-prs adversarial-review; do \
		if [ -d "skills/$$skill" ]; then \
			echo "   Installing $$skill skill..."; \
			mkdir -p "$$SKILLS_DIR/$$skill"; \
			if [ "$(SPEC)" = "simple" ] && [ -f "skills/$$skill/SKILL.simple.md" ]; then \
				cp "skills/$$skill/SKILL.simple.md" "$$SKILLS_DIR/$$skill/SKILL.md"; \
			else \
				cp "skills/$$skill/SKILL.md" "$$SKILLS_DIR/$$skill/SKILL.md"; \
			fi; \
		else \
			echo "   ⚠️  Skill $$skill not found, skipping..."; \
		fi; \
	done; \
	SKILLS_DIR="$$HOME/.claude/skills"; \
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
	SKILLS_DIR="$$HOME/.claude/skills"; \
	if command -v codex >/dev/null 2>&1; then \
		echo "   Installing talk-to-codex skill (codex CLI detected)..."; \
		mkdir -p "$$SKILLS_DIR/talk-to-codex"; \
		cp "skills/talk-to-codex/SKILL.md" "$$SKILLS_DIR/talk-to-codex/SKILL.md"; \
	else \
		echo "   ⏭️  Skipping talk-to-codex (codex CLI not installed)"; \
	fi
	@echo "✅ Skills installed to ~/.claude/skills/"
	@echo ""
	@echo "Available workflow commands:"
	@echo "  /bob:work        - Team-based workflow (requires enable-agent-teams)"
	@echo "  /bob:explore     - Team-based exploration with adversarial challenge"
	@echo "  /bob:audit       - Spec audit + optional Go structural analysis"
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
				if [ "$(SPEC)" = "simple" ] && [ -f "$$agent_dir/SKILL.simple.md" ]; then \
					cp "$$agent_dir/SKILL.simple.md" "$$AGENTS_DIR/$$agent/SKILL.md"; \
				else \
					cp "$$agent_dir/SKILL.md" "$$AGENTS_DIR/$$agent/SKILL.md"; \
				fi; \
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
	@echo "Orchestrators:"
	@echo "  workflow-coder                - EXECUTE phase coordinator"
	@echo "  review-consolidator           - Multi-domain code review"
	@echo ""
	@echo "Workers - Implementation:"
	@echo "  workflow-brainstormer         - Research & creative ideation"
	@echo "  workflow-planner              - Implementation planning"
	@echo "  workflow-implementer          - Code implementation (TDD)"
	@echo "  workflow-task-reviewer        - Task completion validation"
	@echo "  workflow-code-quality         - Go idioms & best practices"
	@echo "  workflow-tester               - Test execution and quality checks"
	@echo ""
	@echo "Workers - Operations:"
	@echo "  commit-agent                  - Git operations & PR creation"
	@echo "  monitor-agent                 - CI/CD & PR monitoring"
	@echo ""
	@echo "Workers - Teams:"
	@echo "  team-coder                    - Concurrent coder teammate"
	@echo "  team-reviewer                 - Concurrent reviewer teammate"
	@echo "  team-analyst                  - Concurrent analyst teammate (exploration)"
	@echo "  team-challenger               - Concurrent challenger teammate (exploration)"
	@echo "  Explore                       - Codebase exploration"

# Install Go LSP plugin
install-lsp:
	@echo "🔧 Installing Go LSP plugin..."
	@if [ -f "scripts/install-lsp.sh" ]; then \
		bash scripts/install-lsp.sh; \
	else \
		echo "   ⚠️  LSP installation script not found, skipping..."; \
	fi

# Install Claude plugins
install-plugins:
	@echo "🔌 Installing Claude plugins..."
	@if ! command -v claude >/dev/null 2>&1; then \
		echo "❌ Error: claude command not found"; \
		exit 1; \
	fi
	@echo "   Adding grafana/ai-kit to marketplace..."
	@claude plugin marketplace add grafana/ai-kit
	@echo "   Installing grafana-engineering@grafana-ai-kit..."
	@claude plugin install grafana-engineering@grafana-ai-kit
	@echo "✅ Claude plugins installed"

# Install everything (skills, agents, LSP, personality) - PRIMARY COMMAND
# Usage: make install [PERSONALITY=pirate|cartoon_pirate]
install: install-skills install-agents install-lsp install-plugins allow
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
	@echo "Installed:"
	@echo "  ✓ Go LSP plugin (if available)"
	@echo "  ✓ Claude plugins (grafana-engineering@grafana-ai-kit)"
	@echo ""
	@echo "Optional (not installed by default):"
	@echo "  - Pre-commit hooks → Run 'make hooks' to install"
	@echo "  - Personality → Run 'make install PERSONALITY=pirate' or 'make install PERSONALITY=cartoon_pirate'"
	@echo ""
	@echo "🔄 Restart Claude to activate all components"
	@echo ""
	@echo "Quick start:"
	@echo "  /bob:work \"Add new feature\"         - Start team-based workflow (run 'make enable-agent-teams' first)"

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
		for skill in work brainstorming explore writing-plans; do \
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
		for skill in work brainstorming explore writing-plans; do \
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

# Add no-Python language preference to ~/.claude/CLAUDE.md
install-no-python:
	@echo "🐍 Adding no-Python preference to ~/.claude/CLAUDE.md..."
	@CLAUDE_MD="$$HOME/.claude/CLAUDE.md"; \
	if [ ! -f "config/user-claude-no-python.md" ]; then \
		echo "❌ Error: config/user-claude-no-python.md not found"; \
		exit 1; \
	fi; \
	if [ -f "$$CLAUDE_MD" ] && grep -q "Do not write Python code" "$$CLAUDE_MD" 2>/dev/null; then \
		echo "✅ No-Python preference already present in $$CLAUDE_MD"; \
	else \
		echo "" >> "$$CLAUDE_MD"; \
		cat config/user-claude-no-python.md >> "$$CLAUDE_MD"; \
		echo "✅ Added no-Python preference to $$CLAUDE_MD"; \
	fi
	@echo ""
	@echo "Preference added:"
	@echo "  - Do not write Python; prefer Go or CLI scripts"
	@echo ""
	@echo "🔄 Restart Claude for changes to take effect"

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
	@FISH_CONFIG="$$HOME/.config/fish/config.fish"; \
	SHELL_RC=""; \
	if [ -f "$$FISH_CONFIG" ] || echo "$$SHELL" | grep -q "fish"; then \
		if grep -q "^function worktree" "$$FISH_CONFIG" 2>/dev/null; then \
			echo "✅ Fish function already exists in $$FISH_CONFIG"; \
		else \
			echo "Adding worktree function to $$FISH_CONFIG..."; \
			mkdir -p "$$HOME/.config/fish"; \
			echo "" >> "$$FISH_CONFIG"; \
			echo "# Git worktree helper function - creates worktree in ../<repo>-worktrees/<branch> and cd's to it" >> "$$FISH_CONFIG"; \
			echo "function worktree" >> "$$FISH_CONFIG"; \
			echo "    set -l branch \$$argv[1]" >> "$$FISH_CONFIG"; \
			echo "    create-worktree \$$branch; and cd (git rev-parse --show-toplevel)/../(basename (git rev-parse --show-toplevel))-worktrees/\$$branch" >> "$$FISH_CONFIG"; \
			echo "end" >> "$$FISH_CONFIG"; \
			echo "✅ Added worktree function to $$FISH_CONFIG"; \
		fi; \
	elif [ -n "$$ZSH_VERSION" ] || [ -f "$$HOME/.zshrc" ]; then \
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
	fi
	@echo ""
	@echo "Usage:"
	@echo "  worktree <branch-name>    - Create worktree and switch to it"
	@echo ""
	@echo "🔄 Reload your shell to use the worktree command:"
	@if echo "$$SHELL" | grep -q "fish"; then \
		echo "  source ~/.config/fish/config.fish"; \
	elif [ -f "$$HOME/.zshrc" ]; then \
		echo "  source ~/.zshrc"; \
	elif [ -f "$$HOME/.bashrc" ]; then \
		echo "  source ~/.bashrc"; \
	fi
	@if ! echo "$$PATH" | grep -q "$$HOME/.local/bin"; then \
		echo ""; \
		echo "⚠️  Warning: ~/.local/bin is not in your PATH"; \
		if echo "$$SHELL" | grep -q "fish"; then \
			echo "Add this to your ~/.config/fish/config.fish:"; \
			echo "  fish_add_path ~/.local/bin"; \
		else \
			echo "Add this to your ~/.bashrc or ~/.zshrc:"; \
			echo "  export PATH=\"\$$HOME/.local/bin:\$$PATH\""; \
		fi; \
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
	@echo "  /bob:work \"Add new feature\" - Start team-based workflow"
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
# install-bob-plugin:
# 	@echo "🔌 Building and installing bob Zellij plugin..."
# 	@command -v cargo >/dev/null 2>&1 || { echo "❌ cargo not found. Install Rust: https://rustup.rs"; exit 1; }
# 	@command -v zellij >/dev/null 2>&1 || { echo "❌ zellij not found. Install: https://zellij.dev"; exit 1; }
# 	@echo "   Building WASM plugin (this may take a while)..."
# 	@cargo build --release --target wasm32-wasip1 \
# 		--manifest-path cmd/bob-plugin/Cargo.toml
# 	@mkdir -p "$$HOME/.local/share/bob"
# 	@cp cmd/bob-plugin/target/wasm32-wasip1/release/bob_plugin.wasm \
# 		"$$HOME/.local/share/bob/bob-plugin.wasm"
# 	@mkdir -p "$$HOME/.config/zellij/layouts"
# 	@cp config/bob.kdl "$$HOME/.config/zellij/layouts/bob.kdl"
# 	@mkdir -p "$$HOME/.local/bin"
# 	@cp scripts/bob "$$HOME/.local/bin/bob"
# 	@chmod +x "$$HOME/.local/bin/bob"
# 	@cp scripts/statusline-command.sh "$$HOME/.claude/statusline-command.sh"
# 	@chmod +x "$$HOME/.claude/statusline-command.sh"
# 	@echo "✅ bob plugin installed"
# 	@echo "   Run 'bob' from any git repository to start"
# 	@echo "   Make sure ~/.local/bin is in your PATH"


LLAMA_VERSION ?= b8533
EMBED_MODEL_URL ?= https://huggingface.co/nomic-ai/nomic-embed-text-v1.5-GGUF/resolve/main/nomic-embed-text-v1.5.Q8_0.gguf
LLAMA_URL ?= https://github.com/ggml-org/llama.cpp/releases/download/$(LLAMA_VERSION)/llama-$(LLAMA_VERSION)-bin-ubuntu-x64.tar.gz

install-navigator:
	@echo "🧭 Building and installing navigator..."
	@if ! command -v go >/dev/null 2>&1; then \
		echo "❌ Error: go not found"; \
		echo "   Please install Go: https://go.dev/dl/"; \
		exit 1; \
	fi
	@mkdir -p "$$HOME/.local/bin"
	go build -o navigator ./cmd/navigator/
	install -m 0755 navigator ~/.local/bin/navigator
	@rm -f navigator
	@echo "✅ navigator binary installed"
	@echo ""
	@echo "   Downloading llama.cpp shared libraries..."
	@mkdir -p "$$HOME/.bob/navigator/lib"
	@if [ -f "$$HOME/.bob/navigator/lib/libllama.so" ]; then \
		echo "   ⏭️  llama.cpp libs already present"; \
	else \
		curl -sL "$(LLAMA_URL)" | tar xz -C "$$HOME/.bob/navigator/lib" --strip-components=1 --wildcards '*/lib*.so*' && \
		echo "   ✅ llama.cpp libs downloaded"; \
	fi
	@echo ""
	@echo "   Downloading nomic-embed-text model (~140MB)..."
	@mkdir -p "$$HOME/.bob/navigator/models"
	@if [ -f "$$HOME/.bob/navigator/models/nomic-embed-text-v1.5.Q8_0.gguf" ]; then \
		echo "   ⏭️  model already present"; \
	else \
		curl -sL -o "$$HOME/.bob/navigator/models/nomic-embed-text-v1.5.Q8_0.gguf" "$(EMBED_MODEL_URL)" && \
		echo "   ✅ nomic-embed-text model downloaded"; \
	fi
	@echo ""
	@echo "   Database: ~/.bob/navigator/thoughts.db"
	@echo ""
	@echo "   Registering navigator MCP server..."
	@if claude mcp list 2>/dev/null | grep -q "^navigator"; then \
		claude mcp remove navigator 2>/dev/null; \
	fi
	@claude mcp add --scope user navigator navigator && \
		echo "   ✅ navigator registered (stdio)"
	@echo "   Set NAVIGATOR_API_KEY in your shell profile to enable consult"
	@$(MAKE) allow
	@if ! echo "$$PATH" | grep -q "$$HOME/.local/bin"; then \
		echo ""; \
		echo "⚠️  Warning: ~/.local/bin is not in your PATH"; \
		if echo "$$SHELL" | grep -q "fish"; then \
			echo "Add to ~/.config/fish/config.fish:"; \
			echo "  fish_add_path ~/.local/bin"; \
		else \
			echo "Add to ~/.bashrc or ~/.zshrc:"; \
			echo "  export PATH=\"\$$HOME/.local/bin:\$$PATH\""; \
		fi; \
	fi

install-engram:
	@echo "🧠 Installing engram persistent memory..."
	@if ! command -v go >/dev/null 2>&1; then \
		echo "❌ Error: go not found"; \
		echo "   Please install Go: https://go.dev/dl/"; \
		exit 1; \
	fi
	@if command -v engram >/dev/null 2>&1; then \
		echo "   ⏭️  engram binary already installed"; \
	else \
		go install github.com/Gentleman-Programming/engram/cmd/engram@latest && \
		echo "✅ engram binary installed"; \
	fi
	@echo ""
	@echo "   Registering engram Claude Code plugin..."
	@if claude plugin list 2>/dev/null | grep -q "engram"; then \
		echo "   ⏭️  engram plugin already installed"; \
	else \
		claude plugin marketplace add Gentleman-Programming/engram && \
		claude plugin install engram && \
		echo "   ✅ engram plugin installed"; \
	fi
	@echo ""
	@echo "   Engram data: ~/.engram/engram.db"
	@echo "   TUI:         engram tui"
	@echo "   Restart Claude Code to activate"

# Install the bob-agents pi extension (project-local .pi/extensions/) and
# copy skills to the user-global ~/.pi/agent/skills/ so pi loads them as
# /bob:* commands from any project.
# Usage: make install-pi [SPEC=simple]
install-pi:
	@echo "🐦 Installing Bob pi components..."
	@echo ""
	@echo "🔌 Extension"
	@EXT_DIR=".pi/extensions/bob-agents"; \
	if [ -f "$$EXT_DIR/index.ts" ]; then \
		echo "   ✓ Already present: $$EXT_DIR"; \
	else \
		echo "   ⚠️  Extension source missing from $$EXT_DIR"; \
		echo "   Run this from the bob repo root."; \
		exit 1; \
	fi
	@echo ""
	@echo "📚 Skills (installing to ~/.pi/agent/skills/)"
	@SKILLS_DIR="$$HOME/.pi/agent/skills"; \
	mkdir -p "$$SKILLS_DIR"; \
	SKILL_COUNT=0; \
	for skill_dir in skills/*; do \
		[ -d "$$skill_dir" ] || continue; \
		skill=$$(basename "$$skill_dir"); \
		if [ "$(SPEC)" = "simple" ] && [ -f "$$skill_dir/SKILL.simple.md" ]; then \
			SRC="$$skill_dir/SKILL.simple.md"; \
		elif [ -f "$$skill_dir/SKILL.md" ]; then \
			SRC="$$skill_dir/SKILL.md"; \
		elif [ -f "$$skill_dir/SKILL.md.template" ]; then \
			GIT_HASH=$$(git rev-parse HEAD); \
			GIT_DATE=$$(git log -1 --format=%cd --date=format:'%Y-%m-%d %H:%M:%S'); \
			GIT_BRANCH=$$(git rev-parse --abbrev-ref HEAD); \
			GIT_REMOTE=$$(git config --get remote.origin.url || echo "local"); \
			INSTALL_DATE=$$(date '+%Y-%m-%d %H:%M:%S'); \
			BOB_REPO_PATH=$$(pwd); \
			SKILL_COUNT_ALL=$$(find skills -name "SKILL.md" | wc -l); \
			AGENT_COUNT_ALL=$$(find agents -name "SKILL.md" 2>/dev/null | wc -l || echo "0"); \
			mkdir -p "$$SKILLS_DIR/$$skill"; \
			sed -e "s|{{GIT_HASH}}|$$GIT_HASH|g" \
			    -e "s|{{GIT_DATE}}|$$GIT_DATE|g" \
			    -e "s|{{GIT_BRANCH}}|$$GIT_BRANCH|g" \
			    -e "s|{{GIT_REMOTE}}|$$GIT_REMOTE|g" \
			    -e "s|{{INSTALL_DATE}}|$$INSTALL_DATE|g" \
			    -e "s|{{BOB_REPO_PATH}}|$$BOB_REPO_PATH|g" \
			    -e "s|{{SKILL_COUNT}}|$$SKILL_COUNT_ALL|g" \
			    -e "s|{{AGENT_COUNT}}|$$AGENT_COUNT_ALL|g" \
			    "$$skill_dir/SKILL.md.template" > "$$SKILLS_DIR/$$skill/SKILL.md"; \
			echo "   Installing $$skill (from template)..."; \
			SKILL_COUNT=$$((SKILL_COUNT + 1)); \
			continue; \
		else \
			continue; \
		fi; \
		echo "   Installing $$skill..."; \
		mkdir -p "$$SKILLS_DIR/$$skill"; \
		cp "$$SRC" "$$SKILLS_DIR/$$skill/SKILL.md"; \
		SKILL_COUNT=$$((SKILL_COUNT + 1)); \
	done; \
	echo "✅ $$SKILL_COUNT skills installed to $$SKILLS_DIR"
	@echo ""
	@echo "✅ Pi installation complete!"
	@echo ""
	@echo "Installed:"
	@echo "  ✓ Extension → .pi/extensions/bob-agents/ (project-local, auto-discovered)"
	@echo "  ✓ Skills    → ~/.pi/agent/skills/ (user-global, available in all projects)"
	@echo ""
	@echo "Available skill commands in pi:"
	@find $$HOME/.pi/agent/skills -name "SKILL.md" -exec grep -m1 '^name:' {} \; 2>/dev/null \
		| sed 's/name: */  \//' | sort
	@echo ""
	@echo "Available tools in pi (from the extension):"
	@echo "  subagent              — spawn agents (single / parallel / chain)"
	@echo "  agent_status          — list running agents and their status"
	@echo "  mailbox_read          — read orchestrator mailbox"
	@echo "  mailbox_send_as       — send a message to a specific agent"
	@echo "  mailbox_broadcast     — broadcast to all active agents"
	@echo "  TaskCreate/List/Get/Update — shared task board"
	@echo "  /agents               — show agent status in pi UI"
	@echo ""
	@echo "🔄 Reload pi (/reload) or restart to activate"

clean:
	@echo "🧹 Cleaning temporary files..."
	@find . -name "*.tmp" -delete 2>/dev/null || true
	@find . -name ".DS_Store" -delete 2>/dev/null || true
	@echo "✅ Clean complete"
