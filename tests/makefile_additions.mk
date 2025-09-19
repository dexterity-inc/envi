# Test targets to add to the main Makefile

.PHONY: test test-unit test-integration test-security test-coverage test-race test-benchmark test-clean

# Run all tests
test: test-unit test-integration

# Run only unit tests
test-unit:
	@echo "Running unit tests..."
	go test ./tests/unit/... -v

# Run integration tests (requires GITHUB_TOKEN)
test-integration:
	@echo "Running integration tests..."
	@if [ -z "$(GITHUB_TOKEN)" ]; then \
		echo "Warning: GITHUB_TOKEN not set, some integration tests will be skipped"; \
	fi
	go test ./tests/integration/... -v

# Run security-focused tests
test-security:
	@echo "Running security tests..."
	go test ./tests/unit/security/... -v
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed, skipping security scan"; \
	fi

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test ./tests/... -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out | grep total:

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	go test ./tests/... -race -v

# Run benchmark tests
test-benchmark:
	@echo "Running benchmark tests..."
	go test ./tests/... -bench=. -benchmem

# Clean test artifacts
test-clean:
	@echo "Cleaning test artifacts..."
	rm -f coverage.out coverage.html
	go clean -testcache

# Run all test validations (for CI)
test-ci: test-unit test-security test-coverage test-race
	@echo "All CI tests completed successfully"

# Quick test (unit tests only, no coverage)
test-quick:
	@echo "Running quick tests..."
	go test ./tests/unit/... -short