"""Core compilation logic for prompt files."""

import yaml
from pathlib import Path
from typing import Dict, Any, Optional
from .library import LibraryManager
from .resolver import ImportResolver
from .targets import raw, claude, cursor, aider, copilot
from .validator import PromptValidator


class PromptCompiler:
    """Compiles .prompt files into target-specific formats."""

    SUPPORTED_TARGETS = {
        "raw": raw.format_prompt,
        "claude": claude.format_prompt,
        "cursor": cursor.format_prompt,
        "aider": aider.format_prompt,
        "copilot": copilot.format_prompt,
    }

    def __init__(self, project_dir: Optional[Path] = None):
        """
        Initialize compiler.

        Args:
            project_dir: Project directory for resolving local prompts
        """
        self.library_manager = LibraryManager(project_dir)
        self.resolver = ImportResolver(self.library_manager)

    def compile(
        self,
        prompt_file: Path,
        target: str = "raw",
        debug: bool = False,
        validate: bool = True
    ) -> str:
        """
        Compile a prompt file to the target format.

        Args:
            prompt_file: Path to .prompt file
            target: Target format (raw, claude, cursor, etc.)
            debug: Show debug information
            validate: Validate the prompt file before compiling (default: True)

        Returns:
            Compiled prompt content

        Raises:
            ValueError: If target is not supported or validation fails
            FileNotFoundError: If prompt file or imports not found
            yaml.YAMLError: If prompt file has invalid YAML
        """
        if target not in self.SUPPORTED_TARGETS:
            raise ValueError(
                f"Unsupported target '{target}'. "
                f"Supported targets: {', '.join(self.SUPPORTED_TARGETS.keys())}"
            )

        # Load and parse prompt file
        spec = self._load_spec(prompt_file)

        # Validate if requested
        if validate:
            is_valid, errors = PromptValidator.validate(spec)
            if not is_valid:
                error_msg = "Validation failed:\n" + "\n".join(f"  - {e}" for e in errors)
                raise ValueError(error_msg)

        # Resolve imports
        imports = spec.get("imports", [])
        imports_content = self.resolver.resolve(imports) if imports else ""

        if debug:
            print(f"[DEBUG] Resolved imports: {self.resolver.get_resolution_order()}")
            print(f"[DEBUG] Imports content length: {len(imports_content)} chars")

        # Extract other sections
        context = spec.get("context", {})
        features = spec.get("features", [])
        constraints = spec.get("constraints", [])

        # Get target formatter and compile
        formatter = self.SUPPORTED_TARGETS[target]
        compiled = formatter(imports_content, context, features, constraints)

        return compiled

    def _load_spec(self, prompt_file: Path) -> Dict[str, Any]:
        """
        Load and parse a .prompt file.

        Args:
            prompt_file: Path to .prompt file

        Returns:
            Parsed spec dictionary

        Raises:
            FileNotFoundError: If file doesn't exist
            yaml.YAMLError: If YAML is invalid
        """
        if not prompt_file.exists():
            raise FileNotFoundError(f"Prompt file not found: {prompt_file}")

        content = prompt_file.read_text(encoding="utf-8")

        try:
            spec = yaml.safe_load(content)
            if spec is None:
                spec = {}
            return spec
        except yaml.YAMLError as e:
            raise yaml.YAMLError(f"Invalid YAML in {prompt_file}: {e}")

    def get_supported_targets(self) -> list:
        """Get list of supported target formats."""
        return list(self.SUPPORTED_TARGETS.keys())
