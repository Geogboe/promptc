package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Geogboe/promptc/internal/compiler"
	"github.com/Geogboe/promptc/internal/library"
)

// TestScenario_NewDeveloperStartsProject simulates a developer workflow:
// 1. Init project with template
// 2. Compile spec
// 3. Add custom features and recompile
// 4. Verify output contains expected content
func TestScenario_NewDeveloperStartsProject(t *testing.T) {
	dir := t.TempDir()

	t.Log("Step 1: Initialize new web API project")
	specFile := filepath.Join(dir, "api.spec.promptc")
	if err := compiler.InitProject(specFile, "web-api"); err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	t.Log("Step 2: Compile")
	comp := compiler.NewCompiler(dir)
	outPath, err := comp.Compile(specFile, nil)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	data, _ := os.ReadFile(outPath)
	result := string(data)

	if !strings.Contains(result, "# Project Instructions") {
		t.Error("output missing expected header")
	}
	if !strings.Contains(result, "REST API") {
		t.Error("output missing REST API content from web-api template")
	}

	t.Log("Step 3: Update spec with custom features")
	customContent := `imports:
  - patterns.rest_api
  - patterns.database
  - constraints.security
  - constraints.code_quality

context:
  language: go
  framework: fiber
  database: postgresql
  auth: jwt

features:
  - auth: "User authentication with JWT tokens"
  - users: "User CRUD operations"
  - posts: "Blog post CRUD operations"
  - search: "Full-text search"

constraints:
  - no_hardcoded_secrets
  - comprehensive_test_coverage
`
	if err := os.WriteFile(specFile, []byte(customContent), 0644); err != nil {
		t.Fatal(err)
	}

	t.Log("Step 4: Recompile")
	outPath, err = comp.Compile(specFile, nil)
	if err != nil {
		t.Fatalf("Recompile failed: %v", err)
	}

	data, _ = os.ReadFile(outPath)
	updated := string(data)

	for _, expected := range []string{
		"User authentication with JWT tokens",
		"User CRUD operations",
		"postgresql",
		"fiber",
		"REST API",
		"Database",
		"Security",
	} {
		if !strings.Contains(updated, expected) {
			t.Errorf("updated output missing: %s", expected)
		}
	}

	t.Log("✓ Developer workflow successful")
}

// TestScenario_TeamSharedLibraries simulates team library sharing with overrides.
func TestScenario_TeamSharedLibraries(t *testing.T) {
	dir := t.TempDir()

	// Dev 1: has team standards library
	dev1Dir := filepath.Join(dir, "dev1-project")
	dev1TeamDir := filepath.Join(dev1Dir, "prompts", "team")
	if err := os.MkdirAll(dev1TeamDir, 0755); err != nil {
		t.Fatal(err)
	}

	teamStandards := `# Team Coding Standards

## Code Review Requirements
- All PRs require 2 approvals
- No direct commits to main
- All code must have tests
`
	if err := os.WriteFile(filepath.Join(dev1TeamDir, "standards.prompt"), []byte(teamStandards), 0644); err != nil {
		t.Fatal(err)
	}

	dev1Spec := filepath.Join(dev1Dir, "app.spec.promptc")
	dev1Content := `imports:
  - patterns.rest_api
  - team.standards

context:
  language: go

features:
  - api: "REST API endpoints"
`
	if err := os.WriteFile(dev1Spec, []byte(dev1Content), 0644); err != nil {
		t.Fatal(err)
	}

	comp1 := compiler.NewCompiler(dev1Dir)
	outPath1, err := comp1.Compile(dev1Spec, nil)
	if err != nil {
		t.Fatalf("Dev1 compile failed: %v", err)
	}

	data1, _ := os.ReadFile(outPath1)
	if !strings.Contains(string(data1), "Team Coding Standards") {
		t.Error("dev1 output missing team standards")
	}
	if !strings.Contains(string(data1), "2 approvals") {
		t.Error("dev1 output missing team standards content")
	}

	// Dev 2: project-specific override
	dev2Dir := filepath.Join(dir, "dev2-project")
	dev2TeamDir := filepath.Join(dev2Dir, "prompts", "team")
	if err := os.MkdirAll(dev2TeamDir, 0755); err != nil {
		t.Fatal(err)
	}

	projectStandards := `# Project-Specific Standards

## This project uses different rules
- Fast-track PRs: only 1 approval needed
- Experimental features allowed
`
	if err := os.WriteFile(filepath.Join(dev2TeamDir, "standards.prompt"), []byte(projectStandards), 0644); err != nil {
		t.Fatal(err)
	}

	dev2Spec := filepath.Join(dev2Dir, "app.spec.promptc")
	dev2Content := `imports:
  - patterns.rest_api
  - team.standards

context:
  language: python

features:
  - prototype: "Quick prototype"
`
	if err := os.WriteFile(dev2Spec, []byte(dev2Content), 0644); err != nil {
		t.Fatal(err)
	}

	comp2 := compiler.NewCompiler(dev2Dir)
	outPath2, err := comp2.Compile(dev2Spec, nil)
	if err != nil {
		t.Fatalf("Dev2 compile failed: %v", err)
	}

	data2, _ := os.ReadFile(outPath2)
	result2 := string(data2)

	if !strings.Contains(result2, "Project-Specific Standards") {
		t.Error("dev2 should use project-specific standards")
	}
	if !strings.Contains(result2, "1 approval") {
		t.Error("dev2 missing project override content")
	}
	if strings.Contains(result2, "2 approvals") {
		t.Error("dev2 should NOT have team standards (overridden by project)")
	}

	t.Log("✓ Team shared libraries and overrides working correctly")
}

