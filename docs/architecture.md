# promptc Architecture

## Overview

promptc is a spec-driven prompt compiler for agentic programming. It compiles `.spec.promptc` YAML files into a single generic `instructions.promptc` output, optionally fetching external resources and running an agent inside a sandbox.

## Component Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLI (cmd/promptc)                        │
│   validate │ compile │ build │ list │ init                      │
└────────────┬────────┬────────┴──────────────────────────────────┘
             │        │
             ▼        ▼
    ┌──────────────┐  ┌────────────────────────────────────────┐
    │  validator   │  │              compiler                  │
    │  Validate()  │  │  loadSpec → validate → resolve → fmt  │
    └──────────────┘  └──────────┬─────────────────────────────┘
                                 │ imports
             ┌───────────────────┼───────────────────┐
             ▼                   ▼                   ▼
    ┌──────────────┐   ┌──────────────────┐  ┌──────────────┐
    │   library    │   │    resolver      │  │  formatter   │
    │  Resolve()   │◄──│  Resolve()       │  │  Format()    │
    │  ListLibs()  │   │  (cycle detect)  │  │              │
    └──────────────┘   └──────────────────┘  └──────────────┘
         ▲
    ┌────┴──────────┐
    │ defaults/     │  ← embedded via go:embed
    │ patterns/     │
    │ constraints/  │
    └───────────────┘

Build pipeline (promptc build):

    ┌──────────────┐   ┌──────────────┐   ┌──────────────────┐
    │   compiler   │──►│   fetcher    │──►│    sandbox       │
    │  compile()   │   │  FetchAll()  │   │  Run(agent cmd)  │
    └──────────────┘   └──────────────┘   └──────────────────┘
         │                   │                     │
         ▼                   ▼                     ▼
    instructions        resources/          agent writes
    .promptc            <name>.md           generated code
                        <name>/             to outputDir/
```

## Package Dependency Graph

```
cmd/promptc
    └── internal/compiler
            ├── internal/validator
            ├── internal/library
            │       └── internal/defaults (embed)
            ├── internal/resolver
            │       └── internal/library
            └── internal/formatter

internal/builder
    ├── internal/compiler
    ├── internal/fetcher
    └── internal/sandbox

internal/progress  (no internal deps)
internal/fetcher   (no internal deps)
internal/sandbox   (no internal deps, external: docker SDK / bwrap)
```

## Data Flow

### `promptc validate <spec>`
1. Load spec YAML from file
2. Run `validator.Validate(spec)` against all top-level keys
3. Exit 0 on success, exit 1 with errors on failure

### `promptc compile <spec>`
1. Load + validate spec
2. Extract imports list; call `resolver.Resolve(imports)` → resolved library content string
3. Extract context, features, constraints from spec
4. Call `formatter.Format(spec, importsContent, meta)` → markdown string
5. Write to `<outputDir>/instructions.promptc`

### `promptc build <spec>`
1. All steps from `compile`
2. Call `fetcher.FetchAll(resources, outputDir/resources/)` concurrently
3. Construct `sandbox.Config` from spec.build section
4. Call `sandbox.Run(ctx, agent.command, agent.args)` with outputDir as workspace
5. Agent reads `instructions.promptc` + `resources/`, writes generated code to workspace

## Sandbox Provider Abstraction

```go
type Sandbox interface {
    Run(ctx context.Context, cmd string, args []string) error
}
```

Implementations:
- **docker**: Docker Go SDK (`github.com/docker/docker/client`), mounts outputDir as `/workspace`
- **bubblewrap**: shells out to `bwrap` binary (Linux only, build tag `linux`)
- **none**: direct exec with warning, no isolation

Selection: `sandbox.New(config)` returns the correct implementation based on `config.Type`.

## Output Directory Structure

```
<outputDir>/                   # default: ./<specname>/
    instructions.promptc       # compiled instructions (markdown)
    resources/                 # fetched resources (only with build/compile --fetch)
        <name>.md              # from URL resources
        <name>/                # from git resources
```

## Security Model

- **Path traversal**: `validateImportName()` in library package blocks `../` and absolute paths
- **File size limits**: 1MB for spec files, 10MB for library files
- **Symlink protection**: library loader resolves symlinks and verifies they stay within the prompts directory
- **Sandbox isolation**: Docker/bubblewrap contain agent execution; `none` mode warns explicitly
- **Resource caching**: SHA-256 hash of URL/ref before storing in `~/.cache/promptc/`
