# ADR-005: Spec and Output File Naming

## Status
Accepted

## Context

We need to choose file extensions/names for:
1. **Input spec files** (previously `.prompt`)
2. **Compiled output files** (previously stdout or arbitrary `-o` path)

### Input: Candidates considered

| Option | Notes |
|--------|-------|
| `.prompt` | Original. Ambiguous — "prompt" is overloaded (UI prompt, LLM prompt, bash prompt) |
| `.promptc` | Short, unique, matches tool name. No `.spec` disambiguator. |
| `.spec.promptc` | Double extension: `.spec` signals "this is a spec file"; `.promptc` signals "processed by promptc". Mirrors patterns like `.test.ts`, `.config.js` |
| `.promptcspec` | Single token, no ambiguity. Ugly. |

### Output: Candidates considered

| Option | Notes |
|--------|-------|
| `instructions.promptc` | Mirrors input extension, clearly outputs from the tool. Somewhat verbose. |
| `instructions.md` | Generic markdown — loses the promptc identity; could be confused with docs |
| `AGENTS.md` | Familiar to Claude Code users, but couples us to one agent tool |
| `<specname>.out` | Conventional but doesn't convey content type |
| `out.promptc` | Short but not descriptive |

## Decision

**Input**: `.spec.promptc` — the double extension explicitly communicates both the file's purpose (spec) and the tool that processes it (promptc).

**Output**: `instructions.promptc` — kept to maintain the promptc identity in the output. The filename is stable and predictable regardless of the input spec name, which makes it easy to reference in agent configs (e.g., `CLAUDE.md: see instructions.promptc`).

## Consequences

**Easier:**
- Clear visual distinction between spec files (`*.spec.promptc`) and output files (`instructions.promptc`) in the same directory
- The `.promptc` suffix on both input and output makes all promptc-related files greppable
- Users can check `.gitignore` for `*.promptc` to ignore output, or `*.spec.promptc` to track specs

**Harder:**
- Double extension is slightly unusual; may confuse editors that associate extensions with file type
- `instructions.promptc` is not conventional markdown so editors won't syntax-highlight it by default (workaround: configure editor to treat `.promptc` as markdown)

**Trade-off accepted:** The naming convention is self-documenting and consistent. Editor syntax highlighting can be fixed with a `.editorconfig` or editor-specific mapping.
