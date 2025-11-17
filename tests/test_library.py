"""Tests for the library manager."""

import pytest
from pathlib import Path
from promptc.library import LibraryManager
import tempfile


class TestLibraryManager:
    """Test cases for LibraryManager."""

    def test_initialization(self):
        """Test library manager initialization."""
        manager = LibraryManager()
        assert manager is not None
        assert manager.project_dir is not None
        assert manager.search_paths is not None

    def test_search_paths_order(self):
        """Test that search paths are in correct order."""
        manager = LibraryManager()

        # Should have built-in path at minimum
        assert len(manager.search_paths) >= 1

        # Last path should be built-in defaults
        assert "defaults" in str(manager.search_paths[-1])

    def test_resolve_builtin_library(self):
        """Test resolving a built-in library."""
        manager = LibraryManager()

        path = manager.resolve("patterns.rest_api")

        assert path is not None
        assert path.exists()
        assert path.name == "rest_api.prompt"

    def test_resolve_nonexistent_library(self):
        """Test resolving a nonexistent library."""
        manager = LibraryManager()

        path = manager.resolve("nonexistent.library")

        assert path is None

    def test_load_builtin_library(self):
        """Test loading a built-in library."""
        manager = LibraryManager()

        content = manager.load("patterns.rest_api")

        assert content is not None
        assert len(content) > 0
        assert "REST API" in content

    def test_load_nonexistent_library_raises(self):
        """Test that loading nonexistent library raises error."""
        manager = LibraryManager()

        with pytest.raises(FileNotFoundError):
            manager.load("nonexistent.library")

    def test_list_libraries(self):
        """Test listing all available libraries."""
        manager = LibraryManager()

        libraries = manager.list_libraries()

        assert libraries is not None
        assert "built-in" in libraries
        assert len(libraries["built-in"]) > 0

        # Should have our default libraries
        builtin = libraries["built-in"]
        assert "patterns.rest_api" in builtin
        assert "patterns.testing" in builtin
        assert "constraints.security" in builtin
        assert "constraints.code_quality" in builtin

    def test_project_local_override(self, tmp_path):
        """Test that project-local prompts override built-ins."""
        # Create project-local prompts directory
        prompts_dir = tmp_path / "prompts" / "patterns"
        prompts_dir.mkdir(parents=True)

        # Create a custom rest_api.prompt
        custom_prompt = prompts_dir / "rest_api.prompt"
        custom_prompt.write_text("Custom REST API content")

        # Initialize manager with project dir
        manager = LibraryManager(project_dir=tmp_path)

        # Should resolve to our custom file
        path = manager.resolve("patterns.rest_api")
        content = path.read_text() if path else ""

        assert "Custom REST API content" in content
