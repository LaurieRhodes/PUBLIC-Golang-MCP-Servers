package editor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStrReplace(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "editor-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create an edit manager
	em, err := NewEditManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create edit manager: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	originalContent := "Hello World\nThis is a test\nGoodbye World"
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test successful replacement
	err = em.StrReplace(testFile, "This is a test", "This is modified")
	if err != nil {
		t.Errorf("StrReplace failed: %v", err)
	}

	// Verify content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	expected := "Hello World\nThis is modified\nGoodbye World"
	if string(content) != expected {
		t.Errorf("Content mismatch. Expected:\n%s\nGot:\n%s", expected, string(content))
	}

	// Test string not found
	err = em.StrReplace(testFile, "nonexistent", "replacement")
	if err == nil {
		t.Error("Expected error for nonexistent string, got nil")
	}

	// Test multiple occurrences
	multiContent := "foo bar foo"
	if err := os.WriteFile(testFile, []byte(multiContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	err = em.StrReplace(testFile, "foo", "baz")
	if err == nil {
		t.Error("Expected error for multiple occurrences, got nil")
	}
}

func TestInsert(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "editor-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create an edit manager
	em, err := NewEditManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create edit manager: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	originalContent := "Line 1\nLine 2\nLine 3"
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test insert after line 1
	err = em.Insert(testFile, 1, "Inserted Line")
	if err != nil {
		t.Errorf("Insert failed: %v", err)
	}

	// Verify content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	expected := "Line 1\nInserted Line\nLine 2\nLine 3"
	if string(content) != expected {
		t.Errorf("Content mismatch. Expected:\n%s\nGot:\n%s", expected, string(content))
	}

	// Test insert at beginning (line 0)
	originalContent = "Line 1\nLine 2\nLine 3"
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	err = em.Insert(testFile, 0, "First Line")
	if err != nil {
		t.Errorf("Insert at beginning failed: %v", err)
	}

	content, err = os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	expected = "First Line\nLine 1\nLine 2\nLine 3"
	if string(content) != expected {
		t.Errorf("Content mismatch. Expected:\n%s\nGot:\n%s", expected, string(content))
	}

	// Test insert at end
	originalContent = "Line 1\nLine 2\nLine 3"
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	err = em.Insert(testFile, 3, "Last Line")
	if err != nil {
		t.Errorf("Insert at end failed: %v", err)
	}

	content, err = os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	expected = "Line 1\nLine 2\nLine 3\nLast Line"
	if string(content) != expected {
		t.Errorf("Content mismatch. Expected:\n%s\nGot:\n%s", expected, string(content))
	}

	// Test invalid line number
	err = em.Insert(testFile, 100, "Invalid")
	if err == nil {
		t.Error("Expected error for invalid line number, got nil")
	}
}

func TestUndoEdit(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "editor-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create an edit manager
	em, err := NewEditManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create edit manager: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	originalContent := "Original Content\nLine 2\nLine 3"
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Make an edit
	err = em.StrReplace(testFile, "Original Content", "Modified Content")
	if err != nil {
		t.Fatalf("StrReplace failed: %v", err)
	}

	// Verify content changed
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if !containsString(string(content), "Modified Content") {
		t.Error("Content was not modified")
	}

	// Undo the edit
	err = em.UndoEdit(testFile)
	if err != nil {
		t.Errorf("UndoEdit failed: %v", err)
	}

	// Verify content restored
	content, err = os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != originalContent {
		t.Errorf("Content not restored. Expected:\n%s\nGot:\n%s", originalContent, string(content))
	}

	// Test undo with no history
	err = em.UndoEdit(testFile)
	if err == nil {
		t.Error("Expected error for undo with no history, got nil")
	}
}

func TestMultipleEditsAndUndo(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "editor-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create an edit manager
	em, err := NewEditManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create edit manager: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	originalContent := "Line 1\nLine 2\nLine 3"
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Make multiple edits
	err = em.StrReplace(testFile, "Line 1", "Modified Line 1")
	if err != nil {
		t.Fatalf("First StrReplace failed: %v", err)
	}

	err = em.Insert(testFile, 1, "Inserted Line")
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	err = em.StrReplace(testFile, "Line 2", "Modified Line 2")
	if err != nil {
		t.Fatalf("Second StrReplace failed: %v", err)
	}

	// Undo last edit
	err = em.UndoEdit(testFile)
	if err != nil {
		t.Errorf("First undo failed: %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if !containsString(string(content), "Line 2") {
		t.Error("Last edit was not undone properly")
	}

	// Undo second-to-last edit
	err = em.UndoEdit(testFile)
	if err != nil {
		t.Errorf("Second undo failed: %v", err)
	}

	content, err = os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if containsString(string(content), "Inserted Line") {
		t.Error("Insert was not undone properly")
	}

	// Undo first edit
	err = em.UndoEdit(testFile)
	if err != nil {
		t.Errorf("Third undo failed: %v", err)
	}

	content, err = os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != originalContent {
		t.Errorf("All edits not undone. Expected:\n%s\nGot:\n%s", originalContent, string(content))
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
			findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
