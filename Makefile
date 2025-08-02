# CraftOps Makefile

.PHONY: help build install clean test lint format run dev

# Build configuration
BINARY_NAME=craftops
BUILD_DIR=build
DIST_DIR=dist
VERSION=2.0.1

# Go build flags
LDFLAGS=-ldflags "-X craftops/internal/cli.Version=$(VERSION) -s -w"
BUILD_FLAGS=-trimpath

# Default target
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  install       - Install the binary to GOPATH/bin"
	@echo "  install-system- Install system-wide to /usr/local/bin (requires sudo)"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  lint          - Run linting checks"
	@echo "  format        - Format code"
	@echo "  run           - Run the application"
	@echo "  dev           - Install development dependencies"
	@echo "  package       - Create distribution packages for all platforms"

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/craftops

# Install the binary
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) ./cmd/craftops
	@echo "Creating aliases..."
	@if [ -w "$(shell go env GOPATH)/bin" ]; then \
		echo "Installed $(BINARY_NAME)"; \
	else \
		echo "Note: Run 'sudo ln -sf $(shell go env GOPATH)/bin/$(BINARY_NAME) /usr/local/bin/craftops' to create system aliases"; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)/
	rm -rf $(DIST_DIR)/
	go clean -cache
	go clean -testcache

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linting
lint:
	@echo "Running linting checks..."
	go vet ./...
	go fmt ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping advanced linting"; \
	fi

# Format code
format:
	@echo "Formatting code..."
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

# Run the application
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Install development dependencies
dev:
	@echo "Installing development dependencies..."
	go mod tidy
	go mod download
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@if ! command -v goimports >/dev/null 2>&1; then \
		echo "Installing goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
	fi

# Install system-wide (requires sudo)
install-system: build
	@echo "Installing $(BINARY_NAME) system-wide..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "Creating system aliases..."
	# No aliases needed - craftops is short and memorable
	@echo "✅ System installation complete!"
	@echo "Available command: $(BINARY_NAME)"

# Create distribution packages
package: clean build
	@echo "Creating distribution packages..."
	@mkdir -p $(DIST_DIR)
	
	# Linux amd64
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/craftops
	
	# Linux arm64
	GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/craftops
	
	# macOS amd64
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/craftops
	
	# macOS arm64
	GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/craftops
	

	
	@echo "Distribution packages created in $(DIST_DIR)/"