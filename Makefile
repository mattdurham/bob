# Belayin' Pin Bob - Captain of Your Agents
# Makefile for installing Bob workflow skills and subagents

SPEC ?= full
CODEX_HOME ?= $(HOME)/.codex
WLLR_HOME ?= $(HOME)/.wllr

.PHONY: help all install install-skills install-agents install-lsp install-guidance install-statusline install-worktree install-personality install-plugins allow hooks enable-agent-teams resolve-copilot ci clean install-no-python install-engram install-pi install-pi-skills install-codex-skills install-wllr install-wllr-skills check-agents

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
	@echo "  make install-engram           - Install engram persistent memory binary + Claude Code plugin"
	@echo "  make install-pi               - Install bob-agents pi extension + skills to .pi/"
	@echo "  make install-pi-skills        - Install Bob skills for pi to ~/.pi/agent/skills/"
	@echo "  make install-codex-skills     - Install Bob skills for Codex to ~/.codex/skills/"
	@echo "  make install-wllr             - Install Bob skills for wllr to ~/.wllr/skills/"
	@echo "  make install-wllr-skills      - Install Bob skills for wllr to ~/.wllr/skills/"
	@# make install-bob-plugin is intentionally hidden; the zellij plugin is not part of the default workflow.
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
	VARIANTS="normal simple"; [ "$(SPEC)" = "simple" ] && VARIANTS="simple"; \
	for skill in bob-work bob-work-agents bob-work-teams bob-explore bob-explore-teams bob-audit bob-code-review bob-cleanup bob-cleanup-teams bob-design bob-generate-overview bob-generate-feature-page bob-generate-okf bob-stage-prs bob-adversarial-review bob-postmortem bob-premortem bob-challenge-idea bob-operational bob-internal-brainstorming bob-internal-writing-plans bob-internal-go-coding; do \
		if [ -d "skills/$$skill" ]; then \
			for variant in $$VARIANTS; do \
				SRC="skills/$$skill/SKILL.md"; SUFFIX=""; \
				if [ "$$variant" = "simple" ]; then SRC="skills/$$skill/SKILL.simple.md"; SUFFIX="-simple"; fi; \
				[ -f "$$SRC" ] || continue; \
				DEST="$$skill$$SUFFIX"; echo "   Installing $$DEST skill..."; \
				mkdir -p "$$SKILLS_DIR/$$DEST"; \
				sed "s/^name: .*/name: $$DEST/" "$$SRC" > "$$SKILLS_DIR/$$DEST/SKILL.md"; \
			done; \
		else \
			echo "   ⚠️  Skill $$skill not found, skipping..."; \
		fi; \
	done; \
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
	    skills/bob-version/SKILL.md.template > "$$SKILLS_DIR/bob-version/SKILL.md"; \
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

