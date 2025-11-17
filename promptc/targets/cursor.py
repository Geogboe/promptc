"""Cursor formatter - outputs prompt in Cursor rules format."""

from typing import Dict, Any


def format_prompt(
    imports_content: str,
    context: Dict[str, Any],
    features: list,
    constraints: list
) -> str:
    """
    Format prompt for Cursor (.cursorrules format).

    Args:
        imports_content: Combined content from all imports
        context: Context dictionary
        features: List of features
        constraints: List of constraints

    Returns:
        Formatted prompt for Cursor
    """
    parts = []

    # Add imports content
    if imports_content:
        parts.append(imports_content)

    # Add context
    if context:
        parts.append("Project Context:")
        for key, value in context.items():
            parts.append(f"- {key}: {value}")

    # Add constraints
    if constraints:
        parts.append("Constraints:")
        for constraint in constraints:
            if isinstance(constraint, dict):
                for key, value in constraint.items():
                    if key != "exclude":
                        parts.append(f"- {key}: {value}")
            else:
                parts.append(f"- {constraint}")

    # Add features
    if features:
        parts.append("Features to implement:")
        for feature in features:
            if isinstance(feature, dict):
                for name, description in feature.items():
                    parts.append(f"- {name}: {description}")
            else:
                parts.append(f"- {feature}")

    return "\n\n".join(parts)
