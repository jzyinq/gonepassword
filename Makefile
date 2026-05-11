SHELL := /bin/bash

fixer: ## run static analysis
	@echo "Static analysis..."
	@golangci-lint run --config .golangci.yml --output.text.path stdout --output.text.colors=true --concurrency 8

tests:
	go test -v -coverprofile=/tmp/godtools.out ./
