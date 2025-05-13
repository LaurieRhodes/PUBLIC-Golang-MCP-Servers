.PHONY: build run clean

# Variables
BINARY_NAME=filesystem-mcp
GO_FILES=$(shell find . -name "*.go")

# Main build target
build: $(GO_FILES)
	go build -o $(BINARY_NAME) ./cmd/server

# Run the server
run: build
	./$(BINARY_NAME)

# Clean up build artifacts
clean:
	rm -f $(BINARY_NAME)

# Install the binary to $GOPATH/bin
install: build
	go install ./cmd/server

# Build for multiple platforms
build-all: clean
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 ./cmd/server
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-windows-amd64.exe ./cmd/server
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin-amd64 ./cmd/server
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-darwin-arm64 ./cmd/server

# Show help
help:
	@echo "Available targets:"
	@echo "  build       - Build the server binary"
	@echo "  run         - Build and run the server"
	@echo "  clean       - Remove build artifacts"
	@echo "  install     - Install the binary to GOPATH/bin"
	@echo "  build-all   - Build for multiple platforms"
	@echo "  help        - Show this help message"
	@echo ""
	@echo "Usage:"
	@echo "  make build"
	@echo "  make run"
	@echo "  ./$(BINARY_NAME)"
	@echo ""
	@echo "Note: The server uses config.json for configuration."