# Companion-reference check. Every "[agent-directory]/<path>" token in an agent
# prompt must name a regular file the installers will actually ship; the gate
# validates the quoted token and its file, not the surrounding Read/command
# syntax (that stays a review responsibility). The agent directory
# itself is the manifest: everything committed under agents/<name>/ ships to all
# three runtimes, except the SKILL*.md prompt variants (installed separately with
# the marker rendered) and top-level hidden entries (rejected here so they
# cannot silently not-ship). Conventions: no spaces in companion filenames (the token grammar
# stops at whitespace), and nothing prunes previously installed files. No
# symlinks anywhere under agents/ — cp -R would ship the link itself, so the
# check rejects them outright. Companions are copied verbatim
# and never rendered, so the literal marker may not appear inside one.
#
# The render below uses sed with a destination byte guard refusing | & and
# backslash, the three sed metacharacter bytes this raw interpolation cannot
# preserve safely, plus a double quote, which sed carries fine but the rendered
# token's own quoting cannot (measured: & duplicates
# the matched text; | breaks the expression after the shell already truncated
# the target at exit 0). If a real machine ever hits the guard on those sed
# bytes, the known drop-in is an awk ENVIRON render: byte-exact, no
# metacharacter semantics, at the cost of appending a newline to a prompt
# lacking a trailing one (all current prompt files end with one). The awk
# route does not help a double-quoted destination — the quote would render
# byte-exactly and still break the token's own quoting.
check-agents:
	@fail=0; \
	for p in agents/*/SKILL*.md; do \
		[ -f "$$p" ] || continue; \
		dir=$${p%/*}; \
		occ=$$(grep -o '\[agent-directory\]' "$$p" | wc -l); \
		ok=$$(grep -o '"\[agent-directory\]/[A-Za-z0-9._/-]\{1,\}"' "$$p" | wc -l); \
		if [ "$$occ" != "$$ok" ]; then \
			echo "❌ $$p: a marker reference is not a quoted \"[agent-directory]/<path>\" token:"; \
			grep -n '\[agent-directory\]' "$$p" | sed 's/^/     /'; \
			fail=1; \
		fi; \
		for t in $$(grep -o '"\[agent-directory\]/[A-Za-z0-9._/-]\{1,\}"' "$$p" | sed 's|^"\[agent-directory\]/||; s|"$$||' | sort -u); do \
			first=$${t%%/*}; \
			case "$$first" in \
				SKILL*.md) echo "❌ $$p: [agent-directory]/$$t names a prompt variant, which never ships as a companion"; fail=1; continue ;; \
				.*) echo "❌ $$p: [agent-directory]/$$t names a hidden entry, which never ships"; fail=1; continue ;; \
			esac; \
			case "/$$t/" in */../*) echo "❌ $$p: [agent-directory]/$$t must stay inside the agent directory"; fail=1; continue ;; esac; \
			[ -f "$$dir/$$t" ] || { echo "❌ $$p: [agent-directory]/$$t is not a shipped regular file ($$dir/$$t)"; fail=1; }; \
		done; \
	done; \
	for e in agents/*/.[!.]* agents/*/..?*; do \
		[ -e "$$e" ] || [ -L "$$e" ] || continue; \
		echo "❌ $$e: hidden entries under an agent directory are never installed; rename or remove"; \
		fail=1; \
	done; \
	links=$$(find agents -type l); \
	if [ -n "$$links" ]; then \
		echo "$$links" | sed 's/^/❌ /; s/$$/: no symlinks anywhere under agents\/ (cp -R would ship the link itself); commit the real file/'; \
		fail=1; \
	fi; \
	bad=$$(find agents -mindepth 2 -type f | grep -v '^agents/[^/]*/SKILL[^/]*\.md$$' | while read -r f; do grep -qF '[agent-directory]' "$$f" && echo "$$f"; done); \
	if [ -n "$$bad" ]; then \
		echo "$$bad" | sed 's/^/❌ /; s/$$/: companions are copied verbatim and never rendered; reference files relative to the script itself/'; \
		fail=1; \
	fi; \
	if [ "$$fail" != 0 ]; then echo "❌ agent companion check failed"; exit 1; fi; \
	echo "✅ agent companion references check out"

# Install specialized subagents
install-agents: check-agents
	@echo "🤖 Installing workflow subagents..."
	@AGENTS_DIR="$$HOME/.claude/agents"; \
	mkdir -p "$$AGENTS_DIR"; \
	AGENT_COUNT=0; \
	if [ -d "agents" ]; then \
		for agent_dir in agents/*; do \
			if [ -d "$$agent_dir" ] && [ -f "$$agent_dir/SKILL.md" ]; then \
				agent=$$(basename "$$agent_dir"); \
				echo "   Installing $$agent agent..."; \
				DEST_DIR="$$AGENTS_DIR/$$agent"; \
				case "$$DEST_DIR" in *'|'*|*'&'*|*'\'*|*'"'*) echo "❌ $$DEST_DIR contains | & \" or a backslash, which cannot be rendered safely into quoted marker references"; exit 1 ;; esac; \
				mkdir -p "$$DEST_DIR" || exit 1; \
				if [ "$(SPEC)" = "simple" ] && [ -f "$$agent_dir/SKILL.simple.md" ]; then \
					SRC="$$agent_dir/SKILL.simple.md"; \
				else \
					SRC="$$agent_dir/SKILL.md"; \
				fi; \
				sed "s|\[agent-directory\]|$$DEST_DIR|g" "$$SRC" > "$$DEST_DIR/SKILL.md" || { echo "❌ render failed for $$SRC"; exit 1; }; \
				for extra in "$$agent_dir"/*; do \
					case "$${extra##*/}" in SKILL*.md) continue ;; esac; \
					cp -R "$$extra" "$$DEST_DIR/" || { echo "❌ copy failed for $$extra"; exit 1; }; \
				done; \
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
	@echo "🔌 Installing Claude plugins..."; \
	if ! command -v claude >/dev/null 2>&1; then \
		echo "⚠️  Claude CLI not found — skipping Claude plugins"; \
	else \
		echo "   Adding grafana/ai-kit to marketplace..."; \
		claude plugin marketplace add grafana/ai-kit; \
		echo "   Installing grafana-engineering@grafana-ai-kit..."; \
		claude plugin install grafana-engineering@grafana-ai-kit; \
		echo "✅ Claude plugins installed"; \
	fi

