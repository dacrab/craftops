.PHONY: build install clean test lint fmt tidy package

BIN := craftops
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo dev)
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

lint:
	@command -v golangci-lint >/dev/null && golangci-lint run -c .github/.golangci.yml || go vet ./...

fmt:
	gofmt -w .
	@command -v goimports >/dev/null && goimports -w $(shell find . -name '*.go' -not -path './vendor/*') || true

tidy:
	go mod tidy

package: clean
	@mkdir -p dist
	@for os in linux darwin; do \
		for arch in amd64 arm64; do \
			echo "Building $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build $(BUILDFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(BIN)-$$os-$$arch ./cmd/$(BIN); \
		done; \
	done
	@cd dist && sha256sum $(BIN)-* > SHA256SUMS
