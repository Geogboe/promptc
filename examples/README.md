# Examples

These examples show the current `promptc` workflow:

1. Author a `*.spec.promptc` file.
2. Validate it with `promptc validate`.
3. Compile it to a generic `instructions.promptc` file with `promptc compile`.
4. Optionally run `promptc build` to fetch resources and invoke an external agent CLI.

## Example Set

### [01-web-api-compile](./01-web-api-compile)

Compile a realistic API spec into `instructions.promptc`.

### [02-custom-library](./02-custom-library)

Compose built-in libraries with a project-local `prompts/` library.

### [03-build-with-agents](./03-build-with-agents)

Use `promptc build` with real `claude` and `codex` specs plus fetched resources.

## Notes

- The output is always generic `instructions.promptc`, not agent-specific formatter output.
- Build examples keep agent configuration in the spec, but the agent CLIs themselves remain external tools managed outside `promptc`.
- The committed `instructions.promptc` files are example snapshots and regression fixtures.
