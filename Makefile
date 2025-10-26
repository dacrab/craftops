# CraftOps Makefile (streamlined)

.PHONY: all help build install install-system clean test test-race lint fmt run dev package

# Build configuration
GO?=go
BINARY_NAME=craftops
BUILD_DIR=build
DIST_DIR=dist
VERSION?=$(shell git describe --tags --always 2>/dev/null || echo 2.0.1)

# Go build flags
LDFLAGS=-ldflags "-X craftops/internal/cli.Version=$(VERSION) -s -w"
BUILD_FLAGS=-trimpath

# Default target
help:
	@echo "Targets: build, install, install-system, clean, test, test-race, lint, fmt, run, dev, package"

# Build the binary
build:
	@echo "Building $(BINARY_NAME) ($(VERSION))..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/craftops

# Install the binary
install:
	$(GO) install $(LDFLAGS) ./cmd/craftops

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR) $(DIST_DIR) coverage.out coverage.html
	$(GO) clean -cache -testcache

# Run tests
test:
	$(GO) test -v ./...

test-race:
	$(GO) test -race -covermode=atomic -coverpkg=./... -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Run linting
lint:
	$(GO) vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then golangci-lint run; else echo "golangci-lint not installed"; fi

# Format code
fmt:
	$(GO) fmt ./...
	@if command -v goimports >/dev/null 2>&1; then goimports -w .; fi

# Run the application
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Install development dependencies
dev:
	$(GO) mod tidy
	@if ! command -v golangci-lint >/dev/null 2>&1; then $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; fi
	@if ! command -v goimports >/dev/null 2>&1; then $(GO) install golang.org/x/tools/cmd/goimports@latest; fi

# Install system-wide (requires sudo)
install-system: build
	sudo install -m0755 $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

# Create distribution packages
OS_ARCH := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64

package: clean
	@mkdir -p $(DIST_DIR)
	@set -e; for oa in $(OS_ARCH); do \
	  os=$${oa%-*}; arch=$${oa#*-}; \
	  echo "Building $$os/$$arch"; \
	  GOOS=$$os GOARCH=$$arch $(GO) build $(BUILD_FLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-$$os-$$arch ./cmd/craftops; \
	done