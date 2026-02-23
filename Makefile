.PHONY: build install clean test lint fmt package tidy verify

BIN := craftops
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo dev)
# Go 1.25.7 optimizations: added -buildvcs and PGO support
LDFLAGS := -s -w -X craftops/internal/cli.Version=$(VERSION)
BUILDFLAGS := -trimpath -buildvcs=auto

build:
	@mkdir -p build
	go build $(BUILDFLAGS) -ldflags "$(LDFLAGS)" -o build/$(BIN) ./cmd/$(BIN)

install:
	go install $(BUILDFLAGS) -ldflags "$(LDFLAGS)" ./cmd/$(BIN)

clean:
	rm -rf build dist

test:
	go test -race -cover -coverprofile=coverage.out ./...

test-verbose:
	go test -race -cover -coverprofile=coverage.out -v ./...

lint:
	@command -v golangci-lint >/dev/null && golangci-lint run -c .github/.golangci.yml || go vet ./...

fmt:
	go fmt ./...
	@command -v goimports >/dev/null && goimports -w . || true

tidy:
	go mod tidy -go=1.25.7

verify: fmt tidy
	@git diff --exit-code go.mod go.sum || (echo "go.mod or go.sum changed after tidy" && exit 1)

package: clean
	@mkdir -p dist
	@for os in linux darwin; do \
		for arch in amd64 arm64; do \
			echo "Building $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build $(BUILDFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(BIN)-$$os-$$arch ./cmd/$(BIN); \
		done; \
	done
