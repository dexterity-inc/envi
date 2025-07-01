# Makefile for envi project
.PHONY: help test test-unit test-integration test-coverage test-benchmark clean build install lint format vet

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Test targets
test: test-unit ## Run all tests

test-unit: ## Run unit tests for all packages
	@echo "Running unit tests..."
	go test -v ./internal/utils ./internal/config ./internal/encryption ./internal/cmd

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	go test -v ./internal/cmd -run "Integration"

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-benchmark: ## Run benchmark tests
	@echo "Running benchmark tests..."
	go test -bench=. -benchmem ./internal/utils ./internal/config ./internal/encryption ./internal/cmd

test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	go test -race ./internal/...

test-verbose: ## Run tests with verbose output
	@echo "Running tests with verbose output..."
	go test -v -count=1 ./internal/...

# Build targets
build: ## Build the binary
	@echo "Building envi..."
	go build -o bin/envi ./cmd/envi

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 go build -o bin/envi-linux-amd64 ./cmd/envi
	GOOS=darwin GOARCH=amd64 go build -o bin/envi-darwin-amd64 ./cmd/envi
	GOOS=darwin GOARCH=arm64 go build -o bin/envi-darwin-arm64 ./cmd/envi
	GOOS=windows GOARCH=amd64 go build -o bin/envi-windows-amd64.exe ./cmd/envi

install: build ## Install the binary to GOPATH/bin
	@echo "Installing envi..."
	go install ./cmd/envi

# Code quality targets
lint: ## Run golangci-lint
	@echo "Running linter..."
	golangci-lint run

format: ## Format code using gofmt
	@echo "Formatting code..."
	gofmt -w .

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

# Development targets
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

clean: ## Clean build artifacts and test files
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -testcache

dev-setup: deps ## Set up development environment
	@echo "Setting up development environment..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Testing utilities
test-utils: ## Test only utils package
	go test -v ./internal/utils

test-config: ## Test only config package  
	go test -v ./internal/config

test-encryption: ## Test only encryption package
	go test -v ./internal/encryption

test-cmd: ## Test only cmd package
	go test -v ./internal/cmd

# Documentation
docs: ## Generate documentation
	@echo "Generating documentation..."
	go doc -all ./internal/... > docs/api.md

# Release targets
release-test: test lint vet ## Run all checks before release
	@echo "All checks passed! Ready for release."

version: ## Show version information
	@echo "Go version: $$(go version)"
	@echo "Git commit: $$(git rev-parse --short HEAD)"
	@echo "Git branch: $$(git branch --show-current)"

# Docker targets (if you want to add Docker support)
docker-build: ## Build Docker image
	docker build -t envi:latest .

docker-test: ## Run tests in Docker
	docker run --rm -v $$(pwd):/app -w /app golang:1.21 make test

# Performance profiling
profile-cpu: ## Run CPU profiling
	go test -cpuprofile=cpu.prof -bench=. ./internal/...
	go tool pprof cpu.prof

profile-mem: ## Run memory profiling  
	go test -memprofile=mem.prof -bench=. ./internal/...
	go tool pprof mem.prof

# Git hooks setup
setup-hooks: ## Set up git pre-commit hooks
	@echo "Setting up git hooks..."
	echo '#!/bin/sh\nmake test lint' > .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit
	@echo "Pre-commit hook installed" 