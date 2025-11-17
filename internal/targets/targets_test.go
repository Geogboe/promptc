package targets

import (
	"strings"
	"testing"
)

func TestGetFormatter(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		shouldError bool
	}{
		{"valid raw", "raw", false},
		{"valid claude", "claude", false},
		{"valid cursor", "cursor", false},
		{"valid aider", "aider", false},
		{"valid copilot", "copilot", false},
		{"invalid target", "invalid", true},
		{"empty target", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter, err := GetFormatter(tt.target)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for target '%s', got none", tt.target)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for target '%s', got: %v", tt.target, err)
				}
				if formatter == nil {
					t.Errorf("Expected formatter for target '%s', got nil", tt.target)
				}
			}
		})
	}
}

func TestGetSupportedTargets(t *testing.T) {
	targets := GetSupportedTargets()

	expectedTargets := []string{"raw", "claude", "cursor", "aider", "copilot"}
	if len(targets) != len(expectedTargets) {
		t.Errorf("Expected %d targets, got %d", len(expectedTargets), len(targets))
	}

	for _, expected := range expectedTargets {
		found := false
		for _, target := range targets {
			if target == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected target '%s' not found in supported targets", expected)
		}
	}
}

func TestFormatRaw(t *testing.T) {
	imports := "# Imported Guidelines\n- Use best practices"
	context := map[string]interface{}{
		"language":  "go",
		"framework": "fiber",
	}
	features := []interface{}{
		map[string]interface{}{"auth": "User authentication"},
		"basic feature",
	}
	constraints := []interface{}{
		"no_hardcoded_secrets",
		map[string]interface{}{"testing": "required"},
	}

	result := FormatRaw(imports, context, features, constraints)

	// Verify imports are included
	if !strings.Contains(result, "Imported Guidelines") {
		t.Error("Raw format should contain imports")
	}

	// Verify context is included
	if !strings.Contains(result, "## Context") {
		t.Error("Raw format should contain Context section")
	}
	if !strings.Contains(result, "language") {
		t.Error("Raw format should contain language")
	}

	// Verify features are included
	if !strings.Contains(result, "## Features") {
		t.Error("Raw format should contain Features section")
	}
	if !strings.Contains(result, "auth") {
		t.Error("Raw format should contain feature")
	}

	// Verify constraints are included
	if !strings.Contains(result, "## Constraints") {
		t.Error("Raw format should contain Constraints section")
	}
	if !strings.Contains(result, "no_hardcoded_secrets") {
		t.Error("Raw format should contain constraint")
	}
}

func TestFormatClaude(t *testing.T) {
	imports := "# Guidelines\nFollow best practices"
	context := map[string]interface{}{
		"language": "python",
	}
	features := []interface{}{
		map[string]interface{}{"api": "REST API"},
	}
	constraints := []interface{}{
		"security_first",
	}

	result := FormatClaude(imports, context, features, constraints)

	// Verify Claude-specific formatting
	if !strings.Contains(result, "# Project Context") {
		t.Error("Claude format should contain Project Context heading")
	}
	if !strings.Contains(result, "# Constraints and Requirements") {
		t.Error("Claude format should contain Constraints heading")
	}
	if !strings.Contains(result, "# Features to Implement") {
		t.Error("Claude format should contain Features heading")
	}

	// Verify bold formatting
	if !strings.Contains(result, "**language**") {
		t.Error("Claude format should use bold for context keys")
	}
}

func TestFormatCursor(t *testing.T) {
	imports := "Guidelines"
	context := map[string]interface{}{
		"language": "javascript",
	}
	features := []interface{}{
		"feature1",
	}
	constraints := []interface{}{
		"constraint1",
	}

	result := FormatCursor(imports, context, features, constraints)

	// Verify Cursor-specific formatting
	if !strings.Contains(result, "Project Context:") {
		t.Error("Cursor format should contain 'Project Context:'")
	}
	if !strings.Contains(result, "Constraints:") {
		t.Error("Cursor format should contain 'Constraints:'")
	}
	if !strings.Contains(result, "Features to implement:") {
		t.Error("Cursor format should contain 'Features to implement:'")
	}
}

func TestFormatAider(t *testing.T) {
	imports := "Guidelines content"
	context := map[string]interface{}{
		"language": "rust",
	}
	features := []interface{}{
		"feature",
	}
	constraints := []interface{}{
		"constraint",
	}

	result := FormatAider(imports, context, features, constraints)

	// Verify Aider-specific formatting
	if !strings.Contains(result, "# Aider Instructions") {
		t.Error("Aider format should contain title")
	}
	if !strings.Contains(result, "## Guidelines") {
		t.Error("Aider format should contain Guidelines section")
	}
	if !strings.Contains(result, "## Project Context") {
		t.Error("Aider format should contain Project Context section")
	}
	if !strings.Contains(result, "## Requirements and Constraints") {
		t.Error("Aider format should contain Requirements section")
	}
	if !strings.Contains(result, "## Features to Implement") {
		t.Error("Aider format should contain Features section")
	}
}

