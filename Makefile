GO ?= go
VERSION ?= dev
LDFLAGS := -X github.com/relux-works/curator/internal/version.value=$(VERSION)

.PHONY: build test fmt lint vet check

build:
	$(GO) build -ldflags '$(LDFLAGS)' -o bin/curator ./cmd/curator

test:
	$(GO) test ./...

fmt:
	gofmt -l -w .

vet:
	$(GO) vet ./...

lint:
	golangci-lint run

check: vet test
	@test -z "$$(gofmt -l .)" || { echo 'gofmt: files need formatting:'; gofmt -l .; exit 1; }
