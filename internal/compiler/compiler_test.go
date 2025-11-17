package compiler

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCompiler(t *testing.T) {
	compiler := NewCompiler("")
	if compiler == nil {
		t.Fatal("NewCompiler returned nil")
	}

	if compiler.LibraryManager == nil {
		t.Error("LibraryManager is nil")
	}

	if compiler.Resolver == nil {
		t.Error("Resolver is nil")
	}
}

func TestCompile(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.prompt")

	content := `context:
  language: go
  framework: cobra

features:
  - cli_tool: "A command line tool"

constraints:
  - comprehensive_error_handling
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	compiler := NewCompiler(tmpDir)

	// Test compilation
	result, err := compiler.Compile(testFile, &CompileOptions{
		Target:   "raw",
		Validate: true,
	})

	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	if result == "" {
		t.Error("Compile returned empty result")
	}

	if !containsString(result, "go") {
		t.Error("Result does not contain context")
	}
}

func TestCompileInvalidFile(t *testing.T) {
	compiler := NewCompiler("")

	_, err := compiler.Compile("nonexistent.prompt", &CompileOptions{
		Target: "raw",
	})

	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestCompileInvalidTarget(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.prompt")

	content := `features:
  - test: "test feature"
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	compiler := NewCompiler(tmpDir)

	_, err := compiler.Compile(testFile, &CompileOptions{
		Target: "invalid",
	})

	if err == nil {
		t.Error("Expected error for invalid target")
	}
}

func TestGetTemplates(t *testing.T) {
	templates := GetTemplates()

	if len(templates) == 0 {
		t.Error("No templates found")
	}

	expectedTemplates := []string{"basic", "web-api", "cli-tool"}
	for _, name := range expectedTemplates {
		if _, ok := templates[name]; !ok {
			t.Errorf("Missing template: %s", name)
		}
	}
}

func TestInitProject(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.prompt")

	err := InitProject(testFile, "basic")
	if err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Project file was not created")
	}

	// Test that it fails if file exists
	err = InitProject(testFile, "basic")
	if err == nil {
		t.Error("Expected error when file already exists")
	}
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && s != "" && substr != "" &&
		(s == substr || len(s) >= len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
