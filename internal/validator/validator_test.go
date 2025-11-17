package validator

import (
	"strings"
	"testing"
)

func TestValidateValidSpec(t *testing.T) {
	spec := map[string]interface{}{
		"imports": []interface{}{
			"patterns.rest_api",
			"constraints.security",
		},
		"context": map[string]interface{}{
			"language":  "go",
			"framework": "fiber",
		},
		"features": []interface{}{
			map[string]interface{}{"auth": "User authentication"},
			"basic feature",
		},
		"constraints": []interface{}{
			"no_hardcoded_secrets",
			map[string]interface{}{"testing": "required"},
		},
	}

	isValid, errors := Validate(spec)
	if !isValid {
		t.Errorf("Expected valid spec, got errors: %v", errors)
	}
	if len(errors) > 0 {
		t.Errorf("Expected no errors, got: %v", errors)
	}
}

func TestValidateEmptySpec(t *testing.T) {
	spec := map[string]interface{}{}

	isValid, errors := Validate(spec)
	if !isValid {
		t.Errorf("Expected valid empty spec, got errors: %v", errors)
	}
}

func TestValidateInvalidTopLevelKey(t *testing.T) {
	spec := map[string]interface{}{
		"invalid_key": "value",
		"features": []interface{}{
			"feature1",
		},
	}

	isValid, errors := Validate(spec)
	if isValid {
		t.Error("Expected invalid spec with unknown top-level key")
	}
	if len(errors) == 0 {
		t.Error("Expected errors for invalid top-level key")
	}

	found := false
	for _, err := range errors {
		if strings.Contains(err, "Unknown top-level key") && strings.Contains(err, "invalid_key") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about unknown top-level key, got: %v", errors)
	}
}

func TestValidateImportsNotList(t *testing.T) {
	spec := map[string]interface{}{
		"imports": "not a list",
	}

	isValid, errors := Validate(spec)
	if isValid {
		t.Error("Expected invalid spec when imports is not a list")
	}

	found := false
	for _, err := range errors {
		if strings.Contains(err, "imports") && strings.Contains(err, "must be a list") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about imports being a list, got: %v", errors)
	}
}

func TestValidateImportsInvalidString(t *testing.T) {
	spec := map[string]interface{}{
		"imports": []interface{}{
			"valid.import",
			123, // Invalid: not a string
		},
	}

	isValid, errors := Validate(spec)
	if isValid {
		t.Error("Expected invalid spec when import is not a string")
	}

	found := false
	for _, err := range errors {
		if strings.Contains(err, "must be a string") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about import being a string, got: %v", errors)
	}
}

func TestValidateImportsInvalidCharacters(t *testing.T) {
	tests := []struct {
		name       string
		importName string
	}{
		{"slash", "invalid/import"},
		{"backslash", "invalid\\import"},
		{"space", "invalid import"},
		{"special chars", "invalid@import"},
		{"path traversal", "../etc/passwd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := map[string]interface{}{
				"imports": []interface{}{tt.importName},
			}

			isValid, errors := Validate(spec)
			if isValid {
				t.Errorf("Expected invalid spec for import '%s'", tt.importName)
			}
			if len(errors) == 0 {
				t.Errorf("Expected errors for import '%s'", tt.importName)
			}
		})
	}
}

func TestValidateImportsExclusion(t *testing.T) {
	spec := map[string]interface{}{
		"imports": []interface{}{
			"valid.import",
			"!exclude.this", // Exclusion should be valid
		},
	}

	isValid, _ := Validate(spec)
	if !isValid {
		t.Error("Expected exclusions to be valid")
	}
}

func TestValidateContextNotMap(t *testing.T) {
	spec := map[string]interface{}{
		"context": "not a map",
	}

	isValid, errors := Validate(spec)
	if isValid {
		t.Error("Expected invalid spec when context is not a map")
	}

	found := false
	for _, err := range errors {
		if strings.Contains(err, "context") && strings.Contains(err, "must be a dictionary") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about context being a dictionary, got: %v", errors)
	}
}

