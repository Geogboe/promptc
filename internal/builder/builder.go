// Package builder orchestrates the full build pipeline:
// compile → fetch resources → run agent in sandbox.
package builder

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Geogboe/promptc/internal/compiler"
	"github.com/Geogboe/promptc/internal/fetcher"
	"github.com/Geogboe/promptc/internal/progress"
	"github.com/Geogboe/promptc/internal/sandbox"
)

// BuildOptions controls build behaviour.
type BuildOptions struct {
	OutputDir       string // default: ./<specname>/
	SkipValidate    bool
	SandboxOverride string // overrides spec.build.sandbox.type
	DryRun          bool   // compile + fetch but do not run the agent
	NoCache         bool   // skip resource cache; always re-fetch
}

// Build runs the full pipeline for specFile:
//
//  1. Compile spec → instructions.promptc in outputDir
//  2. Fetch resources → outputDir/resources/
//  3. Run agent inside sandbox with outputDir as workspace
//
// If opts.DryRun is true, step 3 is skipped.
func Build(ctx context.Context, specFile string, opts BuildOptions) error {
	projectDir := filepath.Dir(specFile)
	if projectDir == "." {
		if cwd, err := os.Getwd(); err == nil {
			projectDir = cwd
		}
	}

	comp := compiler.NewCompiler(projectDir)

	// Parse the full spec to access resources and build config
	spec, err := comp.ParseSpec(specFile)
	if err != nil {
		return fmt.Errorf("parsing spec: %w", err)
	}

	// Determine output directory
	outputDir := opts.OutputDir
	if outputDir == "" {
		base := filepath.Base(specFile)
		// strip .spec.promptc or last extension
		name := base
		for ext := filepath.Ext(name); ext != ""; ext = filepath.Ext(name) {
			name = name[:len(name)-len(ext)]
			if ext == ".spec" {
				break
			}
		}
		outputDir = filepath.Join(projectDir, name)
	}

	// Step 1: Compile
	progress.Step("compiling %s", filepath.Base(specFile))
	outPath, err := comp.Compile(specFile, &compiler.CompileOptions{
		OutputDir:    outputDir,
		SkipValidate: opts.SkipValidate,
	})
	if err != nil {
		return fmt.Errorf("compile: %w", err)
	}
	progress.Done("compiled → %s", outPath)

	// Step 2: Fetch resources
	if len(spec.Resources) > 0 {
		resourcesDir := filepath.Join(outputDir, "resources")
		progress.Step("fetching %d resource(s)", len(spec.Resources))

		fetchResources := make([]fetcher.Resource, len(spec.Resources))
		for i, r := range spec.Resources {
			fetchResources[i] = fetcher.Resource{
				Name: r.Name,
				URL:  r.URL,
				Git:  r.Git,
				Ref:  r.Ref,
				Path: r.Path,
			}
		}

		errs := fetcher.FetchAllCached(ctx, fetchResources, resourcesDir, opts.NoCache)
		if len(errs) > 0 {
			// Collect all errors but don't fail the whole build
			for _, e := range errs {
				progress.Fail("fetch error: %v", e)
			}
			return errors.Join(errs...)
		}
		progress.Done("resources ready in %s", resourcesDir)
	}

	if opts.DryRun {
		progress.Done("dry run complete — skipping agent")
		return nil
	}

	// Step 3: Run agent in sandbox
	if spec.Build == nil {
		return fmt.Errorf("spec has no 'build' section; cannot run agent (use --dry-run to skip)")
	}
	if spec.Build.Agent.Command == "" {
		return fmt.Errorf("spec.build.agent.command is empty")
	}

	sandboxType := spec.Build.Sandbox.Type
	if opts.SandboxOverride != "" {
		sandboxType = opts.SandboxOverride
	}
	if sandboxType == "" {
		sandboxType = "none"
	}

	sbx, err := sandbox.New(sandbox.Config{
		Type:    sandboxType,
		Image:   spec.Build.Sandbox.Image,
		WorkDir: outputDir,
	})
	if err != nil {
		return fmt.Errorf("creating sandbox: %w", err)
	}

	progress.Step("running %s in %s sandbox", spec.Build.Agent.Command, sandboxType)
	if err := sbx.Run(ctx, spec.Build.Agent.Command, spec.Build.Agent.Args); err != nil {
		if err == context.Canceled || err == context.DeadlineExceeded {
			progress.Fail("build interrupted — workspace preserved at %s", outputDir)
			return err
		}
		return fmt.Errorf("agent exited with error: %w", err)
	}

	progress.Done("build complete — output in %s", outputDir)
	return nil
}
