package targets

import (
	"fmt"
	"strings"
)

// Formatter is a function that formats a prompt for a specific target
type Formatter func(importsContent string, context map[string]interface{}, features, constraints []interface{}) string

// GetFormatter returns the formatter for a given target
func GetFormatter(target string) (Formatter, error) {
	formatters := map[string]Formatter{
		"raw":     FormatRaw,
		"claude":  FormatClaude,
		"cursor":  FormatCursor,
		"aider":   FormatAider,
		"copilot": FormatCopilot,
	}

	formatter, ok := formatters[target]
	if !ok {
		return nil, fmt.Errorf("unsupported target '%s'. Supported targets: %s",
			target, strings.Join(GetSupportedTargets(), ", "))
	}

	return formatter, nil
}

// GetSupportedTargets returns a list of supported target formats
func GetSupportedTargets() []string {
	return []string{"raw", "claude", "cursor", "aider", "copilot"}
}

// FormatRaw outputs combined prompt as plain text
func FormatRaw(importsContent string, context map[string]interface{}, features, constraints []interface{}) string {
	var parts []string

	if importsContent != "" {
		parts = append(parts, importsContent)
	}

	if len(context) > 0 {
		parts = append(parts, "## Context")
		for key, value := range context {
			parts = append(parts, fmt.Sprintf("- %s: %v", key, value))
		}
	}

	if len(features) > 0 {
		parts = append(parts, "## Features")
		for _, feature := range features {
			parts = append(parts, formatListItem(feature))
		}
	}

	if len(constraints) > 0 {
		parts = append(parts, "## Constraints")
		for _, constraint := range constraints {
			item := formatListItem(constraint)
			if !strings.Contains(item, "exclude:") {
				parts = append(parts, item)
			}
		}
	}

	return strings.Join(parts, "\n\n")
}

// FormatClaude outputs prompt in Claude-friendly markdown format
func FormatClaude(importsContent string, context map[string]interface{}, features, constraints []interface{}) string {
	var parts []string

	if importsContent != "" {
		parts = append(parts, importsContent)
	}

	if len(context) > 0 {
		parts = append(parts, "# Project Context")
		var items []string
		for key, value := range context {
			items = append(items, fmt.Sprintf("- **%s**: %v", key, value))
		}
		parts = append(parts, strings.Join(items, "\n"))
	}

	if len(constraints) > 0 {
		parts = append(parts, "# Constraints and Requirements")
		var items []string
		for _, constraint := range constraints {
			item := formatConstraintItem(constraint)
			if item != "" {
				items = append(items, item)
			}
		}
		parts = append(parts, strings.Join(items, "\n"))
	}

	if len(features) > 0 {
		parts = append(parts, "# Features to Implement")
		var items []string
		for _, feature := range features {
			items = append(items, formatFeatureItem(feature))
		}
		parts = append(parts, strings.Join(items, "\n"))
	}

	return strings.Join(parts, "\n\n")
}

// FormatCursor outputs prompt in Cursor rules format
func FormatCursor(importsContent string, context map[string]interface{}, features, constraints []interface{}) string {
	var parts []string

	if importsContent != "" {
		parts = append(parts, importsContent)
	}

	if len(context) > 0 {
		parts = append(parts, "Project Context:")
		for key, value := range context {
			parts = append(parts, fmt.Sprintf("- %s: %v", key, value))
		}
	}

	if len(constraints) > 0 {
		parts = append(parts, "Constraints:")
		for _, constraint := range constraints {
			item := formatListItem(constraint)
			if !strings.Contains(item, "exclude:") {
				parts = append(parts, item)
			}
		}
	}

	if len(features) > 0 {
		parts = append(parts, "Features to implement:")
		for _, feature := range features {
			parts = append(parts, formatListItem(feature))
		}
	}

	return strings.Join(parts, "\n\n")
}

