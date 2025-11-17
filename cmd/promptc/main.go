package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Geogboe/promptc/internal/compiler"
	"github.com/Geogboe/promptc/internal/library"
	"github.com/Geogboe/promptc/internal/targets"
	"github.com/spf13/cobra"
)

var version = "0.2.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "promptc",
		Short: "Prompt compiler for agentic programming",
		Long:  `Compile LLM instructions into different formats. Manage reusable prompt libraries.`,
	}

	rootCmd.Version = version

	// Compile command
	var compileTarget string
	var compileOutput string
	var compileDebug bool
	var compileNoValidate bool

	compileCmd := &cobra.Command{
		Use:   "compile <input.prompt>",
		Short: "Compile a .prompt file to target format",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputPath := args[0]

			// Determine project directory
			projectDir := filepath.Dir(inputPath)
			if projectDir == "." {
				var err error
				projectDir, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get working directory: %w", err)
				}
			}

			// Initialize compiler
			comp := compiler.NewCompiler(projectDir)

			// Compile
			if compileDebug {
				fmt.Printf("[DEBUG] Input file: %s\n", inputPath)
				fmt.Printf("[DEBUG] Project dir: %s\n", projectDir)
				fmt.Printf("[DEBUG] Target: %s\n", compileTarget)
			}

			result, err := comp.Compile(inputPath, &compiler.CompileOptions{
				Target:   compileTarget,
				Debug:    compileDebug,
				Validate: !compileNoValidate,
			})
			if err != nil {
				return err
			}

			// Output
			if compileOutput != "" {
				// Create parent directories if needed
				if err := os.MkdirAll(filepath.Dir(compileOutput), 0755); err != nil {
					return fmt.Errorf("failed to create output directory: %w", err)
				}

				if err := os.WriteFile(compileOutput, []byte(result), 0644); err != nil {
					return fmt.Errorf("failed to write output file: %w", err)
				}

				fmt.Printf("Compiled prompt written to: %s\n", compileOutput)
			} else {
				fmt.Println(result)
			}

			return nil
		},
	}

	compileCmd.Flags().StringVarP(&compileTarget, "target", "t", "raw", "Target format (raw, claude, cursor, aider, copilot)")
	compileCmd.Flags().StringVarP(&compileOutput, "output", "o", "", "Output file (default: stdout)")
	compileCmd.Flags().BoolVar(&compileDebug, "debug", false, "Show debug information")
	compileCmd.Flags().BoolVar(&compileNoValidate, "no-validate", false, "Skip validation")

	// List command
	var listVerbose bool

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all available prompt libraries",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := library.NewManager("")
			libs := manager.ListLibraries()

			fmt.Println("Available Prompt Libraries")
			fmt.Println(strings.Repeat("=", 50))
			fmt.Println()

			if len(libs.BuiltIn) > 0 {
				fmt.Println("BUILT-IN:")
				for _, lib := range libs.BuiltIn {
					if listVerbose {
						// Try to get first line as description
						content, err := manager.Resolve(lib)
						if err == nil {
							lines := strings.Split(content, "\n")
							if len(lines) > 0 {
								firstLine := strings.TrimPrefix(lines[0], "# ")
								fmt.Printf("  - %s\n", lib)
								fmt.Printf("    %s\n", firstLine)
								continue
							}
						}
					}
					fmt.Printf("  - %s\n", lib)
				}
				fmt.Println()
			}

			if len(libs.Global) > 0 {
				fmt.Println("GLOBAL:")
				for _, lib := range libs.Global {
					fmt.Printf("  - %s\n", lib)
				}
				fmt.Println()
			}

			if len(libs.Project) > 0 {
				fmt.Println("PROJECT:")
				for _, lib := range libs.Project {
					fmt.Printf("  - %s\n", lib)
				}
				fmt.Println()
			}

			if len(libs.BuiltIn) == 0 && len(libs.Global) == 0 && len(libs.Project) == 0 {
				fmt.Println("No libraries found.")
			}

			return nil
		},
	}

	listCmd.Flags().BoolVarP(&listVerbose, "verbose", "v", false, "Show detailed information")

	// Init command
	var initTemplate string

	initCmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Initialize a new .prompt file",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := "project.prompt"
			if len(args) > 0 {
				name = args[0]
			}

			if err := compiler.InitProject(name, initTemplate); err != nil {
				return err
			}

			fmt.Printf("Created %s\n\n", name)
			fmt.Println("Next steps:")
			fmt.Printf("  1. Edit %s to customize your project\n", name)
			fmt.Printf("  2. Compile with: promptc compile %s --target=claude\n\n", name)
			fmt.Printf("Available targets: %s\n", strings.Join(targets.GetSupportedTargets(), ", "))

			return nil
		},
	}

	initCmd.Flags().StringVarP(&initTemplate, "template", "t", "basic", "Template to use (basic, web-api, cli-tool)")

	// Add commands
	rootCmd.AddCommand(compileCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
