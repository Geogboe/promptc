# promptc - Prompt Compiler for Agentic Programming

A fast, zero-dependency CLI tool (written in Go) that compiles `.spec.promptc` files into a single `instructions.promptc` output. Stop copy-pasting instructions between projects and tools.

## The Problem

As developers using AI coding assistants, we face a few recurring problems:

- **Scattered instructions**: The same guidance gets duplicated across repos and tool-specific files.
- **No reusability**: Common patterns such as REST APIs, testing, and security are copied by hand.
- **Inconsistency**: Instructions drift between projects over time.
- **No composition**: It is hard to build on shared prompt libraries.

## The Solution

`promptc` gives you a simple compiler flow:

1. Write instructions once in `.spec.promptc`
2. Import reusable prompt libraries
3. Compile to a single model-agnostic `instructions.promptc`
4. Optionally run `promptc build` to fetch resources and execute an agent in a sandbox

### Mental Model

```
[Reusable Prompt Libraries] + [Project Spec] → Compiler → instructions.promptc
```

## Installation

### Shell script (Linux / macOS) — Recommended

```sh
curl -fsSL https://raw.githubusercontent.com/Geogboe/promptc/main/install.sh | sh
```

Installs to `/usr/local/bin` if writable, otherwise `~/.local/bin`. Override with `PROMPTC_INSTALL_DIR`.

### PowerShell (Windows)

```powershell
irm https://raw.githubusercontent.com/Geogboe/promptc/main/install.ps1 | iex
```

Installs to `%USERPROFILE%\.local\bin` and adds it to your user PATH automatically. Override with `$env:PROMPTC_INSTALL_DIR`.
The install scripts resolve published release assets from the GitHub release metadata instead of guessing archive names.

### go install (requires Go 1.21+)

```sh
go install github.com/Geogboe/promptc/cmd/promptc@latest
```

### Manual download

Download the latest release archive for your platform from the [releases page](https://github.com/Geogboe/promptc/releases/latest).

### Build from Source

```bash
git clone https://github.com/Geogboe/promptc.git
cd promptc
task build
# Or: go build -o bin/promptc.exe ./cmd/promptc
```

## Examples

The `examples/` directory is the canonical place to learn promptc through real workflows.
Start with [examples/README.md](examples/README.md) for runnable examples and expected outputs.

## Quick Start

### 1. Initialize a new project

```bash
promptc init myapp.spec.promptc --template=web-api
```

### 2. Edit your `.spec.promptc` file

```yaml
imports:
  - patterns.rest_api
  - constraints.security

context:
  language: go
  framework: fiber
  database: postgresql

features:
  - api_endpoint: "User authentication and JWT tokens"

constraints:
  - no_hardcoded_secrets
  - comprehensive_test_coverage
```

### 3. Validate and compile

```bash
promptc validate myapp.spec.promptc
promptc compile myapp.spec.promptc
```

## Features

### Built-in Prompt Libraries (8 total)

**Patterns:**
- `patterns.rest_api` - REST API best practices
- `patterns.testing` - Testing guidelines
- `patterns.database` - Database design and queries
- `patterns.async_programming` - Async/await patterns

**Constraints:**
- `constraints.security` - Security best practices
- `constraints.code_quality` - Code quality standards
- `constraints.performance` - Performance optimization
- `constraints.accessibility` - WCAG compliance

### Project Templates

Initialize projects with templates:

```bash
promptc init myapp.spec.promptc --template=basic      # Basic template
promptc init myapp.spec.promptc --template=web-api    # Web API template
promptc init myapp.spec.promptc --template=cli-tool   # CLI tool template
```

## Commands

### Validate Specs

```bash
promptc validate myapp.spec.promptc
```

### Compile Prompts

```bash
promptc compile myapp.spec.promptc
promptc compile myapp.spec.promptc --output=./output
promptc compile myapp.spec.promptc --debug
promptc compile myapp.spec.promptc --no-validate
```

### Build Projects

```bash
promptc build myapp.spec.promptc --dry-run
promptc build myapp.spec.promptc
```

### List Libraries

```bash
promptc list
promptc list --verbose
```

### Initialize Projects

```bash
promptc init [name.spec.promptc] --template=<template>
```

## Advanced Usage

### Import Resolution

Prompts are resolved from multiple locations in order:

1. **Project-local**: `./prompts/` in your project
2. **Global**: `~/.prompts/` in your home directory
3. **Built-in**: Embedded in the binary

### Exclusions

Exclude specific imports with the `!` prefix:

```yaml
imports:
  - patterns.rest_api
  - !patterns.rest_api.verbose_logging
```

### Custom Libraries

Create your own reusable libraries:

**~/.prompts/company/standards.prompt:**
```
# Company Coding Standards

## Code Review
- All PRs require 2 approvals
- Must pass CI/CD pipeline
```

Use in any project:

```yaml
imports:
  - company.standards
```

## Use Cases

### 1. Team Knowledge Base

Share prompt libraries across your team:

```bash
# Create team prompts repository
mkdir ~/.prompts/team
# Everyone uses the same standards
imports:
  - team.api_standards
```

## Why Go?

The Go rewrite provides significant advantages:

- ✅ **Single binary** - No Python runtime or dependencies needed
- ✅ **Fast startup** - Instant compilation (< 10ms for most operations)
- ✅ **Cross-platform** - Easy distribution for Linux, macOS, Windows
- ✅ **Small footprint** - ~8MB static binary
- ✅ **Embedded resources** - Built-in libraries compiled into binary

## Technical Details

- **Language**: Go 1.21+
- **Binary Size**: ~8MB (static binary)
- **Dependencies**: Zero runtime dependencies
- **Built-in Libraries**: Embedded using go:embed
- **Performance**: Sub-10ms compilation for most operations
- **CLI Framework**: Cobra
- **YAML Parser**: gopkg.in/yaml.v3

## Development

### Building

```bash
task build          # Build for current platform
task build-all      # Build for all platforms
task test           # Run tests
task clean          # Clean build artifacts
```

### Testing

```bash
go test ./...               # Run all tests
go test -v ./internal/...   # Verbose output
```

## Roadmap

- [x] Core compilation engine
- [x] Generic `instructions.promptc` output
- [x] Built-in prompt libraries (8 libraries)
- [x] Project templates
- [x] Validation system
- [x] Go rewrite for performance
- [x] Build pipeline with resource fetching and sandboxed agent execution
- [x] Examples directory for real workflows
- [ ] Watch mode for auto-recompilation
- [ ] Library versioning
- [ ] Published library packages
- [ ] VSCode extension

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Why promptc?

As AI-assisted development becomes standard, managing instructions across different tools is a growing pain. `promptc` treats prompts as code:

- **Fast & Portable** - Single binary, no dependencies
- **Version controlled** - Track changes to your instructions
- **Composable** - Build on reusable libraries
- **Testable** - See exactly what you're sending to the LLM
- **Shareable** - Collaborate on prompt engineering

Just like we don't copy-paste code between projects, we shouldn't copy-paste prompts. Compile them instead.