// FormatAider outputs prompt in Aider configuration format
func FormatAider(importsContent string, context map[string]interface{}, features, constraints []interface{}) string {
	var parts []string

	parts = append(parts, "# Aider Instructions")
	parts = append(parts, "")

	if importsContent != "" {
		parts = append(parts, "## Guidelines")
		parts = append(parts, "")
		parts = append(parts, importsContent)
		parts = append(parts, "")
	}

	if len(context) > 0 {
		parts = append(parts, "## Project Context")
		parts = append(parts, "")
		for key, value := range context {
			parts = append(parts, fmt.Sprintf("- %s: %v", key, value))
		}
		parts = append(parts, "")
	}

	if len(constraints) > 0 {
		parts = append(parts, "## Requirements and Constraints")
		parts = append(parts, "")
		for _, constraint := range constraints {
			item := formatConstraintItem(constraint)
			if item != "" {
				parts = append(parts, item)
			}
		}
		parts = append(parts, "")
	}

	if len(features) > 0 {
		parts = append(parts, "## Features to Implement")
		parts = append(parts, "")
		for _, feature := range features {
			parts = append(parts, formatListItem(feature))
		}
	}

	return strings.Join(parts, "\n")
}

// FormatCopilot outputs prompt for GitHub Copilot
func FormatCopilot(importsContent string, context map[string]interface{}, features, constraints []interface{}) string {
	var parts []string

	parts = append(parts, "# GitHub Copilot Instructions")
	parts = append(parts, "")
	parts = append(parts, "This document provides instructions for GitHub Copilot when working on this project.")
	parts = append(parts, "")

	if importsContent != "" {
		parts = append(parts, "## Development Guidelines")
		parts = append(parts, "")
		parts = append(parts, importsContent)
		parts = append(parts, "")
	}

	if len(context) > 0 {
		parts = append(parts, "## Project Information")
		parts = append(parts, "")
		parts = append(parts, "This project uses:")
		for key, value := range context {
			title := strings.ReplaceAll(key, "_", " ")
			title = strings.Title(title)
			parts = append(parts, fmt.Sprintf("- **%s**: %v", title, value))
		}
		parts = append(parts, "")
	}

	if len(constraints) > 0 {
		parts = append(parts, "## Coding Standards and Constraints")
		parts = append(parts, "")
		parts = append(parts, "When generating code, adhere to these requirements:")
		parts = append(parts, "")
		for _, constraint := range constraints {
			item := formatConstraintItem(constraint)
			if item != "" {
				parts = append(parts, item)
			}
		}
		parts = append(parts, "")
	}

	if len(features) > 0 {
		parts = append(parts, "## Current Development Focus")
		parts = append(parts, "")
		parts = append(parts, "The following features are being implemented:")
		parts = append(parts, "")
		for _, feature := range features {
			if m, ok := feature.(map[string]interface{}); ok {
				for name, desc := range m {
					parts = append(parts, fmt.Sprintf("### %s", name))
					parts = append(parts, fmt.Sprintf("%v", desc))
					parts = append(parts, "")
				}
			} else {
				parts = append(parts, fmt.Sprintf("- %v", feature))
			}
		}
	}

	return strings.Join(parts, "\n")
}

func formatListItem(item interface{}) string {
	switch v := item.(type) {
	case map[string]interface{}:
		for key, value := range v {
			return fmt.Sprintf("- %s: %v", key, value)
		}
	case string:
		return fmt.Sprintf("- %s", v)
	default:
		return fmt.Sprintf("- %v", v)
	}
	return ""
}

func formatFeatureItem(item interface{}) string {
	switch v := item.(type) {
	case map[string]interface{}:
		for key, value := range v {
			return fmt.Sprintf("- **%s**: %v", key, value)
		}
	case string:
		return fmt.Sprintf("- %s", v)
	default:
		return fmt.Sprintf("- %v", v)
	}
	return ""
}

func formatConstraintItem(item interface{}) string {
	switch v := item.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if key == "exclude" {
				return ""
			}
			title := strings.ReplaceAll(key, "_", " ")
			title = strings.Title(title)
			return fmt.Sprintf("- %s: %v", title, value)
		}
	case string:
		title := strings.ReplaceAll(v, "_", " ")
		title = strings.Title(title)
		return fmt.Sprintf("- %s", title)
	default:
		return fmt.Sprintf("- %v", v)
	}
	return ""
}
