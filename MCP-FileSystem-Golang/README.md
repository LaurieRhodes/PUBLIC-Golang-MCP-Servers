# Secure Filesystem MCP Server

<div align="center">

![Model Context Protocol](https://img.shields.io/badge/MCP-Filesystem-blue)
![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-MIT-green)

</div>

## 🚀 Overview

This is a secure Model Context Protocol (MCP) server implementation that provides controlled filesystem access for AI models. It allows Large Language Models to securely read, write, and manipulate files within explicitly defined allowed directories.

> ⚠️ **IMPORTANT SECURITY NOTICE**: Compiled executables are intentionally not included in this repository. Users should always review the source code and compile it themselves to ensure security. Never run precompiled executables from untrusted sources when dealing with filesystem access.

This project is a Go implementation of the original [Filesystem MCP server](https://github.com/modelcontextprotocol/servers/tree/main/src/filesystem) from the Model Context Protocol project developed by Anthropic.

## 📁 Repository

This project is part of the [PUBLIC-Golang-MCP-Servers](https://github.com/LaurieRhodes/PUBLIC-Golang-MCP-Servers) repository, which contains various MCP server implementations in Go.

## ✨ Features

- **Security-First Design**: Comprehensive path validation and containment
- **Multiple Directory Support**: Configure multiple allowed directories
- **Comprehensive Tools**:
  - Read and write files securely
  - Create and list directories
  - Move/rename files and directories
  - Search for files matching patterns
  - Get detailed file metadata
- **Robust Error Handling**: Detailed error messages for troubleshooting
- **Configuration**: Simple JSON-based configuration

## 🔒 Security

Security is a primary focus of this implementation:

- **Strict Path Validation**: Prevents access outside allowed directories
- **Symlink Protection**: Ensures symlinks don't lead outside allowed directories
- **Path Normalization**: Consistent security checks to prevent path traversal attacks
- **Parent Directory Validation**: Verifies parent directories of files being created

## 🧰 Available Tools

| Tool Name                  | Description                          |
| -------------------------- | ------------------------------------ |
| `read_file`                | Read the complete contents of a file |
| `read_multiple_files`      | Read multiple files at once          |
| `write_file`               | Create or overwrite a file           |
| `create_directory`         | Create a new directory               |
| `list_directory`           | List contents of a directory         |
| `move_file`                | Move or rename files and directories |
| `search_files`             | Search for files matching a pattern  |
| `get_file_info`            | Get metadata about a file            |
| `list_allowed_directories` | List all allowed directories         |

## ⚙️ Configuration

The server uses a `config.json` file which should be placed in the same directory as the executable or in the current working directory:

```json
{
  "allowedDirectories": [
    "C:\\Users\\Username\\AppData\\Roaming\\Claude",
    "C:\\Users\\Username\\Documents",
    "D:\\Projects"
  ]
}
```

If the `config.json` file doesn't exist, a default one will be created with the current directory as the allowed directory.

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- Basic understanding of MCP architecture

### Building from Source

#### Standard Build (Windows/macOS)

```bash
# Clone the repository
git clone https://github.com/LaurieRhodes/PUBLIC-Golang-MCP-Servers.git
cd PUBLIC-Golang-MCP-Servers/MCP-FileSystem-Golang

# Build the server
go build -o filesystem-mcp ./cmd/server

# Run the server (will use config.json)
./filesystem-mcp
```

#### Static Build for Linux (Recommended)

On Linux systems, it's recommended to build with static linking to avoid shared library dependency issues (e.g., `libgo.so.23` errors):

```bash
# Clone the repository
git clone https://github.com/LaurieRhodes/PUBLIC-Golang-MCP-Servers.git
cd PUBLIC-Golang-MCP-Servers/MCP-FileSystem-Golang

# Build with static linking (no external dependencies)
CGO_ENABLED=0 go build -o filesystem-mcp -ldflags="-s -w" ./cmd/server

# Verify static linking (should show "not a dynamic executable")
ldd filesystem-mcp

# Make executable and run
chmod +x filesystem-mcp
./filesystem-mcp
```

**Why static linking?** When you compile Go programs on Linux with dynamic linking, they depend on specific versions of shared libraries (like `libgo.so.23`). Static linking produces a self-contained binary that runs on any Linux system without requiring these libraries to be installed.

### MCP Client Configuration

In your MCP client configuration, set up the filesystem server like this:

#### Windows Example
```json
{
  "servers": {
    "filesystem": {
      "command": "C:\\path\\to\\filesystem-mcp.exe",
      "args": []
    }
  }
}
```

#### Linux Example
```json
{
  "servers": {
    "filesystem": {
      "command": "/home/username/path/to/filesystem-mcp",
      "args": []
    }
  }
}
```

Note that unlike the Node.js version, allowed directories are specified in the `config.json` file, not as command-line arguments.

## 📊 Implementation Details

This server is built with Go and follows the Model Context Protocol specifications:

- **Transport**: Uses stdio for communication (reading JSON-RPC messages from stdin and writing responses to stdout)
- **Modular Design**: Clean separation between MCP protocol handling and filesystem operations
- **Comprehensive Error Handling**: Detailed error messages for easier debugging

## 🔍 Tool Schema Examples

### read_file

```json
{
  "type": "object",
  "properties": {
    "path": {
      "type": "string"
    }
  },
  "required": ["path"]
}
```

### write_file

```json
{
  "type": "object",
  "properties": {
    "path": {
      "type": "string"
    },
    "content": {
      "type": "string"
    }
  },
  "required": ["path", "content"]
}
```

See the code for full schema definitions of all tools.

## 📜 License

This MCP server is licensed under the original MIT License. This means you are free to use, modify, and distribute the software, subject to the terms and conditions of the MIT License. For more details, please see the LICENSE file in the project repository.

## 👏 Attribution

This project is a port of the original [Filesystem MCP server](https://github.com/modelcontextprotocol/servers/tree/main/src/filesystem) developed by Anthropic, PBC, which is part of the Model Context Protocol project. The original Node.js implementation is available at `@modelcontextprotocol/server-filesystem`.

## 🤝 Contributing

This source code is provided as example code and not intended to become an active project.
