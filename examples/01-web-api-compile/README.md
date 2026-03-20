# Web API Compile Example

This example shows a compile-only workflow for a Go web API.

## Run It

```powershell
promptc validate .\coffee-shop-api.spec.promptc
promptc compile .\coffee-shop-api.spec.promptc
Get-Content .\coffee-shop-api\instructions.promptc
```

## What It Demonstrates

- Built-in pattern and constraint imports
- Project context and feature descriptions
- Stable compiled output committed in `coffee-shop-api/instructions.promptc`