# Install everything (skills, agents, LSP, personality) - PRIMARY COMMAND
# Usage: make install [PERSONALITY=pirate|cartoon_pirate]
install: install-skills install-agents install-lsp install-plugins allow
	@if command -v codex >/dev/null 2>&1; then \
		echo ""; \
		$(MAKE) install-codex-skills SPEC=$(SPEC) CODEX_HOME="$(CODEX_HOME)"; \
	else \
		echo "⏭️  Codex CLI not installed — skipping Codex skills"; \
	fi
	@if command -v wllr >/dev/null 2>&1; then \
		echo ""; \
		$(MAKE) install-wllr-skills SPEC=$(SPEC) WLLR_HOME="$(WLLR_HOME)"; \
	else \
		echo "⏭️  wllr CLI not installed — skipping wllr skills"; \
	fi
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
ci: check-agents
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

# Install bob-agents pi extension and skills into the project-local .pi/ directory.
# Extensions are already in .pi/extensions/ (auto-discovered by pi when run from this repo).
# Skills are copied to .pi/skills/ so pi loads them as /bob:* commands.
# Usage: make install-pi [SPEC=simple]
install-pi: check-agents
	@echo "🐦 Installing Bob pi components under .pi/..."
	@echo ""
	@echo "📦 LSP Support"
	@if command -v pi >/dev/null 2>&1; then \
		echo "   Installing https://github.com/apmantza/pi-lens..."; \
		pi install https://github.com/apmantza/pi-lens; \
		echo "   ✓ LSP support installed"; \
	else \
		echo "   ⚠️  pi not found in PATH — skipping LSP install"; \
		echo "   Install pi first, then run: pi install https://github.com/apmantza/pi-lens"; \
	fi
	@echo ""
	@echo "🔌 Extension"
	@echo "   ℹ️  bob-agents retired — pi-subagents handles agent spawning natively"
	@mkdir -p "$$HOME/.pi/agent/extensions"
	@echo "📡 OTel Extension"
	@cp extensions/otel.ts "$$HOME/.pi/agent/extensions/otel.ts"; \
	echo "   ✓ Copied extensions/otel.ts → $$HOME/.pi/agent/extensions/otel.ts"
	@echo "🔧 Bash Compact Extension"
	@cp extensions/quiet-thoughts.ts "$$HOME/.pi/agent/extensions/quiet-thoughts.ts"; \
	echo "   ✓ Copied extensions/quiet-thoughts.ts → $$HOME/.pi/agent/extensions/quiet-thoughts.ts"
	@cp extensions/tps.ts "$$HOME/.pi/agent/extensions/tps.ts"; \
	echo "   ✓ Copied extensions/tps.ts → $$HOME/.pi/agent/extensions/tps.ts"
	@echo ""
	@echo "📚 Skills"
	@PI_TRANSFORM='s/subagent_type:/agent:/g; s/run_in_background: true/background: true/g; s/taskId:/id:/g; s/status: "completed"/status: "done"/g'; \
	SKILLS_DIR="$$HOME/.pi/agent/skills"; \
	mkdir -p "$$SKILLS_DIR"; \
	VARIANTS="normal simple"; [ "$(SPEC)" = "simple" ] && VARIANTS="simple"; \
	for skill in bob-work bob-work-agents bob-work-teams bob-explore bob-explore-teams bob-audit bob-code-review bob-cleanup bob-cleanup-teams bob-design bob-generate-overview bob-generate-feature-page bob-generate-okf bob-stage-prs bob-adversarial-review bob-postmortem bob-premortem bob-challenge-idea bob-operational bob-internal-brainstorming bob-internal-writing-plans bob-internal-go-coding; do \
		if [ -d "skills/$$skill" ]; then \
			for variant in $$VARIANTS; do \
				SRC="skills/$$skill/SKILL.pi.md"; SUFFIX=""; \
				if [ "$$variant" = "simple" ]; then SRC="skills/$$skill/SKILL.simple.md"; SUFFIX="-simple"; fi; \
				if [ "$$variant" = "normal" ] && [ ! -f "$$SRC" ]; then SRC="skills/$$skill/SKILL.md"; fi; [ -f "$$SRC" ] || continue; \
				RAW=$$(grep -m1 '^name:' "$$SRC" | sed 's/^name: *//'); DEST=$$(echo "$$RAW" | tr ':' '-'); \
				[ -z "$$DEST" ] && DEST="$$skill"; DEST="$$DEST$$SUFFIX"; echo "   Installing $$DEST..."; \
				mkdir -p "$$SKILLS_DIR/$$DEST"; \
				sed "s/^name: .*/name: $$DEST/; $$PI_TRANSFORM" "$$SRC" | bash scripts/sanitize-native-team-skill.sh > "$$SKILLS_DIR/$$DEST/SKILL.md"; \
			done; \
		else \
			echo "   ⚠️  Skill $$skill not found, skipping..."; \
		fi; \
	done; \
	GIT_HASH=$$(git rev-parse HEAD); \
	GIT_DATE=$$(git log -1 --format=%cd --date=format:'%Y-%m-%d %H:%M:%S'); \
	GIT_BRANCH=$$(git rev-parse --abbrev-ref HEAD); \
	GIT_REMOTE=$$(git config --get remote.origin.url || echo "local"); \
	INSTALL_DATE=$$(date '+%Y-%m-%d %H:%M:%S'); \
	BOB_REPO_PATH=$$(pwd); \
	SKILL_COUNT=$$(find skills -name "SKILL.md" -o -name "SKILL.md.template" | wc -l); \
	AGENT_COUNT=$$(find agents -name "SKILL.md" 2>/dev/null | wc -l || echo "0"); \
	mkdir -p "$$SKILLS_DIR/bob-version"; \
	sed -e "s|{{GIT_HASH}}|$$GIT_HASH|g" \
	    -e "s|{{GIT_DATE}}|$$GIT_DATE|g" \
	    -e "s|{{GIT_BRANCH}}|$$GIT_BRANCH|g" \
	    -e "s|{{GIT_REMOTE}}|$$GIT_REMOTE|g" \
	    -e "s|{{INSTALL_DATE}}|$$INSTALL_DATE|g" \
	    -e "s|{{BOB_REPO_PATH}}|$$BOB_REPO_PATH|g" \
	    -e "s|{{SKILL_COUNT}}|$$SKILL_COUNT|g" \
	    -e "s|{{AGENT_COUNT}}|$$AGENT_COUNT|g" \
	    skills/bob-version/SKILL.md.template | sed 's/^name: .*/name: bob-version/' | bash scripts/sanitize-native-team-skill.sh > "$$SKILLS_DIR/bob-version/SKILL.md"; \
	if command -v codex >/dev/null 2>&1; then \
		mkdir -p "$$SKILLS_DIR/talk-to-codex"; \
		sed "$$PI_TRANSFORM" "skills/talk-to-codex/SKILL.md" | bash scripts/sanitize-native-team-skill.sh > "$$SKILLS_DIR/talk-to-codex/SKILL.md"; \
	fi; \
	echo "✅ Skills installed to $$SKILLS_DIR"
	@echo ""
	@echo "🤖 Agents"
	@PI_TRANSFORM='s/subagent_type:/agent:/g; s/run_in_background: true/background: true/g; s/taskId:/id:/g; s/status: "completed"/status: "done"/g'; \
	AGENTS_DIR="$$HOME/.pi/agent/agents"; \
	mkdir -p "$$AGENTS_DIR"; \
	AGENT_COUNT=0; \
	for agent_dir in agents/*; do \
		[ -d "$$agent_dir" ] || continue; \
		agent=$$(basename "$$agent_dir"); \
		if [ -f "$$agent_dir/SKILL.pi.md" ]; then \
			SRC="$$agent_dir/SKILL.pi.md"; \
			NEED_TRANSFORM=0; \
		elif [ "$(SPEC)" = "simple" ] && [ -f "$$agent_dir/SKILL.simple.md" ]; then \
			SRC="$$agent_dir/SKILL.simple.md"; \
			NEED_TRANSFORM=1; \
		elif [ -f "$$agent_dir/SKILL.md" ]; then \
			SRC="$$agent_dir/SKILL.md"; \
			NEED_TRANSFORM=1; \
		else \
			continue; \
		fi; \
		echo "   Installing $$agent..."; \
		DEST_DIR="$$AGENTS_DIR/$$agent"; \
		case "$$DEST_DIR" in *'|'*|*'&'*|*'\'*|*'"'*) echo "❌ $$DEST_DIR contains | & \" or a backslash, which cannot be rendered safely into quoted marker references"; exit 1 ;; esac; \
		mkdir -p "$$DEST_DIR" || exit 1; \
		if [ "$$NEED_TRANSFORM" = "1" ]; then \
			sed "$$PI_TRANSFORM; s|\[agent-directory\]|$$DEST_DIR|g" "$$SRC" > "$$DEST_DIR/SKILL.md" || { echo "❌ render failed for $$SRC"; exit 1; }; \
		else \
			sed "s|\[agent-directory\]|$$DEST_DIR|g" "$$SRC" > "$$DEST_DIR/SKILL.md" || { echo "❌ render failed for $$SRC"; exit 1; }; \
		fi; \
		for extra in "$$agent_dir"/*; do \
			case "$${extra##*/}" in SKILL*.md) continue ;; esac; \
			cp -R "$$extra" "$$DEST_DIR/" || { echo "❌ copy failed for $$extra"; exit 1; }; \
		done; \
		AGENT_COUNT=$$((AGENT_COUNT + 1)); \
	done; \
	echo "✅ $$AGENT_COUNT agents installed to $$AGENTS_DIR"
	@echo ""
	@echo "✅ Pi installation complete!"
	@echo ""
	@echo "Installed to ~/.pi/agent/:"
	@echo "  ✓ Extension → bob-agents retired (pi-subagents used instead)"
	@echo "  ✓ OTel      → ~/.pi/agent/extensions/otel.ts"
	@echo "  ✓ Quiet     → ~/.pi/agent/extensions/quiet-thoughts.ts"
	@echo "  ✓ TPS       → ~/.pi/agent/extensions/tps.ts"
	@echo "  ✓ Skills    → ~/.pi/agent/skills/"
	@echo "  ✓ Agents    → ~/.pi/agent/agents/"
	@echo "  ✓ LSP       → https://github.com/apmantza/pi-lens"
	@echo ""
	@echo "Available skill commands in pi:"
	@find "$$HOME/.pi/agent/skills" -name "SKILL.md" -exec grep -m1 '^name:' {} \; 2>/dev/null \
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

