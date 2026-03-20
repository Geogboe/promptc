// Package compiler compiles .spec.promptc files into instructions.promptc output.
package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Geogboe/promptc/internal/formatter"
	"github.com/Geogboe/promptc/internal/library"
	"github.com/Geogboe/promptc/internal/resolver"
	"github.com/Geogboe/promptc/internal/validator"
	"gopkg.in/yaml.v3"
)

// Version is injected at build time via -ldflags.
var Version = "dev"

// ParsedSpec is the fully structured representation of a .spec.promptc file.
type ParsedSpec struct {
	Imports     []string
	Context     map[string]interface{}
	Features    []interface{}
	Constraints []interface{}
	Resources   []SpecResource
	Build       *SpecBuild
}

// SpecResource represents one entry in the resources section.
type SpecResource struct {
	Name string
	URL  string
	Git  string
	Ref  string
	Path string
}

// SpecBuild holds the build pipeline configuration.
type SpecBuild struct {
	Agent   SpecAgent
	Sandbox SpecSandbox
}

// SpecAgent describes the agent CLI to invoke.
type SpecAgent struct {
	Command string
	Args    []string
}

// SpecSandbox describes the sandbox environment.
type SpecSandbox struct {
	Type  string
	Image string
}

// CompileOptions controls compiler behaviour.
type CompileOptions struct {
	OutputDir    string // default: CWD/<specname-without-ext>/
	SkipValidate bool
	Debug        bool
}

// Compiler compiles .spec.promptc files into instructions.promptc.
type Compiler struct {
	LibraryManager *library.Manager
	Resolver       *resolver.Resolver
}

// NewCompiler creates a Compiler rooted at projectDir for library resolution.
func NewCompiler(projectDir string) *Compiler {
	mgr := library.NewManager(projectDir)
	return &Compiler{
		LibraryManager: mgr,
		Resolver:       resolver.NewResolver(mgr),
	}
}

