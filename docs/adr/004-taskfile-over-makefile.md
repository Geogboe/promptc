# ADR-004: Taskfile over Makefile

## Status
Accepted

## Context

promptc used a `Makefile` for build automation. Makefiles have significant friction on Windows (requires MinGW/WSL/Git Bash) and have syntax quirks (tab indentation, implicit rules, variable expansion differences) that make them hard to maintain across platforms.

## Decision

Replace `Makefile` with `Taskfile.yml` using [go-task](https://taskfile.dev).

## Rationale

1. **Cross-platform**: Taskfile ships bundled coreutils for Windows (`cp`, `rm`, `mkdir`), so tasks work identically on Linux, macOS, and Windows without Git Bash or WSL.
2. **YAML syntax**: Familiar to Go developers who already use YAML for CI configs; no tab indentation traps.
3. **Matrix builds**: `for` loop with structured items enables clean cross-platform matrix builds without repeating commands.
4. **Internal tasks**: The `internal: true` flag hides implementation tasks from `task --list`, keeping the public API clean.
5. **Source/generates tracking**: Taskfile can skip tasks when sources haven't changed (like Make), using checksums or timestamps.
6. **Local CI parity**: The `ci` task runs the same steps as GitHub Actions, so "passes locally" actually means something.
7. **Go install**: `go install github.com/go-task/task/v3/cmd/task@latest` — fits naturally into Go developer workflows.

## Consequences

**Easier:**
- Windows contributors can run `task build` without MinGW or WSL
- `task ci` provides a single command that mirrors GitHub Actions exactly
- Matrix build targets are defined as data (list of objects) rather than repeated commands

**Harder:**
- Requires installing `task` (one-time: `go install github.com/go-task/task/v3/cmd/task@latest`)
- Team members familiar with Make must learn new syntax

**Trade-off accepted:** The install cost is minimal for a Go project. Cross-platform consistency is worth the switch.
