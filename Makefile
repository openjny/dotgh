.PHONY: build test test-v test-cover lint fmt clean release-check release-dry help

# Version information
VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +"%Y-%m-%d")
LDFLAGS := -X github.com/openjny/dotgh/internal/version.Version=$(VERSION) \
           -X github.com/openjny/dotgh/internal/version.Commit=$(COMMIT) \
           -X github.com/openjny/dotgh/internal/version.Date=$(DATE)

# Build binary
build:
	go build -ldflags "$(LDFLAGS)" -o dotgh ./cmd/dotgh

# Run all tests
test:
	go test ./...

# Run tests with verbose output
test-v:
	go test -v ./...

# Run tests with coverage report
test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@rm -f coverage.out

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Clean build artifacts
clean:
	rm -f dotgh dotgh.exe
	rm -rf dist/

# Validate goreleaser configuration
release-check:
	goreleaser check

# Run goreleaser in snapshot mode (dry run)
release-dry:
	goreleaser release --snapshot --clean

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  test          - Run all tests"
	@echo "  test-v        - Run tests with verbose output"
	@echo "  test-cover    - Run tests with coverage report"
	@echo "  lint          - Run golangci-lint"
	@echo "  fmt           - Format code"
	@echo "  clean         - Remove build artifacts"
	@echo "  release-check - Validate goreleaser configuration"
	@echo "  release-dry   - Run goreleaser in snapshot mode"
	@echo "  help          - Show this help message"
