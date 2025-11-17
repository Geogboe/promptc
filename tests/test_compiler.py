"""Tests for the prompt compiler."""

import pytest
from pathlib import Path
from promptc.compiler import PromptCompiler
import tempfile
import yaml


class TestPromptCompiler:
    """Test cases for PromptCompiler."""

    @pytest.fixture
    def temp_prompt_file(self, tmp_path):
        """Create a temporary prompt file."""
        prompt_file = tmp_path / "test.prompt"
        spec = {
            "context": {"language": "python"},
            "features": [{"test": "Test feature"}],
            "constraints": ["test_constraint"]
        }
        prompt_file.write_text(yaml.dump(spec))
        return prompt_file

    def test_compiler_initialization(self):
        """Test compiler initialization."""
        compiler = PromptCompiler()
        assert compiler is not None
        assert compiler.library_manager is not None
        assert compiler.resolver is not None

    def test_get_supported_targets(self):
        """Test getting supported targets."""
        compiler = PromptCompiler()
        targets = compiler.get_supported_targets()

        assert "raw" in targets
        assert "claude" in targets
        assert "cursor" in targets
        assert "aider" in targets
        assert "copilot" in targets

    def test_compile_nonexistent_file(self):
        """Test compilation of nonexistent file."""
        compiler = PromptCompiler()

        with pytest.raises(FileNotFoundError):
            compiler.compile(Path("nonexistent.prompt"))

    def test_compile_invalid_target(self, temp_prompt_file):
        """Test compilation with invalid target."""
        compiler = PromptCompiler()

        with pytest.raises(ValueError, match="Unsupported target"):
            compiler.compile(temp_prompt_file, target="invalid")

    def test_compile_basic_spec(self, temp_prompt_file):
        """Test compilation of a basic spec."""
        compiler = PromptCompiler()

        result = compiler.compile(temp_prompt_file, target="raw")

        assert result is not None
        assert "python" in result
        assert "Test feature" in result

    def test_compile_all_targets(self, temp_prompt_file):
        """Test compilation to all supported targets."""
        compiler = PromptCompiler()
        targets = compiler.get_supported_targets()

        for target in targets:
            result = compiler.compile(temp_prompt_file, target=target)
            assert result is not None
            assert len(result) > 0

    def test_compile_with_validation(self, tmp_path):
        """Test compilation with validation enabled."""
        compiler = PromptCompiler()

        # Create invalid prompt file
        prompt_file = tmp_path / "invalid.prompt"
        spec = {
            "features": []  # Empty features - invalid
        }
        prompt_file.write_text(yaml.dump(spec))

        with pytest.raises(ValueError, match="Validation failed"):
            compiler.compile(prompt_file, validate=True)

    def test_compile_without_validation(self, tmp_path):
        """Test compilation with validation disabled."""
        compiler = PromptCompiler()

        # Create invalid prompt file
        prompt_file = tmp_path / "invalid.prompt"
        spec = {
            "features": []  # Empty features - invalid
        }
        prompt_file.write_text(yaml.dump(spec))

        # Should not raise with validation disabled
        result = compiler.compile(prompt_file, validate=False)
        assert result is not None
