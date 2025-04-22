.PHONY: build clean install dev test-version

# Get the current Git commit hash
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
# Get the current date in ISO 8601 format
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
# Version from git tag or dev
VERSION := $(shell git describe --tags 2>/dev/null || echo "dev")
# Test version for development
TEST_VERSION ?= 1.0.15

# Go build flags
LDFLAGS := -s -w \
	-X github.com/dexterity-inc/envi/internal/version.Version=$(VERSION) \
	-X github.com/dexterity-inc/envi/internal/version.Commit=$(COMMIT) \
	-X github.com/dexterity-inc/envi/internal/version.BuildDate=$(DATE)

# Test version build flags
TEST_LDFLAGS := -s -w \
	-X github.com/dexterity-inc/envi/internal/version.Version=$(TEST_VERSION) \
	-X github.com/dexterity-inc/envi/internal/version.Commit=$(COMMIT) \
	-X github.com/dexterity-inc/envi/internal/version.BuildDate=$(DATE)

build:
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/envi ./cmd/envi

dev:
	mkdir -p bin
	go build -o bin/envi ./cmd/envi

test-version:
	mkdir -p bin
	go build -ldflags "$(TEST_LDFLAGS)" -o bin/envi ./cmd/envi
	@echo "\nTesting version command with version $(TEST_VERSION):"
	./bin/envi version

install: build
	cp bin/envi $(GOPATH)/bin/envi

clean:
	rm -rf bin/ 