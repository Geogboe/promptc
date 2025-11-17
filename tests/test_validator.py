"""Tests for the prompt validator."""

import pytest
from promptc.validator import PromptValidator


class TestPromptValidator:
    """Test cases for PromptValidator."""

    def test_valid_basic_spec(self):
        """Test validation of a valid basic spec."""
        spec = {
            "imports": ["patterns.rest_api"],
            "context": {"language": "python"},
            "features": [{"feature1": "description"}],
            "constraints": ["no_hardcoded_secrets"]
        }

        is_valid, errors = PromptValidator.validate(spec)

        assert is_valid
        assert len(errors) == 0

    def test_empty_spec(self):
        """Test validation of an empty spec."""
        spec = {}

        is_valid, errors = PromptValidator.validate(spec)

        assert is_valid
        assert len(errors) == 0

    def test_invalid_top_level_key(self):
        """Test detection of invalid top-level keys."""
        spec = {
            "invalid_key": "value"
        }

        is_valid, errors = PromptValidator.validate(spec)

        assert not is_valid
        assert any("Unknown top-level key" in e for e in errors)

    def test_invalid_imports_type(self):
        """Test detection of invalid imports type."""
        spec = {
            "imports": "should be a list"
        }

        is_valid, errors = PromptValidator.validate(spec)

        assert not is_valid
        assert any("imports' must be a list" in e for e in errors)

    def test_invalid_import_name(self):
        """Test detection of invalid import names."""
        spec = {
            "imports": ["valid.import", "invalid-import"]
        }

        is_valid, errors = PromptValidator.validate(spec)

        assert not is_valid
        assert any("Invalid import name" in e for e in errors)

    def test_invalid_context_type(self):
        """Test detection of invalid context type."""
        spec = {
            "context": ["should", "be", "dict"]
        }

        is_valid, errors = PromptValidator.validate(spec)

        assert not is_valid
        assert any("context' must be a dictionary" in e for e in errors)

    def test_invalid_context_value_type(self):
        """Test detection of invalid context value types."""
        spec = {
            "context": {
                "valid": "string",
                "invalid": {"nested": "dict"}
            }
        }

        is_valid, errors = PromptValidator.validate(spec)

        assert not is_valid
        assert any("must be a simple type" in e for e in errors)

    def test_empty_features_list(self):
        """Test detection of empty features list."""
        spec = {
            "features": []
        }

        is_valid, errors = PromptValidator.validate(spec)

        assert not is_valid
        assert any("features' list is empty" in e for e in errors)

    def test_valid_features_formats(self):
        """Test valid feature formats."""
        spec = {
            "features": [
                "simple_string_feature",
                {"key_value": "description"}
            ]
        }

        is_valid, errors = PromptValidator.validate(spec)

        assert is_valid
        assert len(errors) == 0

    def test_invalid_feature_type(self):
        """Test detection of invalid feature types."""
        spec = {
            "features": [123, True, None]
        }

        is_valid, errors = PromptValidator.validate(spec)

        assert not is_valid
        assert len(errors) >= 3

    def test_valid_constraints(self):
        """Test valid constraint formats."""
        spec = {
            "constraints": [
                "simple_constraint",
                {"key": "value"},
                {"exclude": "something"}
            ]
        }

        is_valid, errors = PromptValidator.validate(spec)

        assert is_valid
        assert len(errors) == 0
