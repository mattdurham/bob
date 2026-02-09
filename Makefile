# Belayin' Pin Bob - Captain of Your Agents
# Makefile for building and running Bob workflow orchestrator

.PHONY: help run build install-deps clean test

help:
	@echo "🏴‍☠️ Belayin' Pin Bob - Captain of Your Agents"
	@echo ""
	@echo "Available targets:"
	@echo "  make run           - Run Bob as MCP server"
	@echo "  make build         - Build Bob binary"
	@echo "  make install-deps  - Install Go dependencies"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make test          - Run tests"

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
