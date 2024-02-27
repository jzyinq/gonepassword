SHELL := /bin/bash

fixer: ## run static analysis
	@echo "Static analysis..."
	@golangci-lint run --config .golangci.yml --out-format=colored-line-number --concurrency 8