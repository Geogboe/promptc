"""Aider formatter - outputs prompt in Aider configuration format."""

from typing import Dict, Any


def format_prompt(
    imports_content: str,
    context: Dict[str, Any],
    features: list,
    constraints: list
) -> str:
    """
    Format prompt for Aider (.aider.conf.yml or .aider.txt format).

    Args:
        imports_content: Combined content from all imports
        context: Context dictionary
        features: List of features
        constraints: List of constraints

    Returns:
        Formatted prompt for Aider
    """
    parts = []

    # Aider works well with a clear system message format
    parts.append("# Aider Instructions")
    parts.append("")

    # Add imports content as the foundation
    if imports_content:
        parts.append("## Guidelines")
        parts.append("")
        parts.append(imports_content)
        parts.append("")

    # Add context information
    if context:
        parts.append("## Project Context")
        parts.append("")
        for key, value in context.items():
            parts.append(f"- {key}: {value}")
        parts.append("")

    # Add constraints
    if constraints:
        parts.append("## Requirements and Constraints")
        parts.append("")
        for constraint in constraints:
            if isinstance(constraint, dict):
                for key, value in constraint.items():
                    if key != "exclude":
                        parts.append(f"- {key.replace('_', ' ').title()}: {value}")
            else:
                parts.append(f"- {constraint.replace('_', ' ').title()}")
        parts.append("")

    # Add features
    if features:
        parts.append("## Features to Implement")
        parts.append("")
        for feature in features:
            if isinstance(feature, dict):
                for name, description in feature.items():
                    parts.append(f"- {name}: {description}")
            else:
                parts.append(f"- {feature}")

    return "\n".join(parts)
