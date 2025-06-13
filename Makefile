# ESI Emulator Makefile

# Variables
BINARY_NAME=esi-emulator
BUILD_DIR=build
MAIN_FILE=main.go
EXAMPLES_FILE=cmd/examples/main.go

# Ensure Go workspaces don't interfere
export GOWORK=off

# Default target
.PHONY: all
all: build

# Build the main application
.PHONY: build
build:
	@echo "🔨 Building ESI Emulator..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
.PHONY: build-all
build-all:
	@echo "🔨 Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_FILE)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_FILE)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_FILE)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_FILE)
	@echo "✅ Multi-platform build complete"

# Run the application in development mode
.PHONY: run
run:
	@echo "🚀 Running ESI Emulator in development mode..."
	go run $(MAIN_FILE) -mode development -debug

# Run the application in different modes
.PHONY: run-fastly
run-fastly:
	@echo "🚀 Running ESI Emulator in Fastly mode..."
	go run $(MAIN_FILE) -mode fastly

.PHONY: run-akamai
run-akamai:
	@echo "🚀 Running ESI Emulator in Akamai mode..."
	go run $(MAIN_FILE) -mode akamai

.PHONY: run-w3c
run-w3c:
	@echo "🚀 Running ESI Emulator in W3C mode..."
	go run $(MAIN_FILE) -mode w3c

# Run examples
.PHONY: examples
examples:
	@echo "📚 Running examples..."
	go run $(EXAMPLES_FILE)

# Install dependencies
.PHONY: deps
deps:
	@echo "📦 Installing dependencies..."
	go mod tidy
	go mod download

# Run tests
.PHONY: test
test:
	@echo "🧪 Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report generated: coverage.html"

# Run linting
.PHONY: lint
lint:
	@echo "🔍 Running linting..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		go vet ./...; \
	fi

# Format code
.PHONY: fmt
fmt:
	@echo "✨ Formatting code..."
	go fmt ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Run the built binary
.PHONY: run-binary
run-binary: build
	@echo "🚀 Running built binary..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Install the binary to GOPATH/bin
.PHONY: install
install:
	@echo "📦 Installing to GOPATH/bin..."
	go install $(MAIN_FILE)

# Development server with auto-reload (requires air)
.PHONY: dev
dev:
	@echo "🔥 Starting development server with auto-reload..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "⚠️  air not installed. Run: go install github.com/cosmtrek/air@latest"; \
		echo "🔄 Falling back to regular run..."; \
		$(MAKE) run; \
	fi

# Docker build
.PHONY: docker-build
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t esi-emulator .

# Docker run
.PHONY: docker-run
docker-run: docker-build
	@echo "🐳 Running Docker container..."
	docker run -p 3000:3000 esi-emulator

# Show help
.PHONY: help
help:
	@echo "ESI Emulator - Available Commands:"
	@echo ""
	@echo "  make build         - Build the application"
	@echo "  make run           - Run in development mode"
	@echo "  make run-fastly    - Run in Fastly mode"
	@echo "  make run-akamai    - Run in Akamai mode"
	@echo "  make run-w3c       - Run in W3C mode"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-run    - Run Docker container"
	@echo "  make help          - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build && ./build/esi-emulator -help"

# Default help when no target is specified
.DEFAULT_GOAL := help 