# Install only Bob skills for pi. pi has native team support, so generated
# skills omit Claude Code's experimental team environment-variable guidance.
# Usage: make install-pi-skills [SPEC=simple]
install-pi-skills:
	@echo "📚 Installing Bob skills for pi..."
	@PI_TRANSFORM='s/subagent_type:/agent:/g; s/run_in_background: true/background: true/g; s/taskId:/id:/g; s/status: "completed"/status: "done"/g'; \
	SKILLS_DIR="$$HOME/.pi/agent/skills"; \
	mkdir -p "$$SKILLS_DIR"; \
	VARIANTS="normal simple"; [ "$(SPEC)" = "simple" ] && VARIANTS="simple"; \
	for skill in bob-work bob-work-agents bob-work-teams bob-explore bob-explore-teams bob-audit bob-code-review bob-cleanup bob-cleanup-teams bob-design bob-generate-overview bob-generate-feature-page bob-generate-okf bob-stage-prs bob-adversarial-review bob-postmortem bob-premortem bob-challenge-idea bob-operational bob-internal-brainstorming bob-internal-writing-plans bob-internal-go-coding; do \
		if [ -d "skills/$$skill" ]; then \
			for variant in $$VARIANTS; do \
				SRC="skills/$$skill/SKILL.pi.md"; SUFFIX=""; \
				if [ "$$variant" = "simple" ]; then SRC="skills/$$skill/SKILL.simple.md"; SUFFIX="-simple"; fi; \
				if [ "$$variant" = "normal" ] && [ ! -f "$$SRC" ]; then SRC="skills/$$skill/SKILL.md"; fi; [ -f "$$SRC" ] || continue; \
				RAW=$$(grep -m1 '^name:' "$$SRC" | sed 's/^name: *//'); DEST=$$(echo "$$RAW" | tr ':' '-'); \
				[ -z "$$DEST" ] && DEST="$$skill"; DEST="$$DEST$$SUFFIX"; echo "   Installing $$DEST..."; \
				mkdir -p "$$SKILLS_DIR/$$DEST"; \
				sed "s/^name: .*/name: $$DEST/; $$PI_TRANSFORM" "$$SRC" | bash scripts/sanitize-native-team-skill.sh > "$$SKILLS_DIR/$$DEST/SKILL.md"; \
			done; \
		else \
			echo "   ⚠️  Skill $$skill not found, skipping..."; \
		fi; \
	done; \
	GIT_HASH=$$(git rev-parse HEAD); \
	GIT_DATE=$$(git log -1 --format=%cd --date=format:'%Y-%m-%d %H:%M:%S'); \
	GIT_BRANCH=$$(git rev-parse --abbrev-ref HEAD); \
	GIT_REMOTE=$$(git config --get remote.origin.url || echo "local"); \
	INSTALL_DATE=$$(date '+%Y-%m-%d %H:%M:%S'); \
	BOB_REPO_PATH=$$(pwd); \
	SKILL_COUNT=$$(find skills -name "SKILL.md" -o -name "SKILL.md.template" | wc -l); \
	AGENT_COUNT=$$(find agents -name "SKILL.md" 2>/dev/null | wc -l || echo "0"); \
	mkdir -p "$$SKILLS_DIR/bob-version"; \
	sed -e "s|{{GIT_HASH}}|$$GIT_HASH|g" \
	    -e "s|{{GIT_DATE}}|$$GIT_DATE|g" \
	    -e "s|{{GIT_BRANCH}}|$$GIT_BRANCH|g" \
	    -e "s|{{GIT_REMOTE}}|$$GIT_REMOTE|g" \
	    -e "s|{{INSTALL_DATE}}|$$INSTALL_DATE|g" \
	    -e "s|{{BOB_REPO_PATH}}|$$BOB_REPO_PATH|g" \
	    -e "s|{{SKILL_COUNT}}|$$SKILL_COUNT|g" \
	    -e "s|{{AGENT_COUNT}}|$$AGENT_COUNT|g" \
	    skills/bob-version/SKILL.md.template | sed 's/^name: .*/name: bob-version/' | bash scripts/sanitize-native-team-skill.sh > "$$SKILLS_DIR/bob-version/SKILL.md"; \
	if command -v codex >/dev/null 2>&1; then \
		mkdir -p "$$SKILLS_DIR/talk-to-codex"; \
		sed "$$PI_TRANSFORM" "skills/talk-to-codex/SKILL.md" | bash scripts/sanitize-native-team-skill.sh > "$$SKILLS_DIR/talk-to-codex/SKILL.md"; \
	fi; \
	echo "✅ Bob pi skills installed to $$SKILLS_DIR"; \
	echo ""; \
	echo "Available skill commands in pi:"; \
	find "$$SKILLS_DIR" -name "SKILL.md" -exec grep -m1 '^name:' {} \; 2>/dev/null \
		| sed 's/name: */  \//' | sort

