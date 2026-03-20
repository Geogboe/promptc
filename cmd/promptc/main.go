package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Geogboe/promptc/internal/builder"
	"github.com/Geogboe/promptc/internal/compiler"
	"github.com/Geogboe/promptc/internal/library"
	"github.com/spf13/cobra"
)

// version is injected at build time via -ldflags "-X main.version=..."
var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	if err := newRootCmd().ExecuteContext(ctx); err != nil {
		stop()
		os.Exit(1)
	}
	stop()
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:     "promptc",
		Short:   "Prompt compiler for agentic programming",
		Long:    "Compile .spec.promptc files into agent instructions. Manage reusable prompt libraries.",
		Version: version,
	}
	compiler.Version = version

	root.AddCommand(
		newValidateCmd(),
		newCompileCmd(),
		newBuildCmd(),
		newListCmd(),
		newInitCmd(),
	)
	return root
}

// ── validate ──────────────────────────────────────────────────────────────────

func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <spec.spec.promptc>",
		Short: "Validate a .spec.promptc file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			specFile := args[0]
			comp := compiler.NewCompiler(dirOf(specFile))

			// Compile with SkipValidate=false writes output; we only want validation.
			// Re-use ParseSpec + Validate directly via a no-output path.
			_, err := comp.Compile(specFile, &compiler.CompileOptions{
				OutputDir:    os.TempDir(), // discard output
				SkipValidate: false,
			})
			if err != nil {
				// Print validation errors cleanly
				fmt.Fprintln(os.Stderr, err.Error())
				return fmt.Errorf("spec is invalid")
			}
			fmt.Printf("✓ %s is valid\n", filepath.Base(specFile))
			return nil
		},
	}
}

// ── compile ───────────────────────────────────────────────────────────────────

func newCompileCmd() *cobra.Command {
	var outputDir string
	var noValidate bool
	var debug bool

	cmd := &cobra.Command{
		Use:   "compile <spec.spec.promptc>",
		Short: "Compile a spec to instructions.promptc",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			specFile := args[0]
			comp := compiler.NewCompiler(dirOf(specFile))

			outPath, err := comp.Compile(specFile, &compiler.CompileOptions{
				OutputDir:    outputDir,
				SkipValidate: noValidate,
				Debug:        debug,
			})
			if err != nil {
				return err
			}
			fmt.Printf("✓ Compiled to %s\n", outPath)
			return nil
		},
	}
	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory (default: ./<specname>/)")
	cmd.Flags().BoolVar(&noValidate, "no-validate", false, "Skip validation")
	cmd.Flags().BoolVar(&debug, "debug", false, "Show debug information")
	return cmd
}

// ── build ─────────────────────────────────────────────────────────────────────

func newBuildCmd() *cobra.Command {
	var outputDir string
	var noValidate bool
	var sandboxOverride string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "build <spec.spec.promptc>",
		Short: "Compile, fetch resources, and run agent in sandbox",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			specFile := args[0]
			return builder.Build(cmd.Context(), specFile, builder.BuildOptions{
				OutputDir:       outputDir,
				SkipValidate:    noValidate,
				SandboxOverride: sandboxOverride,
				DryRun:          dryRun,
			})
		},
	}
	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory")
	cmd.Flags().BoolVar(&noValidate, "no-validate", false, "Skip validation")
	cmd.Flags().StringVar(&sandboxOverride, "sandbox", "", "Override sandbox type (docker, bubblewrap, none)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Compile and fetch but do not run the agent")
	return cmd
}

// ── list ──────────────────────────────────────────────────────────────────────

func newListCmd() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available prompt libraries",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}
			mgr := library.NewManager(cwd)
			libs := mgr.ListLibraries()

			fmt.Println("Available Prompt Libraries")
			fmt.Println(strings.Repeat("=", 50))
			fmt.Println()

			printLibs := func(label string, names []string) {
				if len(names) == 0 {
					return
				}
				fmt.Printf("%s:\n", label)
				for _, name := range names {
					if verbose {
						content, err := mgr.Resolve(name)
						if err == nil {
							lines := strings.SplitN(content, "\n", 2)
							fmt.Printf("  - %s\n", name)
							fmt.Printf("    %s\n", strings.TrimPrefix(strings.TrimSpace(lines[0]), "# "))
							continue
						}
					}
					fmt.Printf("  - %s\n", name)
				}
				fmt.Println()
			}

			printLibs("BUILT-IN", libs.BuiltIn)
			printLibs("GLOBAL", libs.Global)
			printLibs("PROJECT", libs.Project)

			if len(libs.BuiltIn)+len(libs.Global)+len(libs.Project) == 0 {
				fmt.Println("No libraries found.")
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show library descriptions")
	return cmd
}

// ── init ──────────────────────────────────────────────────────────────────────

func newInitCmd() *cobra.Command {
	var templateName string

	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Create a new .spec.promptc file from a template",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := "project.spec.promptc"
			if len(args) > 0 {
				name = args[0]
				if !strings.HasSuffix(name, ".spec.promptc") {
					name += ".spec.promptc"
				}
			}

			if err := compiler.InitProject(name, templateName); err != nil {
				return err
			}

			fmt.Printf("Created %s\n\n", name)
			fmt.Println("Next steps:")
			fmt.Printf("  1. Edit %s to customise your project\n", name)
			fmt.Printf("  2. Validate: promptc validate %s\n", name)
			fmt.Printf("  3. Compile:  promptc compile  %s\n", name)
			fmt.Printf("  4. Build:    promptc build    %s --sandbox none --dry-run\n", name)
			return nil
		},
	}
	cmd.Flags().StringVarP(&templateName, "template", "t", "basic", "Template (basic, web-api, cli-tool)")
	return cmd
}

// dirOf returns the directory of a file path, resolving to CWD for bare filenames.
func dirOf(path string) string {
	dir := filepath.Dir(path)
	if dir == "." {
		if cwd, err := os.Getwd(); err == nil {
			return cwd
		}
	}
	return dir
}
