"""Claude formatter - outputs prompt in Claude-friendly markdown format."""

from typing import Dict, Any


def format_prompt(
    imports_content: str,
    context: Dict[str, Any],
    features: list,
    constraints: list
) -> str:
    """
    Format prompt for Claude (markdown format for .claude/instructions.md).

    Args:
        imports_content: Combined content from all imports
        context: Context dictionary
        features: List of features
        constraints: List of constraints

    Returns:
        Formatted prompt in Claude markdown format
    """
    parts = []

    # Add imports content as the foundation
    if imports_content:
        parts.append(imports_content)

    # Add context information
    if context:
        parts.append("# Project Context")
        context_items = []
        for key, value in context.items():
            context_items.append(f"- **{key}**: {value}")
        parts.append("\n".join(context_items))

    # Add constraints
    if constraints:
        parts.append("# Constraints and Requirements")
        constraint_items = []
        for constraint in constraints:
            if isinstance(constraint, dict):
                for key, value in constraint.items():
                    if key != "exclude":
                        constraint_items.append(f"- {key.replace('_', ' ').title()}: {value}")
            else:
                constraint_items.append(f"- {constraint.replace('_', ' ').title()}")
        parts.append("\n".join(constraint_items))

    # Add features
    if features:
        parts.append("# Features to Implement")
        feature_items = []
        for feature in features:
            if isinstance(feature, dict):
                for name, description in feature.items():
                    feature_items.append(f"- **{name}**: {description}")
            else:
                feature_items.append(f"- {feature}")
        parts.append("\n".join(feature_items))

    return "\n\n".join(parts)