func TestValidateContextInvalidValueType(t *testing.T) {
	spec := map[string]interface{}{
		"context": map[string]interface{}{
			"valid":   "string",
			"invalid": []string{"array", "not", "allowed"},
		},
	}

	isValid, errors := Validate(spec)
	if isValid {
		t.Error("Expected invalid spec when context has complex value")
	}

	found := false
	for _, err := range errors {
		if strings.Contains(err, "simple type") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about simple types, got: %v", errors)
	}
}

func TestValidateContextValidTypes(t *testing.T) {
	spec := map[string]interface{}{
		"context": map[string]interface{}{
			"string":  "value",
			"int":     42,
			"int64":   int64(100),
			"float":   3.14,
			"bool":    true,
			"nil":     nil,
		},
		"features": []interface{}{"feature"}, // Need at least one feature
	}

	isValid, errors := Validate(spec)
	if !isValid {
		t.Errorf("Expected valid spec with all valid context types, got errors: %v", errors)
	}
}

func TestValidateFeaturesNotList(t *testing.T) {
	spec := map[string]interface{}{
		"features": "not a list",
	}

	isValid, errors := Validate(spec)
	if isValid {
		t.Error("Expected invalid spec when features is not a list")
	}

	found := false
	for _, err := range errors {
		if strings.Contains(err, "features") && strings.Contains(err, "must be a list") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about features being a list, got: %v", errors)
	}
}

func TestValidateFeaturesEmpty(t *testing.T) {
	spec := map[string]interface{}{
		"features": []interface{}{},
	}

	isValid, errors := Validate(spec)
	if isValid {
		t.Error("Expected invalid spec when features list is empty")
	}

	found := false
	for _, err := range errors {
		if strings.Contains(err, "features") && strings.Contains(err, "empty") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about empty features, got: %v", errors)
	}
}

func TestValidateFeaturesInvalidMap(t *testing.T) {
	spec := map[string]interface{}{
		"features": []interface{}{
			map[string]interface{}{
				"key1": "value1",
				"key2": "value2", // Should have exactly one key
			},
		},
	}

	isValid, errors := Validate(spec)
	if isValid {
		t.Error("Expected invalid spec when feature map has multiple keys")
	}

	found := false
	for _, err := range errors {
		if strings.Contains(err, "exactly one key-value pair") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about exactly one key-value pair, got: %v", errors)
	}
}

func TestValidateFeaturesInvalidDescriptionType(t *testing.T) {
	spec := map[string]interface{}{
		"features": []interface{}{
			map[string]interface{}{
				"feature": 123, // Should be string
			},
		},
	}

	isValid, errors := Validate(spec)
	if isValid {
		t.Error("Expected invalid spec when feature description is not a string")
	}

	found := false
	for _, err := range errors {
		if strings.Contains(err, "description") && strings.Contains(err, "must be a string") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about description being a string, got: %v", errors)
	}
}

func TestValidateFeaturesEmptyString(t *testing.T) {
	spec := map[string]interface{}{
		"features": []interface{}{
			"",           // Empty string
			"   ",        // Whitespace only
			"valid feature",
		},
	}

	isValid, errors := Validate(spec)
	if isValid {
		t.Error("Expected invalid spec when feature is empty string")
	}

	if len(errors) < 2 {
		t.Errorf("Expected at least 2 errors for empty strings, got: %v", errors)
	}
}

func TestValidateFeaturesInvalidType(t *testing.T) {
	spec := map[string]interface{}{
		"features": []interface{}{
			123, // Not a string or map
		},
	}

	isValid, errors := Validate(spec)
	if isValid {
		t.Error("Expected invalid spec when feature is invalid type")
	}

	found := false
	for _, err := range errors {
		if strings.Contains(err, "must be a string or dictionary") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about string or dictionary, got: %v", errors)
	}
}

func TestValidateConstraintsNotList(t *testing.T) {
	spec := map[string]interface{}{
		"constraints": "not a list",
	}

	isValid, errors := Validate(spec)
	if isValid {
		t.Error("Expected invalid spec when constraints is not a list")
	}

	found := false
	for _, err := range errors {
		if strings.Contains(err, "constraints") && strings.Contains(err, "must be a list") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about constraints being a list, got: %v", errors)
	}
}

