.PHONY: build install clean test lint fmt dev package

VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-s -w -X craftops/internal/cli.Version=$(VERSION)"

build:
	@mkdir -p build
	go build -trimpath $(LDFLAGS) -o build/craftops ./cmd/craftops

install:
	go install $(LDFLAGS) ./cmd/craftops

install-system: build
	sudo install -m755 build/craftops /usr/local/bin/

clean:
	rm -rf build dist coverage.out

test:
	go test -race -cover ./...

lint:
	go vet ./...
	@command -v golangci-lint >/dev/null && golangci-lint run -c .github/.golangci.yml || true

fmt:
	go fmt ./...
	@command -v goimports >/dev/null && goimports -w . || true

dev:
	go mod tidy
	@command -v golangci-lint >/dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

package: clean
	@mkdir -p dist
	@for os in linux darwin; do \
		for arch in amd64 arm64; do \
			echo "Building $$os/$$arch"; \
			GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -trimpath $(LDFLAGS) -o dist/craftops-$$os-$$arch ./cmd/craftops; \
		done; \
	done