// Compile validates, resolves, and formats specFile, writing the result to
// <outputDir>/instructions.promptc. Returns the output path on success.
func (c *Compiler) Compile(specFile string, opts *CompileOptions) (string, error) {
	if opts == nil {
		opts = &CompileOptions{}
	}

	raw, err := c.loadSpec(specFile)
	if err != nil {
		return "", err
	}

	if !opts.SkipValidate {
		isValid, errs := validator.Validate(raw)
		if !isValid {
			msg := "validation failed:\n"
			for _, e := range errs {
				msg += fmt.Sprintf("  - %s\n", e)
			}
			return "", fmt.Errorf("%s", msg)
		}
	}

	spec := parseSpec(raw)

	importsContent := ""
	if len(spec.Imports) > 0 {
		importsContent, err = c.Resolver.Resolve(spec.Imports)
		if err != nil {
			return "", err
		}
	}

	if opts.Debug {
		fmt.Printf("[DEBUG] spec file: %s\n", specFile)
		fmt.Printf("[DEBUG] resolved imports: %v\n", c.Resolver.GetResolutionOrder())
		fmt.Printf("[DEBUG] imports content length: %d chars\n", len(importsContent))
	}

	outputDir := opts.OutputDir
	if outputDir == "" {
		base := filepath.Base(specFile)
		// strip .spec.promptc or any extension
		name := strings.TrimSuffix(base, filepath.Ext(base))
		name = strings.TrimSuffix(name, ".spec")
		outputDir = filepath.Join(filepath.Dir(specFile), name)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	fmtSpec := formatter.Spec{
		Context:     spec.Context,
		Features:    spec.Features,
		Constraints: spec.Constraints,
	}
	meta := formatter.FormatterMeta{
		SpecFile:   filepath.Base(specFile),
		Version:    Version,
		CompiledAt: time.Now(),
	}
	output := formatter.Format(fmtSpec, importsContent, meta)

	outPath := filepath.Join(outputDir, "instructions.promptc")
	if err := os.WriteFile(outPath, []byte(output), 0644); err != nil {
		return "", fmt.Errorf("failed to write output file: %w", err)
	}

	return outPath, nil
}

// ParseSpec parses a spec file and returns the ParsedSpec without compiling.
// Used by builder to access resources and build config.
func (c *Compiler) ParseSpec(specFile string) (*ParsedSpec, error) {
	raw, err := c.loadSpec(specFile)
	if err != nil {
		return nil, err
	}
	spec := parseSpec(raw)
	return &spec, nil
}

func (c *Compiler) loadSpec(specFile string) (map[string]interface{}, error) {
	info, err := os.Stat(specFile)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("spec file not found: %s", specFile)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat spec file: %w", err)
	}

	const maxSize = 1 * 1024 * 1024 // 1MB
	if info.Size() > maxSize {
		return nil, fmt.Errorf("spec file too large: %d bytes (max %d bytes)", info.Size(), maxSize)
	}

	content, err := os.ReadFile(specFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(content, &raw); err != nil {
		return nil, fmt.Errorf("invalid YAML in %s: %w", specFile, err)
	}
	if raw == nil {
		raw = make(map[string]interface{})
	}
	return raw, nil
}

// parseSpec converts the raw YAML map into a structured ParsedSpec.
func parseSpec(raw map[string]interface{}) ParsedSpec {
	spec := ParsedSpec{
		Context: make(map[string]interface{}),
	}

	if v, ok := raw["imports"].([]interface{}); ok {
		for _, imp := range v {
			if s, ok := imp.(string); ok {
				spec.Imports = append(spec.Imports, s)
			}
		}
	}

	if v, ok := raw["context"].(map[string]interface{}); ok {
		spec.Context = v
	}

	if v, ok := raw["features"].([]interface{}); ok {
		spec.Features = v
	}

	if v, ok := raw["constraints"].([]interface{}); ok {
		spec.Constraints = v
	}

	if v, ok := raw["resources"].([]interface{}); ok {
		for _, item := range v {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			r := SpecResource{}
			if s, ok := m["name"].(string); ok {
				r.Name = s
			}
			if s, ok := m["url"].(string); ok {
				r.URL = s
			}
			if s, ok := m["git"].(string); ok {
				r.Git = s
			}
			if s, ok := m["ref"].(string); ok {
				r.Ref = s
			}
			if s, ok := m["path"].(string); ok {
				r.Path = s
			}
			spec.Resources = append(spec.Resources, r)
		}
	}

	if bRaw, ok := raw["build"].(map[string]interface{}); ok {
		build := &SpecBuild{}
		if aRaw, ok := bRaw["agent"].(map[string]interface{}); ok {
			build.Agent.Command, _ = aRaw["command"].(string)
			if argsRaw, ok := aRaw["args"].([]interface{}); ok {
				for _, a := range argsRaw {
					if s, ok := a.(string); ok {
						build.Agent.Args = append(build.Agent.Args, s)
					}
				}
			}
		}
		if sRaw, ok := bRaw["sandbox"].(map[string]interface{}); ok {
			build.Sandbox.Type, _ = sRaw["type"].(string)
			build.Sandbox.Image, _ = sRaw["image"].(string)
		}
		spec.Build = build
	}

	return spec
}

// Template represents a project template.
type Template struct {
	Name    string
	Content string
}

// GetTemplates returns available project templates.
func GetTemplates() map[string]string {
	return map[string]string{
		"basic": `# Basic project template
imports:
  - constraints.code_quality

context:
  language: go

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
  language: go
  framework: fiber
  database: postgresql
  testing: testify

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
  language: go
  testing: testify

features:
  - cli_command: "Description of your CLI command"

constraints:
  - clear_help_messages
  - comprehensive_test_coverage
  - cross_platform_compatibility
`,
	}
}

// InitProject creates a new .spec.promptc file from a template.
func InitProject(name, templateName string) error {
	templates := GetTemplates()
	content, ok := templates[templateName]
	if !ok {
		names := make([]string, 0, len(templates))
		for n := range templates {
			names = append(names, n)
		}
		return fmt.Errorf("unknown template %q. Available: %s",
			templateName, strings.Join(names, ", "))
	}

	if _, err := os.Stat(name); err == nil {
		return fmt.Errorf("file %s already exists", name)
	}

	if err := os.WriteFile(name, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}
