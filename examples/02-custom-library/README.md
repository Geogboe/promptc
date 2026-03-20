# Custom Library Example

This example shows how project-local prompts under `./prompts/` are resolved alongside built-in libraries.

## Run It

```powershell
promptc validate .\team-service.spec.promptc
promptc compile .\team-service.spec.promptc
Get-Content .\team-service\instructions.promptc
```

## What It Demonstrates

- Project-local library resolution from `./prompts`
- Team standards layered with built-in prompt content
- A committed output snapshot in `team-service/instructions.promptc`
