#!/bin/bash

# Set a test version for development
export TEST_VERSION="1.0.15"

echo "Building Envi CLI with test version $TEST_VERSION..."
go build -ldflags "-s -w \
  -X github.com/dexterity-inc/envi/internal/version.Version=$TEST_VERSION \
  -X github.com/dexterity-inc/envi/internal/version.Commit=$(git rev-parse --short HEAD) \
  -X github.com/dexterity-inc/envi/internal/version.BuildDate=$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
  -o ./bin/envi ./cmd/envi

echo -e "\nTesting version short flag:"
./bin/envi -v

echo -e "\nTesting version long flag:"
./bin/envi --version 