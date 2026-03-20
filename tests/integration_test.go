package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Geogboe/promptc/internal/compiler"
	"github.com/Geogboe/promptc/internal/library"
)

// TestEndToEndWorkflow tests the complete workflow from init to compile.
func TestEndToEndWorkflow(t *testing.T) {
	dir := t.TempDir()

	// Init a project
	specFile := filepath.Join(dir, "myapp.spec.promptc")
	if err := compiler.InitProject(specFile, "web-api"); err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}
	if _, err := os.Stat(specFile); os.IsNotExist(err) {
		t.Fatal("spec file was not created")
	}

	// Compile it
	comp := compiler.NewCompiler(dir)
	outPath, err := comp.Compile(specFile, &compiler.CompileOptions{SkipValidate: false})
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	if outPath == "" {
		t.Fatal("Compile returned empty output path")
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	out := string(data)

	if !strings.Contains(out, "# Project Instructions") {
		t.Error("output missing project instructions header")
	}
	if !strings.Contains(out, "REST API") {
		t.Error("output missing REST API content from imports")
	}
}

// TestCustomLibraryResolution tests custom library resolution.
func TestCustomLibraryResolution(t *testing.T) {
	dir := t.TempDir()

	promptsDir := filepath.Join(dir, "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatal(err)
	}

	customLib := filepath.Join(promptsDir, "custom.prompt")
	customContent := "# Custom Library\nThis is a custom prompt library."
	if err := os.WriteFile(customLib, []byte(customContent), 0644); err != nil {
		t.Fatal(err)
	}

	specFile := filepath.Join(dir, "test.spec.promptc")
	specContent := `imports:
  - custom

features:
  - test: "Test feature"
`
	if err := os.WriteFile(specFile, []byte(specContent), 0644); err != nil {
		t.Fatal(err)
	}

	comp := compiler.NewCompiler(dir)
	outPath, err := comp.Compile(specFile, &compiler.CompileOptions{SkipValidate: false})
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	data, _ := os.ReadFile(outPath)
	if !strings.Contains(string(data), "Custom Library") {
		t.Error("custom library content not found in output")
	}
}

// TestInvalidSpecFile tests handling of invalid spec files.
func TestInvalidSpecFile(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "invalid YAML",
			content: "imports:\n  - test\ncontext:\n  invalid yaml here: [",
		},
		{
			name:    "empty features",
			content: "features: []",
		},
		{
			name:    "invalid import name",
			content: "imports:\n  - \"../../../etc/passwd\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specFile := filepath.Join(dir, tt.name+".spec.promptc")
			if err := os.WriteFile(specFile, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			comp := compiler.NewCompiler(dir)
			_, err := comp.Compile(specFile, &compiler.CompileOptions{SkipValidate: false})
			if err == nil {
				t.Errorf("expected error for %s, got none", tt.name)
			}
		})
	}
}

// TestLibraryListingIntegration tests library listing across all sources.
func TestLibraryListingIntegration(t *testing.T) {
	dir := t.TempDir()

	promptsDir := filepath.Join(dir, "prompts", "mylibs")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(promptsDir, "project.prompt"), []byte("# Project Library"), 0644); err != nil {
		t.Fatal(err)
	}

	mgr := library.NewManager(dir)
	libs := mgr.ListLibraries()

	if len(libs.BuiltIn) == 0 {
		t.Error("no built-in libraries found")
	}

	found := false
	for _, lib := range libs.Project {
		if strings.Contains(lib, "project") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("project library not found. Got: %v", libs.Project)
	}
}

// TestCompilationWithMultipleImports tests complex import scenarios.
func TestCompilationWithMultipleImports(t *testing.T) {
	dir := t.TempDir()

	specFile := filepath.Join(dir, "complex.spec.promptc")
	specContent := `imports:
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
	if err := os.WriteFile(specFile, []byte(specContent), 0644); err != nil {
		t.Fatal(err)
	}

	comp := compiler.NewCompiler(dir)
	outPath, err := comp.Compile(specFile, nil)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	data, _ := os.ReadFile(outPath)
	out := string(data)

	for _, required := range []string{
		"REST API", "Testing", "Database", "Security", "Code Quality",
	} {
		if !strings.Contains(out, required) {
			t.Errorf("output missing required content: %s", required)
		}
	}

	if !strings.Contains(out, "User authentication with JWT") {
		t.Error("output missing feature content")
	}
}

// TestOutputFileCreated tests that output file is written correctly.
func TestOutputFileCreated(t *testing.T) {
	dir := t.TempDir()

	specFile := filepath.Join(dir, "test.spec.promptc")
	specContent := `features:
  - test: "Test feature"
`
	if err := os.WriteFile(specFile, []byte(specContent), 0644); err != nil {
		t.Fatal(err)
	}

	outDir := filepath.Join(dir, "output")
	comp := compiler.NewCompiler(dir)
	outPath, err := comp.Compile(specFile, &compiler.CompileOptions{OutputDir: outDir})
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Verify the output path is as expected
	expected := filepath.Join(outDir, "instructions.promptc")
	if outPath != expected {
		t.Errorf("output path = %q, want %q", outPath, expected)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	if len(data) == 0 {
		t.Error("output file is empty")
	}
}
