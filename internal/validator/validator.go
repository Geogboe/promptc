package validator

import (
	"fmt"
	"strings"
	"unicode"
)

// Validate validates a prompt specification
func Validate(spec map[string]interface{}) (bool, []string) {
	var errors []string

	// Check for valid top-level keys
	validKeys := map[string]bool{
		"imports":     true,
		"context":     true,
		"features":    true,
		"constraints": true,
	}

	for key := range spec {
		if !validKeys[key] {
			errors = append(errors, fmt.Sprintf("Unknown top-level key: '%s'. Valid keys are: imports, context, features, constraints", key))
		}
	}

	// Validate imports
	if imports, ok := spec["imports"]; ok {
		errors = append(errors, validateImports(imports)...)
	}

	// Validate context
	if context, ok := spec["context"]; ok {
		errors = append(errors, validateContext(context)...)
	}

	// Validate features
	if features, ok := spec["features"]; ok {
		errors = append(errors, validateFeatures(features)...)
	}

	// Validate constraints
	if constraints, ok := spec["constraints"]; ok {
		errors = append(errors, validateConstraints(constraints)...)
	}

	return len(errors) == 0, errors
}

func validateImports(imports interface{}) []string {
	var errors []string

	importsList, ok := imports.([]interface{})
	if !ok {
		return []string{"'imports' must be a list"}
	}

	for i, imp := range importsList {
		impStr, ok := imp.(string)
		if !ok {
			errors = append(errors, fmt.Sprintf("Import at index %d must be a string", i))
			continue
		}

		if impStr != "" && !strings.HasPrefix(impStr, "!") {
			// Check for valid import name format
			if !isValidImportName(impStr) {
				errors = append(errors, fmt.Sprintf("Invalid import name '%s': use only letters, numbers, dots, and underscores", impStr))
			}
		}
	}

	return errors
}

func validateContext(context interface{}) []string {
	var errors []string

	contextMap, ok := context.(map[string]interface{})
	if !ok {
		return []string{"'context' must be a dictionary"}
	}

	// Context values should be simple types
	for key, value := range contextMap {
		switch value.(type) {
		case string, int, int64, float64, bool, nil:
			// Valid types
		default:
			errors = append(errors, fmt.Sprintf("Context value for '%s' must be a simple type (string, number, boolean)", key))
		}
	}

	return errors
}

func validateFeatures(features interface{}) []string {
	var errors []string

	featuresList, ok := features.([]interface{})
	if !ok {
		return []string{"'features' must be a list"}
	}

	if len(featuresList) == 0 {
		return []string{"'features' list is empty - add at least one feature"}
	}

	for i, feature := range featuresList {
		switch v := feature.(type) {
		case map[string]interface{}:
			if len(v) != 1 {
				errors = append(errors, fmt.Sprintf("Feature at index %d should have exactly one key-value pair", i))
			}
			for _, value := range v {
				if _, ok := value.(string); !ok {
					errors = append(errors, fmt.Sprintf("Feature description at index %d must be a string", i))
				}
			}
		case string:
			if strings.TrimSpace(v) == "" {
				errors = append(errors, fmt.Sprintf("Feature at index %d is an empty string", i))
			}
		default:
			errors = append(errors, fmt.Sprintf("Feature at index %d must be a string or dictionary", i))
		}
	}

	return errors
}

func validateConstraints(constraints interface{}) []string {
	var errors []string

	constraintsList, ok := constraints.([]interface{})
	if !ok {
		return []string{"'constraints' must be a list"}
	}

	for i, constraint := range constraintsList {
		switch v := constraint.(type) {
		case map[string]interface{}:
			for key, value := range v {
				if key == "exclude" {
					continue
				}
				switch value.(type) {
				case string, bool:
					// Valid types
				default:
					errors = append(errors, fmt.Sprintf("Constraint value for '%s' must be a string or boolean", key))
				}
			}
		case string:
			if strings.TrimSpace(v) == "" {
				errors = append(errors, fmt.Sprintf("Constraint at index %d is an empty string", i))
			}
		default:
			errors = append(errors, fmt.Sprintf("Constraint at index %d must be a string or dictionary", i))
		}
	}

	return errors
}

func isValidImportName(name string) bool {
	for _, c := range name {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '.' && c != '_' {
			return false
		}
	}
	return true
}
