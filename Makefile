.PHONY: help build build-cli run-cli test test-coverage lint format install install-cli clean release-snapshot

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

##@ Build
build: build-cli ## Build all binaries

build-cli: ## Build the CLI binary (daco)
	@go build -ldflags="$(LDFLAGS)" -o bin/daco ./cmd/daco

run-cli: ## Run the CLI (e.g. make run-cli ARGS="--help")
	@go run -ldflags="$(LDFLAGS)" ./cmd/daco $(ARGS)

install: install-cli ## Install all binaries to $GOPATH/bin

install-cli: ## Install the CLI to $GOPATH/bin
	@go install -ldflags="$(LDFLAGS)" ./cmd/daco
	@echo "installed: $$(go env GOPATH)/bin/daco"

##@ Test & Quality
test: ## Run tests with race detection + per-package coverage
	@go test -race -cover ./...

test-coverage: ## Run tests and emit an HTML coverage report (opens coverage.html)
	@go test -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | tail -1
	@go tool cover -html=coverage.out -o coverage.html
	@echo "coverage report: coverage.html"

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
	@rm -rf bin/ tmp/ dist/ coverage.out coverage.html
