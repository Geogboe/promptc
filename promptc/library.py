"""Library management for prompt libraries."""

import os
from pathlib import Path
from typing import Optional, List


class LibraryManager:
    """Manages prompt library resolution and loading."""

    def __init__(self, project_dir: Optional[Path] = None):
        """
        Initialize library manager.

        Args:
            project_dir: Project directory to look for local prompts
        """
        self.project_dir = project_dir or Path.cwd()
        self.search_paths = self._get_search_paths()

    def _get_search_paths(self) -> List[Path]:
        """Get ordered list of paths to search for prompt libraries."""
        paths = []

        # 1. Project-local prompts
        project_prompts = self.project_dir / "prompts"
        if project_prompts.exists():
            paths.append(project_prompts)

        # 2. Global user prompts
        home_prompts = Path.home() / ".prompts"
        if home_prompts.exists():
            paths.append(home_prompts)

        # 3. Built-in defaults (in package)
        package_dir = Path(__file__).parent
        default_prompts = package_dir / "defaults"
        if default_prompts.exists():
            paths.append(default_prompts)

        return paths

    def resolve(self, import_name: str) -> Optional[Path]:
        """
        Resolve an import name to a file path.

        Args:
            import_name: Import name like "patterns.rest_api" or "company.standards"

        Returns:
            Path to the prompt file, or None if not found
        """
        # Convert dot notation to path (patterns.rest_api -> patterns/rest_api.prompt)
        relative_path = import_name.replace(".", os.sep) + ".prompt"

        for search_path in self.search_paths:
            candidate = search_path / relative_path
            if candidate.exists():
                return candidate

        return None

    def load(self, import_name: str) -> str:
        """
        Load prompt content from a library.

        Args:
            import_name: Import name to load

        Returns:
            Content of the prompt file

        Raises:
            FileNotFoundError: If the import cannot be resolved
        """
        path = self.resolve(import_name)
        if path is None:
            raise FileNotFoundError(
                f"Cannot resolve import '{import_name}'. "
                f"Searched in: {', '.join(str(p) for p in self.search_paths)}"
            )

        return path.read_text(encoding="utf-8")

    def list_libraries(self) -> dict:
        """
        List all available prompt libraries organized by source.

        Returns:
            Dictionary with sources as keys and lists of libraries as values
        """
        libraries = {
            "project": [],
            "global": [],
            "built-in": []
        }

        for search_path in self.search_paths:
            source_key = self._get_source_key(search_path)
            libs = self._scan_directory(search_path)
            libraries[source_key].extend(libs)

        return libraries

    def _get_source_key(self, path: Path) -> str:
        """Get the source key for a search path."""
        if "defaults" in str(path):
            return "built-in"
        elif path.name == "prompts" and path.parent == self.project_dir:
            return "project"
        else:
            return "global"

    def _scan_directory(self, base_path: Path, prefix: str = "") -> List[str]:
        """
        Recursively scan directory for .prompt files.

        Args:
            base_path: Base directory to scan
            prefix: Prefix for library names (for nested directories)

        Returns:
            List of library import names
        """
        libraries = []

        if not base_path.exists():
            return libraries

        for item in sorted(base_path.iterdir()):
            if item.is_file() and item.suffix == ".prompt":
                # Convert to import name
                lib_name = item.stem
                if prefix:
                    lib_name = f"{prefix}.{lib_name}"
                libraries.append(lib_name)
            elif item.is_dir() and not item.name.startswith("."):
                # Recursively scan subdirectories
                subdir_prefix = item.name if not prefix else f"{prefix}.{item.name}"
                libraries.extend(self._scan_directory(item, subdir_prefix))

        return libraries
