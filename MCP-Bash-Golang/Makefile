.PHONY: build clean test install

# Binary name
BINARY=bash-mcp
VERSION=1.0.0

# Build flags
LDFLAGS=-ldflags="-s -w"

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY) v$(VERSION)..."
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY) ./cmd/server
	@echo "✓ Build complete: $(BINARY)"

# Build for Linux (static)
build-linux:
	@echo "Building for Linux (static)..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-linux ./cmd/server
	@echo "✓ Linux build complete"

# Build for macOS
build-mac:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-mac ./cmd/server
	@echo "✓ macOS build complete"

# Build for Windows
build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY).exe ./cmd/server
	@echo "✓ Windows build complete"

# Build for all platforms
build-all: build-linux build-mac build-windows
	@echo "✓ All builds complete"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY) $(BINARY)-linux $(BINARY)-mac $(BINARY).exe
	@echo "✓ Clean complete"

# Run tests (placeholder)
test:
	@echo "Running tests..."
	go test ./...

# Install (copy to /usr/local/bin)
install: build
	@echo "Installing $(BINARY) to /usr/local/bin..."
	sudo cp $(BINARY) /usr/local/bin/
	@echo "✓ Installation complete"

# Create config if it doesn't exist
config:
	@if [ ! -f config.json ]; then \
		cp config.example.json config.json; \
		echo "✓ Created config.json from example"; \
	else \
		echo "config.json already exists"; \
	fi

# Setup: build and create config
setup: build config
	@echo "✓ Setup complete"
	@echo ""
	@echo "To use with Claude Desktop, add this to claude_desktop_config.json:"
	@echo '{'
	@echo '  "mcpServers": {'
	@echo '    "bash": {'
	@echo '      "command": "'$(shell pwd)/$(BINARY)'"'
	@echo '    }'
	@echo '  }'
	@echo '}'

# Help
help:
	@echo "Bash MCP Server - Makefile targets:"
	@echo ""
	@echo "  make build         - Build for current platform"
	@echo "  make build-linux   - Build for Linux"
	@echo "  make build-mac     - Build for macOS"
	@echo "  make build-windows - Build for Windows"
	@echo "  make build-all     - Build for all platforms"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make test          - Run tests"
	@echo "  make install       - Install to /usr/local/bin"
	@echo "  make config        - Create config.json from example"
	@echo "  make setup         - Build and create config"
	@echo "  make help          - Show this help"
