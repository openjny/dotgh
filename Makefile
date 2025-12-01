.PHONY: build test lint fmt clean help

# Build binary
build:
	go build -o dotgh ./cmd/dotgh

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
