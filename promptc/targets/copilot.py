"""GitHub Copilot formatter - outputs prompt for .github/copilot-instructions.md."""

from typing import Dict, Any


def format_prompt(
    imports_content: str,
    context: Dict[str, Any],
    features: list,
    constraints: list
) -> str:
    """
    Format prompt for GitHub Copilot (.github/copilot-instructions.md).

    Args:
        imports_content: Combined content from all imports
        context: Context dictionary
        features: List of features
        constraints: List of constraints

    Returns:
        Formatted prompt for GitHub Copilot
    """
    parts = []

    parts.append("# GitHub Copilot Instructions")
    parts.append("")
    parts.append("This document provides instructions for GitHub Copilot when working on this project.")
    parts.append("")

    # Add imports content
    if imports_content:
        parts.append("## Development Guidelines")
        parts.append("")
        parts.append(imports_content)
        parts.append("")

    # Add context
    if context:
        parts.append("## Project Information")
        parts.append("")
        parts.append("This project uses:")
        for key, value in context.items():
            parts.append(f"- **{key.replace('_', ' ').title()}**: {value}")
        parts.append("")

    # Add constraints
    if constraints:
        parts.append("## Coding Standards and Constraints")
        parts.append("")
        parts.append("When generating code, adhere to these requirements:")
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
        parts.append("## Current Development Focus")
        parts.append("")
        parts.append("The following features are being implemented:")
        parts.append("")
        for feature in features:
            if isinstance(feature, dict):
                for name, description in feature.items():
                    parts.append(f"### {name}")
                    parts.append(f"{description}")
                    parts.append("")
            else:
                parts.append(f"- {feature}")

    return "\n".join(parts)
