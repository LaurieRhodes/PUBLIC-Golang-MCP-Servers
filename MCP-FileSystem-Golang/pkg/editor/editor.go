package editor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// EditHistory tracks file edits for undo functionality
type EditHistory struct {
	FilePath     string
	OriginalHash string
	BackupPath   string
	Timestamp    time.Time
}

// EditManager manages file editing operations with undo capability
type EditManager struct {
	history      []EditHistory
	historyMutex sync.RWMutex
	backupDir    string
}

// NewEditManager creates a new EditManager
func NewEditManager(backupDir string) (*EditManager, error) {
	if backupDir == "" {
		// Use system temp directory
		backupDir = filepath.Join(os.TempDir(), "mcp-filesystem-backups")
	}

	// Ensure backup directory exists
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	return &EditManager{
		history:   make([]EditHistory, 0),
		backupDir: backupDir,
	}, nil
}

// createBackup creates a backup of a file before editing
func (em *EditManager) createBackup(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file for backup: %w", err)
	}

	// Create a unique backup filename
	timestamp := time.Now().UnixNano()
	backupName := fmt.Sprintf("%s_%d.bak", filepath.Base(filePath), timestamp)
	backupPath := filepath.Join(em.backupDir, backupName)

	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to write backup: %w", err)
	}

	return backupPath, nil
}

// addToHistory adds an edit to the history
func (em *EditManager) addToHistory(filePath, backupPath string) {
	em.historyMutex.Lock()
	defer em.historyMutex.Unlock()

	entry := EditHistory{
		FilePath:   filePath,
		BackupPath: backupPath,
		Timestamp:  time.Now(),
	}

	em.history = append(em.history, entry)

	// Keep only the last 100 edits
	if len(em.history) > 100 {
		// Remove old backup file
		if err := os.Remove(em.history[0].BackupPath); err != nil {
			// Log error but continue
			fmt.Fprintf(os.Stderr, "Warning: failed to remove old backup: %v\n", err)
		}
		em.history = em.history[1:]
	}
}

// StrReplace performs an exact string match and replace in a file
func (em *EditManager) StrReplace(filePath, oldStr, newStr string) error {
	// Read the entire file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	fileContent := string(content)

	// Check if old string exists
	if !strings.Contains(fileContent, oldStr) {
		return fmt.Errorf("string not found in file: %q", oldStr)
	}

	// Count occurrences
	count := strings.Count(fileContent, oldStr)
	if count > 1 {
		return fmt.Errorf("string appears %d times in file; it must appear exactly once for str_replace", count)
	}

	// Create backup before modifying
	backupPath, err := em.createBackup(filePath)
	if err != nil {
		return err
	}

	// Perform replacement
	newContent := strings.Replace(fileContent, oldStr, newStr, 1)

	// Write the modified content
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Add to history
	em.addToHistory(filePath, backupPath)

	return nil
}

// Insert inserts text after a specified line number
func (em *EditManager) Insert(filePath string, lineNumber int, text string) error {
	// Read file line by line
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Validate line number (1-indexed)
	if lineNumber < 0 || lineNumber > len(lines) {
		return fmt.Errorf("invalid line number %d; file has %d lines (use 0 to insert at beginning, %d to append)", 
			lineNumber, len(lines), len(lines))
	}

	// Create backup
	backupPath, err := em.createBackup(filePath)
	if err != nil {
		return err
	}

	// Insert text after the specified line
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:lineNumber]...)
	newLines = append(newLines, text)
	newLines = append(newLines, lines[lineNumber:]...)

	// Write back to file
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Add to history
	em.addToHistory(filePath, backupPath)

	return nil
}

