.PHONY: build install clean test lint fmt package

BIN := craftops
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo dev)
LDFLAGS := -s -w -X craftops/internal/cli.Version=$(VERSION)

build:
	@mkdir -p build
	go build -trimpath -ldflags "$(LDFLAGS)" -o build/$(BIN) ./cmd/$(BIN)

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/$(BIN)

clean:
	rm -rf build dist

test:
	go test -race ./...

lint:
	@command -v golangci-lint >/dev/null && golangci-lint run || go vet ./...

fmt:
	go fmt ./...

package: clean
	@mkdir -p dist
	@for os in linux darwin; do \
		for arch in amd64 arm64; do \
			echo "Building $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/$(BIN)-$$os-$$arch ./cmd/$(BIN); \
		done; \
	done
