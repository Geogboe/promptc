# ADR-006: GitHub as v2 Library Registry

## Status
Proposed (v2, not yet implemented)

## Context

promptc ships with 8 built-in libraries (4 patterns, 4 constraints). As the ecosystem grows, a registry for community-contributed libraries is needed. Options:

1. **Dedicated registry server**: Custom API, requires infrastructure, maintenance, and trust model
2. **GitHub topics**: Use GitHub's existing search API to find repos tagged with a topic (e.g., `promptc-library`)
3. **Central curated list**: A single repo (`promptc-community/libraries`) with a directory, like awesome-lists

## Decision

Use **GitHub topics** as the registry mechanism (v2).

- `promptc search <term>` → queries GitHub API: `topic:promptc-library <term>`
- `promptc install <user/repo>` → clones to `~/.prompts/<user>/<repo>/` or `./prompts/<user>/<repo>/`
- `promptc publish` → adds `promptc-library` topic to the current GitHub repo via API

No registry server required. Discovery is GitHub search. Installation is git clone. Publishing is a topic tag.

## Import Resolution Order (v2)

```
1. Project-local: ./prompts/
2. User-installed: ~/.prompts/<user>/<repo>/
3. Global: ~/.prompts/
4. Built-in (embedded)
```

Import syntax:
```yaml
imports:
  - user/repo/patterns.rest_api   # installed: user/repo prefix
  - patterns.rest_api              # built-in: no prefix
```

## Consequences

**Easier:**
- Zero infrastructure to maintain — GitHub is the registry
- Library publishing is a one-command GitHub topic tag
- Search is GitHub search — full text search, star counts, etc.
- Installation is git clone — versioning via tags/branches/commits

**Harder:**
- Search quality depends on GitHub topic adoption
- No version pinning in v2 (addressed in v3 with `user/repo@v1.2.3` syntax)
- GitHub API rate limits may affect search for unauthenticated users (60 req/hour)
- No quality control: any repo with the topic is discoverable

**Trade-off accepted:** GitHub as registry is zero-infrastructure and familiar to developers. The lack of version pinning in v2 is acceptable for early adopters; v3 adds semver import syntax.
