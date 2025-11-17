package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Geogboe/promptc/internal/compiler"
	"github.com/Geogboe/promptc/internal/library"
)

// TestScenario_NewDeveloperStartsWebAPIProject simulates a real developer workflow:
// 1. Initialize new project with web-api template
// 2. Compile for Claude to start development
// 3. Add custom features
// 4. Recompile and verify
// 5. Switch to different AI tools for different contexts
func TestScenario_NewDeveloperStartsWebAPIProject(t *testing.T) {
	tmpDir := t.TempDir()

	// Step 1: Developer initializes a new web API project
	t.Log("Step 1: Initialize new web API project")
	promptFile := filepath.Join(tmpDir, "api.prompt")
	err := compiler.InitProject(promptFile, "web-api")
	if err != nil {
		t.Fatalf("Failed to initialize project: %v", err)
	}

	// Step 2: Compile for Claude to start development
	t.Log("Step 2: Compile for Claude AI")
	comp := compiler.NewCompiler(tmpDir)
	claudeResult, err := comp.Compile(promptFile, &compiler.CompileOptions{
		Target:   "claude",
		Validate: true,
	})
	if err != nil {
		t.Fatalf("Failed to compile for Claude: %v", err)
	}

	// Verify Claude-specific formatting
	if !strings.Contains(claudeResult, "# Project Context") {
		t.Error("Claude output missing expected formatting")
	}
	if !strings.Contains(claudeResult, "REST API") {
		t.Error("Claude output missing REST API content from template")
	}

	// Step 3: Developer adds custom features to the prompt
	t.Log("Step 3: Add custom features")
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
  - comments: "Nested comments on posts"
  - search: "Full-text search for posts"

constraints:
  - no_hardcoded_secrets
  - comprehensive_test_coverage
  - proper_error_handling
  - api_documentation
`
	if err := os.WriteFile(promptFile, []byte(customContent), 0644); err != nil {
		t.Fatalf("Failed to update prompt file: %v", err)
	}

	// Step 4: Recompile with custom features
	t.Log("Step 4: Recompile with custom features")
	updatedResult, err := comp.Compile(promptFile, &compiler.CompileOptions{
		Target:   "claude",
		Validate: true,
	})
	if err != nil {
		t.Fatalf("Failed to recompile: %v", err)
	}

	// Verify all custom features are included
	expectedContent := []string{
		"User authentication with JWT tokens",
		"User CRUD operations",
		"Blog post CRUD operations",
		"postgresql",
		"fiber",
		"REST API",
		"Database",
		"Security",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(updatedResult, expected) {
			t.Errorf("Updated result missing expected content: %s", expected)
		}
	}

	// Step 5: Switch to Cursor for code completion
	t.Log("Step 5: Compile for Cursor (different context)")
	cursorResult, err := comp.Compile(promptFile, &compiler.CompileOptions{
		Target:   "cursor",
		Validate: true,
	})
	if err != nil {
		t.Fatalf("Failed to compile for Cursor: %v", err)
	}

	// Verify Cursor-specific formatting
	if !strings.Contains(cursorResult, "Project Context:") {
		t.Error("Cursor output missing expected formatting")
	}

	// Step 6: Generate Aider instructions for focused refactoring
	t.Log("Step 6: Compile for Aider (focused refactoring)")
	aiderResult, err := comp.Compile(promptFile, &compiler.CompileOptions{
		Target:   "aider",
		Validate: true,
	})
	if err != nil {
		t.Fatalf("Failed to compile for Aider: %v", err)
	}

	// Verify Aider-specific formatting
	if !strings.Contains(aiderResult, "# Aider Instructions") {
		t.Error("Aider output missing expected formatting")
	}

	t.Log("✓ Complete developer workflow successful")
}

// TestScenario_TeamSharedLibraries simulates a team using shared prompt libraries:
// 1. Team creates shared libraries in a central location
// 2. Individual developers use shared libraries
// 3. Developers can override with project-specific versions
// 4. Library precedence works correctly
func TestScenario_TeamSharedLibraries(t *testing.T) {
	tmpDir := t.TempDir()

	// Step 1: Create team shared library in dev1's project
	t.Log("Step 1: Create team shared libraries")
	// Note: In real scenario, this would be in a global location or git submodule
	// For testing, we'll create it in dev1's project prompts directory

	// Step 2: Developer 1 uses team library in their project
	t.Log("Step 2: Developer 1 uses team library")
	dev1Dir := filepath.Join(tmpDir, "dev1-project")
	if err := os.MkdirAll(dev1Dir, 0755); err != nil {
		t.Fatalf("Failed to create dev1 dir: %v", err)
	}

	// Create team library in dev1's project
	dev1TeamPromptsDir := filepath.Join(dev1Dir, "prompts/team")
	if err := os.MkdirAll(dev1TeamPromptsDir, 0755); err != nil {
		t.Fatalf("Failed to create dev1 team prompts dir: %v", err)
	}

	teamStandards := `# Team Coding Standards

## Code Review Requirements
- All PRs require 2 approvals
- No direct commits to main
- All code must have tests

## Style Guide
- Use conventional commits
- Follow project linting rules
- Document all public APIs
`
	teamLib := filepath.Join(dev1TeamPromptsDir, "standards.prompt")
	if err := os.WriteFile(teamLib, []byte(teamStandards), 0644); err != nil {
		t.Fatalf("Failed to create team library: %v", err)
	}

	dev1Prompt := filepath.Join(dev1Dir, "app.prompt")
	dev1Content := `imports:
  - patterns.rest_api
  - team.standards

context:
  language: go
  developer: "Dev 1"

features:
  - api: "REST API endpoints"
`
	if err := os.WriteFile(dev1Prompt, []byte(dev1Content), 0644); err != nil {
		t.Fatalf("Failed to create dev1 prompt: %v", err)
	}

	// Compile dev1's prompt
	comp1 := compiler.NewCompiler(dev1Dir)
	result1, err := comp1.Compile(dev1Prompt, &compiler.CompileOptions{
		Target:   "raw",
		Validate: true,
	})
	if err != nil {
		t.Fatalf("Failed to compile dev1 prompt: %v", err)
	}

	// Verify team standards are included
	if !strings.Contains(result1, "Team Coding Standards") {
		t.Error("Dev1 result missing team standards")
	}
	if !strings.Contains(result1, "2 approvals") {
		t.Error("Dev1 result missing team standards content")
	}

	// Step 3: Developer 2 creates project-specific override
	t.Log("Step 3: Developer 2 creates local override")
	dev2Dir := filepath.Join(tmpDir, "dev2-project")
	dev2PromptsDir := filepath.Join(dev2Dir, "prompts/team")
	if err := os.MkdirAll(dev2PromptsDir, 0755); err != nil {
		t.Fatalf("Failed to create dev2 prompts dir: %v", err)
	}

	// Dev2 has project-specific standards (overrides team)
	projectStandards := `# Project-Specific Standards

## This project uses different rules
- Fast-track PRs: only 1 approval needed
- Experimental features allowed
- Prototyping focused
`
	projectLib := filepath.Join(dev2PromptsDir, "standards.prompt")
	if err := os.WriteFile(projectLib, []byte(projectStandards), 0644); err != nil {
		t.Fatalf("Failed to create project library: %v", err)
	}

	dev2Prompt := filepath.Join(dev2Dir, "app.prompt")
	dev2Content := `imports:
  - patterns.rest_api
  - team.standards

context:
  language: python
  developer: "Dev 2"

features:
  - prototype: "Quick prototype"
`
	if err := os.WriteFile(dev2Prompt, []byte(dev2Content), 0644); err != nil {
		t.Fatalf("Failed to create dev2 prompt: %v", err)
	}

	// Compile dev2's prompt
	comp2 := compiler.NewCompiler(dev2Dir)
	result2, err := comp2.Compile(dev2Prompt, &compiler.CompileOptions{
		Target:   "raw",
		Validate: true,
	})
	if err != nil {
		t.Fatalf("Failed to compile dev2 prompt: %v", err)
	}

	// Verify project-specific override is used (NOT team standards)
	if !strings.Contains(result2, "Project-Specific Standards") {
		t.Error("Dev2 result should use project-specific standards")
	}
	if !strings.Contains(result2, "1 approval") {
		t.Error("Dev2 result missing project override content")
	}
	if strings.Contains(result2, "2 approvals") {
		t.Error("Dev2 result should NOT contain team standards (should be overridden)")
	}

	t.Log("✓ Team shared libraries and overrides working correctly")
}

// TestScenario_BuildingAppStepByStep simulates building an app iteratively:
// 1. Start with basic structure
// 2. Add authentication
// 3. Add database
// 4. Add testing requirements
// 5. Each step compiles and builds on previous
func TestScenario_BuildingAppStepByStep(t *testing.T) {
	tmpDir := t.TempDir()
	promptFile := filepath.Join(tmpDir, "app.prompt")
	comp := compiler.NewCompiler(tmpDir)

	// Step 1: Start with basic structure
	t.Log("Step 1: Basic app structure")
	step1 := `context:
  language: go
  type: web-api

features:
  - structure: "Basic project structure with main.go"
`
	if err := os.WriteFile(promptFile, []byte(step1), 0644); err != nil {
		t.Fatalf("Failed to write step1: %v", err)
	}

	result1, err := comp.Compile(promptFile, &compiler.CompileOptions{
		Target:   "claude",
		Validate: true,
	})
	if err != nil {
		t.Fatalf("Step 1 failed: %v", err)
	}
	if !strings.Contains(result1, "Basic project structure") {
		t.Error("Step 1 missing expected content")
	}

	// Step 2: Add authentication
	t.Log("Step 2: Add authentication")
	step2 := `imports:
  - patterns.rest_api

context:
  language: go
  type: web-api
  auth: jwt

features:
  - structure: "Basic project structure with main.go"
  - auth: "JWT authentication with login/register endpoints"
  - middleware: "Auth middleware for protected routes"
`
	if err := os.WriteFile(promptFile, []byte(step2), 0644); err != nil {
		t.Fatalf("Failed to write step2: %v", err)
	}

	result2, err := comp.Compile(promptFile, &compiler.CompileOptions{
		Target:   "claude",
		Validate: true,
	})
	if err != nil {
		t.Fatalf("Step 2 failed: %v", err)
	}
	if !strings.Contains(result2, "JWT authentication") {
		t.Error("Step 2 missing auth content")
	}
	if !strings.Contains(result2, "REST API") {
		t.Error("Step 2 missing imported content")
	}

	// Step 3: Add database
	t.Log("Step 3: Add database integration")
	step3 := `imports:
  - patterns.rest_api
  - patterns.database

context:
  language: go
  type: web-api
  auth: jwt
  database: postgresql
  orm: gorm

features:
  - structure: "Basic project structure with main.go"
  - auth: "JWT authentication with login/register endpoints"
  - middleware: "Auth middleware for protected routes"
  - users: "User model and CRUD operations"
  - migrations: "Database migrations setup"
`
	if err := os.WriteFile(promptFile, []byte(step3), 0644); err != nil {
		t.Fatalf("Failed to write step3: %v", err)
	}

	result3, err := comp.Compile(promptFile, &compiler.CompileOptions{
		Target:   "claude",
		Validate: true,
	})
	if err != nil {
		t.Fatalf("Step 3 failed: %v", err)
	}
	if !strings.Contains(result3, "Database") {
		t.Error("Step 3 missing database content")
	}
	if !strings.Contains(result3, "User model") {
		t.Error("Step 3 missing user model feature")
	}

	// Step 4: Add comprehensive testing
	t.Log("Step 4: Add testing requirements")
	step4 := `imports:
  - patterns.rest_api
  - patterns.database
  - patterns.testing
  - constraints.code_quality

context:
  language: go
  type: web-api
  auth: jwt
  database: postgresql
  orm: gorm
  testing: testify

features:
  - structure: "Basic project structure with main.go"
  - auth: "JWT authentication with login/register endpoints"
  - middleware: "Auth middleware for protected routes"
  - users: "User model and CRUD operations"
  - migrations: "Database migrations setup"
  - tests: "Unit tests for all handlers and services"
  - integration_tests: "Integration tests for API endpoints"

constraints:
  - comprehensive_test_coverage
  - proper_error_handling
  - no_hardcoded_secrets
`
	if err := os.WriteFile(promptFile, []byte(step4), 0644); err != nil {
		t.Fatalf("Failed to write step4: %v", err)
	}

	result4, err := comp.Compile(promptFile, &compiler.CompileOptions{
		Target:   "claude",
		Validate: true,
	})
	if err != nil {
		t.Fatalf("Step 4 failed: %v", err)
	}

	// Verify all accumulated content is present
	expectedFinal := []string{
		"REST API",
		"Database",
		"Testing",
		"Code Quality",
		"JWT authentication",
		"User model",
		"Unit tests",
		"Integration tests",
		"Comprehensive Test Coverage", // Constraints are title-cased
	}

	for _, expected := range expectedFinal {
		if !strings.Contains(result4, expected) {
			t.Errorf("Final result missing: %s", expected)
		}
	}

	t.Log("✓ Iterative app building successful")
}

// TestScenario_MultiFileProjectWithDifferentTargets simulates a complex project:
// 1. Multiple .prompt files for different parts of the app
// 2. Frontend prompt for Copilot
// 3. Backend prompt for Claude
// 4. Database prompt for Aider
// 5. DevOps prompt for Cursor
func TestScenario_MultiFileProjectWithDifferentTargets(t *testing.T) {
	tmpDir := t.TempDir()

	// Create project-wide custom library
	t.Log("Step 1: Create project-wide custom libraries")
	projectPromptsDir := filepath.Join(tmpDir, "prompts/project")
	if err := os.MkdirAll(projectPromptsDir, 0755); err != nil {
		t.Fatalf("Failed to create project prompts dir: %v", err)
	}

	projectStandards := `# Project Standards

## Architecture
- Microservices architecture
- Event-driven communication
- RESTful APIs

## Tech Stack
- Backend: Go + Fiber
- Frontend: React + TypeScript
- Database: PostgreSQL
- Cache: Redis
- Queue: RabbitMQ
`
	projectStandardsFile := filepath.Join(projectPromptsDir, "standards.prompt")
	if err := os.WriteFile(projectStandardsFile, []byte(projectStandards), 0644); err != nil {
		t.Fatalf("Failed to create project standards: %v", err)
	}

	// Create compiler AFTER prompts directory exists (so search paths include it)
	comp := compiler.NewCompiler(tmpDir)

	// Frontend prompt for GitHub Copilot
	t.Log("Step 2: Create frontend prompt for Copilot")
	frontendPrompt := filepath.Join(tmpDir, "frontend.prompt")
	frontendContent := `imports:
  - project.standards
  - constraints.accessibility
  - constraints.code_quality

context:
  language: typescript
  framework: react
  styling: tailwindcss
  state: zustand

features:
  - dashboard: "User dashboard with widgets"
  - forms: "Form handling with validation"
  - api_client: "Type-safe API client"
  - auth_ui: "Login and registration forms"

constraints:
  - accessibility_wcag_aa
  - responsive_design
  - type_safety
`
	if err := os.WriteFile(frontendPrompt, []byte(frontendContent), 0644); err != nil {
		t.Fatalf("Failed to create frontend prompt: %v", err)
	}

	frontendResult, err := comp.Compile(frontendPrompt, &compiler.CompileOptions{
		Target:   "copilot",
		Validate: true,
	})
	if err != nil {
		t.Fatalf("Failed to compile frontend prompt: %v", err)
	}

	if !strings.Contains(frontendResult, "GitHub Copilot Instructions") {
		t.Error("Frontend result missing Copilot formatting")
	}
	if !strings.Contains(frontendResult, "React") && !strings.Contains(frontendResult, "react") {
		t.Error("Frontend result missing React context")
	}
	if !strings.Contains(frontendResult, "Microservices") {
		t.Error("Frontend result missing project standards")
	}

	// Backend prompt for Claude
	t.Log("Step 3: Create backend prompt for Claude")
	backendPrompt := filepath.Join(tmpDir, "backend.prompt")
	backendContent := `imports:
  - project.standards
  - patterns.rest_api
  - patterns.database
  - constraints.security
  - constraints.performance

context:
  language: go
  framework: fiber
  database: postgresql
  orm: gorm
  auth: jwt
  cache: redis

features:
  - auth_service: "Authentication service with JWT"
  - user_service: "User management service"
  - api_gateway: "API gateway with rate limiting"
  - event_publisher: "Publish events to RabbitMQ"
  - caching: "Redis caching layer"

constraints:
  - no_hardcoded_secrets
  - comprehensive_test_coverage
  - api_documentation
  - rate_limiting
  - input_validation
`
	if err := os.WriteFile(backendPrompt, []byte(backendContent), 0644); err != nil {
		t.Fatalf("Failed to create backend prompt: %v", err)
	}

	backendResult, err := comp.Compile(backendPrompt, &compiler.CompileOptions{
		Target:   "claude",
		Validate: true,
	})
	if err != nil {
		t.Fatalf("Failed to compile backend prompt: %v", err)
	}

	if !strings.Contains(backendResult, "# Project Context") {
		t.Error("Backend result missing Claude formatting")
	}
	if !strings.Contains(backendResult, "REST API") {
		t.Error("Backend result missing REST API patterns")
	}
	if !strings.Contains(backendResult, "Authentication service") {
		t.Error("Backend result missing auth feature")
	}

	// Database prompt for Aider (focused refactoring)
	t.Log("Step 4: Create database prompt for Aider")
	databasePrompt := filepath.Join(tmpDir, "database.prompt")
	databaseContent := `imports:
  - project.standards
  - patterns.database

context:
  database: postgresql
  orm: gorm
  migrations: golang-migrate

features:
  - schema: "Database schema design"
  - indexes: "Optimized indexes for queries"
  - migrations: "Versioned migrations"
  - relations: "Foreign key relationships"
  - seeds: "Seed data for development"

constraints:
  - normalized_design
  - proper_indexing
  - migration_rollback_support
`
	if err := os.WriteFile(databasePrompt, []byte(databaseContent), 0644); err != nil {
		t.Fatalf("Failed to create database prompt: %v", err)
	}

	databaseResult, err := comp.Compile(databasePrompt, &compiler.CompileOptions{
		Target:   "aider",
		Validate: true,
	})
	if err != nil {
		t.Fatalf("Failed to compile database prompt: %v", err)
	}

	if !strings.Contains(databaseResult, "# Aider Instructions") {
		t.Error("Database result missing Aider formatting")
	}
	if !strings.Contains(databaseResult, "Database") {
		t.Error("Database result missing database patterns")
	}

	// DevOps prompt for Cursor
	t.Log("Step 5: Create devops prompt for Cursor")
	devopsPrompt := filepath.Join(tmpDir, "devops.prompt")
	devopsContent := `imports:
  - project.standards

context:
  deployment: docker
  orchestration: kubernetes
  ci_cd: github_actions
  monitoring: prometheus
  logging: elk_stack

features:
  - dockerfile: "Multi-stage Docker builds"
  - k8s_manifests: "Kubernetes deployment manifests"
  - ci_pipeline: "GitHub Actions CI/CD pipeline"
  - monitoring: "Prometheus metrics and Grafana dashboards"
  - logging: "Centralized logging with ELK"

constraints:
  - zero_downtime_deployment
  - automated_rollback
  - health_checks
`
	if err := os.WriteFile(devopsPrompt, []byte(devopsContent), 0644); err != nil {
		t.Fatalf("Failed to create devops prompt: %v", err)
	}

	devopsResult, err := comp.Compile(devopsPrompt, &compiler.CompileOptions{
		Target:   "cursor",
		Validate: true,
	})
	if err != nil {
		t.Fatalf("Failed to compile devops prompt: %v", err)
	}

	if !strings.Contains(devopsResult, "Project Context:") {
		t.Error("DevOps result missing Cursor formatting")
	}
	if !strings.Contains(devopsResult, "kubernetes") {
		t.Error("DevOps result missing k8s context")
	}

	t.Log("✓ Multi-file project with different targets successful")
}

// TestScenario_ExclusionsInComplexProject tests import exclusions in real scenarios
func TestScenario_ExclusionsInComplexProject(t *testing.T) {
	tmpDir := t.TempDir()

	// Create custom libraries with overlapping content
	promptsDir := filepath.Join(tmpDir, "prompts")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatalf("Failed to create prompts dir: %v", err)
	}

	// Base library with common patterns
	baseLib := `# Base Patterns

## Common
- Use error handling
- Log all operations
- Validate inputs
`
	if err := os.WriteFile(filepath.Join(promptsDir, "base.prompt"), []byte(baseLib), 0644); err != nil {
		t.Fatalf("Failed to create base library: %v", err)
	}

	// Extended library that includes base
	extendedLib := `# Extended Patterns

## Advanced
- Use caching
- Implement retry logic
- Use circuit breakers
`
	if err := os.WriteFile(filepath.Join(promptsDir, "extended.prompt"), []byte(extendedLib), 0644); err != nil {
		t.Fatalf("Failed to create extended library: %v", err)
	}

	// Test 1: Without exclusion (both included)
	t.Log("Test 1: Import both libraries")
	prompt1 := filepath.Join(tmpDir, "test1.prompt")
	content1 := `imports:
  - base
  - extended

features:
  - feature: "Test feature"
`
	if err := os.WriteFile(prompt1, []byte(content1), 0644); err != nil {
		t.Fatalf("Failed to create prompt1: %v", err)
	}

	comp := compiler.NewCompiler(tmpDir)
	result1, err := comp.Compile(prompt1, &compiler.CompileOptions{
		Target:   "raw",
		Validate: true,
	})
	if err != nil {
		t.Fatalf("Failed to compile test1: %v", err)
	}

	if !strings.Contains(result1, "Base Patterns") {
		t.Error("Result1 should contain base patterns")
	}
	if !strings.Contains(result1, "Extended Patterns") {
		t.Error("Result1 should contain extended patterns")
	}

	// Test 2: With exclusion (only extended, exclude base)
	t.Log("Test 2: Import extended but exclude base")
	prompt2 := filepath.Join(tmpDir, "test2.prompt")
	content2 := `imports:
  - extended
  - "!base"

features:
  - feature: "Test feature"
`
	if err := os.WriteFile(prompt2, []byte(content2), 0644); err != nil {
		t.Fatalf("Failed to create prompt2: %v", err)
	}

	result2, err := comp.Compile(prompt2, &compiler.CompileOptions{
		Target:   "raw",
		Validate: true,
	})
	if err != nil {
		t.Fatalf("Failed to compile test2: %v", err)
	}

	if strings.Contains(result2, "Base Patterns") {
		t.Error("Result2 should NOT contain base patterns (excluded)")
	}
	if !strings.Contains(result2, "Extended Patterns") {
		t.Error("Result2 should contain extended patterns")
	}

	t.Log("✓ Exclusions working correctly in complex projects")
}

// TestScenario_LibraryDiscoveryAndListing tests library management
func TestScenario_LibraryDiscoveryAndListing(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multi-level library structure
	t.Log("Creating complex library structure")

	// Project-local libraries
	projectPromptsDir := filepath.Join(tmpDir, "prompts")
	if err := os.MkdirAll(filepath.Join(projectPromptsDir, "frontend"), 0755); err != nil {
		t.Fatalf("Failed to create frontend prompts: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(projectPromptsDir, "backend"), 0755); err != nil {
		t.Fatalf("Failed to create backend prompts: %v", err)
	}

	// Frontend libraries
	os.WriteFile(filepath.Join(projectPromptsDir, "frontend", "react.prompt"), []byte("# React patterns"), 0644)
	os.WriteFile(filepath.Join(projectPromptsDir, "frontend", "vue.prompt"), []byte("# Vue patterns"), 0644)

	// Backend libraries
	os.WriteFile(filepath.Join(projectPromptsDir, "backend", "api.prompt"), []byte("# API patterns"), 0644)
	os.WriteFile(filepath.Join(projectPromptsDir, "backend", "database.prompt"), []byte("# Database patterns"), 0644)

	// List all libraries
	mgr := library.NewManager(tmpDir)
	libs := mgr.ListLibraries()

	// Verify built-in libraries are present
	if len(libs.BuiltIn) == 0 {
		t.Error("Should have built-in libraries")
	}

	// Verify project libraries are discovered
	projectLibs := libs.Project
	expectedLibs := []string{"frontend.react", "frontend.vue", "backend.api", "backend.database"}

	for _, expected := range expectedLibs {
		found := false
		for _, lib := range projectLibs {
			if strings.Contains(lib, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected library not found: %s. Got: %v", expected, projectLibs)
		}
	}

	// Test using nested library
	t.Log("Testing nested library import")
	promptFile := filepath.Join(tmpDir, "app.prompt")
	promptContent := `imports:
  - frontend.react
  - backend.api

features:
  - app: "Full-stack app"
`
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("Failed to create prompt: %v", err)
	}

	comp := compiler.NewCompiler(tmpDir)
	result, err := comp.Compile(promptFile, &compiler.CompileOptions{
		Target:   "raw",
		Validate: true,
	})
	if err != nil {
		t.Fatalf("Failed to compile with nested imports: %v", err)
	}

	if !strings.Contains(result, "React patterns") {
		t.Error("Should contain React patterns from nested library")
	}
	if !strings.Contains(result, "API patterns") {
		t.Error("Should contain API patterns from nested library")
	}

	t.Log("✓ Library discovery and nested imports working correctly")
}
