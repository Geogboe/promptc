# ADR-001: Remove Agent-Specific Targets

## Status
Accepted

## Context

promptc v0.2 compiled `.prompt` specs into 5 target formats: `raw`, `claude`, `cursor`, `aider`, `copilot`. Each formatter produced slightly different markdown tailored to how each tool consumed system prompts.

This approach had several problems:

1. **Maintenance burden**: Every schema change (new section, new field) required updating 5 formatters in sync.
2. **False specificity**: The formatting differences between targets were cosmetic — all tools accept generic markdown. The "claude format" was just markdown with `#` headings; "cursor format" was the same content with different heading text. No target used any agent-specific API or file format.
3. **Fragile coupling**: promptc shouldn't need to track which heading style each AI tool prefers, as tools change their preferences independently of the spec.
4. **Scope creep risk**: The target system invited requests for more targets (Gemini, Copilot Workspace, etc.), each adding maintenance surface with minimal user value.

## Decision

Remove all 5 target formatters. Produce a single `instructions.promptc` output using clean, generic markdown with an HTML comment header identifying the source spec and version.

The output is model-agnostic: any AI tool that accepts a context file or system prompt can consume it. Users copy or symlink `instructions.promptc` to wherever their tool expects it (`.cursorrules`, `CLAUDE.md`, etc.).

## Consequences

**Easier:**
- New spec fields (e.g., `resources`, `build`) only need one formatter update
- Output is human-readable and diffable without knowing which target was used
- No decision paralysis for users ("which target should I use?")

**Harder:**
- Users who relied on target-specific formatting must adapt (no backwards compatibility in v1)
- Tooling that ingested the old target output format must be updated

**Trade-off accepted:** The cosmetic formatting differences between targets provided near-zero value relative to the maintenance cost. Generic markdown is correct for every AI tool we're aware of.
