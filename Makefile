.PHONY: help build run test lint format install clean

help: ## Show this help
	@echo "\033[93mDaco CLI\033[0m"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

##@ Build
build: ## Build the CLI binary
	@go build -o bin/daco ./cmd/daco

run: ## Run the CLI (e.g. make run ARGS="init --help")
	@go run ./cmd/daco $(ARGS)

install: ## Install to $GOPATH/bin
	@go install ./cmd/daco

##@ Test & Quality
test: ## Run tests
	@go test -v -race ./...

lint: ## Run linter (requires: golangci-lint)
	@golangci-lint run ./...

format: ## Format code
	@gofmt -s -w .
	@goimports -w .

##@ Cleanup
clean: ## Remove build artifacts
	@rm -rf bin/ tmp/
