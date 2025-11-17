package library

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	manager := NewManager("")
	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if len(manager.SearchPaths) == 0 {
		t.Error("SearchPaths is empty")
	}
}

func TestResolveBuiltInLibrary(t *testing.T) {
	manager := NewManager("")

	content, err := manager.Resolve("patterns.rest_api")
	if err != nil {
		t.Fatalf("Failed to resolve built-in library: %v", err)
	}

	if content == "" {
		t.Error("Resolved content is empty")
	}

	if !contains(content, "REST API") {
		t.Error("Content does not contain expected text")
	}
}

func TestResolveNonexistentLibrary(t *testing.T) {
	manager := NewManager("")

	_, err := manager.Resolve("nonexistent.library")
	if err == nil {
		t.Error("Expected error for nonexistent library")
	}
}

func TestListLibraries(t *testing.T) {
	manager := NewManager("")

	libs := manager.ListLibraries()
	if libs == nil {
		t.Fatal("ListLibraries returned nil")
	}

	if len(libs.BuiltIn) == 0 {
		t.Error("No built-in libraries found")
	}

	// Check for expected libraries
	expectedLibs := []string{
		"patterns.rest_api",
		"patterns.testing",
		"patterns.database",
		"patterns.async_programming",
		"constraints.security",
		"constraints.code_quality",
		"constraints.performance",
		"constraints.accessibility",
	}

	for _, expected := range expectedLibs {
		found := false
		for _, lib := range libs.BuiltIn {
			if lib == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected library not found: %s", expected)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
