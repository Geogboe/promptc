"""Command-line interface for promptc."""

import sys
import argparse
from pathlib import Path
from .compiler import PromptCompiler
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

Supported targets: raw, claude, cursor
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
        choices=["raw", "claude", "cursor"],
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

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        sys.exit(1)

    if args.command == "compile":
        try:
            compile_command(args)
        except Exception as e:
            print(f"Error: {e}", file=sys.stderr)
            if args.debug:
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


if __name__ == "__main__":
    main()
