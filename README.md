# promptc - Prompt Compiler for Agentic Programming

A fast, zero-dependency CLI tool (written in Go) that compiles LLM instructions into different formats. Stop copy-pasting instructions between Claude, Cursor, Aider, and other AI coding tools.

**New in v0.2.0**: Complete rewrite in Go for better performance, single binary distribution, and no runtime dependencies!

## The Problem

As developers using AI coding assistants, we face several challenges:

- **Scattered instructions**: Different tools need different formats (.cursorrules, .claude/instructions.md, etc.)
- **No reusability**: Common patterns (REST APIs, testing, security) are copy-pasted across projects
- **Inconsistency**: Instructions drift between tools and projects
- **No composition**: Can't build on shared knowledge bases

## The Solution

`promptc` provides a **compiler for prompts** that lets you:

1. Write instructions once in a simple YAML format
2. Import reusable prompt libraries (REST API patterns, security constraints, etc.)
3. Compile to any target format (Claude, Cursor, Aider, Copilot)
4. Share and version prompt libraries across projects

### Mental Model

```
[Reusable Prompt Libraries] + [Project Spec] → Compiler → [Target-specific prompt]
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
# Or: go build -o bin/promptc cmd/promptc/main.go
```

## Quick Start

### 1. Initialize a new project

```bash
promptc init myapp.prompt --template=web-api
```

### 2. Edit your .prompt file

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

### 3. Compile to your target format

```bash
# Compile for Claude
promptc compile myapp.prompt --target=claude --output=.claude/instructions.md

# Compile for Cursor
promptc compile myapp.prompt --target=cursor --output=.cursorrules

# Compile for Aider
promptc compile myapp.prompt --target=aider --output=.aider.txt

# Compile for GitHub Copilot
promptc compile myapp.prompt --target=copilot --output=.github/copilot-instructions.md
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

### Supported Targets (5 formats)

- **raw** - Plain text output
- **claude** - Markdown format for Claude
- **cursor** - Cursor rules format
- **aider** - Aider configuration format
- **copilot** - GitHub Copilot instructions

### Project Templates

Initialize projects with templates:

```bash
promptc init myapp.prompt --template=basic      # Basic template
promptc init myapp.prompt --template=web-api    # Web API template
promptc init myapp.prompt --template=cli-tool   # CLI tool template
```

## Commands

### List Available Libraries

```bash
# List all available prompt libraries
promptc list

# Show descriptions
promptc list --verbose
```

### Compile Prompts

```bash
# Compile to stdout
promptc compile myapp.prompt

# Compile to file
promptc compile myapp.prompt --target=claude --output=instructions.md

# Debug mode
promptc compile myapp.prompt --debug

# Skip validation
promptc compile myapp.prompt --no-validate
```

### Initialize Projects

```bash
promptc init [name] --template=<template>
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

### 1. Consistent Instructions Across Tools

```bash
promptc compile myapp.prompt --target=claude --output=.claude/instructions.md
promptc compile myapp.prompt --target=cursor --output=.cursorrules
promptc compile myapp.prompt --target=aider --output=.aider.txt
```

### 2. Team Knowledge Base

Share prompt libraries across your team:

```bash
# Create team prompts repository
mkdir ~/.prompts/team
# Everyone uses the same standards
imports:
  - team.api_standards
```

### 3. Project Templates

Create templates for common project types and share them.

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

### Project Structure

```
promptc/
├── cmd/promptc/          # Main entry point
├── internal/
│   ├── compiler/         # Core compilation logic
│   ├── library/          # Library management (with go:embed)
│   ├── resolver/         # Import resolution
│   ├── validator/        # Validation logic
│   └── targets/          # Target-specific formatters
├── Makefile              # Build automation
├── go.mod                # Go module definition
└── README.md             # This file
```

## Roadmap

- [x] Core compilation engine
- [x] Multiple target formats (5 targets)
- [x] Built-in prompt libraries (8 libraries)
- [x] Project templates
- [x] Validation system
- [x] Go rewrite for performance
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
