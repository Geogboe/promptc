# promptc Roadmap

## v1 (current) — Core Refactor

Focus: Clean foundation, single output format, resource fetching, sandboxed build pipeline.

### Done
- [x] Spec format: `.spec.promptc` YAML files
- [x] Generic `instructions.promptc` output (model-agnostic markdown)
- [x] 5 CLI commands: `validate`, `compile`, `build`, `list`, `init`
- [x] Resource fetching: HTTP URLs → markdown, git repos → directory
- [x] Sandbox providers: Docker (moby SDK), bubblewrap (Linux), none
- [x] Build pipeline: compile + fetch + sandbox + agent
- [x] Signal handling: Ctrl+C → graceful sandbox cleanup
- [x] Resource caching: SHA-256 content hash, `~/.cache/promptc/`
- [x] CI/CD: GitHub Actions (lint, test, build, secret scan)
- [x] Release: goreleaser cross-platform binaries, release-please changelog
- [x] Taskfile: replaces Makefile, matrix cross-compile

### Spec fields (v1)
```yaml
imports:      # library resolution (unchanged from v0)
context:      # key-value project metadata
features:     # list of feature descriptions
constraints:  # list of constraint strings/dicts
resources:    # NEW: fetch web/git resources
build:        # NEW: sandbox + agent config
```

---

## v2 — Library Registry

Focus: Discoverable, installable prompt libraries via GitHub.

### Planned Commands
- `promptc search <term>` — query GitHub for repos with topic `promptc-library`
- `promptc install <user/repo>` — clone to `~/.prompts/<user>/<repo>/`
- `promptc publish` — add `promptc-library` topic to current repo

### Install Resolution Order (v2)
1. Project-local `./prompts/`
2. User-installed `~/.prompts/<user>/<repo>/`
3. Global `~/.prompts/`
4. Built-in (embedded)

### Import Syntax
```yaml
imports:
  - user/repo/patterns.rest_api  # installed library (user/repo prefix)
  - patterns.rest_api             # built-in (no prefix)
```

### Windows Sandbox (v2)
- Evaluate `github.com/microsoft/hcsshim` for native Windows container support
- Target: Docker Desktop not required on Windows for sandbox

### Other v2 Items
- `promptc cache clear` — evict resource cache entries
- Watch mode: `promptc compile --watch` re-compiles on spec change
- Resource cache TTL configuration

---

## v3 — Advanced Features

Focus: Library versioning, remote specs, multi-agent builds.

### Planned
- Library versioning: `user/repo@v1.2.3` import syntax
- Remote spec references: compile a spec from a URL
- Multi-agent builds: parallel agent runs for different spec sections
- `promptc upgrade` — update installed libraries to latest versions
- Spec composition: one spec can import another spec's features/constraints
- Dry-run diff: show what `build` would change without running agent
