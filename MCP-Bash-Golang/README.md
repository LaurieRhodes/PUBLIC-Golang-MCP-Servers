# Bash MCP Server

<div align="center">

![Model Context Protocol](https://img.shields.io/badge/MCP-Bash-blue)
![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-MIT-green)

</div>

## 🚀 Overview

This is a Model Context Protocol (MCP) server implementation that provides bash command execution for AI models. It allows Large Language Models to execute bash commands in a persistent, stateful session.

> ⚠️ **IMPORTANT SECURITY NOTICE**: This server executes arbitrary bash commands on your system. Only use it in trusted environments and with AI models you trust. Review all commands before execution in production environments.

This server addresses the issue where Claude Sonnet models have been trained to expect a `bash_tool` as part of Anthropic's computer use feature, but this tool is not available in standard Claude Desktop MCP environments.

## 📁 Repository

This project is part of the [PUBLIC-Golang-MCP-Servers](https://github.com/LaurieRhodes/PUBLIC-Golang-MCP-Servers) repository, which contains various MCP server implementations in Go.



## 🔒 Security Considerations

**WARNING**: This server executes bash commands on your system. Security considerations:

- ✅ Commands run with the same permissions as the server process
- ✅ No sudo or root access by default
- ✅ Configurable timeout prevents infinite loops
- ✅ Session can be restarted if compromised
- ❌ No command whitelisting (commands are not filtered)
- ❌ No sandboxing (commands have full system access within user permissions)

**Best Practices**:

- Run the server with minimal necessary permissions
- Use in development/testing environments only
- Review AI-generated commands before production use
- Consider using Docker or VMs for additional isolation
- Monitor server logs for unexpected commands

## 🧰 Available Tool

### bash

Execute bash commands in a persistent session.

| Parameter | Required | Type    | Description                              |
| --------- | -------- | ------- | ---------------------------------------- |
| command   | Yes      | string  | The bash command to execute              |
| restart   | No       | boolean | Set to true to restart the session first |

**Supported Features**:

- Pipelines: `ls | grep pattern`
- Command chaining: `cd /tmp && ls -la`
- Environment variables: `export VAR=value`
- Background processes: `sleep 10 &`
- File I/O redirection: `echo "text" > file.txt`
- Command substitution: `echo $(date)`
- Conditional execution: `test -f file && cat file`

**Unsupported Features**:

- Interactive commands: `vim`, `less`, `top`
- Commands requiring user input: password prompts
- `sudo` without NOPASSWD configuration

## ⚙️ Configuration

The server uses a `config.json` file in the same directory as the MCP server:

```json
{
  "commandTimeout": 120,
  "enabled": true
}
```

**Configuration Options**:

- `commandTimeout`: Maximum execution time in seconds (default: 120)
- `enabled`: Set to false to disable the bash tool entirely

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- Bash shell available on the system
- Basic understanding of MCP architecture

### Building from Source

#### Linux (Static Build - Recommended)

```bash
cd MCP-Bash-Golang
CGO_ENABLED=0 go build -o bash-mcp -ldflags="-s -w" ./cmd/server
chmod +x bash-mcp
```

#### macOS/Windows

```bash
cd MCP-Bash-Golang
go build -o bash-mcp ./cmd/server
```

### MCP Client Configuration

#### Claude Desktop Configuration

Edit `~/.config/Claude/claude_desktop_config.json` (Linux/macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "bash": {
      "command": "/full/path/to/bash-mcp",
      "args": []
    }
  }
}
```

## 📊 Implementation Details

### Architecture

```
bash-mcp
├── cmd/server/
│   └── main.go           # Server entry point and handlers
├── pkg/
│   ├── bash/
│   │   └── bash.go       # Bash session management
│   ├── mcp/
│   │   ├── types.go      # MCP type definitions
│   │   ├── server.go     # MCP server implementation
│   │   └── transport.go  # stdio transport
│   └── config/
│       └── config.go     # Configuration management
```

### Session Persistence

The bash session maintains state between commands

## 📚 Related Documentation

- [Anthropic Bash Tool Documentation](https://docs.claude.com/en/docs/agents-and-tools/tool-use/bash-tool)
- [Model Context Protocol](https://modelcontextprotocol.io/)
- [GitHub Issue #4027](https://github.com/cline/cline/issues/4027) - Background on Claude tool training

# 

## 📜 License

MIT License (same as other MCP servers in this repository)

## 👏 Attribution

This server addresses community-identified gaps in Claude Desktop's MCP tool availability, specifically the missing `bash_tool` that Claude models have been trained to use.

## ⚠️ Disclaimer

This tool executes arbitrary bash commands. Use responsibly and only in trusted environments. The authors are not responsible for any damage caused by misuse of this tool.

## 🤝 Contributing

This source code is provided as example code. Feel free to fork and extend for your needs.

---

**Version**: 1.0.0  
**Status**: Production Ready  
**Platform**: Linux, macOS (Windows with WSL/bash)
