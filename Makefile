# Belayin' Pin Bob - Captain of Your Agents
# Makefile for installing Bob workflow skills and subagents

.PHONY: help install install-skills install-agents install-lsp install-guidance clean

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
	@echo "  make install-guidance PATH=/path - Copy AGENTS.md & CLAUDE.md to repo"
	@echo "  make clean                    - Clean temporary files"
	@echo ""
	@echo "Quick start:"
	@echo "  make install                  - Install everything"
	@echo "  /work \"feature description\" - Start a workflow"

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
	@echo "✅ Skills installed to ~/.claude/skills/"
	@echo ""
	@echo "Available workflow commands:"
	@echo "  /work            - Full development workflow"
	@echo "  /code-review     - Code review workflow"
	@echo "  /performance     - Performance optimization"
	@echo "  /explore         - Codebase exploration"
	@echo "  /brainstorming   - Creative ideation"
	@echo "  /writing-plans   - Implementation planning"

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

# Clean temporary files
clean:
	@echo "🧹 Cleaning temporary files..."
	@find . -name "*.tmp" -delete 2>/dev/null || true
	@find . -name ".DS_Store" -delete 2>/dev/null || true
	@echo "✅ Clean complete"
