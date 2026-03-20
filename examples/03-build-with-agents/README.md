# Build With Agents Example

This example shows how `promptc build` compiles a spec, fetches external resources, and then hands the workspace to an external agent CLI.

## Files

- `claude.spec.promptc`: build config for `claude`
- `codex.spec.promptc`: build config for `codex`
- `claude/instructions.promptc` and `codex/instructions.promptc`: committed compile snapshots

## Run It

Compile only:

```powershell
promptc compile .\claude.spec.promptc
promptc compile .\codex.spec.promptc
```

Dry run the build pipeline without invoking an agent:

```powershell
promptc build .\codex.spec.promptc --dry-run
```

Run the full build with a real agent CLI installed:

```powershell
promptc build .\codex.spec.promptc
promptc build .\claude.spec.promptc
```

## Notes

- These examples use `sandbox.type: none` so the host-installed `claude` and `codex` CLIs can run directly.
- The fetched `resources/` directory and any agent-generated files are intentionally not committed.
- Non-interactive agent behavior can vary by local CLI version and auth state. The specs are real starting points, but you may need to adjust flags if your installed CLI has different defaults.
