package library

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathTraversalPrevention(t *testing.T) {
	manager := NewManager("")

	tests := []struct {
		name        string
		importName  string
		shouldError bool
	}{
		{
			name:        "valid import",
			importName:  "patterns.rest_api",
			shouldError: false,
		},
		{
			name:        "path traversal with double dots",
			importName:  "../../../etc/passwd",
			shouldError: true,
		},
		{
			name:        "path traversal in middle",
			importName:  "patterns../../../etc/passwd",
			shouldError: true,
		},
		{
			name:        "absolute path unix",
			importName:  "/etc/passwd",
			shouldError: true,
		},
		{
			name:        "absolute path windows",
			importName:  "C:\\Windows\\System32",
			shouldError: true,
		},
		{
			name:        "backslash attempt",
			importName:  "..\\..\\etc\\passwd",
			shouldError: true,
		},
		{
			name:        "null byte",
			importName:  "test\x00",
			shouldError: true,
		},
		{
			name:        "special chars",
			importName:  "test;rm -rf /",
			shouldError: true,
		},
		{
			name:        "valid with hyphen",
			importName:  "my-pattern.test",
			shouldError: false,
		},
		{
			name:        "valid with underscore",
			importName:  "my_pattern.test",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := manager.Resolve(tt.importName)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for input '%s', but got none", tt.importName)
				}
			} else {
				// For valid imports, we expect them to not be found (since they don't exist)
				// but the error should be "cannot resolve" not a path traversal error
				if err != nil && (err == ErrPathTraversal || err == ErrInvalidImportName) {
					t.Errorf("Got path traversal/invalid error for valid input '%s': %v", tt.importName, err)
				}
			}
		})
	}
}

func TestSymlinkSecurity(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	promptsDir := filepath.Join(tmpDir, "prompts")
	outsideDir := filepath.Join(tmpDir, "outside")

	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatalf("Failed to create prompts dir: %v", err)
	}
	if err := os.MkdirAll(outsideDir, 0755); err != nil {
		t.Fatalf("Failed to create outside dir: %v", err)
	}

	// Create a file outside the prompts directory
	outsideFile := filepath.Join(outsideDir, "secret.prompt")
	if err := os.WriteFile(outsideFile, []byte("secret content"), 0644); err != nil {
		t.Fatalf("Failed to create outside file: %v", err)
	}

	// Create a symlink inside prompts pointing to the outside file
	symlinkPath := filepath.Join(promptsDir, "evil.prompt")
	if err := os.Symlink(outsideFile, symlinkPath); err != nil {
		t.Skipf("Cannot create symlink (may not have permissions): %v", err)
	}

	// Try to access via symlink
	manager := NewManager(tmpDir)

	_, err := manager.Resolve("evil")
	if err == nil {
		t.Error("Expected error when following symlink outside allowed directory")
	}
}

func TestFileSizeLimit(t *testing.T) {
	tmpDir := t.TempDir()
	promptsDir := filepath.Join(tmpDir, "prompts")

	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatalf("Failed to create prompts dir: %v", err)
	}

	// Create a file that exceeds the size limit
	largeFile := filepath.Join(promptsDir, "large.prompt")
	const fileSize = 11 * 1024 * 1024 // 11MB (over the 10MB limit)

	f, err := os.Create(largeFile)
	if err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	// Write dummy data
	if err := f.Truncate(fileSize); err != nil {
		f.Close()
		t.Fatalf("Failed to truncate file: %v", err)
	}
	f.Close()

	manager := NewManager(tmpDir)

	// Test loadFromFilesystem directly to verify size limit works
	_, err = manager.loadFromFilesystem(promptsDir, "large.prompt")
	if err == nil {
		t.Error("Expected error for file exceeding size limit")
	}

	if err != nil && !contains(err.Error(), "too large") {
		t.Errorf("Expected 'too large' error, got: %v", err)
	}
}

func TestValidImportNameFunction(t *testing.T) {
	tests := []struct {
		name        string
		importName  string
		expectError bool
	}{
		{"valid simple", "test", false},
		{"valid with dot", "test.example", false},
		{"valid with underscore", "test_example", false},
		{"valid with hyphen", "test-example", false},
		{"valid complex", "my-lib.sub_module.test", false},
		{"empty string", "", true},
		{"path traversal", "../test", true},
		{"absolute path", "/etc/passwd", true},
		{"windows path", "C:\\test", true},
		{"special char semicolon", "test;cmd", true},
		{"special char pipe", "test|cmd", true},
		{"special char ampersand", "test&cmd", true},
		{"space", "test example", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateImportName(tt.importName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for '%s', got none", tt.importName)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for '%s', got: %v", tt.importName, err)
			}
		})
	}
}

func TestPathCleaningOnInit(t *testing.T) {
	// Test that paths are properly cleaned when creating manager
	tests := []string{
		"/test/path/./with/./dots",
		"/test/path/../cleaned",
		"./relative/path",
	}

	for _, path := range tests {
		manager := NewManager(path)

		// ProjectDir should be cleaned
		if manager.ProjectDir != filepath.Clean(path) {
			t.Errorf("ProjectDir not cleaned: got %s, want %s", manager.ProjectDir, filepath.Clean(path))
		}

		// Search paths should be cleaned
		for _, sp := range manager.SearchPaths {
			if sp != "defaults" && sp != filepath.Clean(sp) {
				t.Errorf("Search path not cleaned: %s", sp)
			}
		}
	}
}
