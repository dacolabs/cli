.PHONY: help setup build run test clean release-snapshot

VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
LDFLAGS := -s -w -X main.version=$(VERSION)

help: ## Show this help
	@echo "Daco CLI (Catalog)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: ## Configure git hooks
	@git config core.hooksPath .githooks
	@echo "Git hooks configured"

build: ## Build the CLI binary
	@go build -ldflags="$(LDFLAGS)" -o bin/daco .

run: ## Run the CLI (e.g. make run ARGS="--help")
	@go run -ldflags="$(LDFLAGS)" . $(ARGS)

test: ## Run tests
	@go test -v -race ./...

clean: ## Remove build artifacts
	@rm -rf bin/

release-snapshot: ## Local GoReleaser dry run
	@goreleaser release --snapshot --clean
