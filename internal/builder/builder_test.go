package builder

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuild_DryRunNoResources(t *testing.T) {
	dir := t.TempDir()
	specFile := filepath.Join(dir, "app.spec.promptc")
	content := `features:
  - hello: "World"
`
	if err := os.WriteFile(specFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	err := Build(context.Background(), specFile, BuildOptions{
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("Build dry run failed: %v", err)
	}

	// instructions.promptc should exist
	outDir := filepath.Join(dir, "app")
	outFile := filepath.Join(outDir, "instructions.promptc")
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	if !strings.Contains(string(data), "World") {
		t.Error("output missing feature content")
	}
}

func TestBuild_NoBuildSection_NoAgent(t *testing.T) {
	dir := t.TempDir()
	specFile := filepath.Join(dir, "app.spec.promptc")
	content := `features:
  - f: "test"
`
	if err := os.WriteFile(specFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Without dry-run and without build section, should fail
	err := Build(context.Background(), specFile, BuildOptions{DryRun: false})
	if err == nil {
		t.Fatal("expected error when no build section and not dry-run")
	}
	if !strings.Contains(err.Error(), "build") {
		t.Errorf("error should mention build section: %v", err)
	}
}

func TestBuild_CustomOutputDir(t *testing.T) {
	dir := t.TempDir()
	specFile := filepath.Join(dir, "app.spec.promptc")
	content := `features:
  - f: "test"
`
	if err := os.WriteFile(specFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	outDir := filepath.Join(dir, "my-custom-output")
	err := Build(context.Background(), specFile, BuildOptions{
		OutputDir: outDir,
		DryRun:    true,
	})
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outDir, "instructions.promptc")); os.IsNotExist(err) {
		t.Error("output not in custom output directory")
	}
}

func TestBuild_ContextCancelled(t *testing.T) {
	dir := t.TempDir()
	specFile := filepath.Join(dir, "app.spec.promptc")
	content := `features:
  - f: "test"
build:
  agent:
    command: sleep
    args: ["30"]
  sandbox:
    type: none
`
	if err := os.WriteFile(specFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before build starts

	err := Build(ctx, specFile, BuildOptions{})
	if err == nil {
		t.Fatal("expected error when context is cancelled")
	}
}

func TestBuild_InvalidSandboxType(t *testing.T) {
	dir := t.TempDir()
	specFile := filepath.Join(dir, "app.spec.promptc")
	content := `features:
  - f: "test"
build:
  agent:
    command: echo
  sandbox:
    type: none
`
	if err := os.WriteFile(specFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	err := Build(context.Background(), specFile, BuildOptions{
		SandboxOverride: "invalid-sandbox",
	})
	if err == nil {
		t.Fatal("expected error for invalid sandbox type")
	}
	if !strings.Contains(err.Error(), "sandbox") {
		t.Errorf("error should mention sandbox: %v", err)
	}
}
