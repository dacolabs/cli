.PHONY: help setup build build-cli build-api run-cli run-api test lint format install install-cli install-api clean release-snapshot

VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w \
	-X github.com/dacolabs/daco/internal/version.Version=$(VERSION) \
	-X github.com/dacolabs/daco/internal/version.Commit=$(COMMIT) \
	-X github.com/dacolabs/daco/internal/version.Date=$(DATE)

help: ## Show this help
	@echo "\033[93mDaco\033[0m"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

##@ Setup
setup: ## Set up development environment (git hooks)
	@git config core.hooksPath .githooks
	@echo "Git hooks configured"

##@ Build
build: build-cli build-api ## Build both binaries (daco + daco-api)

build-cli: ## Build the CLI binary (daco)
	@go build -ldflags="$(LDFLAGS)" -o bin/daco ./cmd/cli

build-api: ## Build the API server binary (daco-api)
	@go build -ldflags="$(LDFLAGS)" -o bin/daco-api ./cmd/api

run-cli: ## Run the CLI (e.g. make run-cli ARGS="--help")
	@go run -ldflags="$(LDFLAGS)" ./cmd/cli $(ARGS)

run-api: ## Run the API server
	@go run -ldflags="$(LDFLAGS)" ./cmd/api

install: install-cli install-api ## Install both binaries to $GOPATH/bin

install-cli: ## Install the CLI to $GOPATH/bin
	@go install -ldflags="$(LDFLAGS)" ./cmd/cli
	@echo "installed: $$(go env GOPATH)/bin/daco"

install-api: ## Install the API server to $GOPATH/bin
	@go install -ldflags="$(LDFLAGS)" ./cmd/api
	@echo "installed: $$(go env GOPATH)/bin/daco-api"

##@ Test & Quality
test: ## Run tests with race detection
	@go test -v -race ./...

lint: ## Run linter (requires: golangci-lint)
	@golangci-lint run ./...

format: ## Format code
	@gofmt -s -w .
	@goimports -w .

##@ Release
release-snapshot: ## Build release artifacts locally (for testing; requires: goreleaser)
	@goreleaser release --snapshot --clean

##@ Cleanup
clean: ## Remove build artifacts
	@rm -rf bin/ tmp/ dist/
