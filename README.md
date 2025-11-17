# promptc - Prompt Compiler for Agentic Programming

A tool that manages and compiles LLM instructions/context into different formats. Stop copy-pasting instructions between Claude, Cursor, Aider, and other AI coding tools.

## The Problem

As developers using AI coding assistants, we face several challenges:

- **Scattered instructions**: Different tools need different formats (.cursorrules, .claude/instructions.md, .aider.conf.yml)
- **No reusability**: Common patterns (REST APIs, testing, security) are copy-pasted across projects
- **Inconsistency**: Instructions drift between tools and projects
- **No composition**: Can't build on shared knowledge bases

## The Solution

`promptc` provides a **compiler for prompts** that lets you:

1. Write instructions once in a simple YAML format
2. Import reusable prompt libraries (REST API patterns, security constraints, etc.)
3. Compile to any target format (Claude, Cursor, Aider, etc.)
4. Share and version prompt libraries across projects

### Mental Model

```
[Reusable Prompt Libraries] + [Project Spec] → Compiler → [Target-specific prompt]
```

## Installation

```bash
pip install promptc
```

Or install from source:

```bash
git clone https://github.com/yourusername/promptc.git
cd promptc
pip install -e .
```

## Quick Start

### 1. Create a .prompt file

Create `myapp.prompt`:

```yaml
imports:
  - patterns.rest_api      # Built-in REST API best practices
  - constraints.security   # Built-in security constraints

context:
  language: python3.11
  framework: fastapi
  database: postgresql

features:
  - find_coffee_shops: "User finds nearby independent coffee shops"
  - save_favorites: "User saves favorite locations"
  - view_history: "User views their search history"

constraints:
  - no_hardcoded_secrets
  - comprehensive_error_handling
  - exclude: chain_stores
```

### 2. Compile to your target format

```bash
# Compile for Claude
promptc compile myapp.prompt --target=claude --output=.claude/instructions.md

# Compile for Cursor
promptc compile myapp.prompt --target=cursor --output=.cursorrules

# Just see the output
promptc compile myapp.prompt --target=raw
```

### 3. Use with your AI tool

The compiled instructions are now ready to use with Claude, Cursor, or any other AI coding assistant.

## Features

### Reusable Prompt Libraries

Create reusable prompt libraries that can be shared across projects:

**prompts/patterns/rest_api.prompt:**
```
# REST API Best Practices

When building REST endpoints:
- Use GET for reading, POST for creating
- Return 200 for success, 404 for not found
- Include request IDs for debugging
- Validate all inputs
```

Import them in your project:

```yaml
imports:
  - patterns.rest_api
  - patterns.testing
  - constraints.security
```

### Import Resolution

Prompts are resolved from multiple locations in order:

1. **Project-local**: `./prompts/` in your project
2. **Global**: `~/.prompts/` in your home directory
3. **Built-in**: Shipped with promptc

This allows you to:
- Override built-in libraries with project-specific versions
- Share libraries across all your projects via `~/.prompts/`
- Use well-tested built-in libraries

### Exclusions

Exclude specific imports with the `!` prefix:

```yaml
imports:
  - patterns.rest_api
  - !patterns.rest_api.verbose_logging  # Exclude verbose logging
```

### Built-in Libraries

promptc ships with several useful prompt libraries:

**Patterns:**
- `patterns.rest_api` - REST API best practices
- `patterns.testing` - Testing guidelines and best practices

**Constraints:**
- `constraints.security` - Security best practices
- `constraints.code_quality` - Code quality standards

### Supported Targets

- **raw** - Plain text output (default)
- **claude** - Markdown format for Claude (.claude/instructions.md)
- **cursor** - Cursor rules format (.cursorrules)

More targets coming soon!

## Advanced Usage

### Debug Mode

See how imports are resolved:

```bash
promptc compile myapp.prompt --target=claude --debug
```

Output:
```
[DEBUG] Input file: myapp.prompt
[DEBUG] Project dir: /home/user/myproject
[DEBUG] Target: claude
[DEBUG] Resolved imports: ['patterns.rest_api', 'constraints.security']
[DEBUG] Imports content length: 1847 chars
```

### Complex Example

**myapp.prompt:**
```yaml
imports:
  - patterns.rest_api
  - patterns.testing
  - constraints.security
  - constraints.code_quality

context:
  language: typescript
  framework: express
  database: postgresql
  testing: jest
  deployment: docker

features:
  - user_registration: "Users can create accounts with email/password"
  - profile_management: "Users can update their profile information"
  - password_reset: "Users can reset forgotten passwords via email"
  - api_authentication: "API uses JWT tokens for authentication"

constraints:
  - no_hardcoded_secrets
  - comprehensive_test_coverage
  - type_safety_required
  - docker_compose_for_local_dev
```

Compile it:

```bash
promptc compile myapp.prompt --target=claude --output=.claude/instructions.md
```

## Project Structure

When using promptc in your project:

```
your-project/
├── myapp.prompt           # Your project prompt spec
├── prompts/               # Project-local prompt libraries (optional)
│   ├── patterns/
│   └── constraints/
├── .claude/
│   └── instructions.md    # Generated (don't edit manually)
└── .cursorrules          # Generated (don't edit manually)
```

## Creating Prompt Libraries

Create reusable libraries for your organization:

**~/.prompts/company/standards.prompt:**
```
# Company Coding Standards

## Code Review
- All PRs require 2 approvals
- Must pass CI/CD pipeline
- Code coverage must not decrease

## Git Workflow
- Use conventional commits
- Squash merge to main
- Delete branches after merge
```

Use in any project:

```yaml
imports:
  - company.standards
```

## Use Cases

### 1. Consistent Instructions Across Tools

Write once, use everywhere:

```bash
# Generate for all your tools
promptc compile myapp.prompt --target=claude --output=.claude/instructions.md
promptc compile myapp.prompt --target=cursor --output=.cursorrules
```

### 2. Team Knowledge Base

Share prompt libraries across your team:

```bash
# Clone team prompts
git clone git@github.com:yourteam/prompts.git ~/.prompts/team

# Everyone on the team can now use them
imports:
  - team.api_standards
  - team.testing_policy
```

### 3. Project Templates

Create templates for common project types:

```yaml
# web-api-template.prompt
imports:
  - patterns.rest_api
  - patterns.testing
  - constraints.security
  - constraints.code_quality

context:
  type: web_api

features:
  - Replace this with your features
```

## Roadmap

- [ ] More targets (Aider, Copilot, etc.)
- [ ] `promptc init` - Interactive project setup
- [ ] `promptc add <library>` - Add prompt libraries
- [ ] Library versioning and dependencies
- [ ] Published library packages (npm/pip style)
- [ ] VSCode extension for .prompt files
- [ ] Watch mode for auto-recompilation
- [ ] Prompt testing and validation

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

- **Version controlled** - Track changes to your instructions
- **Composable** - Build on reusable libraries
- **Testable** - See exactly what you're sending to the LLM
- **Shareable** - Collaborate on prompt engineering

Just like we don't copy-paste code between projects, we shouldn't copy-paste prompts. Compile them instead.
