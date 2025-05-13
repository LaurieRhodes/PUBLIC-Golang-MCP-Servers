# Brave Search MCP Server (Golang)

<div align="center">

![Model Context Protocol](https://img.shields.io/badge/MCP-Brave_Search-blue)
![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-MIT-green)

</div>

A Go implementation of a Model Context Protocol (MCP) server that integrates with the Brave Search API, providing both web and local search capabilities.

> ⚠️ **IMPORTANT SECURITY NOTICE**: Compiled executables are intentionally not included in this repository. Users should always review the source code and compile it themselves to ensure security. Never run precompiled executables from untrusted sources when dealing with API keys and sensitive information.

## 📁 Repository

This project is part of the [PUBLIC-Golang-MCP-Servers](https://github.com/LaurieRhodes/PUBLIC-Golang-MCP-Servers) repository, which contains various MCP server implementations in Go.

## ✨ Features

- **Web Search**: General queries, news, articles, with pagination and freshness controls
- **Local Search**: Find businesses, restaurants, and services with detailed information
- **MCP Protocol Support**: Full compliance with Model Context Protocol for AI assistant integration
- **Configuration File**: Simple JSON configuration for API keys and settings
- **Flexible Deployment**: Can be used as a standalone CLI tool or integrated with Claude Desktop

## 🧰 MCP Tools Provided

### brave_web_search

Executes web searches with pagination and filtering.

**Inputs:**

- `query` (string): Search terms
- `count` (number, optional): Results per page (max 20, default 10)
- `offset` (number, optional): Pagination offset (max 9, default 0)

### brave_local_search

Searches for local businesses and services.

**Inputs:**

- `query` (string): Local search terms
- `count` (number, optional): Number of results (max 20, default 5)

Automatically falls back to web search if no local results found.

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher (for building from source)
- A Brave Search API key

### Building from Source

```bash
# Clone the repository
git clone https://github.com/LaurieRhodes/PUBLIC-Golang-MCP-Servers.git
cd PUBLIC-Golang-MCP-Servers/MCP-Brave-Search-Golang

# Build the server
go build -o brave-search-mcp.exe ./cmd/server

# Run the server (will use config.json)
./brave-search-mcp
```

### Configuration

When you first run the application, it will create a `config.json` file in the same directory as the executable. Edit this file to add your Brave Search API key:

```json
{
  "braveApiKey": "YOUR_API_KEY_HERE",
  "rateLimit": {
    "perSecond": 1,
    "perMonth": 15000
  }
}
```

#### Getting an API Key

1. Sign up for a [Brave Search API account](https://brave.com/search/api/)
2. Choose a plan (Free tier available with 2,000 queries/month)
3. Generate your API key [from the developer dashboard](https://api.search.brave.com/app/keys)

### Usage

After configuration, simply run the executable:

```bash
./brave-search-mcp
```

The server will start and listen on stdin/stdout for MCP protocol messages.

### Usage with Claude Desktop

Add this to your Claude Desktop configuration:

```json
{
  "mcpServers": {
    "brave-search": {
      "command": "C:\\path\\to\\brave-search-mcp",
      "args": []
    }
  }
}
```

Replace `C:\\path\\to\\brave-search-mcp` with the actual path to the executable file.

## 📊 MCP Protocol Implementation

This server implements the Model Context Protocol (MCP), which enables AI assistants to access tools and resources through a standardized interface. Key components of the implementation include:

- **Initialization Handshake**: Proper protocol handshake for establishing connections
- **Tool Discovery**: Support for `tools/list` method to discover available tools
- **Tool Execution**: Support for `tools/call` method to execute tools
- **JSON-RPC 2.0**: Compliant with JSON-RPC 2.0 message format

## 📂 Project Structure

```
brave-search-mcp/
├── cmd/
│   └── server/            # Application entry point
│       └── main.go
├── pkg/
│   ├── brave/             # Brave API client implementation
│   │   ├── local_search.go
│   │   └── web_search.go
│   └── config/            # Configuration handling
│       └── config.go
├── internal/
│   └── ratelimit/         # Rate limiting implementation
│       └── ratelimit.go
├── go.mod                 # Go module definition
├── config.example.json    # Example configuration
├── Makefile               # Build automation
├── README.md              # Project documentation
└── LICENSE                # MIT License
```

## 🔧 Troubleshooting

- **API Key Issues**: If you see authentication errors, make sure your API key is correct in the config.json file.
- **Rate Limiting**: The Brave Search API has rate limits. The server includes built-in rate limiting to help avoid exceeding these limits.
- **Gzip Compression**: The server handles gzip-compressed responses from the Brave API automatically.

## 📜 License

This MCP server is licensed under the original project MIT License. See the LICENSE file for details.

## 👏 Attribution

This project is a Go implementation of the original [Brave Search MCP server](https://github.com/modelcontextprotocol/servers/tree/main/src/brave-search) developed by Anthropic, PBC. The original server is part of the Model Context Protocol project and is available as an NPM package: `@modelcontextprotocol/server-brave-search`.

The implementation follows the same API endpoints and capabilities as the original server, but is written in Go instead of TypeScript/JavaScript.

## 🤝 Contributing

This source code is provided as example code and not intended to become an active project.
