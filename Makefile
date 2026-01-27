.PHONY: help setup build run test lint format install clean release-snapshot

VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w \
	-X github.com/dacolabs/cli/internal/version.Version=$(VERSION) \
	-X github.com/dacolabs/cli/internal/version.Commit=$(COMMIT) \
	-X github.com/dacolabs/cli/internal/version.Date=$(DATE)

help: ## Show this help
	@echo "\033[93mDaco CLI\033[0m"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

##@ Setup
setup: ## Set up development environment
	@git config core.hooksPath .githooks
	@echo "Git hooks configured"

##@ Build
build: ## Build the CLI binary
	@go build -ldflags="$(LDFLAGS)" -o bin/daco ./cmd/daco

run: ## Run the CLI (e.g. make run ARGS="init --help")
	@go run -ldflags="$(LDFLAGS)" ./cmd/daco $(ARGS)

install: ## Install to $GOPATH/bin
	@go install -ldflags="$(LDFLAGS)" ./cmd/daco

##@ Test & Quality
test: ## Run tests
	@go test -v -race ./...

lint: ## Run linter (requires: golangci-lint)
	@golangci-lint run ./...

format: ## Format code
	@gofmt -s -w .
	@goimports -w .

##@ Release
release-snapshot: ## Build release artifacts locally (for testing)
	@goreleaser release --snapshot --clean

##@ Cleanup
clean: ## Remove build artifacts
	@rm -rf bin/ tmp/ dist/
