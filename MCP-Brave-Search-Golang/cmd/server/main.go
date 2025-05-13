package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/LaurieRhodes/PUBLIC-Golang-MCP-Servers/MCP-Brave-Search-Golang/internal/ratelimit"
	"github.com/LaurieRhodes/PUBLIC-Golang-MCP-Servers/MCP-Brave-Search-Golang/pkg/brave"
	"github.com/LaurieRhodes/PUBLIC-Golang-MCP-Servers/MCP-Brave-Search-Golang/pkg/config"
)

// JSONRPCMessage represents a JSON-RPC message
type JSONRPCMessage struct {
	JsonRPC string          `json:"jsonrpc"`
	ID      string          `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ErrorMessage   `json:"error,omitempty"`
}

// ErrorMessage represents an error in a JSON-RPC message
type ErrorMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Main server state
var (
	initialized bool
	apiKey      string
	rateLimiter *ratelimit.RateLimiter
)

func main() {
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "Shutting down...")
		os.Exit(0)
	}()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Store configuration in global variables
	apiKey = cfg.BraveAPIKey
	rateLimiter = ratelimit.NewRateLimiter(ratelimit.RateLimits{
		PerSecond: cfg.RateLimit.PerSecond,
		PerMonth:  cfg.RateLimit.PerMonth,
	})

	// Start the server
	fmt.Fprintln(os.Stderr, "Brave Search MCP Server starting...")
	RunServer()
}

// RunServer starts the MCP server
func RunServer() {
	// Create scanner for stdin
	scanner := bufio.NewScanner(os.Stdin)
	// Create writer for stdout
	writer := bufio.NewWriter(os.Stdout)

	// Process requests
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue // Skip empty lines
		}

		fmt.Fprintf(os.Stderr, "Received: %s\n", line)

		// Parse the message
		var message JSONRPCMessage
		if err := json.Unmarshal([]byte(line), &message); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing message: %v\n", err)
			continue
		}

		// Process the message
		var responseMsg *JSONRPCMessage

		switch message.Method {
		case "initialize":
			responseMsg = handleInitialize(message)
		case "initialized":
			initialized = true
			continue // No response for notification
		case "tools/list":
			responseMsg = handleToolsList(message)
		case "tools/call":
			responseMsg = handleToolsCall(message)
		case "list_tools": // Backward compatibility
			responseMsg = handleToolsList(message)
		case "call_tool": // Backward compatibility
			responseMsg = handleToolsCall(message)
		default:
			responseMsg = &JSONRPCMessage{
				JsonRPC: "2.0",
				ID:      message.ID,
				Error: &ErrorMessage{
					Code:    -32601,
					Message: "Method not supported: " + message.Method,
				},
			}
		}

		// Send response if applicable
		if responseMsg != nil {
			responseBytes, err := json.Marshal(responseMsg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling response: %v\n", err)
				continue
			}

			fmt.Fprintf(os.Stderr, "Sending: %s\n", string(responseBytes))
			_, err = writer.WriteString(string(responseBytes) + "\n")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing response: %v\n", err)
				continue
			}
			err = writer.Flush()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error flushing response: %v\n", err)
				continue
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
	}
}

// handleInitialize handles the initialize request
func handleInitialize(message JSONRPCMessage) *JSONRPCMessage {
	// Parse the params
	var params map[string]interface{}
	if err := json.Unmarshal(message.Params, &params); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing initialize params: %v\n", err)
		return &JSONRPCMessage{
			JsonRPC: "2.0",
			ID:      message.ID,
			Error: &ErrorMessage{
				Code:    -32700,
				Message: "Parse error",
			},
		}
	}

	// Extract client info
	clientInfo := params["clientInfo"].(map[string]interface{})
	fmt.Fprintf(os.Stderr, "Client info: %s %s\n", clientInfo["name"], clientInfo["version"])

	// Get protocol version
	protocolVersion := params["protocolVersion"].(string)
	fmt.Fprintf(os.Stderr, "Protocol version: %s\n", protocolVersion)

	// Create server info
	serverInfo := map[string]interface{}{
		"name":    "brave-search-mcp",
		"version": "0.1.0",
	}

	// Create capabilities
	capabilities := map[string]interface{}{
		"tools": map[string]interface{}{
			"list": true,
			"call": true,
		},
	}

	// Create result
	result := map[string]interface{}{
		"protocolVersion": protocolVersion,
		"serverInfo":      serverInfo,
		"capabilities":    capabilities,
	}

	// Marshal result to JSON
	resultBytes, err := json.Marshal(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling result: %v\n", err)
		return &JSONRPCMessage{
			JsonRPC: "2.0",
			ID:      message.ID,
			Error: &ErrorMessage{
				Code:    -32603,
				Message: "Internal error",
			},
		}
	}

	// Set initialized flag
	initialized = true

	// Return response
	return &JSONRPCMessage{
		JsonRPC: "2.0",
		ID:      message.ID,
		Result:  resultBytes,
	}
}

// handleToolsList handles the tools/list request
func handleToolsList(message JSONRPCMessage) *JSONRPCMessage {
	// If not initialized, reject the request
	if !initialized {
		return &JSONRPCMessage{
			JsonRPC: "2.0",
			ID:      message.ID,
			Error: &ErrorMessage{
				Code:    -32002,
				Message: "Server not initialized",
			},
		}
	}

	// Create web search tool
	webSearchTool := map[string]interface{}{
		"name":        brave.WebSearchTool["name"],
		"description": brave.WebSearchTool["description"],
		"inputSchema": brave.WebSearchTool["inputSchema"],
	}

	// Create local search tool
	localSearchTool := map[string]interface{}{
		"name":        brave.LocalSearchTool["name"],
		"description": brave.LocalSearchTool["description"],
		"inputSchema": brave.LocalSearchTool["inputSchema"],
	}

	// Create tools list
	toolsList := map[string]interface{}{
		"tools": []interface{}{
			webSearchTool,
			localSearchTool,
		},
	}

	// Marshal result to JSON
	resultBytes, err := json.Marshal(toolsList)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling result: %v\n", err)
		return &JSONRPCMessage{
			JsonRPC: "2.0",
			ID:      message.ID,
			Error: &ErrorMessage{
				Code:    -32603,
				Message: "Internal error",
			},
		}
	}

	// Return response
	return &JSONRPCMessage{
		JsonRPC: "2.0",
		ID:      message.ID,
		Result:  resultBytes,
	}
}

// handleToolsCall handles the tools/call request
func handleToolsCall(message JSONRPCMessage) *JSONRPCMessage {
	// If not initialized, reject the request
	if !initialized {
		return &JSONRPCMessage{
			JsonRPC: "2.0",
			ID:      message.ID,
			Error: &ErrorMessage{
				Code:    -32002,
				Message: "Server not initialized",
			},
		}
	}

	// Parse the params
	var params map[string]interface{}
	if err := json.Unmarshal(message.Params, &params); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing call params: %v\n", err)
		return &JSONRPCMessage{
			JsonRPC: "2.0",
			ID:      message.ID,
			Error: &ErrorMessage{
				Code:    -32700,
				Message: "Parse error",
			},
		}
	}

	// Extract tool name
	toolName, ok := params["name"].(string)
	if !ok {
		return &JSONRPCMessage{
			JsonRPC: "2.0",
			ID:      message.ID,
			Error: &ErrorMessage{
				Code:    -32602,
				Message: "Invalid params: missing tool name",
			},
		}
	}

	// Extract arguments
	arguments, ok := params["arguments"]
	if !ok {
		return &JSONRPCMessage{
			JsonRPC: "2.0",
			ID:      message.ID,
			Error: &ErrorMessage{
				Code:    -32602,
				Message: "Invalid params: missing arguments",
			},
		}
	}

	// Marshal arguments to JSON
	argumentsBytes, err := json.Marshal(arguments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling arguments: %v\n", err)
		return &JSONRPCMessage{
			JsonRPC: "2.0",
			ID:      message.ID,
			Error: &ErrorMessage{
				Code:    -32603,
				Message: "Internal error",
			},
		}
	}

	// Process the tool call
	var response map[string]interface{}

	switch toolName {
	case "brave_web_search":
		// Parse web search arguments
		var args struct {
			Query  string `json:"query"`
			Count  int    `json:"count"`
			Offset int    `json:"offset"`
		}
		if err := json.Unmarshal(argumentsBytes, &args); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing web search arguments: %v\n", err)
			return &JSONRPCMessage{
				JsonRPC: "2.0",
				ID:      message.ID,
				Error: &ErrorMessage{
					Code:    -32602,
					Message: "Invalid params: " + err.Error(),
				},
			}
		}

		// Set default count if needed
		if args.Count <= 0 {
			args.Count = 10
		}

		// Perform web search
		results, err := brave.WebSearch(apiKey, args.Query, args.Count, args.Offset, rateLimiter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Web search error: %v\n", err)
			response = map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "Error: " + err.Error(),
					},
				},
				"isError": true,
			}
		} else {
			fmt.Fprintf(os.Stderr, "Web search success\n")
			response = map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": results,
					},
				},
				"isError": false,
			}
		}

	case "brave_local_search":
		// Parse local search arguments
		var args struct {
			Query string `json:"query"`
			Count int    `json:"count"`
		}
		if err := json.Unmarshal(argumentsBytes, &args); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing local search arguments: %v\n", err)
			return &JSONRPCMessage{
				JsonRPC: "2.0",
				ID:      message.ID,
				Error: &ErrorMessage{
					Code:    -32602,
					Message: "Invalid params: " + err.Error(),
				},
			}
		}

		// Set default count if needed
		if args.Count <= 0 {
			args.Count = 5
		}

		// Perform local search
		results, err := brave.LocalSearch(apiKey, args.Query, args.Count, rateLimiter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Local search error: %v\n", err)
			response = map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "Error: " + err.Error(),
					},
				},
				"isError": true,
			}
		} else {
			fmt.Fprintf(os.Stderr, "Local search success\n")
			response = map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": results,
					},
				},
				"isError": false,
			}
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown tool: %s\n", toolName)
		response = map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": "Unknown tool: " + toolName,
				},
			},
			"isError": true,
		}
	}

	// Marshal response to JSON
	resultBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling result: %v\n", err)
		return &JSONRPCMessage{
			JsonRPC: "2.0",
			ID:      message.ID,
			Error: &ErrorMessage{
				Code:    -32603,
				Message: "Internal error",
			},
		}
	}

	// Return response
	return &JSONRPCMessage{
		JsonRPC: "2.0",
		ID:      message.ID,
		Result:  resultBytes,
	}
}
