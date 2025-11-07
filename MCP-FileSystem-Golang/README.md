# Secure Filesystem MCP Server

<div align="center">

![Model Context Protocol](https://img.shields.io/badge/MCP-Filesystem-blue)
![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-MIT-green)

</div>

## 🚀 Overview

This is a secure Model Context Protocol (MCP) server implementation that provides controlled filesystem access for AI models. It allows Large Language Models to securely read, write, and manipulate files within explicitly defined allowed directories.

> ⚠️ **IMPORTANT SECURITY NOTICE**: Compiled executables are intentionally not included in this repository. Users should always review the source code and compile it themselves to ensure security. Never run precompiled executables from untrusted sources when dealing with filesystem access.

This project is a Go implementation of the original [Filesystem MCP server](https://github.com/modelcontextprotocol/servers/tree/main/src/filesystem) from the Model Context Protocol project developed by Anthropic, **with additional editor tools** to match Claude's trained expectations.

## 📁 Repository

This project is part of the [PUBLIC-Golang-MCP-Servers](https://github.com/LaurieRhodes/PUBLIC-Golang-MCP-Servers) repository, which contains various MCP server implementations in Go.

## ✨ Features

- **Filesystem Tools**:
  - Read and write files securely
  - Create and list directories
  - Move/rename files and directories
  - Search for files matching patterns
  - Get detailed file metadata
- **Editor Tools** (NEW):
  - `str_replace`: Surgical string replacement with validation
  - `insert`: Insert text at specific line numbers
  - `undo_edit`: Rollback file changes with automatic backups

## 🔧 Editor Tools Extension

This server includes **editor tools** that align with Claude Sonnet's trained expectations. These tools address [issue #4027](https://github.com/cline/cline/issues/4027) where Claude attempts to use editing tools that don't exist in standard MCP Desktop environments.

**Why this matters**: Claude Sonnet has been trained with certain file editing primitives as part of Anthropic's computer use feature. When these aren't available, users experience failed tool calls and degraded workflows. These editor tools bridge that gap.



## 🧰 Available Tools

### Filesystem Tools

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

### Editor Tools

| Tool Name     | Description                                             |
| ------------- | ------------------------------------------------------- |
| `str_replace` | Replace exact string in file (must appear once)         |
| `insert`      | Insert text after specified line number                 |
| `undo_edit`   | Undo last edit to a file (automatic backup restoration) |

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

Note that unlike the Node.js version, allowed directories are specified in the `config.json` file **in the same directory as the compiled MCP server**, not as command-line arguments. This allows a modular portability of the MCP tooling between different GenAI tools rather than creating a dependency on a single tool like Claude Desktop.

In your MCP client configuration, set up the filesystem server like this:

#### Windows Example

```json
{
  "mcpServers": {
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
  "mcpServers": {
    "filesystem": {
      "command": "/home/username/path/to/filesystem-mcp",
      "args": []
    }
  }
}
```





## 📊 Implementation Details

This server is built with Go and follows the Model Context Protocol specifications:

- **Transport**: Uses stdio for communication (reading JSON-RPC messages from stdin and writing responses to stdout)
- **Modular Design**: Clean separation between MCP protocol handling, filesystem operations, and editor operations
- **Comprehensive Error Handling**: Detailed error messages for easier debugging
- **Automatic Backups**: Editor operations create timestamped backups before modifications

## 🔍 Tool Schema Examples



## 📚 Additional Documentation

- [GitHub Issue #4027](https://github.com/cline/cline/issues/4027) - Background on the Claude tool training issue

## 📜 License

This MCP server is licensed under the original Anthropic MIT License. This means you are free to use, modify, and distribute the software, subject to the terms and conditions of the MIT License. For more details, please see the LICENSE file in the project repository.

## 👏 Attribution

This project is a port of the original [Filesystem MCP server](https://github.com/modelcontextprotocol/servers/tree/main/src/filesystem) developed by Anthropic, PBC, which is part of the Model Context Protocol project. The original Node.js implementation is available at `@modelcontextprotocol/server-filesystem`.

The editor tools extension addresses community-identified gaps in Claude Desktop's MCP tool availability.

## 🤝 Contributing

This source code is provided as example code and not intended to become an active project. Feel free to fork and extend for your needs.
