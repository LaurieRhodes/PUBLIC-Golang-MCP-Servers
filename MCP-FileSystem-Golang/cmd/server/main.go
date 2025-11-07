package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/LaurieRhodes/PUBLIC-Golang-MCP-Servers/MCP-FileSystem-Golang/pkg/config"
	"github.com/LaurieRhodes/PUBLIC-Golang-MCP-Servers/MCP-FileSystem-Golang/pkg/editor"
	"github.com/LaurieRhodes/PUBLIC-Golang-MCP-Servers/MCP-FileSystem-Golang/pkg/filesystem"
	"github.com/LaurieRhodes/PUBLIC-Golang-MCP-Servers/MCP-FileSystem-Golang/pkg/mcp"
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

	// Create the file manager with allowed directories from config
	fileManager := filesystem.NewFileManager(cfg.AllowedDirectories)

	// Create the edit manager for undo functionality
	backupDir := filepath.Join(os.TempDir(), "mcp-filesystem-backups")
	editManager, err := editor.NewEditManager(backupDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating edit manager: %v\n", err)
		os.Exit(1)
	}

	// Create and configure the MCP server
	server := mcp.NewServer(
		mcp.ServerInfo{
			Name:    "secure-filesystem-server",
			Version: "0.3.0",
		},
		mcp.ServerConfig{
			Capabilities: mcp.ServerCapabilities{
				Tools: map[string]interface{}{
					"list": true,
					"call": true,
				},
			},
		},
	)

	// Set up handlers
	setupServerHandlers(server, fileManager, editManager)

	// Start the server with stdio transport
	transport := mcp.NewStdioTransport()
	fmt.Fprintf(os.Stderr, "Secure MCP Filesystem Server v0.3.0 starting on stdin/stdout\n")
	fmt.Fprintf(os.Stderr, "Allowed directories: %v\n", cfg.AllowedDirectories)
	fmt.Fprintf(os.Stderr, "Edit backup directory: %s\n", backupDir)
	
	err = server.Connect(transport)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}

	// The server is now running and processing requests via the transport
	// It will continue running until stdin is closed or the process is terminated
	select {} // Wait forever
}

// setupServerHandlers sets up the request handlers for the server
func setupServerHandlers(server *mcp.Server, fileManager *filesystem.FileManager, editManager *editor.EditManager) {
	// Handler for tools/list
	server.SetRequestHandler("tools/list", func(params json.RawMessage) (json.RawMessage, error) {
		// Combine filesystem and editor tools
		allTools := make([]mcp.Tool, 0, len(filesystem.FilesystemTools)+len(editor.EditorTools))
		
		// Add filesystem tools
		for _, toolDef := range filesystem.FilesystemTools {
			inputSchema, err := json.Marshal(toolDef.InputSchema)
			if err != nil {
				continue
			}
			
			allTools = append(allTools, mcp.Tool{
				Name:        toolDef.Name,
				Description: toolDef.Description,
				InputSchema: inputSchema,
			})
		}
		
		// Add editor tools
		for _, toolDef := range editor.EditorTools {
			inputSchema, err := json.Marshal(toolDef.InputSchema)
			if err != nil {
				continue
			}
			
			allTools = append(allTools, mcp.Tool{
				Name:        toolDef.Name,
				Description: toolDef.Description,
				InputSchema: inputSchema,
			})
		}
		
		response := mcp.ListToolsResponse{
			Tools: allTools,
		}
		
		return json.Marshal(response)
	})

	// Handler for list_tools (backward compatibility)
	server.SetRequestHandler("list_tools", func(params json.RawMessage) (json.RawMessage, error) {
		handler := server.GetHandler("tools/list")
		return handler(params)
	})
	
	// Handler for tools/call
	server.SetRequestHandler("tools/call", func(params json.RawMessage) (json.RawMessage, error) {
		var request mcp.CallToolRequest
		if err := json.Unmarshal(params, &request); err != nil {
			return nil, fmt.Errorf("invalid call parameters: %w", err)
		}
		
		// Process the tool call
		return handleToolCall(request, fileManager, editManager)
	})

	// Handler for call_tool (backward compatibility)
	server.SetRequestHandler("call_tool", func(params json.RawMessage) (json.RawMessage, error) {
		handler := server.GetHandler("tools/call")
		return handler(params)
	})
}

