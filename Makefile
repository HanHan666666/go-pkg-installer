# go-pkg-installer Makefile

.PHONY: all build test lint clean cover fmt vet tidy help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOFMT=gofmt
GOMOD=$(GOCMD) mod
GOLINT=golangci-lint

# Build parameters
BINARY_NAME=installer
BINARY_CLI=installer-cli
BUILD_DIR=build
CMD_DIR=cmd

# Test parameters
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Default target
all: tidy fmt vet lint test build

# Build the GUI application
build: build-gui

build-gui:
	@echo "Building GUI installer..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)/installer/...

# Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race ./...

# Run tests with coverage
cover:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

# Run tests for a specific package
test-pkg:
	@echo "Running tests for $(PKG)..."
	$(GOTEST) -v -race ./$(PKG)/...

# Run linter
lint:
	@echo "Running linter..."
	@if command -v $(GOLINT) >/dev/null 2>&1; then \
		$(GOLINT) run ./...; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# Check formatting (for CI)
fmt-check:
	@echo "Checking code formatting..."
	@test -z "$$($(GOFMT) -s -l .)" || (echo "Code not formatted. Run 'make fmt'" && exit 1)

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download

# Verify dependencies
verify:
	@echo "Verifying dependencies..."
	$(GOMOD) verify

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	$(GOCMD) clean

# Generate (for go generate directives)
generate:
	@echo "Running go generate..."
	$(GOCMD) generate ./...

# Run the GUI installer (for development)
run-gui: build-gui
	@echo "Running GUI installer..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run the CLI installer (for development)
run-cli: build-cli
	@echo "Running CLI installer..."
	./$(BUILD_DIR)/$(BINARY_CLI)

# Install tools
tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Show help
help:
	@echo "Available targets:"
	@echo "  all        - Run tidy, fmt, vet, lint, test, build"
	@echo "  build      - Build both GUI and CLI binaries"
	@echo "  build-gui  - Build GUI installer"
	@echo "  build-cli  - Build CLI installer"
	@echo "  test       - Run all tests"
	@echo "  cover      - Run tests with coverage report"
	@echo "  lint       - Run golangci-lint"
	@echo "  vet        - Run go vet"
	@echo "  fmt        - Format code"
	@echo "  fmt-check  - Check code formatting (CI)"
	@echo "  tidy       - Tidy go.mod"
	@echo "  deps       - Download dependencies"
	@echo "  clean      - Clean build artifacts"
	@echo "  tools      - Install development tools"
	@echo "  help       - Show this help"
