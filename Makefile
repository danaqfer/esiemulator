# Edge Computing Emulator Suite Makefile
# Makefile for building and running the emulator suite

# Variables
BINARY_NAME = edge-emulator
ESI_GENERATOR_NAME = ESIcontainergenerator
BUILD_DIR = bin
MAIN_PATH = cmd/edge-emulator/main.go
ESI_GENERATOR_PATH = cmd/ESIcontainergenerator/main.go
PROJECT_ROOT = $(shell pwd)

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod
BINARY_UNIX = $(BINARY_NAME)_unix

# Build flags
LDFLAGS = -ldflags="-s -w"

# Default target
.DEFAULT_GOAL := help

# Create build directory
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Build the application
.PHONY: build
build: $(BUILD_DIR)
	@echo "Building Edge Computing Emulator Suite..."
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/edge-emulator
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build ESI Container Generator
.PHONY: build-esi-generator
build-esi-generator: $(BUILD_DIR)
	@echo "Building ESI Container Generator..."
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(ESI_GENERATOR_NAME) ./cmd/ESIcontainergenerator
	@echo "Build complete: $(BUILD_DIR)/$(ESI_GENERATOR_NAME)"

# Build all tools
.PHONY: build-all-tools
build-all-tools: build build-esi-generator
	@echo "All tools built successfully"

# Build for multiple platforms
.PHONY: build-all
build-all: build-linux build-windows build-darwin

.PHONY: build-linux
build-linux: $(BUILD_DIR)
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(ESI_GENERATOR_NAME)-linux-amd64 $(ESI_GENERATOR_PATH)

.PHONY: build-windows
build-windows: $(BUILD_DIR)
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(ESI_GENERATOR_NAME)-windows-amd64.exe $(ESI_GENERATOR_PATH)

.PHONY: build-darwin
build-darwin: $(BUILD_DIR)
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(ESI_GENERATOR_NAME)-darwin-amd64 $(ESI_GENERATOR_PATH)

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
.PHONY: lint
lint:
	@echo "Running linter checks..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, running go vet instead..."; \
		$(GOCMD) vet ./...; \
	fi

# Format code
.PHONY: format
format:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f *.test
	rm -f *.out
	rm -f coverage.html
	rm -f coverage.out

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	$(GOMOD) tidy
	$(GOMOD) download

# Run ESI emulator (Akamai mode)
.PHONY: run
run:
	@echo "Running ESI Emulator in Akamai mode..."
	EMULATOR_MODE=esi ESI_MODE=akamai $(GOCMD) run $(MAIN_PATH) -mode=esi -esi-mode=akamai

# Run ESI emulator (Fastly mode)
.PHONY: run-fastly
run-fastly:
	@echo "Running ESI Emulator in Fastly mode..."
	EMULATOR_MODE=esi ESI_MODE=fastly $(GOCMD) run $(MAIN_PATH) -mode=esi -esi-mode=fastly

# Run ESI emulator (W3C mode)
.PHONY: run-w3c
run-w3c:
	@echo "Running ESI Emulator in W3C mode..."
	EMULATOR_MODE=esi ESI_MODE=w3c $(GOCMD) run $(MAIN_PATH) -mode=esi -esi-mode=w3c

# Run Property Manager emulator
.PHONY: run-property-manager
run-property-manager:
	@echo "Running Property Manager Emulator..."
	EMULATOR_MODE=property-manager $(GOCMD) run $(MAIN_PATH) -mode=property-manager

