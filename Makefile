.DEFAULT_GOAL := help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

fmt: ## Run formatters
	gofumpt -w .
	goimports -w .
	golines -w .

lint: fmt ## Run linters; runs with GOOS env var for linting on darwin
	golangci-lint run

test: ## Run unit tests
	go test ./... -v