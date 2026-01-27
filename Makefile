.PHONY: build install test test-unit test-integration test-compat test-all test-cover test-race clean release snapshot install-goreleaser completions

# Build variables
BINARY_NAME=brain
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

# Uninstall the binary
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	rm -rf $(LIB_DIR)
	@echo "Uninstall complete"

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
	rm -f coverage.out coverage.html
	go clean
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

# Quick local install (for development)
dev-install: build
	@echo "Installing for development..."
	mkdir -p $(HOME)/.local/bin
	cp $(BUILD_DIR)/$(BINARY_NAME) $(HOME)/.local/bin/
	chmod +x $(HOME)/.local/bin/$(BINARY_NAME)
	@echo "Installed to $(HOME)/.local/bin/$(BINARY_NAME)"
	@echo "Make sure $(HOME)/.local/bin is in your PATH"

# Show help
help:
	@echo "Local Brain - Makefile targets:"
	@echo ""
	@echo "Building:"
	@echo "  make build          - Build the binary"
	@echo "  make install        - Install to $(INSTALL_DIR)"
	@echo "  make dev-install    - Install to ~/.local/bin (development)"
	@echo "  make uninstall      - Remove from $(INSTALL_DIR)"
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
