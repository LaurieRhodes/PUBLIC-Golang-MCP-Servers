#!/bin/bash

# Build script for MCP Filesystem Server with Editor Tools
# This script builds the server for the current platform

set -e  # Exit on error

echo "=================================="
echo "MCP Filesystem Server Builder"
echo "Version: 0.3.0 (with Editor Tools)"
echo "=================================="
echo

# Detect OS
OS=$(uname -s)
ARCH=$(uname -m)

echo "Detected: $OS $ARCH"
echo

# Set output filename based on OS
if [[ "$OS" == "Linux" ]]; then
    OUTPUT="filesystem-mcp"
    BUILD_TYPE="static"
    echo "Building statically-linked binary for Linux..."
    CGO_ENABLED=0 go build -o "$OUTPUT" -ldflags="-s -w" ./cmd/server
    
    echo
    echo "Verifying static linking..."
    ldd "$OUTPUT" 2>&1 | grep -q "not a dynamic executable" && echo "✓ Binary is statically linked" || echo "⚠ Binary may have dynamic dependencies"
    
elif [[ "$OS" == "Darwin" ]]; then
    OUTPUT="filesystem-mcp"
    BUILD_TYPE="standard"
    echo "Building for macOS..."
    go build -o "$OUTPUT" ./cmd/server
    
elif [[ "$OS" =~ "MINGW" ]] || [[ "$OS" =~ "MSYS" ]] || [[ "$OS" =~ "CYGWIN" ]]; then
    OUTPUT="filesystem-mcp.exe"
    BUILD_TYPE="standard"
    echo "Building for Windows..."
    go build -o "$OUTPUT" ./cmd/server
    
else
    echo "Unknown OS: $OS"
    echo "Attempting standard build..."
    OUTPUT="filesystem-mcp"
    BUILD_TYPE="standard"
    go build -o "$OUTPUT" ./cmd/server
fi

# Make executable (Linux/macOS)
if [[ "$OS" != "Linux" ]] && [[ "$OS" != "Darwin" ]]; then
    chmod +x "$OUTPUT" 2>/dev/null || true
fi

echo
echo "=================================="
echo "✓ Build successful!"
echo "=================================="
echo "Output: $OUTPUT"
echo "Build type: $BUILD_TYPE"
echo
echo "Next steps:"
echo "1. Copy config.example.json to config.json"
echo "2. Edit config.json with your allowed directories"
echo "3. Configure Claude Desktop to use: $(pwd)/$OUTPUT"
echo
echo "For testing, run:"
echo "  go test ./... -v"
echo
echo "See README.md and EDITOR_TOOLS.md for more information"
echo