func TestFormatCopilot(t *testing.T) {
	imports := "Development guidelines"
	context := map[string]interface{}{
		"language": "typescript",
		"testing":  "jest",
	}
	features := []interface{}{
		map[string]interface{}{"feature": "Description"},
	}
	constraints := []interface{}{
		"code_quality",
	}

	result := FormatCopilot(imports, context, features, constraints)

	// Verify Copilot-specific formatting
	if !strings.Contains(result, "# GitHub Copilot Instructions") {
		t.Error("Copilot format should contain title")
	}
	if !strings.Contains(result, "## Development Guidelines") {
		t.Error("Copilot format should contain Guidelines section")
	}
	if !strings.Contains(result, "## Project Information") {
		t.Error("Copilot format should contain Project Information section")
	}
	if !strings.Contains(result, "## Coding Standards and Constraints") {
		t.Error("Copilot format should contain Standards section")
	}
	if !strings.Contains(result, "## Current Development Focus") {
		t.Error("Copilot format should contain Development Focus section")
	}
}

func TestFormatRawEmptyContent(t *testing.T) {
	result := FormatRaw("", map[string]interface{}{}, []interface{}{}, []interface{}{})

	if result != "" {
		t.Errorf("Expected empty result for empty input, got: %s", result)
	}
}

func TestFormatRawOnlyImports(t *testing.T) {
	imports := "Only imports content"
	result := FormatRaw(imports, map[string]interface{}{}, []interface{}{}, []interface{}{})

	if !strings.Contains(result, "Only imports") {
		t.Error("Should contain imports content")
	}
	if strings.Contains(result, "##") {
		t.Error("Should not contain section headers when sections are empty")
	}
}

func TestFormatRawExcludeConstraints(t *testing.T) {
	constraints := []interface{}{
		"include_this",
		map[string]interface{}{"exclude": "exclude_value"},
	}

	result := FormatRaw("", map[string]interface{}{}, []interface{}{}, constraints)

	if !strings.Contains(result, "include_this") {
		t.Error("Should include regular constraint")
	}
	if strings.Contains(result, "exclude:") || strings.Contains(result, "exclude_value") {
		t.Error("Should not include excluded constraints")
	}
}

func TestFormatListItem(t *testing.T) {
	tests := []struct {
		name     string
		item     interface{}
		expected string
	}{
		{
			name:     "string item",
			item:     "test",
			expected: "- test",
		},
		{
			name:     "map item",
			item:     map[string]interface{}{"key": "value"},
			expected: "- key: value",
		},
		{
			name:     "number item",
			item:     42,
			expected: "- 42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatListItem(tt.item)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFormatFeatureItem(t *testing.T) {
	tests := []struct {
		name     string
		item     interface{}
		contains string
	}{
		{
			name:     "string feature",
			item:     "feature",
			contains: "- feature",
		},
		{
			name:     "map feature with bold",
			item:     map[string]interface{}{"name": "desc"},
			contains: "**name**",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFeatureItem(tt.item)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected to contain %s, got %s", tt.contains, result)
			}
		})
	}
}

func TestFormatConstraintItem(t *testing.T) {
	tests := []struct {
		name     string
		item     interface{}
		expected string
	}{
		{
			name:     "string constraint with title case",
			item:     "no_hardcoded_secrets",
			expected: "No Hardcoded Secrets",
		},
		{
			name:     "exclude constraint returns empty",
			item:     map[string]interface{}{"exclude": "value"},
			expected: "",
		},
		{
			name:     "map constraint with title case",
			item:     map[string]interface{}{"test_key": "value"},
			expected: "Test Key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatConstraintItem(tt.item)
			if tt.expected == "" {
				if result != "" {
					t.Errorf("Expected empty result, got: %s", result)
				}
			} else {
				if !strings.Contains(result, tt.expected) {
					t.Errorf("Expected to contain %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestAllFormattersHandleNilContext(t *testing.T) {
	formatters := map[string]Formatter{
		"raw":     FormatRaw,
		"claude":  FormatClaude,
		"cursor":  FormatCursor,
		"aider":   FormatAider,
		"copilot": FormatCopilot,
	}

	for name, formatter := range formatters {
		t.Run(name, func(t *testing.T) {
			result := formatter("", map[string]interface{}{}, []interface{}{}, []interface{}{})
			// Should not panic and should return a valid string
			if result == "" {
				// Empty is acceptable for empty input
			}
		})
	}
}

func TestClaudeFormatterOrdering(t *testing.T) {
	imports := "Imports"
	context := map[string]interface{}{"lang": "go"}
	features := []interface{}{"feat"}
	constraints := []interface{}{"const"}

	result := FormatClaude(imports, context, features, constraints)

	// Find positions of each section
	importsPos := strings.Index(result, "Imports")
	contextPos := strings.Index(result, "# Project Context")
	constraintsPos := strings.Index(result, "# Constraints and Requirements")
	featuresPos := strings.Index(result, "# Features to Implement")

	// Verify ordering: imports -> context -> constraints -> features
	if importsPos == -1 || contextPos == -1 || constraintsPos == -1 || featuresPos == -1 {
		t.Error("Missing expected sections in Claude format")
	}

	if !(importsPos < contextPos && contextPos < constraintsPos && constraintsPos < featuresPos) {
		t.Error("Claude format sections are not in expected order")
	}
}

func TestCopilotFormatterTitleCase(t *testing.T) {
	context := map[string]interface{}{
		"programming_language": "go",
		"test_framework":       "testing",
	}

	result := FormatCopilot("", context, []interface{}{}, []interface{}{})

	// Should convert underscores to spaces and title case
	if !strings.Contains(result, "Programming Language") {
		t.Error("Expected title case conversion of context keys")
	}
	if !strings.Contains(result, "Test Framework") {
		t.Error("Expected title case conversion of context keys")
	}
}
