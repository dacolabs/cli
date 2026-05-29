.PHONY: build build-cli build-api install install-cli install-api test lint clean

VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w \
	-X github.com/dacolabs/daco/internal/version.Version=$(VERSION) \
	-X github.com/dacolabs/daco/internal/version.Commit=$(COMMIT) \
	-X github.com/dacolabs/daco/internal/version.Date=$(DATE)

build: build-cli build-api

build-cli:
	@go build -ldflags="$(LDFLAGS)" -o bin/daco ./cmd/cli

build-api:
	@go build -ldflags="$(LDFLAGS)" -o bin/daco-api ./cmd/api

install: install-cli install-api

install-cli:
	@go build -ldflags="$(LDFLAGS)" -o $$(go env GOPATH)/bin/daco ./cmd/cli
	@echo "installed: $$(go env GOPATH)/bin/daco"

install-api:
	@go build -ldflags="$(LDFLAGS)" -o $$(go env GOPATH)/bin/daco-api ./cmd/api
	@echo "installed: $$(go env GOPATH)/bin/daco-api"

test:
	@go test ./...

lint:
	@golangci-lint run --timeout=5m

clean:
	@rm -rf bin/
