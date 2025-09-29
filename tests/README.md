# Test Configuration and Automation

This document outlines the test configuration and CI/CD setup for the envi project.

## Test Structure

```
tests/
├── unit/                    # Unit tests for individual components
│   ├── config/             # Configuration management tests
│   ├── encryption/         # Encryption/decryption tests
│   ├── security/          # Security validation tests
│   ├── tui/               # Terminal UI component tests
│   ├── version/           # Version handling tests
│   └── cmd/               # CLI command tests
├── integration/            # Integration tests
│   ├── cli_commands/      # End-to-end CLI command tests
│   ├── workflows/         # Complete workflow tests
│   └── api_integration/   # GitHub API integration tests
├── fixtures/              # Test data and fixtures
│   ├── test_env_files/    # Sample .env files for testing
│   ├── test_configs/      # Sample configuration files
│   └── mock_responses/    # Mock API responses
├── helpers/               # Test utilities and helpers
└── testdata/             # Additional test data
```

## Running Tests

### Unit Tests Only
```bash
go test ./tests/unit/... -v
```

### Integration Tests
```bash
# Requires GITHUB_TOKEN environment variable
export GITHUB_TOKEN=your_github_token
go test ./tests/integration/... -v
```

### All Tests
```bash
go test ./tests/... -v
```

### With Coverage
```bash
go test ./tests/... -v -cover -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### With Race Detection
```bash
go test ./tests/... -race -v
```

## Test Categories

### 1. Unit Tests
- **Config Package**: Configuration loading, saving, token validation
- **Encryption Package**: AES-256 encryption, PBKDF2 key derivation, masking
- **Security Package**: Path validation, input sanitization, environment variable validation
- **TUI Package**: Terminal interface components, user input handling
- **Version Package**: Version information handling
- **CMD Package**: Individual command logic and flag handling

### 2. Integration Tests
- **Push/Pull Workflows**: Complete GitHub Gist integration
- **Encryption Workflows**: End-to-end encryption and decryption
- **Configuration Workflows**: Setup and token management
- **Sharing Workflows**: Multi-user sharing scenarios
- **Validation Workflows**: File validation and error handling

### 3. Security Tests
- **Path Traversal Protection**: Prevents directory traversal attacks
- **Input Validation**: Validates environment variable names and values
- **Encryption Security**: Tests encryption strength and key derivation
- **Token Security**: Validates GitHub token handling and storage

### 4. Performance Tests
- **Large File Handling**: Tests with large .env files
- **Encryption Performance**: Benchmarks encryption/decryption speed
- **API Rate Limiting**: Tests GitHub API rate limit handling

## Continuous Integration

### GitHub Actions Workflow
Create `.github/workflows/test.yml`:

```yaml
name: Test Suite

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go-version: [1.21, 1.22, 1.23]
    
    runs-on: ${{ matrix.os }}
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run unit tests
      run: go test ./tests/unit/... -v -race -coverprofile=coverage.out
    
    - name: Run integration tests
      if: matrix.os == 'ubuntu-latest' && matrix.go-version == '1.23'
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      run: go test ./tests/integration/... -v
    
    - name: Upload coverage to Codecov
      if: matrix.os == 'ubuntu-latest' && matrix.go-version == '1.23'
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
        flags: unittests
        name: codecov-umbrella

  security:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.23'
    
    - name: Run security tests
      run: go test ./tests/unit/security/... -v
    
    - name: Run gosec Security Scanner
      uses: securecodewarrior/github-action-gosec@master
      with:
        args: '-fmt sarif -out gosec.sarif ./...'
    
    - name: Upload SARIF file
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: gosec.sarif
```

## Test Environment Variables

### Required for Integration Tests
- `GITHUB_TOKEN`: GitHub Personal Access Token with gist permissions

### Optional Test Configuration
- `ENVI_TEST_TIMEOUT`: Test timeout in seconds (default: 30)
- `ENVI_TEST_VERBOSE`: Enable verbose test output
- `ENVI_SKIP_INTEGRATION`: Skip integration tests

## Makefile Targets

Add to existing Makefile:

```makefile
# Test targets
.PHONY: test test-unit test-integration test-security test-coverage test-race

test: test-unit test-integration

test-unit:
	go test ./tests/unit/... -v

test-integration:
	go test ./tests/integration/... -v

test-security:
	go test ./tests/unit/security/... -v
	gosec ./...

test-coverage:
	go test ./tests/... -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-race:
	go test ./tests/... -race -v

test-benchmark:
	go test ./tests/... -bench=. -benchmem

test-clean:
	rm -f coverage.out coverage.html
	go clean -testcache
```

## Test Data Management

### Environment Files
- `basic.env`: Standard environment variables
- `edge_cases.env`: Special characters and edge cases
- `production.env`: Production-like configuration
- `invalid.env`: Invalid formats for error testing

### Configuration Files
- `basic_config.yaml`: Standard configuration
- `empty_config.yaml`: Minimal configuration
- `full_config.yaml`: All options configured

### Mock Data
- GitHub API responses for various scenarios
- Encrypted content samples
- Key file examples

## Best Practices

1. **Test Isolation**: Each test should be independent and not rely on external state
2. **Mock External Dependencies**: Use mocks for GitHub API, file system, keyring
3. **Environment Cleanup**: Always clean up test files and restore environment
4. **Error Testing**: Test both success and failure scenarios
5. **Edge Cases**: Test boundary conditions and invalid inputs
6. **Performance**: Include benchmarks for critical operations
7. **Security**: Validate all security-related functionality thoroughly

## Code Coverage Goals

- **Unit Tests**: > 85% coverage
- **Integration Tests**: > 70% coverage for workflows
- **Security Tests**: 100% coverage for security functions
- **Overall**: > 80% total coverage