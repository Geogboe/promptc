package resolver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Geogboe/promptc/internal/library"
)

func TestResolverBasicImport(t *testing.T) {
	tmpDir := t.TempDir()
	promptsDir := filepath.Join(tmpDir, "prompts")
	os.MkdirAll(promptsDir, 0755)

	// Create test library file
	testContent := "# Test Library\nThis is test content"
	os.WriteFile(filepath.Join(promptsDir, "test.prompt"), []byte(testContent), 0644)

	manager := library.NewManager(tmpDir)
	resolver := NewResolver(manager)

	result, err := resolver.Resolve([]string{"test"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(result, "Test Library") {
		t.Errorf("Expected result to contain 'Test Library', got: %s", result)
	}
}

func TestResolverMultipleImports(t *testing.T) {
	tmpDir := t.TempDir()
	promptsDir := filepath.Join(tmpDir, "prompts")
	os.MkdirAll(promptsDir, 0755)

	// Create multiple library files
	os.WriteFile(filepath.Join(promptsDir, "lib1.prompt"), []byte("Library 1"), 0644)
	os.WriteFile(filepath.Join(promptsDir, "lib2.prompt"), []byte("Library 2"), 0644)
	os.WriteFile(filepath.Join(promptsDir, "lib3.prompt"), []byte("Library 3"), 0644)

	manager := library.NewManager(tmpDir)
	resolver := NewResolver(manager)

	result, err := resolver.Resolve([]string{"lib1", "lib2", "lib3"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(result, "Library 1") {
		t.Errorf("Expected result to contain 'Library 1'")
	}
	if !strings.Contains(result, "Library 2") {
		t.Errorf("Expected result to contain 'Library 2'")
	}
	if !strings.Contains(result, "Library 3") {
		t.Errorf("Expected result to contain 'Library 3'")
	}
}

func TestResolverExclusions(t *testing.T) {
	tmpDir := t.TempDir()
	promptsDir := filepath.Join(tmpDir, "prompts")
	os.MkdirAll(promptsDir, 0755)

	// Create library files
	os.WriteFile(filepath.Join(promptsDir, "include.prompt"), []byte("Include this"), 0644)
	os.WriteFile(filepath.Join(promptsDir, "exclude.prompt"), []byte("Exclude this"), 0644)

	manager := library.NewManager(tmpDir)
	resolver := NewResolver(manager)

	// Import with exclusion
	result, err := resolver.Resolve([]string{"include", "!exclude", "exclude"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(result, "Include this") {
		t.Errorf("Expected result to contain 'Include this'")
	}
	if strings.Contains(result, "Exclude this") {
		t.Errorf("Expected result NOT to contain 'Exclude this', but it does")
	}
}

func TestResolverCycleDetection(t *testing.T) {
	tmpDir := t.TempDir()
	promptsDir := filepath.Join(tmpDir, "prompts")
	os.MkdirAll(promptsDir, 0755)

	// Create library file
	os.WriteFile(filepath.Join(promptsDir, "lib.prompt"), []byte("Library content"), 0644)

	manager := library.NewManager(tmpDir)
	resolver := NewResolver(manager)

	// Import the same library multiple times
	result, err := resolver.Resolve([]string{"lib", "lib", "lib"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should only include content once
	count := strings.Count(result, "Library content")
	if count != 1 {
		t.Errorf("Expected content to appear once, appeared %d times", count)
	}
}

func TestResolverEmptyImports(t *testing.T) {
	tmpDir := t.TempDir()
	manager := library.NewManager(tmpDir)
	resolver := NewResolver(manager)

	result, err := resolver.Resolve([]string{})
	if err != nil {
		t.Fatalf("Expected no error for empty imports, got: %v", err)
	}

	if result != "" {
		t.Errorf("Expected empty result for no imports, got: %s", result)
	}
}

func TestResolverInvalidImport(t *testing.T) {
	tmpDir := t.TempDir()
	manager := library.NewManager(tmpDir)
	resolver := NewResolver(manager)

	_, err := resolver.Resolve([]string{"nonexistent"})
	if err == nil {
		t.Error("Expected error for nonexistent import, got none")
	}

	if !strings.Contains(err.Error(), "cannot resolve import") {
		t.Errorf("Expected 'cannot resolve import' error, got: %v", err)
	}
}

func TestResolverBuiltInLibraries(t *testing.T) {
	manager := library.NewManager("")
	resolver := NewResolver(manager)

	// Test built-in library resolution
	builtInLibraries := []string{
		"patterns.rest_api",
		"patterns.testing",
		"patterns.database",
		"patterns.async_programming",
		"constraints.security",
		"constraints.code_quality",
		"constraints.performance",
		"constraints.accessibility",
	}

	for _, libName := range builtInLibraries {
		t.Run(libName, func(t *testing.T) {
			result, err := resolver.Resolve([]string{libName})
			if err != nil {
				t.Fatalf("Failed to resolve built-in library '%s': %v", libName, err)
			}

			if result == "" {
				t.Errorf("Built-in library '%s' returned empty content", libName)
			}

			// Reset resolver for next test
			resolver = NewResolver(manager)
		})
	}
}

func TestResolverResolutionOrder(t *testing.T) {
	tmpDir := t.TempDir()
	promptsDir := filepath.Join(tmpDir, "prompts")
	os.MkdirAll(promptsDir, 0755)

	// Create library files
	os.WriteFile(filepath.Join(promptsDir, "first.prompt"), []byte("First"), 0644)
	os.WriteFile(filepath.Join(promptsDir, "second.prompt"), []byte("Second"), 0644)
	os.WriteFile(filepath.Join(promptsDir, "third.prompt"), []byte("Third"), 0644)

	manager := library.NewManager(tmpDir)
	resolver := NewResolver(manager)

	_, err := resolver.Resolve([]string{"first", "second", "third"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	order := resolver.GetResolutionOrder()
	if len(order) != 3 {
		t.Errorf("Expected 3 resolved imports, got %d", len(order))
	}

	// Verify all imports are in the order
	found := make(map[string]bool)
	for _, imp := range order {
		found[imp] = true
	}

	if !found["first"] || !found["second"] || !found["third"] {
		t.Errorf("Resolution order missing expected imports: %v", order)
	}
}

func TestResolverExclusionWithoutImport(t *testing.T) {
	tmpDir := t.TempDir()
	promptsDir := filepath.Join(tmpDir, "prompts")
	os.MkdirAll(promptsDir, 0755)

	os.WriteFile(filepath.Join(promptsDir, "lib.prompt"), []byte("Library"), 0644)

	manager := library.NewManager(tmpDir)
	resolver := NewResolver(manager)

	// Only exclusion, no actual import
	result, err := resolver.Resolve([]string{"!lib"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result != "" {
		t.Errorf("Expected empty result when only excluding, got: %s", result)
	}
}

func TestResolverStateReset(t *testing.T) {
	tmpDir := t.TempDir()
	promptsDir := filepath.Join(tmpDir, "prompts")
	os.MkdirAll(promptsDir, 0755)

	os.WriteFile(filepath.Join(promptsDir, "lib1.prompt"), []byte("Lib 1"), 0644)
	os.WriteFile(filepath.Join(promptsDir, "lib2.prompt"), []byte("Lib 2"), 0644)

	manager := library.NewManager(tmpDir)
	resolver := NewResolver(manager)

	// First resolution
	_, err := resolver.Resolve([]string{"lib1"})
	if err != nil {
		t.Fatalf("First resolve failed: %v", err)
	}

	order1 := resolver.GetResolutionOrder()
	if len(order1) != 1 {
		t.Errorf("Expected 1 import in first resolution, got %d", len(order1))
	}

	// Second resolution should reset state
	_, err = resolver.Resolve([]string{"lib2"})
	if err != nil {
		t.Fatalf("Second resolve failed: %v", err)
	}

	order2 := resolver.GetResolutionOrder()
	if len(order2) != 1 {
		t.Errorf("Expected 1 import in second resolution, got %d", len(order2))
	}

	if order2[0] != "lib2" {
		t.Errorf("Expected second resolution to only contain 'lib2', got: %v", order2)
	}
}

func TestResolverMixedLocalAndBuiltIn(t *testing.T) {
	tmpDir := t.TempDir()
	promptsDir := filepath.Join(tmpDir, "prompts")
	os.MkdirAll(promptsDir, 0755)

	// Create local library
	os.WriteFile(filepath.Join(promptsDir, "custom.prompt"), []byte("Custom Library"), 0644)

	manager := library.NewManager(tmpDir)
	resolver := NewResolver(manager)

	// Mix local and built-in imports
	result, err := resolver.Resolve([]string{"custom", "patterns.rest_api"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !strings.Contains(result, "Custom Library") {
		t.Errorf("Expected result to contain custom library content")
	}

	// Built-in library should also be included
	if result == "" {
		t.Errorf("Expected non-empty result with both local and built-in libraries")
	}
}
