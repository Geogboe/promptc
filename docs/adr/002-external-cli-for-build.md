# ADR-002: External CLI for Build Agent

## Status
Accepted

## Context

The `promptc build` command needs to run an AI agent (e.g., `claude`, `aider`, `cursor`) against the compiled spec and fetched resources. There are two broad approaches:

1. **Embed the AI API directly**: Call the Anthropic/OpenAI/etc. API from within promptc, managing the conversation loop, tool use, and file writes internally.
2. **Shell out to an existing agent CLI**: Treat the agent as a black-box process that runs in the workspace directory, reads `instructions.promptc`, and writes files.

## Decision

Shell out to an external agent CLI on PATH.

The spec's `build.agent` section specifies:
```yaml
build:
  agent:
    command: claude          # binary on PATH
    args: ["--dangerously-skip-permissions"]
```

promptc runs this command inside the sandbox with the output directory as the working directory.

## Consequences

**Easier:**
- No API key management in promptc — the agent CLI handles its own credentials
- Users can use any agent that has a CLI (claude, aider, cursor, etc.) without promptc changes
- Agent CLI updates are independent of promptc releases
- Much simpler code: `sandbox.Run(ctx, spec.Build.Agent.Command, spec.Build.Agent.Args)`

**Harder:**
- Requires agent CLI to be installed and on PATH (user setup step)
- Less control over agent behavior; can't inject custom system prompts dynamically
- Harder to test without a real agent CLI installed

**Trade-off accepted:** promptc's job is orchestration (compile → fetch → sandbox → run). Embedding an AI API would make promptc responsible for prompt engineering, token management, and multi-turn conversation — a separate product. The external CLI approach keeps the abstraction clean and lets each tool do what it does best.
