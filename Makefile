.PHONY: build clean install dev

# Get the current Git commit hash
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
# Get the current date in ISO 8601 format
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
# Version from git tag or dev
VERSION := $(shell git describe --tags 2>/dev/null || echo "dev")

# Go build flags
LDFLAGS := -s -w \
	-X github.com/dexterity-inc/envi/internal/version.Version=$(VERSION) \
	-X github.com/dexterity-inc/envi/internal/version.Commit=$(COMMIT) \
	-X github.com/dexterity-inc/envi/internal/version.BuildDate=$(DATE)

build:
	go build -ldflags "$(LDFLAGS)" -o bin/envi ./cmd/envi

dev:
	go build -o bin/envi ./cmd/envi

install: build
	cp bin/envi $(GOPATH)/bin/envi

clean:
	rm -rf bin/ 