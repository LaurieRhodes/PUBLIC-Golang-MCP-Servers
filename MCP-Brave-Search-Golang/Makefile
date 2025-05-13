.PHONY: build test clean run fmt lint all release cross-platform help

# Build the application
build:
	@echo "Building brave-search-mcp..."
	go build -o brave-search-mcp ./cmd/server

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f brave-search-mcp brave-search-mcp.exe
	rm -f brave-search-mcp-*

# Run the application
run: build
	@echo "Running brave-search-mcp..."
	./brave-search-mcp

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Check for lint issues
lint:
	@echo "Linting code..."
	go vet ./...

# Cross-platform builds
cross-platform: clean
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -o brave-search-mcp-linux-amd64 ./cmd/server
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build -o brave-search-mcp-darwin-amd64 ./cmd/server
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build -o brave-search-mcp-windows-amd64.exe ./cmd/server
	@echo "Cross-platform builds complete."

# Create a release directory with binaries and documentation
release: cross-platform
	@echo "Creating release package..."
	mkdir -p release
	cp brave-search-mcp-* release/
	cp README.md LICENSE release/
	cp -r docs release/
	cp config.example.json release/
	@echo "Release package created in ./release directory."

# Default target
all: fmt lint test build

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  test           - Run tests"
	@echo "  clean          - Remove build artifacts"
	@echo "  run            - Build and run the application"
	@echo "  fmt            - Format the code"
	@echo "  lint           - Run linters"
	@echo "  cross-platform - Build for multiple platforms"
	@echo "  release        - Create a release package"
	@echo "  all            - Run fmt, lint, test, and build"
	@echo "  help           - Show this help message"