# Install Bob workflow skills for Codex.
# Codex discovers skills under CODEX_HOME/skills/<name>/SKILL.md.
# Usage: make install-codex-skills [SPEC=simple] [CODEX_HOME=/path/to/.codex]
install-codex-skills:
	@echo "📚 Installing Bob skills for Codex..."
	@SKILLS_DIR="$(CODEX_HOME)/skills"; \
	mkdir -p "$$SKILLS_DIR"; \
	VARIANTS="normal simple"; [ "$(SPEC)" = "simple" ] && VARIANTS="simple"; \
	for skill in bob-work bob-work-agents bob-work-teams bob-explore bob-explore-teams bob-audit bob-code-review bob-cleanup bob-cleanup-teams bob-design bob-generate-overview bob-generate-feature-page bob-generate-okf bob-stage-prs bob-adversarial-review bob-postmortem bob-premortem bob-challenge-idea bob-operational bob-internal-brainstorming bob-internal-writing-plans bob-internal-go-coding; do \
		if [ -d "skills/$$skill" ]; then \
			for variant in $$VARIANTS; do \
				SRC="skills/$$skill/SKILL.codex.md"; SUFFIX=""; \
				if [ "$$variant" = "simple" ]; then SRC="skills/$$skill/SKILL.simple.md"; SUFFIX="-simple"; fi; \
				if [ "$$variant" = "normal" ] && [ ! -f "$$SRC" ]; then SRC="skills/$$skill/SKILL.md"; fi; [ -f "$$SRC" ] || continue; \
				RAW=$$(grep -m1 '^name:' "$$SRC" | sed 's/^name: *//; s/^"//; s/"$$//'); DEST=$$(echo "$$RAW" | tr ':' '-'); \
				[ -z "$$DEST" ] && DEST="$$skill"; DEST="$$DEST$$SUFFIX"; echo "   Installing $$DEST..."; mkdir -p "$$SKILLS_DIR/$$DEST"; \
				sed "s/^name: .*/name: $$DEST/" "$$SRC" | bash scripts/sanitize-native-team-skill.sh > "$$SKILLS_DIR/$$DEST/SKILL.md"; \
			done; \
		fi; \
	done; \
	if [ -f "skills/bob-version/SKILL.md.template" ]; then \
		mkdir -p "$$SKILLS_DIR/bob-version"; \
		sed -e "s|{{GIT_HASH}}|$$(git rev-parse HEAD)|g" -e "s|{{GIT_DATE}}|$$(git log -1 --format=%cd --date=format:'%Y-%m-%d %H:%M:%S')|g" -e "s|{{GIT_BRANCH}}|$$(git rev-parse --abbrev-ref HEAD)|g" -e "s|{{GIT_REMOTE}}|$$(git config --get remote.origin.url || echo local)|g" -e "s|{{INSTALL_DATE}}|$$(date '+%Y-%m-%d %H:%M:%S')|g" -e "s|{{BOB_REPO_PATH}}|$$(pwd)|g" skills/bob-version/SKILL.md.template | sed 's/^name: .*/name: bob-version/' | bash scripts/sanitize-native-team-skill.sh > "$$SKILLS_DIR/bob-version/SKILL.md"; \
	fi; \
	echo "✅ Bob Codex skills installed to $$SKILLS_DIR"