// UndoEdit undoes the last edit made to a specific file
func (em *EditManager) UndoEdit(filePath string) error {
	em.historyMutex.Lock()
	defer em.historyMutex.Unlock()

	// Find the most recent edit for this file
	var lastEditIndex = -1
	for i := len(em.history) - 1; i >= 0; i-- {
		if em.history[i].FilePath == filePath {
			lastEditIndex = i
			break
		}
	}

	if lastEditIndex == -1 {
		return fmt.Errorf("no edit history found for file: %s", filePath)
	}

	entry := em.history[lastEditIndex]

	// Restore from backup
	backupContent, err := os.ReadFile(entry.BackupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	if err := os.WriteFile(filePath, backupContent, 0644); err != nil {
		return fmt.Errorf("failed to restore file: %w", err)
	}

	// Remove the backup file
	if err := os.Remove(entry.BackupPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to remove backup file: %v\n", err)
	}

	// Remove from history
	em.history = append(em.history[:lastEditIndex], em.history[lastEditIndex+1:]...)

	return nil
}

// GetEditHistory returns the edit history for a specific file
func (em *EditManager) GetEditHistory(filePath string) []EditHistory {
	em.historyMutex.RLock()
	defer em.historyMutex.RUnlock()

	var fileHistory []EditHistory
	for _, entry := range em.history {
		if entry.FilePath == filePath {
			fileHistory = append(fileHistory, entry)
		}
	}

	return fileHistory
}

// Tool schemas for editor operations

// StrReplaceSchema defines the schema for str_replace tool input
var StrReplaceSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"path": map[string]interface{}{
			"type":        "string",
			"description": "Path to the file to edit",
		},
		"old_str": map[string]interface{}{
			"type":        "string",
			"description": "The exact string to replace (must appear exactly once in the file)",
		},
		"new_str": map[string]interface{}{
			"type":        "string",
			"description": "The string to replace it with (can be empty to delete)",
		},
	},
	"required": []string{"path", "old_str"},
}

// InsertSchema defines the schema for insert tool input
var InsertSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"path": map[string]interface{}{
			"type":        "string",
			"description": "Path to the file to edit",
		},
		"line_number": map[string]interface{}{
			"type":        "integer",
			"description": "Line number after which to insert (0 for beginning, file line count to append)",
		},
		"text": map[string]interface{}{
			"type":        "string",
			"description": "Text to insert",
		},
	},
	"required": []string{"path", "line_number", "text"},
}

// UndoEditSchema defines the schema for undo_edit tool input
var UndoEditSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"path": map[string]interface{}{
			"type":        "string",
			"description": "Path to the file to undo edits for",
		},
	},
	"required": []string{"path"},
}

// EditorTool defines the schema for an editor tool
type EditorTool struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
}

// EditorTools is a map of editor tool definitions
var EditorTools = map[string]EditorTool{
	"str_replace": {
		Name: "str_replace",
		Description: "Replace an exact string in a file with another string. The old_str must appear " +
			"exactly once in the file. This is the safest way to make surgical edits to files. " +
			"A backup is automatically created before the edit. Use this instead of rewriting entire files " +
			"when making small changes. Only works within allowed directories.",
		InputSchema: StrReplaceSchema,
	},
	"insert": {
		Name: "insert",
		Description: "Insert text after a specified line number in a file (1-indexed). Use line_number=0 " +
			"to insert at the beginning of the file, or line_number equal to the file's line count to append. " +
			"A backup is automatically created before the edit. Only works within allowed directories.",
		InputSchema: InsertSchema,
	},
	"undo_edit": {
		Name: "undo_edit",
		Description: "Undo the last edit made to a specific file. This will restore the file to its state " +
			"before the last str_replace or insert operation. Can be called multiple times to undo multiple " +
			"edits. Only works within allowed directories.",
		InputSchema: UndoEditSchema,
	},
}

// Argument parsing functions

// ParseStrReplaceArgs parses arguments for str_replace
func ParseStrReplaceArgs(args json.RawMessage) (path, oldStr, newStr string, err error) {
	var params struct {
		Path   string `json:"path"`
		OldStr string `json:"old_str"`
		NewStr string `json:"new_str"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		return "", "", "", fmt.Errorf("invalid arguments for str_replace: %w", err)
	}

	if params.Path == "" {
		return "", "", "", fmt.Errorf("path parameter is required")
	}

	if params.OldStr == "" {
		return "", "", "", fmt.Errorf("old_str parameter is required")
	}

	return params.Path, params.OldStr, params.NewStr, nil
}

// ParseInsertArgs parses arguments for insert
func ParseInsertArgs(args json.RawMessage) (path string, lineNumber int, text string, err error) {
	var params struct {
		Path       string `json:"path"`
		LineNumber int    `json:"line_number"`
		Text       string `json:"text"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		return "", 0, "", fmt.Errorf("invalid arguments for insert: %w", err)
	}

	if params.Path == "" {
		return "", 0, "", fmt.Errorf("path parameter is required")
	}

	if params.Text == "" {
		return "", 0, "", fmt.Errorf("text parameter is required")
	}

	return params.Path, params.LineNumber, params.Text, nil
}

// ParseUndoEditArgs parses arguments for undo_edit
func ParseUndoEditArgs(args json.RawMessage) (path string, err error) {
	var params struct {
		Path string `json:"path"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments for undo_edit: %w", err)
	}

	if params.Path == "" {
		return "", fmt.Errorf("path parameter is required")
	}

	return params.Path, nil
}
