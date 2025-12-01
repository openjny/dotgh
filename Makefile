.PHONY: build test lint fmt clean help

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

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Clean build artifacts
clean:
	rm -f dotgh dotgh.exe

# Show help
help:
	@echo "Available targets:"
	@echo "  build   - Build the binary"
	@echo "  test    - Run all tests"
	@echo "  test-v  - Run tests with verbose output"
	@echo "  lint    - Run golangci-lint"
	@echo "  fmt     - Format code"
	@echo "  clean   - Remove build artifacts"
	@echo "  help    - Show this help message"
