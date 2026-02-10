.PHONY: build build-mcp build-mcp-dev build-all install install-mcp install-all dev-install dev-install-mcp uninstall uninstall-mcp uninstall-all test test-unit test-integration test-compat test-all test-cover test-race test-verbose clean fmt lint vet check completions release snapshot install-goreleaser help

# Build variables
BINARY_NAME=brain
MCP_BINARY_NAME=brain-mcp
MCP_DEV_BINARY_NAME=brain-mcp-dev
MCP_DIR=mcp
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
BUILD_DIR=.
INSTALL_DIR?=/usr/local/bin
LIB_DIR?=/usr/local/lib/brain

# Build flags (matching goreleaser)
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build the MCP server
build-mcp:
	@echo "Building $(MCP_BINARY_NAME)..."
	cd $(MCP_DIR) && go build $(LDFLAGS) -o ../$(BUILD_DIR)/$(MCP_BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(MCP_BINARY_NAME)"

# Build the MCP server (dev variant — uses ~/brains-dev/ isolated workspace)
build-mcp-dev:
	@echo "Building $(MCP_DEV_BINARY_NAME)..."
	cd $(MCP_DIR) && go build $(LDFLAGS) -o ../$(BUILD_DIR)/$(MCP_DEV_BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(MCP_DEV_BINARY_NAME)"

# Build both CLI and MCP server
build-all: build build-mcp

# Install the binary
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	mkdir -p $(INSTALL_DIR)
	mkdir -p $(LIB_DIR)
	cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	cp lib/brain-prompt.sh $(LIB_DIR)/
	chmod 755 $(INSTALL_DIR)/$(BINARY_NAME)
	chmod 644 $(LIB_DIR)/brain-prompt.sh
	@echo "Installation complete"
	@echo "Add to your shell config: source $(LIB_DIR)/brain-prompt.sh"

# Install the MCP server
install-mcp: build-mcp
	@echo "Installing $(MCP_BINARY_NAME) to $(INSTALL_DIR)..."
	mkdir -p $(INSTALL_DIR)
	cp $(BUILD_DIR)/$(MCP_BINARY_NAME) $(INSTALL_DIR)/
	chmod 755 $(INSTALL_DIR)/$(MCP_BINARY_NAME)
	@echo "Installation complete: $(INSTALL_DIR)/$(MCP_BINARY_NAME)"
	@echo ""
	@echo "Add to Claude Desktop config:"
	@echo "  macOS: ~/Library/Application Support/Claude/claude_desktop_config.json"
	@echo "  Linux: ~/.config/Claude/claude_desktop_config.json"
	@echo ""
	@echo "Config format:"
	@echo '  {'
	@echo '    "mcpServers": {'
	@echo '      "local-brain": {'
	@echo '        "command": "$(INSTALL_DIR)/$(MCP_BINARY_NAME)"'
	@echo '      }'
	@echo '    }'
	@echo '  }'

# Install both CLI and MCP server
install-all: install install-mcp

# Uninstall the binary
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	rm -rf $(LIB_DIR)
	@echo "Uninstall complete"

# Uninstall the MCP server
uninstall-mcp:
	@echo "Uninstalling $(MCP_BINARY_NAME)..."
	rm -f $(INSTALL_DIR)/$(MCP_BINARY_NAME)
	@echo "Uninstall complete"

# Uninstall both CLI and MCP server
uninstall-all: uninstall uninstall-mcp

# Run unit tests (fast, no integration)
test-unit:
	@echo "Running unit tests..."
	go test -v -short ./pkg/...

# Run integration tests (slower, test multiple components together)
test-integration:
	@echo "Running integration tests..."
	go test -v -run Integration ./...

# Run compatibility tests (verify bash compatibility)
test-compat:
	@echo "Running compatibility tests..."
	go test -v ./test/compat/...

# Run all tests
test-all:
	@echo "Running all tests..."
	go test -v ./...

# Default test target (unit tests only for speed)
test: test-unit

# Generate coverage report
test-cover:
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./pkg/...
	go tool cover -html=coverage.out -o coverage.html
	@echo ""
	@echo "Coverage summary:"
	@go tool cover -func=coverage.out | grep total
	@echo ""
	@echo "Detailed report: coverage.html"

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	go test -race ./pkg/...

# Run tests with verbose output and coverage
test-verbose:
	@echo "Running tests with verbose output..."
	go test -v -cover ./pkg/...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)-dev
	rm -f $(BUILD_DIR)/$(MCP_BINARY_NAME)
	rm -f $(BUILD_DIR)/$(MCP_DEV_BINARY_NAME)
	rm -f coverage.out coverage.html
	go clean
	cd $(MCP_DIR) && go clean
	@echo "Clean complete"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install from https://golangci-lint.run/" && exit 1)
	golangci-lint run

# Vet code
vet:
	@echo "Vetting code..."
	go vet ./...

# Run all checks (fmt, vet, test)
check: fmt vet test-all
	@echo "All checks passed!"

# Generate shell completions
completions:
	@echo "Generating shell completions..."
	@mkdir -p completions
	@$(BUILD_DIR)/$(BINARY_NAME) completion bash > completions/brain.bash
	@$(BUILD_DIR)/$(BINARY_NAME) completion zsh > completions/_brain
	@$(BUILD_DIR)/$(BINARY_NAME) completion fish > completions/brain.fish
	@echo "Completions generated in completions/"

# Install goreleaser (if not installed)
install-goreleaser:
	@echo "Checking for goreleaser..."
	@which goreleaser > /dev/null || (echo "Installing goreleaser..." && go install github.com/goreleaser/goreleaser@latest)

# Create a release (requires goreleaser and git tag)
release: install-goreleaser
	@echo "Creating release with goreleaser..."
	@echo "Current version: $(VERSION)"
	@goreleaser release --clean

# Create a snapshot release (no git tag required, for testing)
snapshot: install-goreleaser build
	@echo "Creating snapshot release..."
	@goreleaser release --snapshot --clean --skip=publish
	@echo ""
	@echo "Snapshot artifacts created in dist/"

# Quick local install (for development) — installs both CLI and MCP dev binaries
dev-install: build-mcp-dev
	@echo "Building $(BINARY_NAME)-dev..."
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-dev .
	@echo "Installing for development..."
	mkdir -p $(HOME)/.local/bin
	cp $(BUILD_DIR)/$(BINARY_NAME)-dev $(HOME)/.local/bin/
	chmod +x $(HOME)/.local/bin/$(BINARY_NAME)-dev
	cp $(BUILD_DIR)/$(MCP_DEV_BINARY_NAME) $(HOME)/.local/bin/
	chmod +x $(HOME)/.local/bin/$(MCP_DEV_BINARY_NAME)
	@echo "Installed to $(HOME)/.local/bin/$(BINARY_NAME)-dev"
	@echo "Installed to $(HOME)/.local/bin/$(MCP_DEV_BINARY_NAME)"
	@echo "Make sure $(HOME)/.local/bin is in your PATH"
	@echo ""
	@echo "Dev mode uses separate isolated paths:"
	@echo "  Config: ~/.config/brain-dev/"
	@echo "  Brains: ~/brains-dev/"
	@echo "  Symlink: ~/brain-dev"
	@echo ""
	@echo "To register the dev MCP server with Claude Code:"
	@echo "  claude mcp add --transport stdio local-brain-dev $(HOME)/.local/bin/$(MCP_DEV_BINARY_NAME)"

# Quick local install for MCP server (dev variant — uses ~/brains-dev/ isolated workspace)
dev-install-mcp: build-mcp-dev
	@echo "Installing $(MCP_DEV_BINARY_NAME) for development..."
	mkdir -p $(HOME)/.local/bin
	cp $(BUILD_DIR)/$(MCP_DEV_BINARY_NAME) $(HOME)/.local/bin/
	chmod +x $(HOME)/.local/bin/$(MCP_DEV_BINARY_NAME)
	@echo "Installed to $(HOME)/.local/bin/$(MCP_DEV_BINARY_NAME)"
	@echo "Make sure $(HOME)/.local/bin is in your PATH"
	@echo ""
	@echo "To register with Claude Code:"
	@echo "  claude mcp add --transport stdio local-brain-dev $(HOME)/.local/bin/$(MCP_DEV_BINARY_NAME)"
	@echo ""
	@echo "Uses isolated dev paths:"
	@echo "  Config: ~/.config/brain-dev/"
	@echo "  Brains: ~/brains-dev/"

# Show help
help:
	@echo "Local Brain - Makefile targets:"
	@echo ""
	@echo "Building:"
	@echo "  make build          - Build the CLI binary"
	@echo "  make build-mcp      - Build the MCP server binary"
	@echo "  make build-mcp-dev  - Build the MCP server dev binary (uses ~/brains-dev/)"
	@echo "  make build-all      - Build both CLI and MCP server"
	@echo "  make install        - Install CLI to $(INSTALL_DIR)"
	@echo "  make install-mcp    - Install MCP server to $(INSTALL_DIR)"
	@echo "  make install-all    - Install both CLI and MCP server"
	@echo "  make dev-install    - Install CLI + MCP dev binaries to ~/.local/bin"
	@echo "  make dev-install-mcp - Install MCP dev binary to ~/.local/bin"
	@echo "  make uninstall      - Remove CLI from $(INSTALL_DIR)"
	@echo "  make uninstall-mcp  - Remove MCP server from $(INSTALL_DIR)"
	@echo "  make uninstall-all  - Remove both CLI and MCP server"
	@echo "  make completions    - Generate shell completions"
	@echo ""
	@echo "Distribution:"
	@echo "  make release        - Create a release with goreleaser (requires git tag)"
	@echo "  make snapshot       - Create snapshot release for testing"
	@echo "  make install-goreleaser - Install goreleaser if not present"
	@echo ""
	@echo "Testing:"
	@echo "  make test           - Run unit tests (fast)"
	@echo "  make test-unit      - Run unit tests"
	@echo "  make test-integration - Run integration tests"
	@echo "  make test-compat    - Run compatibility tests"
	@echo "  make test-all       - Run all tests"
	@echo "  make test-cover     - Generate coverage report"
	@echo "  make test-race      - Run tests with race detection"
	@echo "  make test-verbose   - Run tests with verbose output"
	@echo ""
	@echo "Code quality:"
	@echo "  make fmt            - Format code"
	@echo "  make vet            - Vet code"
	@echo "  make lint           - Run linter (requires golangci-lint)"
	@echo "  make check          - Run all checks (fmt, vet, test)"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean          - Remove build artifacts"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION=$(VERSION)"
	@echo "  COMMIT=$(COMMIT)"
	@echo "  DATE=$(DATE)"
	@echo "  INSTALL_DIR=$(INSTALL_DIR)"
	@echo "  LIB_DIR=$(LIB_DIR)"