# Run examples
.PHONY: examples
examples:
	@echo "Running examples..."
	@if [ -d "cmd/examples" ]; then \
		for file in cmd/examples/*.go; do \
			if [ -f "$$file" ]; then \
				echo "Running example: $$(basename $$file)"; \
				$(GOCMD) run "$$file"; \
			fi; \
		done; \
	else \
		echo "No examples found in cmd/examples"; \
	fi

# Run with debug mode
.PHONY: run-debug
run-debug:
	@echo "Running ESI Emulator in debug mode..."
	EMULATOR_MODE=esi ESI_MODE=akamai DEBUG=true $(GOCMD) run $(MAIN_PATH) -mode=esi -esi-mode=akamai -debug

# Run Property Manager with debug mode
.PHONY: run-property-manager-debug
run-property-manager-debug:
	@echo "Running Property Manager Emulator in debug mode..."
	EMULATOR_MODE=property-manager DEBUG=true $(GOCMD) run $(MAIN_PATH) -mode=property-manager -debug

# Build and run binary
.PHONY: run-binary
run-binary: build
	@echo "Running built binary..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Install the binary
.PHONY: install
install: build
	@echo "Installing binary..."
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# Uninstall the binary
.PHONY: uninstall
uninstall:
	@echo "Uninstalling binary..."
	rm -f /usr/local/bin/$(BINARY_NAME)

# Generate documentation
.PHONY: docs
docs:
	@echo "Generating documentation..."
	@if command -v godoc >/dev/null 2>&1; then \
		godoc -http=:6060; \
	else \
		echo "godoc not found, skipping documentation generation"; \
	fi

# Check for security vulnerabilities
.PHONY: security
security:
	@echo "Checking for security vulnerabilities..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not found, skipping security check"; \
	fi

# Run benchmarks
.PHONY: benchmark
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Show help
.PHONY: help
help:
	@echo "Edge Computing Emulator Suite - Available Commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  build              Build the main application"
	@echo "  build-esi-generator Build ESI Container Generator"
	@echo "  build-all-tools    Build all tools (main app + ESI generator)"
	@echo "  build-all          Build for all platforms (Linux, Windows, macOS)"
	@echo "  build-linux        Build for Linux"
	@echo "  build-windows      Build for Windows"
	@echo "  build-darwin       Build for macOS"
	@echo ""
	@echo "Run Commands:"
	@echo "  run                Run ESI emulator (Akamai mode)"
	@echo "  run-fastly         Run ESI emulator (Fastly mode)"
	@echo "  run-w3c            Run ESI emulator (W3C mode)"
	@echo "  run-property-manager Run Property Manager emulator"
	@echo "  run-debug          Run ESI emulator in debug mode"
	@echo "  run-property-manager-debug Run Property Manager in debug mode"
	@echo "  run-binary         Build and run binary"
	@echo ""
	@echo "ESI Generator Commands:"
	@echo "  ./bin/ESIcontainergenerator -input config.json -output container.html"
	@echo "  ./bin/ESIcontainergenerator -input config.json -verbose"
	@echo "  ./bin/ESIcontainergenerator -help"
	@echo ""
	@echo "Development Commands:"
	@echo "  test               Run tests"
	@echo "  test-coverage      Run tests with coverage"
	@echo "  lint               Run linter checks"
	@echo "  format             Format code"
	@echo "  deps               Install dependencies"
	@echo "  examples           Run example programs"
	@echo "  benchmark          Run benchmarks"
	@echo ""
	@echo "Utility Commands:"
	@echo "  clean              Clean build artifacts"
	@echo "  install            Install binary to /usr/local/bin"
	@echo "  uninstall          Remove binary from /usr/local/bin"
	@echo "  docs               Generate documentation"
	@echo "  security           Check for security vulnerabilities"
	@echo "  help               Show this help"
	@echo ""
	@echo "Environment Variables:"
	@echo "  EMULATOR_MODE      Set to 'esi' or 'property-manager'"
	@echo "  ESI_MODE           Set to 'fastly', 'akamai', 'w3c', or 'development'"
	@echo "  PORT               Server port (default: 3000)"
	@echo "  DEBUG              Enable debug mode"
	@echo ""
	@echo "Examples:"
	@echo "  make build-all-tools"
	@echo "  make run"
	@echo "  make run-property-manager"
	@echo "  make test-coverage"
	@echo "  ./bin/ESIcontainergenerator -input cmd/ESIcontainergenerator/example-config.json -verbose"
	@echo "  EMULATOR_MODE=esi ESI_MODE=fastly make run" 