# Install Bob skills and agent prompts for wllr.
# wllr's built-in skills extension scans ~/.wllr/skills/<name>/SKILL.md.
# Usage: make install-wllr-skills [SPEC=simple] [WLLR_HOME=/path/to/.wllr]
install-wllr install-wllr-skills: check-agents
	@echo "📚 Installing Bob skills for wllr..."
	@SKILLS_DIR="$(WLLR_HOME)/skills"; \
	mkdir -p "$$SKILLS_DIR"; \
	VARIANTS="normal simple"; [ "$(SPEC)" = "simple" ] && VARIANTS="simple"; \
	for skill in bob-work bob-work-agents bob-work-teams bob-explore bob-explore-teams bob-audit bob-code-review bob-cleanup bob-cleanup-teams bob-design bob-generate-overview bob-generate-feature-page bob-generate-okf bob-stage-prs bob-adversarial-review bob-postmortem bob-premortem bob-challenge-idea bob-operational bob-internal-brainstorming bob-internal-writing-plans bob-internal-go-coding; do \
		if [ -d "skills/$$skill" ]; then \
			for variant in $$VARIANTS; do \
				SRC="skills/$$skill/SKILL.md"; SUFFIX=""; \
				if [ "$$variant" = "simple" ]; then SRC="skills/$$skill/SKILL.simple.md"; SUFFIX="-simple"; fi; \
				[ -f "$$SRC" ] || continue; \
				RAW=$$(grep -m1 '^name:' "$$SRC" | sed 's/^name: *//; s/^"//; s/"$$//'); CMD="$$RAW"; \
				case "$$CMD" in bob-internal-*) CMD="bob:internal:$${CMD#bob-internal-}" ;; bob-*) CMD="bob:$${CMD#bob-}" ;; esac; \
				CMD="$$CMD$$SUFFIX"; DEST=$$(echo "$$CMD" | tr ':' '-'); [ -z "$$DEST" ] && DEST="$$skill$$SUFFIX"; \
				echo "   Installing $$CMD skill..."; mkdir -p "$$SKILLS_DIR/$$DEST"; \
				sed "s/^name: .*/name: $$CMD/" "$$SRC" | bash scripts/sanitize-native-team-skill.sh > "$$SKILLS_DIR/$$DEST/SKILL.md"; \
			done; \
		else \
			echo "   ⚠️  Skill $$skill not found, skipping..."; \
		fi; \
	done; \
	echo "   Generating bob:version skill..."; \
	GIT_HASH=$$(git rev-parse HEAD); \
	GIT_DATE=$$(git log -1 --format=%cd --date=format:'%Y-%m-%d %H:%M:%S'); \
	GIT_BRANCH=$$(git rev-parse --abbrev-ref HEAD); \
	GIT_REMOTE=$$(git config --get remote.origin.url || echo "local"); \
	INSTALL_DATE=$$(date '+%Y-%m-%d %H:%M:%S'); \
	BOB_REPO_PATH=$$(pwd); \
	SKILL_COUNT=$$(find skills -name "SKILL.md" -o -name "SKILL.md.template" | wc -l); \
	AGENT_COUNT=$$(find agents -name "SKILL.md" 2>/dev/null | wc -l || echo "0"); \
	HOOKS_STATUS="**Hooks:** not managed by wllr install"; \
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
	    skills/bob-version/SKILL.md.template | bash scripts/sanitize-native-team-skill.sh > "$$SKILLS_DIR/bob-version/SKILL.md"; \
	if command -v codex >/dev/null 2>&1; then \
		echo "   Installing talk-to-codex skill (codex CLI detected)..."; \
		mkdir -p "$$SKILLS_DIR/talk-to-codex"; \
		bash scripts/sanitize-native-team-skill.sh < "skills/talk-to-codex/SKILL.md" > "$$SKILLS_DIR/talk-to-codex/SKILL.md"; \
	else \
		echo "   ⏭️  Skipping talk-to-codex (codex CLI not installed)"; \
	fi; \
	echo "   Installing agent prompt skills for wllr subagents..."; \
	AGENT_COUNT=0; \
	for agent_dir in agents/*; do \
		[ -d "$$agent_dir" ] || continue; \
		agent=$$(basename "$$agent_dir"); \
		if [ "$(SPEC)" = "simple" ] && [ -f "$$agent_dir/SKILL.simple.md" ]; then \
			SRC="$$agent_dir/SKILL.simple.md"; \
		elif [ -f "$$agent_dir/SKILL.md" ]; then \
			SRC="$$agent_dir/SKILL.md"; \
		else \
			continue; \
		fi; \
		echo "   Installing $$agent agent prompt..."; \
		DEST_DIR="$$SKILLS_DIR/$$agent"; \
		case "$$DEST_DIR" in *'|'*|*'&'*|*'\'*|*'"'*) echo "❌ $$DEST_DIR contains | & \" or a backslash, which cannot be rendered safely into quoted marker references"; exit 1 ;; esac; \
		mkdir -p "$$DEST_DIR" || exit 1; \
		sed "s|\[agent-directory\]|$$DEST_DIR|g" "$$SRC" > "$$DEST_DIR/SKILL.md" || { echo "❌ render failed for $$SRC"; exit 1; }; \
		for extra in "$$agent_dir"/*; do \
			case "$${extra##*/}" in SKILL*.md) continue ;; esac; \
			cp -R "$$extra" "$$DEST_DIR/" || { echo "❌ copy failed for $$extra"; exit 1; }; \
		done; \
		AGENT_COUNT=$$((AGENT_COUNT + 1)); \
	done; \
	echo "✅ Bob wllr skills installed to $$SKILLS_DIR"; \
	echo ""; \
	echo "Available wllr skill commands:"; \
	find "$$SKILLS_DIR" -name "SKILL.md" -exec grep -l '^user-invocable: true' {} \; 2>/dev/null \
		| while read -r file; do grep -m1 '^name:' "$$file"; done \
		| sed 's/name: */  \//' | sort; \
	echo ""; \
	echo "🔄 Restart wllr or start a new session to load these skills"

clean:
	@echo "🧹 Cleaning temporary files..."
	@find . -name "*.tmp" -delete 2>/dev/null || true
	@find . -name ".DS_Store" -delete 2>/dev/null || true
	@echo "✅ Clean complete"
