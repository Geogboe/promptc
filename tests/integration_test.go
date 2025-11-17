package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Geogboe/promptc/internal/compiler"
	"github.com/Geogboe/promptc/internal/library"
)

// TestEndToEndWorkflow tests the complete workflow from init to compile
func TestEndToEndWorkflow(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize a project
	promptFile := filepath.Join(tmpDir, "myapp.prompt")
	err := compiler.InitProject(promptFile, "web-api")
	if err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(promptFile); os.IsNotExist(err) {
		t.Fatal("Prompt file was not created")
	}

	// Compile to all targets
	comp := compiler.NewCompiler(tmpDir)
	targets := []string{"raw", "claude", "cursor", "aider", "copilot"}

	for _, target := range targets {
		result, err := comp.Compile(promptFile, &compiler.CompileOptions{
			Target:   target,
			Validate: true,
		})

		if err != nil {
			t.Fatalf("Compilation to %s failed: %v", target, err)
		}

		if result == "" {
			t.Errorf("Compilation to %s produced empty result", target)
		}

		// Verify the result contains expected content
		if !strings.Contains(result, "REST API") {
			t.Errorf("Result for %s missing REST API content", target)
		}
	}
}

// TestCustomLibraryResolution tests custom library resolution
func TestCustomLibraryResolution(t *testing.T) {
	tmpDir := t.TempDir()

	// Create project prompts directory
	promptsDir := filepath.Join(tmpDir, "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatalf("Failed to create prompts dir: %v", err)
	}

	// Create a custom library
	customLib := filepath.Join(promptsDir, "custom.prompt")
	customContent := "# Custom Library\nThis is a custom prompt library."
	if err := os.WriteFile(customLib, []byte(customContent), 0644); err != nil {
		t.Fatalf("Failed to create custom library: %v", err)
	}

	// Create a prompt file that imports the custom library
	promptFile := filepath.Join(tmpDir, "test.prompt")
	promptContent := `imports:
  - custom

features:
  - test: "Test feature"
`
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("Failed to create prompt file: %v", err)
	}

	// Compile and verify custom library is included
	comp := compiler.NewCompiler(tmpDir)
	result, err := comp.Compile(promptFile, &compiler.CompileOptions{
		Target:   "raw",
		Validate: true,
	})

	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	if !strings.Contains(result, "Custom Library") {
		t.Error("Custom library content not found in result")
	}
}

// TestInvalidPromptFile tests handling of invalid prompt files
func TestInvalidPromptFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct{
		name    string
		content string
	}{
		{
			name: "invalid YAML",
			content: `
imports:
  - test
context:
  invalid yaml here: [
`,
		},
		{
			name: "empty features",
			content: `features: []`,
		},
		{
			name: "invalid import name",
			content: `imports:
  - "../../../etc/passwd"
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			promptFile := filepath.Join(tmpDir, tt.name+".prompt")
			if err := os.WriteFile(promptFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			comp := compiler.NewCompiler(tmpDir)
			_, err := comp.Compile(promptFile, &compiler.CompileOptions{
				Target:   "raw",
				Validate: true,
			})

			if err == nil {
				t.Errorf("Expected error for %s, got none", tt.name)
			}
		})
	}
}

// TestLibraryListingIntegration tests library listing across all sources
func TestLibraryListingIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create project-local library
	promptsDir := filepath.Join(tmpDir, "prompts/mylibs")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatalf("Failed to create prompts dir: %v", err)
	}

	projectLib := filepath.Join(promptsDir, "project.prompt")
	if err := os.WriteFile(projectLib, []byte("# Project Library"), 0644); err != nil {
		t.Fatalf("Failed to create project library: %v", err)
	}

	// List libraries
	mgr := library.NewManager(tmpDir)
	libs := mgr.ListLibraries()

	// Should have built-in libraries
	if len(libs.BuiltIn) == 0 {
		t.Error("No built-in libraries found")
	}

	// Should find our project library
	found := false
	for _, lib := range libs.Project {
		if strings.Contains(lib, "project") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Project library not found. Got: %v", libs.Project)
	}
}

// TestCompilationWithMultipleImports tests complex import scenarios
func TestCompilationWithMultipleImports(t *testing.T) {
	tmpDir := t.TempDir()

	promptFile := filepath.Join(tmpDir, "complex.prompt")
	promptContent := `imports:
  - patterns.rest_api
  - patterns.testing
  - patterns.database
  - constraints.security
  - constraints.code_quality

context:
  language: go
  framework: fiber
  database: postgresql
  testing: testify

features:
  - auth: "User authentication with JWT"
  - crud: "CRUD operations for resources"
  - api: "REST API endpoints"

constraints:
  - no_hardcoded_secrets
  - comprehensive_test_coverage
  - proper_error_handling
`

	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("Failed to create prompt file: %v", err)
	}

	comp := compiler.NewCompiler(tmpDir)
	result, err := comp.Compile(promptFile, &compiler.CompileOptions{
		Target:   "claude",
		Validate: true,
		Debug:    false,
	})

	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Verify all imports are included
	requiredContent := []string{
		"REST API",
		"Testing",
		"Database",
		"Security",
		"Code Quality",
	}

	for _, required := range requiredContent {
		if !strings.Contains(result, required) {
			t.Errorf("Result missing required content: %s", required)
		}
	}

	// Verify context is included
	if !strings.Contains(result, "go") {
		t.Error("Result missing context")
	}

	// Verify features are included
	if !strings.Contains(result, "User authentication") {
		t.Error("Result missing features")
	}
}

// TestOutputToFile tests writing compilation results to files
func TestOutputToFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple prompt file
	promptFile := filepath.Join(tmpDir, "test.prompt")
	promptContent := `features:
  - test: "Test feature"
`
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("Failed to create prompt file: %v", err)
	}

	comp := compiler.NewCompiler(tmpDir)

	// Compile and write to file
	result, err := comp.Compile(promptFile, &compiler.CompileOptions{
		Target:   "claude",
		Validate: true,
	})

	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Write to output file
	outputFile := filepath.Join(tmpDir, ".claude/instructions.md")
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	if err := os.WriteFile(outputFile, []byte(result), 0644); err != nil {
		t.Fatalf("Failed to write output file: %v", err)
	}

	// Verify file was written correctly
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if string(content) != result {
		t.Error("Output file content doesn't match compilation result")
	}
}
