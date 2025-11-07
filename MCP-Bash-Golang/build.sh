#!/bin/bash

# Build script for Bash MCP Server

set -e

echo "=================================="
echo "Bash MCP Server Builder"
echo "Version: 1.0.0"
echo "=================================="
echo

# Detect OS
OS=$(uname -s)
ARCH=$(uname -m)

echo "Detected: $OS $ARCH"
echo

# Set output filename based on OS
if [[ "$OS" == "Linux" ]]; then
    OUTPUT="bash-mcp"
    BUILD_TYPE="static"
    echo "Building statically-linked binary for Linux..."
    CGO_ENABLED=0 go build -o "$OUTPUT" -ldflags="-s -w" ./cmd/server
    
    echo
    echo "Verifying static linking..."
    ldd "$OUTPUT" 2>&1 | grep -q "not a dynamic executable" && echo "✓ Binary is statically linked" || echo "⚠ Binary may have dynamic dependencies"
    
elif [[ "$OS" == "Darwin" ]]; then
    OUTPUT="bash-mcp"
    BUILD_TYPE="standard"
    echo "Building for macOS..."
    go build -o "$OUTPUT" ./cmd/server
    
elif [[ "$OS" =~ "MINGW" ]] || [[ "$OS" =~ "MSYS" ]] || [[ "$OS" =~ "CYGWIN" ]]; then
    OUTPUT="bash-mcp.exe"
    BUILD_TYPE="standard"
    echo "Building for Windows..."
    go build -o "$OUTPUT" ./cmd/server
    
else
    echo "Unknown OS: $OS"
    echo "Attempting standard build..."
    OUTPUT="bash-mcp"
    BUILD_TYPE="standard"
    go build -o "$OUTPUT" ./cmd/server
fi

# Make executable (Linux/macOS)
if [[ "$OS" == "Linux" ]] || [[ "$OS" == "Darwin" ]]; then
    chmod +x "$OUTPUT"
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
echo "2. Edit config.json if needed (default: 120s timeout, enabled)"
echo "3. Configure Claude Desktop to use: $(pwd)/$OUTPUT"
echo
echo "Example Claude Desktop config.json entry:"
echo '{'
echo '  "mcpServers": {'
echo '    "bash": {'
echo "      \"command\": \"$(pwd)/$OUTPUT\""
echo '    }'
echo '  }'
echo '}'
echo
echo "⚠️  WARNING: This tool executes arbitrary bash commands."
echo "   Use only in trusted environments!"
echo
