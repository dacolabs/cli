.PHONY: build-cli install-cli test clean

VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w \
	-X github.com/dacolabs/daco/internal/version.Version=$(VERSION) \
	-X github.com/dacolabs/daco/internal/version.Commit=$(COMMIT) \
	-X github.com/dacolabs/daco/internal/version.Date=$(DATE)

build-cli:
	@go build -ldflags="$(LDFLAGS)" -o bin/daco ./cmd/cli

install-cli:
	@go build -ldflags="$(LDFLAGS)" -o $$(go env GOPATH)/bin/daco ./cmd/cli
	@echo "installed: $$(go env GOPATH)/bin/daco"

test:
	@go test ./...

clean:
	@rm -rf bin/
