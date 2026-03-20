package validator

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// validNameRe matches alphanumeric characters plus hyphens. Used for resource names.
var validNameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-]*$`)

// Validate validates a prompt specification.
func Validate(spec map[string]interface{}) (bool, []string) {
	var errors []string

	// Check for valid top-level keys
	validKeys := map[string]bool{
		"imports":     true,
		"context":     true,
		"features":    true,
		"constraints": true,
		"resources":   true,
		"build":       true,
	}

	for key := range spec {
		if !validKeys[key] {
			errors = append(errors, fmt.Sprintf(
				"unknown top-level key: %q. Valid keys: imports, context, features, constraints, resources, build",
				key,
			))
		}
	}

	if imports, ok := spec["imports"]; ok {
		errors = append(errors, validateImports(imports)...)
	}
	if context, ok := spec["context"]; ok {
		errors = append(errors, validateContext(context)...)
	}
	if features, ok := spec["features"]; ok {
		errors = append(errors, validateFeatures(features)...)
	}
	if constraints, ok := spec["constraints"]; ok {
		errors = append(errors, validateConstraints(constraints)...)
	}
	if resources, ok := spec["resources"]; ok {
		errors = append(errors, validateResources(resources)...)
	}
	if build, ok := spec["build"]; ok {
		errors = append(errors, validateBuild(build)...)
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
			errors = append(errors, fmt.Sprintf("import at index %d must be a string", i))
			continue
		}
		if impStr != "" && !strings.HasPrefix(impStr, "!") {
			if !isValidImportName(impStr) {
				errors = append(errors, fmt.Sprintf(
					"invalid import name %q: use only letters, numbers, dots, underscores, and hyphens",
					impStr,
				))
			}
		}
	}

	return errors
}

func validateContext(context interface{}) []string {
	contextMap, ok := context.(map[string]interface{})
	if !ok {
		return []string{"'context' must be a dictionary"}
	}

	var errors []string
	for key, value := range contextMap {
		switch value.(type) {
		case string, int, int64, float64, bool, nil:
			// valid
		default:
			errors = append(errors, fmt.Sprintf(
				"context value for %q must be a simple type (string, number, boolean)", key,
			))
		}
	}
	return errors
}

func validateFeatures(features interface{}) []string {
	featuresList, ok := features.([]interface{})
	if !ok {
		return []string{"'features' must be a list"}
	}
	if len(featuresList) == 0 {
		return []string{"'features' list is empty — add at least one feature"}
	}

	var errors []string
	for i, feature := range featuresList {
		switch v := feature.(type) {
		case map[string]interface{}:
			if len(v) != 1 {
				errors = append(errors, fmt.Sprintf("feature at index %d must have exactly one key-value pair", i))
			}
			for _, value := range v {
				if _, ok := value.(string); !ok {
					errors = append(errors, fmt.Sprintf("feature description at index %d must be a string", i))
				}
			}
		case string:
			if strings.TrimSpace(v) == "" {
				errors = append(errors, fmt.Sprintf("feature at index %d is an empty string", i))
			}
		default:
			errors = append(errors, fmt.Sprintf("feature at index %d must be a string or single-key dictionary", i))
		}
	}
	return errors
}

func validateConstraints(constraints interface{}) []string {
	constraintsList, ok := constraints.([]interface{})
	if !ok {
		return []string{"'constraints' must be a list"}
	}

	var errors []string
	for i, constraint := range constraintsList {
		switch v := constraint.(type) {
		case map[string]interface{}:
			for key, value := range v {
				if key == "exclude" {
					continue
				}
				switch value.(type) {
				case string, bool:
					// valid
				default:
					errors = append(errors, fmt.Sprintf(
						"constraint value for %q must be a string or boolean", key,
					))
				}
			}
		case string:
			if strings.TrimSpace(v) == "" {
				errors = append(errors, fmt.Sprintf("constraint at index %d is an empty string", i))
			}
		default:
			errors = append(errors, fmt.Sprintf("constraint at index %d must be a string or dictionary", i))
		}
	}
	return errors
}

func validateResources(resources interface{}) []string {
	list, ok := resources.([]interface{})
	if !ok {
		return []string{"'resources' must be a list"}
	}

	var errors []string
	for i, item := range list {
		m, ok := item.(map[string]interface{})
		if !ok {
			errors = append(errors, fmt.Sprintf("resource at index %d must be a dictionary", i))
			continue
		}

		// Each resource must have exactly one of url or git
		_, hasURL := m["url"]
		_, hasGit := m["git"]
		if hasURL && hasGit {
			errors = append(errors, fmt.Sprintf("resource at index %d must have either 'url' or 'git', not both", i))
		} else if !hasURL && !hasGit {
			errors = append(errors, fmt.Sprintf("resource at index %d must have either 'url' or 'git'", i))
		}

		// name is required
		nameVal, hasName := m["name"]
		if !hasName {
			errors = append(errors, fmt.Sprintf("resource at index %d is missing required 'name' field", i))
		} else if name, ok := nameVal.(string); !ok {
			errors = append(errors, fmt.Sprintf("resource at index %d 'name' must be a string", i))
		} else if !isValidResourceName(name) {
			errors = append(errors, fmt.Sprintf(
				"resource at index %d 'name' %q must contain only alphanumeric characters and hyphens",
				i, name,
			))
		}

		// ref and path are optional strings
		for _, optField := range []string{"ref", "path"} {
			if val, ok := m[optField]; ok {
				if _, ok := val.(string); !ok {
					errors = append(errors, fmt.Sprintf("resource at index %d %q must be a string", i, optField))
				}
			}
		}

		// url must be a string if present
		if hasURL {
			if _, ok := m["url"].(string); !ok {
				errors = append(errors, fmt.Sprintf("resource at index %d 'url' must be a string", i))
			}
		}
		// git must be a string if present
		if hasGit {
			if _, ok := m["git"].(string); !ok {
				errors = append(errors, fmt.Sprintf("resource at index %d 'git' must be a string", i))
			}
		}
	}
	return errors
}

func validateBuild(build interface{}) []string {
	m, ok := build.(map[string]interface{})
	if !ok {
		return []string{"'build' must be a dictionary"}
	}

	var errors []string

	// Validate agent section
	if agentRaw, ok := m["agent"]; ok {
		agentMap, ok := agentRaw.(map[string]interface{})
		if !ok {
			errors = append(errors, "'build.agent' must be a dictionary")
		} else {
			// command is required
			cmdVal, hasCmd := agentMap["command"]
			if !hasCmd {
				errors = append(errors, "'build.agent.command' is required")
			} else if _, ok := cmdVal.(string); !ok {
				errors = append(errors, "'build.agent.command' must be a string")
			}

			// args is optional list of strings
			if argsRaw, ok := agentMap["args"]; ok {
				argsList, ok := argsRaw.([]interface{})
				if !ok {
					errors = append(errors, "'build.agent.args' must be a list")
				} else {
					for j, arg := range argsList {
						if _, ok := arg.(string); !ok {
							errors = append(errors, fmt.Sprintf("'build.agent.args[%d]' must be a string", j))
						}
					}
				}
			}
		}
	}

	// Validate sandbox section
	if sandboxRaw, ok := m["sandbox"]; ok {
		sandboxMap, ok := sandboxRaw.(map[string]interface{})
		if !ok {
			errors = append(errors, "'build.sandbox' must be a dictionary")
		} else {
			validTypes := map[string]bool{"docker": true, "bubblewrap": true, "none": true}
			typeVal, hasType := sandboxMap["type"]
			if !hasType {
				errors = append(errors, "'build.sandbox.type' is required")
			} else {
				typeStr, ok := typeVal.(string)
				switch {
				case !ok:
					errors = append(errors, "'build.sandbox.type' must be a string")
				case !validTypes[typeStr]:
					errors = append(errors, fmt.Sprintf(
						"'build.sandbox.type' %q is invalid; must be one of: docker, bubblewrap, none",
						typeStr,
					))
				case typeStr == "docker":
					// image is required for docker
					imageVal, hasImage := sandboxMap["image"]
					if !hasImage {
						errors = append(errors, "'build.sandbox.image' is required when sandbox type is docker")
					} else if _, ok := imageVal.(string); !ok {
						errors = append(errors, "'build.sandbox.image' must be a string")
					}
				}
			}
		}
	}

	return errors
}

func isValidImportName(name string) bool {
	for _, c := range name {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '.' && c != '_' && c != '-' {
			return false
		}
	}
	return true
}

func isValidResourceName(name string) bool {
	return validNameRe.MatchString(name)
}
