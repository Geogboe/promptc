"""Validation for .prompt files."""

from typing import Dict, Any, List, Tuple


class PromptValidator:
    """Validates .prompt file specifications."""

    @staticmethod
    def validate(spec: Dict[str, Any]) -> Tuple[bool, List[str]]:
        """
        Validate a prompt specification.

        Args:
            spec: Parsed prompt specification dictionary

        Returns:
            Tuple of (is_valid, list_of_errors)
        """
        errors = []

        # Check for valid top-level keys
        valid_keys = {"imports", "context", "features", "constraints"}
        for key in spec.keys():
            if key not in valid_keys:
                errors.append(f"Unknown top-level key: '{key}'. Valid keys are: {', '.join(valid_keys)}")

        # Validate imports
        if "imports" in spec:
            import_errors = PromptValidator._validate_imports(spec["imports"])
            errors.extend(import_errors)

        # Validate context
        if "context" in spec:
            context_errors = PromptValidator._validate_context(spec["context"])
            errors.extend(context_errors)

        # Validate features
        if "features" in spec:
            feature_errors = PromptValidator._validate_features(spec["features"])
            errors.extend(feature_errors)

        # Validate constraints
        if "constraints" in spec:
            constraint_errors = PromptValidator._validate_constraints(spec["constraints"])
            errors.extend(constraint_errors)

        return len(errors) == 0, errors

    @staticmethod
    def _validate_imports(imports: Any) -> List[str]:
        """Validate imports section."""
        errors = []

        if not isinstance(imports, list):
            errors.append("'imports' must be a list")
            return errors

        for i, imp in enumerate(imports):
            if not isinstance(imp, str):
                errors.append(f"Import at index {i} must be a string, got {type(imp).__name__}")
            elif imp and not imp.startswith("!"):
                # Check for valid import name format (alphanumeric, dots, underscores)
                if not all(c.isalnum() or c in "._" for c in imp):
                    errors.append(f"Invalid import name '{imp}': use only letters, numbers, dots, and underscores")

        return errors

    @staticmethod
    def _validate_context(context: Any) -> List[str]:
        """Validate context section."""
        errors = []

        if not isinstance(context, dict):
            errors.append("'context' must be a dictionary")
            return errors

        # Context values should be simple types (strings, numbers, booleans)
        for key, value in context.items():
            if not isinstance(value, (str, int, float, bool, type(None))):
                errors.append(
                    f"Context value for '{key}' must be a simple type "
                    f"(string, number, boolean), got {type(value).__name__}"
                )

        return errors

    @staticmethod
    def _validate_features(features: Any) -> List[str]:
        """Validate features section."""
        errors = []

        if not isinstance(features, list):
            errors.append("'features' must be a list")
            return errors

        if len(features) == 0:
            errors.append("'features' list is empty - add at least one feature")

        for i, feature in enumerate(features):
            if isinstance(feature, dict):
                # Feature as key-value pair
                if len(feature) != 1:
                    errors.append(f"Feature at index {i} should have exactly one key-value pair")
                for key, value in feature.items():
                    if not isinstance(value, str):
                        errors.append(f"Feature description for '{key}' must be a string")
            elif isinstance(feature, str):
                # Simple string feature
                if not feature.strip():
                    errors.append(f"Feature at index {i} is an empty string")
            else:
                errors.append(f"Feature at index {i} must be a string or dictionary, got {type(feature).__name__}")

        return errors

    @staticmethod
    def _validate_constraints(constraints: Any) -> List[str]:
        """Validate constraints section."""
        errors = []

        if not isinstance(constraints, list):
            errors.append("'constraints' must be a list")
            return errors

        for i, constraint in enumerate(constraints):
            if isinstance(constraint, dict):
                # Constraint as key-value pair
                for key, value in constraint.items():
                    if key == "exclude":
                        # Special case for exclusions
                        continue
                    if not isinstance(value, (str, bool)):
                        errors.append(
                            f"Constraint value for '{key}' must be a string or boolean, "
                            f"got {type(value).__name__}"
                        )
            elif isinstance(constraint, str):
                # Simple string constraint
                if not constraint.strip():
                    errors.append(f"Constraint at index {i} is an empty string")
            else:
                errors.append(
                    f"Constraint at index {i} must be a string or dictionary, "
                    f"got {type(constraint).__name__}"
                )

        return errors
