# promptc - Project Status

## Current State

promptc is a Go-based CLI for compiling `.spec.promptc` files into generic `instructions.promptc` output, with optional resource fetching and sandboxed agent execution via `promptc build`.

## What Works

- `promptc validate` checks spec files
- `promptc compile` produces `instructions.promptc`
- `promptc build` fetches resources and runs an external agent in a sandbox
- `promptc list` discovers built-in and local libraries
- `promptc init` creates new specs from templates
- Built-in prompt libraries are embedded in the binary

## Documentation and Examples

- `README.md` now reflects the current single-output workflow
- `examples/` is the canonical place to learn the tool through real workflows
- `docs/architecture.md` describes the compile/build pipeline and output layout

## Notes

- Release and installer behavior is handled through GitHub releases and the install scripts
- This file is intentionally short so it does not drift into a stale benchmark report