// handleToolCall handles a tool call request
func handleToolCall(request mcp.CallToolRequest, fileManager *filesystem.FileManager, editManager *editor.EditManager) (json.RawMessage, error) {
	var response mcp.CallToolResponse
	
	// Process based on tool name
	switch request.Name {
	// Filesystem tools
	case "read_file":
		path, err := filesystem.ParseReadFileArgs(request.Arguments)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		content, err := fileManager.ReadFile(path)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		response = mcp.CallToolResponse{
			Content: []mcp.ContentItem{
				{Type: "text", Text: content},
			},
		}
	
	case "read_multiple_files":
		paths, err := filesystem.ParseReadMultipleFilesArgs(request.Arguments)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		content, err := fileManager.ReadMultipleFiles(paths)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		response = mcp.CallToolResponse{
			Content: []mcp.ContentItem{
				{Type: "text", Text: content},
			},
		}
	
	case "write_file":
		path, content, err := filesystem.ParseWriteFileArgs(request.Arguments)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		err = fileManager.WriteFile(path, content)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		response = mcp.CallToolResponse{
			Content: []mcp.ContentItem{
				{Type: "text", Text: fmt.Sprintf("Successfully wrote to %s", path)},
			},
		}
	
	case "create_directory":
		path, err := filesystem.ParseCreateDirectoryArgs(request.Arguments)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		err = fileManager.CreateDirectory(path)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		response = mcp.CallToolResponse{
			Content: []mcp.ContentItem{
				{Type: "text", Text: fmt.Sprintf("Successfully created directory %s", path)},
			},
		}
	
	case "list_directory":
		path, err := filesystem.ParseListDirectoryArgs(request.Arguments)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		listing, err := fileManager.ListDirectory(path)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		response = mcp.CallToolResponse{
			Content: []mcp.ContentItem{
				{Type: "text", Text: listing},
			},
		}
	
	case "move_file":
		source, destination, err := filesystem.ParseMoveFileArgs(request.Arguments)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		err = fileManager.MoveFile(source, destination)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		response = mcp.CallToolResponse{
			Content: []mcp.ContentItem{
				{Type: "text", Text: fmt.Sprintf("Successfully moved %s to %s", source, destination)},
			},
		}
	
	case "search_files":
		path, pattern, err := filesystem.ParseSearchFilesArgs(request.Arguments)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		results, err := filesystem.SearchFiles(fileManager, path, pattern)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		var resultText string
		if len(results) > 0 {
			resultText = fmt.Sprintf("%d matches found:\n%s", len(results), strings.Join(results, "\n"))
		} else {
			resultText = "No matches found"
		}
		
		response = mcp.CallToolResponse{
			Content: []mcp.ContentItem{
				{Type: "text", Text: resultText},
			},
		}
	
	case "get_file_info":
		path, err := filesystem.ParseGetFileInfoArgs(request.Arguments)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		info, err := fileManager.GetFileInfo(path)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		response = mcp.CallToolResponse{
			Content: []mcp.ContentItem{
				{Type: "text", Text: info},
			},
		}
	
	case "list_allowed_directories":
		response = mcp.CallToolResponse{
			Content: []mcp.ContentItem{
				{Type: "text", Text: fileManager.ListAllowedDirectories()},
			},
		}
	
	// Editor tools
	case "str_replace":
		path, oldStr, newStr, err := editor.ParseStrReplaceArgs(request.Arguments)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		// Validate path first
		validPath, err := fileManager.ValidatePath(path)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		err = editManager.StrReplace(validPath, oldStr, newStr)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		response = mcp.CallToolResponse{
			Content: []mcp.ContentItem{
				{Type: "text", Text: fmt.Sprintf("Successfully replaced text in %s", path)},
			},
		}
	
	case "insert":
		path, lineNumber, text, err := editor.ParseInsertArgs(request.Arguments)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		// Validate path first
		validPath, err := fileManager.ValidatePath(path)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		err = editManager.Insert(validPath, lineNumber, text)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		response = mcp.CallToolResponse{
			Content: []mcp.ContentItem{
				{Type: "text", Text: fmt.Sprintf("Successfully inserted text at line %d in %s", lineNumber, path)},
			},
		}
	
	case "undo_edit":
		path, err := editor.ParseUndoEditArgs(request.Arguments)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		// Validate path first
		validPath, err := fileManager.ValidatePath(path)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		err = editManager.UndoEdit(validPath)
		if err != nil {
			return createErrorResponse(err.Error())
		}
		
		response = mcp.CallToolResponse{
			Content: []mcp.ContentItem{
				{Type: "text", Text: fmt.Sprintf("Successfully undid last edit to %s", path)},
			},
		}
	
	default:
		return createErrorResponse(fmt.Sprintf("Unknown tool: %s", request.Name))
	}
	
	return json.Marshal(response)
}

// createErrorResponse creates an error response for a tool call
func createErrorResponse(message string) (json.RawMessage, error) {
	response := mcp.CallToolResponse{
		Content: []mcp.ContentItem{
			{Type: "text", Text: fmt.Sprintf("Error: %s", message)},
		},
		IsError: true,
	}
	
	return json.Marshal(response)
}
