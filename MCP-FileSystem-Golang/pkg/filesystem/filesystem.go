package filesystem

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileInfo represents metadata about a file
type FileInfo struct {
	Size        int64     `json:"size"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
	Accessed    time.Time `json:"accessed"`
	IsDirectory bool      `json:"isDirectory"`
	IsFile      bool      `json:"isFile"`
	Permissions string    `json:"permissions"`
}

// FileManager handles filesystem operations with security checks
type FileManager struct {
	allowedDirectories []string
}

// NewFileManager creates a new FileManager with the given allowed directories
func NewFileManager(allowedDirs []string) *FileManager {
	// Normalize all paths consistently
	normalizedDirs := make([]string, len(allowedDirs))
	for i, dir := range allowedDirs {
		normalizedDirs[i] = normalizePath(filepath.Clean(dir))
	}

	return &FileManager{
		allowedDirectories: normalizedDirs,
	}
}

// normalizePath normalizes a path for secure comparison
func normalizePath(path string) string {
	return strings.ToLower(filepath.Clean(path))
}

// expandHomePath expands a ~ prefix to the user's home directory
func expandHomePath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("couldn't get home directory: %w", err)
	}
	
	if path == "~" {
		return home, nil
	}
	
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:]), nil
	}
	
	return path, nil
}

// ValidatePath checks if a path is allowed and returns its absolute path
func (fm *FileManager) ValidatePath(requestedPath string) (string, error) {
	// Expand home path if needed
	expandedPath, err := expandHomePath(requestedPath)
	if err != nil {
		return "", err
	}

	// Get absolute path
	absolute := requestedPath
	if !filepath.IsAbs(expandedPath) {
		absolute = filepath.Join(filepath.Clean(expandedPath))
	} else {
		absolute = filepath.Clean(expandedPath)
	}

	// Check if path is within allowed directories
	normalizedRequested := normalizePath(absolute)
	isAllowed := false

	for _, dir := range fm.allowedDirectories {
		if strings.HasPrefix(normalizedRequested, dir) {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return "", fmt.Errorf("access denied - path outside allowed directories: %s", absolute)
	}

	// Handle symlinks by checking their real path
	realPath, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		// For new files that don't exist yet, verify parent directory
		parentDir := filepath.Dir(absolute)
		
		// Check if parent directory exists
		_, parentErr := os.Stat(parentDir)
		if parentErr != nil {
			return "", fmt.Errorf("parent directory does not exist: %s", parentDir)
		}
		
		// Try to get real path of parent
		realParentPath, parentErr := filepath.EvalSymlinks(parentDir)
		if parentErr != nil {
			return "", fmt.Errorf("error checking parent directory: %w", parentErr)
		}
		
		// Verify parent is in allowed directories
		normalizedParent := normalizePath(realParentPath)
		parentAllowed := false
		
		for _, dir := range fm.allowedDirectories {
			if strings.HasPrefix(normalizedParent, dir) {
				parentAllowed = true
				break
			}
		}
		
		if !parentAllowed {
			return "", fmt.Errorf("access denied - parent directory outside allowed directories")
		}
		
		return absolute, nil
	}

	// Verify the real path is also allowed
	normalizedReal := normalizePath(realPath)
	realPathAllowed := false
	
	for _, dir := range fm.allowedDirectories {
		if strings.HasPrefix(normalizedReal, dir) {
			realPathAllowed = true
			break
		}
	}
	
	if !realPathAllowed {
		return "", fmt.Errorf("access denied - symlink target outside allowed directories")
	}
	
	return realPath, nil
}

// ReadFileSchema defines the schema for read_file tool input
var ReadFileSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"path": map[string]interface{}{
			"type": "string",
		},
	},
	"required": []string{"path"},
}

// ReadMultipleFilesSchema defines the schema for read_multiple_files tool input
var ReadMultipleFilesSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"paths": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "string",
			},
		},
	},
	"required": []string{"paths"},
}

// WriteFileSchema defines the schema for write_file tool input
var WriteFileSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"path": map[string]interface{}{
			"type": "string",
		},
		"content": map[string]interface{}{
			"type": "string",
		},
	},
	"required": []string{"path", "content"},
}

// CreateDirectorySchema defines the schema for create_directory tool input
var CreateDirectorySchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"path": map[string]interface{}{
			"type": "string",
		},
	},
	"required": []string{"path"},
}

// ListDirectorySchema defines the schema for list_directory tool input
var ListDirectorySchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"path": map[string]interface{}{
			"type": "string",
		},
	},
	"required": []string{"path"},
}

// MoveFileSchema defines the schema for move_file tool input
var MoveFileSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"source": map[string]interface{}{
			"type": "string",
		},
		"destination": map[string]interface{}{
			"type": "string",
		},
	},
	"required": []string{"source", "destination"},
}

// SearchFilesSchema defines the schema for search_files tool input
var SearchFilesSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"path": map[string]interface{}{
			"type": "string",
		},
		"pattern": map[string]interface{}{
			"type": "string",
		},
	},
	"required": []string{"path", "pattern"},
}

// GetFileInfoSchema defines the schema for get_file_info tool input
var GetFileInfoSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"path": map[string]interface{}{
			"type": "string",
		},
	},
	"required": []string{"path"},
}

// ListAllowedDirectoriesSchema defines the schema for list_allowed_directories tool input
var ListAllowedDirectoriesSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{},
	"required": []string{},
}

// FilesystemTool defines the schema for a filesystem tool
type FilesystemTool struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
}

// FilesystemTools is a map of tool definitions
var FilesystemTools = map[string]FilesystemTool{
	"read_file": {
		Name: "read_file",
		Description: "Read the complete contents of a file from the file system. " +
			"Handles various text encodings and provides detailed error messages " +
			"if the file cannot be read. Use this tool when you need to examine " +
			"the contents of a single file. Only works within allowed directories.",
		InputSchema: ReadFileSchema,
	},
	"read_multiple_files": {
		Name: "read_multiple_files",
		Description: "Read the contents of multiple files simultaneously. This is more " +
			"efficient than reading files one by one when you need to analyze " +
			"or compare multiple files. Each file's content is returned with its " +
			"path as a reference. Failed reads for individual files won't stop " +
			"the entire operation. Only works within allowed directories.",
		InputSchema: ReadMultipleFilesSchema,
	},
	"write_file": {
		Name: "write_file",
		Description: "Create a new file or completely overwrite an existing file with new content. " +
			"Use with caution as it will overwrite existing files without warning. " +
			"Handles text content with proper encoding. Only works within allowed directories.",
		InputSchema: WriteFileSchema,
	},
	"create_directory": {
		Name: "create_directory",
		Description: "Create a new directory or ensure a directory exists. Can create multiple " +
			"nested directories in one operation. If the directory already exists, " +
			"this operation will succeed silently. Perfect for setting up directory " +
			"structures for projects or ensuring required paths exist. Only works within allowed directories.",
		InputSchema: CreateDirectorySchema,
	},
	"list_directory": {
		Name: "list_directory",
		Description: "Get a detailed listing of all files and directories in a specified path. " +
			"Results clearly distinguish between files and directories with [FILE] and [DIR] " +
			"prefixes. This tool is essential for understanding directory structure and " +
			"finding specific files within a directory. Only works within allowed directories.",
		InputSchema: ListDirectorySchema,
	},
	"move_file": {
		Name: "move_file",
		Description: "Move or rename files and directories. Can move files between directories " +
			"and rename them in a single operation. If the destination exists, the " +
			"operation will fail. Works across different directories and can be used " +
			"for simple renaming within the same directory. Both source and destination must be within allowed directories.",
		InputSchema: MoveFileSchema,
	},
	"search_files": {
		Name: "search_files",
		Description: "Recursively search for files and directories matching a pattern. " +
			"Searches through all subdirectories from the starting path. The search " +
			"is case-insensitive and matches partial names. Returns full paths to all " +
			"matching items. Great for finding files when you don't know their exact location. " +
			"Only searches within allowed directories.",
		InputSchema: SearchFilesSchema,
	},
	"get_file_info": {
		Name: "get_file_info",
		Description: "Retrieve detailed metadata about a file or directory. Returns comprehensive " +
			"information including size, creation time, last modified time, permissions, " +
			"and type. This tool is perfect for understanding file characteristics " +
			"without reading the actual content. Only works within allowed directories.",
		InputSchema: GetFileInfoSchema,
	},
	"list_allowed_directories": {
		Name: "list_allowed_directories",
		Description: "Returns the list of directories that this server is allowed to access. " +
			"Use this to understand which directories are available before trying to access files.",
		InputSchema: ListAllowedDirectoriesSchema,
	},
}

// GetFileStats returns file metadata
func GetFileStats(filePath string) (FileInfo, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return FileInfo{}, err
	}

	// Get file time attributes
	var created, accessed, modified time.Time
	
	// On some file systems, some time attributes might not be available
	// Here's a basic implementation that works cross-platform
	modified = info.ModTime()
	
	// For creation time and access time, we use platform-specific methods
	// In a real implementation, this would use platform-specific syscalls
	// For simplicity, we'll use ModTime for all times here
	created = modified
	accessed = modified

	// Get file permissions in octal format
	permissions := fmt.Sprintf("%o", info.Mode().Perm())

	return FileInfo{
		Size:        info.Size(),
		Created:     created,
		Modified:    modified,
		Accessed:    accessed,
		IsDirectory: info.IsDir(),
		IsFile:      !info.IsDir(),
		Permissions: permissions,
	}, nil
}

// SearchFiles searches for files matching a pattern in a directory tree
func SearchFiles(fm *FileManager, rootPath, pattern string) ([]string, error) {
	// Validate the root path
	validRootPath, err := fm.ValidatePath(rootPath)
	if err != nil {
		return nil, err
	}

	var results []string
	pattern = strings.ToLower(pattern)

	err = filepath.WalkDir(validRootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Skip errors and continue walking
			return nil
		}

		// Try to validate each path
		_, validateErr := fm.ValidatePath(path)
		if validateErr != nil {
			// Skip this path if it's not valid
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if the name matches the pattern
		if strings.Contains(strings.ToLower(d.Name()), pattern) {
			results = append(results, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// ReadFile reads the contents of a file
func (fm *FileManager) ReadFile(path string) (string, error) {
	validPath, err := fm.ValidatePath(path)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(validPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// ReadMultipleFiles reads the contents of multiple files
func (fm *FileManager) ReadMultipleFiles(paths []string) (string, error) {
	var results []string

	for _, filePath := range paths {
		content, err := fm.ReadFile(filePath)
		if err != nil {
			results = append(results, fmt.Sprintf("%s: Error - %s", filePath, err.Error()))
		} else {
			results = append(results, fmt.Sprintf("%s:\n%s", filePath, content))
		}
	}

	return strings.Join(results, "\n---\n"), nil
}

// WriteFile writes content to a file
func (fm *FileManager) WriteFile(path, content string) error {
	validPath, err := fm.ValidatePath(path)
	if err != nil {
		return err
	}

	err = os.WriteFile(validPath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// CreateDirectory creates a directory
func (fm *FileManager) CreateDirectory(path string) error {
	validPath, err := fm.ValidatePath(path)
	if err != nil {
		return err
	}

	err = os.MkdirAll(validPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return nil
}

// ListDirectory lists the contents of a directory
func (fm *FileManager) ListDirectory(path string) (string, error) {
	validPath, err := fm.ValidatePath(path)
	if err != nil {
		return "", err
	}

	entries, err := os.ReadDir(validPath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	var result []string
	for _, entry := range entries {
		prefix := "[FILE]"
		if entry.IsDir() {
			prefix = "[DIR]"
		}
		result = append(result, fmt.Sprintf("%s %s", prefix, entry.Name()))
	}

	return strings.Join(result, "\n"), nil
}

// MoveFile moves or renames a file or directory
func (fm *FileManager) MoveFile(source, destination string) error {
	validSource, err := fm.ValidatePath(source)
	if err != nil {
		return err
	}

	validDest, err := fm.ValidatePath(destination)
	if err != nil {
		return err
	}

	err = os.Rename(validSource, validDest)
	if err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	return nil
}

// GetFileInfo gets information about a file
func (fm *FileManager) GetFileInfo(path string) (string, error) {
	validPath, err := fm.ValidatePath(path)
	if err != nil {
		return "", err
	}

	info, err := GetFileStats(validPath)
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	// Format the file info
	result := []string{
		fmt.Sprintf("size: %d", info.Size),
		fmt.Sprintf("created: %s", info.Created),
		fmt.Sprintf("modified: %s", info.Modified),
		fmt.Sprintf("accessed: %s", info.Accessed),
		fmt.Sprintf("isDirectory: %t", info.IsDirectory),
		fmt.Sprintf("isFile: %t", info.IsFile),
		fmt.Sprintf("permissions: %s", info.Permissions),
	}

	return strings.Join(result, "\n"), nil
}

// ListAllowedDirectories returns the list of allowed directories
func (fm *FileManager) ListAllowedDirectories() string {
	return fmt.Sprintf("Allowed directories:\n%s", strings.Join(fm.allowedDirectories, "\n"))
}

// ParseReadFileArgs parses arguments for read_file
func ParseReadFileArgs(args json.RawMessage) (string, error) {
	var params struct {
		Path string `json:"path"`
	}
	
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments for read_file: %w", err)
	}
	
	if params.Path == "" {
		return "", fmt.Errorf("path parameter is required")
	}
	
	return params.Path, nil
}

// ParseReadMultipleFilesArgs parses arguments for read_multiple_files
func ParseReadMultipleFilesArgs(args json.RawMessage) ([]string, error) {
	var params struct {
		Paths []string `json:"paths"`
	}
	
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("invalid arguments for read_multiple_files: %w", err)
	}
	
	if len(params.Paths) == 0 {
		return nil, fmt.Errorf("paths parameter is required and must not be empty")
	}
	
	return params.Paths, nil
}

// ParseWriteFileArgs parses arguments for write_file
func ParseWriteFileArgs(args json.RawMessage) (string, string, error) {
	var params struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	
	if err := json.Unmarshal(args, &params); err != nil {
		return "", "", fmt.Errorf("invalid arguments for write_file: %w", err)
	}
	
	if params.Path == "" {
		return "", "", fmt.Errorf("path parameter is required")
	}
	
	return params.Path, params.Content, nil
}

// ParseCreateDirectoryArgs parses arguments for create_directory
func ParseCreateDirectoryArgs(args json.RawMessage) (string, error) {
	var params struct {
		Path string `json:"path"`
	}
	
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments for create_directory: %w", err)
	}
	
	if params.Path == "" {
		return "", fmt.Errorf("path parameter is required")
	}
	
	return params.Path, nil
}

// ParseListDirectoryArgs parses arguments for list_directory
func ParseListDirectoryArgs(args json.RawMessage) (string, error) {
	var params struct {
		Path string `json:"path"`
	}
	
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments for list_directory: %w", err)
	}
	
	if params.Path == "" {
		return "", fmt.Errorf("path parameter is required")
	}
	
	return params.Path, nil
}

// ParseMoveFileArgs parses arguments for move_file
func ParseMoveFileArgs(args json.RawMessage) (string, string, error) {
	var params struct {
		Source      string `json:"source"`
		Destination string `json:"destination"`
	}
	
	if err := json.Unmarshal(args, &params); err != nil {
		return "", "", fmt.Errorf("invalid arguments for move_file: %w", err)
	}
	
	if params.Source == "" || params.Destination == "" {
		return "", "", fmt.Errorf("source and destination parameters are required")
	}
	
	return params.Source, params.Destination, nil
}

// ParseSearchFilesArgs parses arguments for search_files
func ParseSearchFilesArgs(args json.RawMessage) (string, string, error) {
	var params struct {
		Path    string `json:"path"`
		Pattern string `json:"pattern"`
	}
	
	if err := json.Unmarshal(args, &params); err != nil {
		return "", "", fmt.Errorf("invalid arguments for search_files: %w", err)
	}
	
	if params.Path == "" || params.Pattern == "" {
		return "", "", fmt.Errorf("path and pattern parameters are required")
	}
	
	return params.Path, params.Pattern, nil
}

// ParseGetFileInfoArgs parses arguments for get_file_info
func ParseGetFileInfoArgs(args json.RawMessage) (string, error) {
	var params struct {
		Path string `json:"path"`
	}
	
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments for get_file_info: %w", err)
	}
	
	if params.Path == "" {
		return "", fmt.Errorf("path parameter is required")
	}
	
	return params.Path, nil
}
