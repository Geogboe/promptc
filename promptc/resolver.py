"""Import resolution for prompt files."""

from typing import List, Set, Dict
from pathlib import Path
from .library import LibraryManager


class ImportResolver:
    """Resolves imports recursively with cycle detection and exclusions."""

    def __init__(self, library_manager: LibraryManager):
        """
        Initialize resolver.

        Args:
            library_manager: Library manager for resolving imports
        """
        self.library_manager = library_manager
        self._visited: Set[str] = set()
        self._exclusions: Set[str] = set()
        self._resolved_content: Dict[str, str] = {}

    def resolve(self, imports: List[str]) -> str:
        """
        Resolve all imports recursively.

        Args:
            imports: List of import names (may include ! prefix for exclusions)

        Returns:
            Combined content from all resolved imports
        """
        self._visited.clear()
        self._exclusions.clear()
        self._resolved_content.clear()

        # First pass: collect exclusions
        for import_name in imports:
            if import_name.startswith("!"):
                self._exclusions.add(import_name[1:])

        # Second pass: resolve imports
        content_parts = []
        for import_name in imports:
            if not import_name.startswith("!"):
                resolved = self._resolve_recursive(import_name)
                if resolved:
                    content_parts.append(resolved)

        return "\n\n".join(content_parts)

    def _resolve_recursive(self, import_name: str) -> str:
        """
        Recursively resolve a single import.

        Args:
            import_name: Import name to resolve

        Returns:
            Combined content from this import and its dependencies
        """
        # Check if excluded
        if import_name in self._exclusions:
            return ""

        # Check if already visited (cycle detection)
        if import_name in self._visited:
            return ""

        # Check if already resolved
        if import_name in self._resolved_content:
            return self._resolved_content[import_name]

        self._visited.add(import_name)

        # Load the content
        try:
            content = self.library_manager.load(import_name)
        except FileNotFoundError as e:
            raise FileNotFoundError(f"Failed to resolve import '{import_name}': {e}")

        # Store and return
        self._resolved_content[import_name] = content
        return content

    def get_resolution_order(self) -> List[str]:
        """
        Get the order in which imports were resolved.

        Returns:
            List of import names in resolution order
        """
        return list(self._visited)