func TestValidateConstraintsExclude(t *testing.T) {
	spec := map[string]interface{}{
		"constraints": []interface{}{
			map[string]interface{}{
				"exclude": "some_value", // Exclude should be allowed
			},
		},
		"features": []interface{}{"feature"}, // Need at least one feature
	}

	isValid, errors := Validate(spec)
	if !isValid {
		t.Errorf("Expected valid spec with exclude constraint, got errors: %v", errors)
	}
}

func TestValidateConstraintsInvalidValueType(t *testing.T) {
	spec := map[string]interface{}{
		"constraints": []interface{}{
			map[string]interface{}{
				"constraint": 123, // Should be string or bool
			},
		},
	}

	isValid, errors := Validate(spec)
	if isValid {
		t.Error("Expected invalid spec when constraint value is invalid type")
	}

	found := false
	for _, err := range errors {
		if strings.Contains(err, "must be a string or boolean") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about string or boolean, got: %v", errors)
	}
}

func TestValidateConstraintsValidTypes(t *testing.T) {
	spec := map[string]interface{}{
		"constraints": []interface{}{
			"string_constraint",
			map[string]interface{}{
				"bool_constraint": true,
				"string_map":      "value",
			},
		},
		"features": []interface{}{"feature"}, // Need at least one feature
	}

	isValid, errors := Validate(spec)
	if !isValid {
		t.Errorf("Expected valid spec with valid constraint types, got errors: %v", errors)
	}
}

func TestValidateConstraintsEmptyString(t *testing.T) {
	spec := map[string]interface{}{
		"constraints": []interface{}{
			"",    // Empty string
			"   ", // Whitespace only
		},
	}

	isValid, errors := Validate(spec)
	if isValid {
		t.Error("Expected invalid spec when constraint is empty string")
	}

	if len(errors) < 2 {
		t.Errorf("Expected at least 2 errors for empty strings, got: %v", errors)
	}
}

func TestValidateConstraintsInvalidType(t *testing.T) {
	spec := map[string]interface{}{
		"constraints": []interface{}{
			123, // Not a string or map
		},
	}

	isValid, errors := Validate(spec)
	if isValid {
		t.Error("Expected invalid spec when constraint is invalid type")
	}

	found := false
	for _, err := range errors {
		if strings.Contains(err, "must be a string or dictionary") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected error about string or dictionary, got: %v", errors)
	}
}

func TestIsValidImportName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"valid simple", "patterns", true},
		{"valid dotted", "patterns.rest_api", true},
		{"valid underscores", "my_custom_lib", true},
		{"valid mixed", "lib_1.pattern_2", true},
		{"invalid slash", "lib/pattern", false},
		{"invalid backslash", "lib\\pattern", false},
		{"invalid space", "lib pattern", false},
		{"invalid special", "lib@pattern", false},
		{"invalid path traversal", "../lib", false},
		{"invalid absolute", "/lib/pattern", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidImportName(tt.input)
			if result != tt.valid {
				t.Errorf("isValidImportName(%s) = %v, want %v", tt.input, result, tt.valid)
			}
		})
	}
}

func TestValidateMultipleErrors(t *testing.T) {
	spec := map[string]interface{}{
		"invalid_key": "value",
		"imports":     "not a list",
		"context":     "not a map",
		"features":    "not a list",
		"constraints": "not a list",
	}

	isValid, errors := Validate(spec)
	if isValid {
		t.Error("Expected invalid spec with multiple errors")
	}

	// Should have at least 5 errors (one for each invalid field)
	if len(errors) < 5 {
		t.Errorf("Expected at least 5 errors, got %d: %v", len(errors), errors)
	}
}

func TestValidateMinimalValidSpec(t *testing.T) {
	spec := map[string]interface{}{
		"features": []interface{}{"single feature"},
	}

	isValid, errors := Validate(spec)
	if !isValid {
		t.Errorf("Expected valid minimal spec, got errors: %v", errors)
	}
}