// TestScenario_IterativeBuild simulates building an app step by step.
func TestScenario_IterativeBuild(t *testing.T) {
	dir := t.TempDir()
	specFile := filepath.Join(dir, "app.spec.promptc")
	comp := compiler.NewCompiler(dir)

	steps := []struct {
		name    string
		content string
		check   string
	}{
		{
			name: "basic structure",
			content: `context:
  language: go
features:
  - structure: "Basic project structure"
`,
			check: "Basic project structure",
		},
		{
			name: "add auth",
			content: `imports:
  - patterns.rest_api
context:
  language: go
  auth: jwt
features:
  - structure: "Basic project structure"
  - auth: "JWT authentication"
`,
			check: "JWT authentication",
		},
		{
			name: "add database",
			content: `imports:
  - patterns.rest_api
  - patterns.database
context:
  language: go
  auth: jwt
  database: postgresql
features:
  - structure: "Basic project structure"
  - auth: "JWT authentication"
  - users: "User model and CRUD"
`,
			check: "User model and CRUD",
		},
	}

	for _, step := range steps {
		t.Log("Step:", step.name)
		if err := os.WriteFile(specFile, []byte(step.content), 0644); err != nil {
			t.Fatal(err)
		}

		outPath, err := comp.Compile(specFile, nil)
		if err != nil {
			t.Fatalf("Step %q compile failed: %v", step.name, err)
		}

		data, _ := os.ReadFile(outPath)
		if !strings.Contains(string(data), step.check) {
			t.Errorf("step %q: output missing %q", step.name, step.check)
		}
	}

	t.Log("✓ Iterative build successful")
}

// TestScenario_ExclusionsInComplexProject tests import exclusions.
func TestScenario_ExclusionsInComplexProject(t *testing.T) {
	dir := t.TempDir()

	promptsDir := filepath.Join(dir, "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatal(err)
	}

	baseLib := "# Base Patterns\n\n## Common\n- Use error handling\n"
	if err := os.WriteFile(filepath.Join(promptsDir, "base.prompt"), []byte(baseLib), 0644); err != nil {
		t.Fatal(err)
	}
	extLib := "# Extended Patterns\n\n## Advanced\n- Use caching\n"
	if err := os.WriteFile(filepath.Join(promptsDir, "extended.prompt"), []byte(extLib), 0644); err != nil {
		t.Fatal(err)
	}

	comp := compiler.NewCompiler(dir)

	// Test 1: both imported
	spec1 := filepath.Join(dir, "both.spec.promptc")
	if err := os.WriteFile(spec1, []byte("imports:\n  - base\n  - extended\nfeatures:\n  - f: \"test\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	outPath1, err := comp.Compile(spec1, nil)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	data1, _ := os.ReadFile(outPath1)
	if !strings.Contains(string(data1), "Base Patterns") || !strings.Contains(string(data1), "Extended Patterns") {
		t.Error("both libraries should be present")
	}

	// Test 2: exclude base
	spec2 := filepath.Join(dir, "exclude.spec.promptc")
	if err := os.WriteFile(spec2, []byte("imports:\n  - extended\n  - \"!base\"\nfeatures:\n  - f: \"test\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	outPath2, err := comp.Compile(spec2, nil)
	if err != nil {
		t.Fatalf("compile with exclusion failed: %v", err)
	}
	data2, _ := os.ReadFile(outPath2)
	if strings.Contains(string(data2), "Base Patterns") {
		t.Error("base should be excluded")
	}
	if !strings.Contains(string(data2), "Extended Patterns") {
		t.Error("extended should still be present")
	}

	t.Log("✓ Exclusions working correctly")
}

// TestScenario_LibraryDiscovery tests library listing with nested structure.
func TestScenario_LibraryDiscovery(t *testing.T) {
	dir := t.TempDir()

	dirs := []string{"prompts/frontend", "prompts/backend"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
			t.Fatal(err)
		}
	}

	files := map[string]string{
		"prompts/frontend/react.prompt":   "# React patterns",
		"prompts/frontend/vue.prompt":     "# Vue patterns",
		"prompts/backend/api.prompt":      "# API patterns",
		"prompts/backend/database.prompt": "# Database patterns",
	}
	for path, content := range files {
		if err := os.WriteFile(filepath.Join(dir, path), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	mgr := library.NewManager(dir)
	libs := mgr.ListLibraries()

	if len(libs.BuiltIn) == 0 {
		t.Error("should have built-in libraries")
	}

	for _, expected := range []string{"frontend.react", "frontend.vue", "backend.api", "backend.database"} {
		found := false
		for _, lib := range libs.Project {
			if strings.Contains(lib, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected library not found: %s. Got: %v", expected, libs.Project)
		}
	}

	// Verify nested import works
	specFile := filepath.Join(dir, "app.spec.promptc")
	if err := os.WriteFile(specFile, []byte("imports:\n  - frontend.react\n  - backend.api\nfeatures:\n  - app: \"Full-stack\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	comp := compiler.NewCompiler(dir)
	outPath, err := comp.Compile(specFile, nil)
	if err != nil {
		t.Fatalf("Compile with nested imports failed: %v", err)
	}

	data, _ := os.ReadFile(outPath)
	if !strings.Contains(string(data), "React patterns") {
		t.Error("output missing React patterns from nested library")
	}
	if !strings.Contains(string(data), "API patterns") {
		t.Error("output missing API patterns from nested library")
	}

	t.Log("✓ Library discovery and nested imports working")
}
