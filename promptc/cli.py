"""Command-line interface for promptc."""

import sys
import argparse
from pathlib import Path
from .compiler import PromptCompiler
from .library import LibraryManager
from .watcher import PromptWatcher
from . import __version__


def main():
    """Main CLI entry point."""
    parser = argparse.ArgumentParser(
        prog="promptc",
        description="Prompt compiler for agentic programming - compile LLM instructions into different formats",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  promptc compile myapp.prompt --target=claude
  promptc compile myapp.prompt --target=cursor --output=.cursorrules
  promptc compile myapp.prompt --target=raw --debug

Supported targets: raw, claude, cursor, aider, copilot
        """
    )

    parser.add_argument(
        "--version",
        action="version",
        version=f"promptc {__version__}"
    )

    subparsers = parser.add_subparsers(dest="command", help="Available commands")

    # Compile command
    compile_parser = subparsers.add_parser(
        "compile",
        help="Compile a .prompt file to target format"
    )
    compile_parser.add_argument(
        "input",
        type=str,
        help="Input .prompt file"
    )
    compile_parser.add_argument(
        "--target",
        type=str,
        default="raw",
        choices=["raw", "claude", "cursor", "aider", "copilot"],
        help="Target format (default: raw)"
    )
    compile_parser.add_argument(
        "--output",
        type=str,
        help="Output file (default: stdout)"
    )
    compile_parser.add_argument(
        "--debug",
        action="store_true",
        help="Show debug information"
    )

    # List command
    list_parser = subparsers.add_parser(
        "list",
        help="List all available prompt libraries"
    )
    list_parser.add_argument(
        "--verbose",
        "-v",
        action="store_true",
        help="Show detailed information"
    )

    # Init command
    init_parser = subparsers.add_parser(
        "init",
        help="Initialize a new .prompt file"
    )
    init_parser.add_argument(
        "name",
        type=str,
        nargs="?",
        default="project.prompt",
        help="Name of the prompt file (default: project.prompt)"
    )
    init_parser.add_argument(
        "--template",
        type=str,
        choices=["basic", "web-api", "cli-tool"],
        default="basic",
        help="Template to use (default: basic)"
    )

    # Watch command
    watch_parser = subparsers.add_parser(
        "watch",
        help="Watch a .prompt file and auto-recompile on changes"
    )
    watch_parser.add_argument(
        "input",
        type=str,
        help="Input .prompt file to watch"
    )
    watch_parser.add_argument(
        "--target",
        type=str,
        default="raw",
        choices=["raw", "claude", "cursor", "aider", "copilot"],
        help="Target format (default: raw)"
    )
    watch_parser.add_argument(
        "--output",
        type=str,
        help="Output file (default: stdout on changes)"
    )
    watch_parser.add_argument(
        "--interval",
        type=float,
        default=1.0,
        help="Check interval in seconds (default: 1.0)"
    )

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        sys.exit(1)

    try:
        if args.command == "compile":
            compile_command(args)
        elif args.command == "list":
            list_command(args)
        elif args.command == "init":
            init_command(args)
        elif args.command == "watch":
            watch_command(args)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        if hasattr(args, 'debug') and args.debug:
            raise
        sys.exit(1)


def compile_command(args):
    """Execute the compile command."""
    input_path = Path(args.input)

    # Determine project directory (parent of input file)
    project_dir = input_path.parent if input_path.parent != Path(".") else Path.cwd()

    # Initialize compiler
    compiler = PromptCompiler(project_dir=project_dir)

    # Compile
    if args.debug:
        print(f"[DEBUG] Input file: {input_path}")
        print(f"[DEBUG] Project dir: {project_dir}")
        print(f"[DEBUG] Target: {args.target}")

    result = compiler.compile(input_path, target=args.target, debug=args.debug)

    # Output
    if args.output:
        output_path = Path(args.output)
        output_path.parent.mkdir(parents=True, exist_ok=True)
        output_path.write_text(result, encoding="utf-8")
        print(f"Compiled prompt written to: {output_path}")
    else:
        print(result)


def list_command(args):
    """Execute the list command."""
    manager = LibraryManager()
    libraries = manager.list_libraries()

    print("Available Prompt Libraries")
    print("=" * 50)
    print()

    for source, libs in libraries.items():
        if libs:
            print(f"{source.upper()}:")
            for lib in sorted(libs):
                if args.verbose:
                    # Try to get the first line of the library as description
                    try:
                        path = manager.resolve(lib)
                        if path:
                            content = path.read_text(encoding="utf-8")
                            first_line = content.split('\n')[0].strip('# ')
                            print(f"  - {lib}")
                            print(f"    {first_line}")
                    except:
                        print(f"  - {lib}")
                else:
                    print(f"  - {lib}")
            print()

    if not any(libraries.values()):
        print("No libraries found.")
        print()
        print("Built-in libraries should be available.")
        print("Check your installation or create libraries in:")
        print(f"  - ./prompts/ (project-local)")
        print(f"  - ~/.prompts/ (global)")


def init_command(args):
    """Execute the init command."""
    output_path = Path(args.name)

    if output_path.exists():
        print(f"Error: {output_path} already exists", file=sys.stderr)
        sys.exit(1)

    # Define templates
    templates = {
        "basic": """# Basic prompt template
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
""",
        "web-api": """# Web API project template
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
""",
        "cli-tool": """# CLI tool project template
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
"""
    }

    content = templates[args.template]
    output_path.write_text(content, encoding="utf-8")

    print(f"Created {output_path}")
    print()
    print("Next steps:")
    print(f"  1. Edit {output_path} to customize your project")
    print(f"  2. Compile with: promptc compile {output_path} --target=claude")
    print()
    print("Available targets: raw, claude, cursor, aider, copilot")


def watch_command(args):
    """Execute the watch command."""
    input_path = Path(args.input)
    output_path = Path(args.output) if args.output else None

    watcher = PromptWatcher(
        prompt_file=input_path,
        target=args.target,
        output=output_path,
        interval=args.interval
    )

    watcher.watch()


if __name__ == "__main__":
    main()
