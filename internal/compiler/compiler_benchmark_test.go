package compiler

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkCompile(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.prompt")

	content := `imports:
  - patterns.rest_api
  - patterns.testing
  - constraints.security

context:
  language: go
  framework: fiber

features:
  - auth: "User authentication"
  - api: "REST API endpoints"

constraints:
  - no_hardcoded_secrets
  - comprehensive_tests
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	compiler := NewCompiler(tmpDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := compiler.Compile(testFile, &CompileOptions{
			Target:   "claude",
			Validate: true,
		})
		if err != nil {
			b.Fatalf("Compile failed: %v", err)
		}
	}
}

func BenchmarkCompileNoValidation(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.prompt")

	content := `context:
  language: go

features:
  - test: "test feature"
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	compiler := NewCompiler(tmpDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := compiler.Compile(testFile, &CompileOptions{
			Target:   "raw",
			Validate: false,
		})
		if err != nil {
			b.Fatalf("Compile failed: %v", err)
		}
	}
}

func BenchmarkLoadSpec(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.prompt")

	content := `context:
  language: go
  framework: fiber
  database: postgres

features:
  - feature1: "Description 1"
  - feature2: "Description 2"

constraints:
  - constraint1
  - constraint2
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	compiler := NewCompiler(tmpDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := compiler.loadSpec(testFile)
		if err != nil {
			b.Fatalf("loadSpec failed: %v", err)
		}
	}
}

func BenchmarkValidation(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.prompt")

	content := `imports:
  - patterns.rest_api

context:
  language: go

features:
  - test: "test feature"

constraints:
  - test_constraint
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	compiler := NewCompiler(tmpDir)
	spec, err := compiler.loadSpec(testFile)
	if err != nil {
		b.Fatalf("Failed to load spec: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compiler.Compile(testFile, &CompileOptions{
			Target:   "raw",
			Validate: true,
		})
	}

	_ = spec
}
