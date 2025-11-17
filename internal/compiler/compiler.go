package compiler

import (
	"fmt"
	"os"
	"strings"

	"github.com/Geogboe/promptc/internal/library"
	"github.com/Geogboe/promptc/internal/resolver"
	"github.com/Geogboe/promptc/internal/targets"
	"github.com/Geogboe/promptc/internal/validator"
	"gopkg.in/yaml.v3"
)

// Compiler compiles .prompt files into target-specific formats
type Compiler struct {
	LibraryManager *library.Manager
	Resolver       *resolver.Resolver
}

// NewCompiler creates a new prompt compiler
func NewCompiler(projectDir string) *Compiler {
	libraryManager := library.NewManager(projectDir)
	return &Compiler{
		LibraryManager: libraryManager,
		Resolver:       resolver.NewResolver(libraryManager),
	}
}

// CompileOptions contains options for compilation
type CompileOptions struct {
	Target   string
	Debug    bool
	Validate bool
}

// Compile compiles a prompt file to the target format
func (c *Compiler) Compile(promptFile string, opts *CompileOptions) (string, error) {
	if opts == nil {
		opts = &CompileOptions{
			Target:   "raw",
			Validate: true,
		}
	}

	// Check if target is supported
	_, err := targets.GetFormatter(opts.Target)
	if err != nil {
		return "", err
	}

	// Load and parse prompt file
	spec, err := c.loadSpec(promptFile)
	if err != nil {
		return "", err
	}

	// Validate if requested
	if opts.Validate {
		isValid, errors := validator.Validate(spec)
		if !isValid {
			errorMsg := "Validation failed:\n"
			for _, e := range errors {
				errorMsg += fmt.Sprintf("  - %s\n", e)
			}
			return "", fmt.Errorf("%s", errorMsg)
		}
	}

	// Resolve imports
	var imports []string
	if importsRaw, ok := spec["imports"]; ok {
		if importsList, ok := importsRaw.([]interface{}); ok {
			for _, imp := range importsList {
				if impStr, ok := imp.(string); ok {
					imports = append(imports, impStr)
				}
			}
		}
	}

	importsContent := ""
	if len(imports) > 0 {
		importsContent, err = c.Resolver.Resolve(imports)
		if err != nil {
			return "", err
		}
	}

	if opts.Debug {
		fmt.Printf("[DEBUG] Resolved imports: %v\n", c.Resolver.GetResolutionOrder())
		fmt.Printf("[DEBUG] Imports content length: %d chars\n", len(importsContent))
	}

	// Extract other sections
	context := make(map[string]interface{})
	if ctx, ok := spec["context"].(map[string]interface{}); ok {
		context = ctx
	}

	var features []interface{}
	if feat, ok := spec["features"].([]interface{}); ok {
		features = feat
	}

	var constraints []interface{}
	if cons, ok := spec["constraints"].([]interface{}); ok {
		constraints = cons
	}

	// Get formatter and compile
	formatter, _ := targets.GetFormatter(opts.Target)
	compiled := formatter(importsContent, context, features, constraints)

	return compiled, nil
}

func (c *Compiler) loadSpec(promptFile string) (map[string]interface{}, error) {
	// Check if file exists and get file info
	fileInfo, err := os.Stat(promptFile)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("prompt file not found: %s", promptFile)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat prompt file: %w", err)
	}

	// Prevent YAML bombs and memory exhaustion with file size limit
	const maxPromptFileSize = 1 * 1024 * 1024 // 1MB limit for .prompt files
	if fileInfo.Size() > maxPromptFileSize {
		return nil, fmt.Errorf("prompt file too large: %d bytes (max %d bytes)", fileInfo.Size(), maxPromptFileSize)
	}

	// Read file
	content, err := os.ReadFile(promptFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read prompt file: %w", err)
	}

	// Parse YAML
	var spec map[string]interface{}
	if err := yaml.Unmarshal(content, &spec); err != nil {
		return nil, fmt.Errorf("invalid YAML in %s: %w", promptFile, err)
	}

	if spec == nil {
		spec = make(map[string]interface{})
	}

	return spec, nil
}

// Template represents a project template
type Template struct {
	Name    string
	Content string
}

// GetTemplates returns available project templates
func GetTemplates() map[string]string {
	return map[string]string{
		"basic": `# Basic prompt template
imports:
  - constraints.code_quality

context:
  language: python
  framework: ""

features:
  - feature_name: "Description of the feature"

constraints:
  - clear_documentation
  - comprehensive_error_handling
`,
		"web-api": `# Web API project template
imports:
  - patterns.rest_api
  - patterns.testing
  - constraints.security
  - constraints.code_quality

context:
  language: python
  framework: fastapi
  database: postgresql
  testing: pytest

features:
  - api_endpoint: "Description of your API endpoint"

constraints:
  - no_hardcoded_secrets
  - comprehensive_test_coverage
  - api_documentation
`,
		"cli-tool": `# CLI tool project template
imports:
  - patterns.testing
  - constraints.code_quality

context:
  language: python
  framework: click
  testing: pytest

features:
  - cli_command: "Description of your CLI command"

constraints:
  - clear_help_messages
  - comprehensive_test_coverage
  - cross_platform_compatibility
`,
	}
}

// InitProject initializes a new .prompt file from a template
func InitProject(name, templateName string) error {
	templates := GetTemplates()
	content, ok := templates[templateName]
	if !ok {
		return fmt.Errorf("unknown template '%s'. Available: %s",
			templateName, strings.Join(getTemplateNames(), ", "))
	}

	// Check if file already exists
	if _, err := os.Stat(name); err == nil {
		return fmt.Errorf("file %s already exists", name)
	}

	// Write file
	if err := os.WriteFile(name, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func getTemplateNames() []string {
	templates := GetTemplates()
	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}
	return names
}
