# ADR-003: Docker as Primary Sandbox

## Status
Accepted

## Context

The `promptc build` command runs an AI agent that writes arbitrary files to the workspace directory. Without sandbox isolation, this means an agent could write outside the workspace, execute system commands, or exfiltrate data. We need a sandbox strategy that:

1. Isolates the agent's filesystem access to the workspace directory
2. Works on the platforms where promptc will be used (Linux, macOS, Windows)
3. Has a reliable Go SDK for process management and streaming

### Options Considered

**Windows Job Objects (`golang.org/x/sys/windows/mkwinsyscall`)**
- Provides CPU/memory resource limits only
- Does NOT provide filesystem isolation
- Not suitable for our use case

**`github.com/microsoft/hcsshim`** (Windows Host Compute System)
- Microsoft's official Go library for Windows containers (HCS API)
- Used by containerd and Docker for Windows
- Reputable and actively maintained
- Provides real filesystem isolation on Windows
- **Problem**: Requires Windows Server 2019+ or Windows 10 1809+ with containers feature enabled; adds significant complexity and binary size; requires the Windows Container feature to be enabled separately from Docker Desktop

**bubblewrap (`bwrap`)**
- Linux kernel namespaces-based sandbox
- Lightweight, no daemon required
- Ships with Flatpak; available on most Linux distros
- **Problem**: Linux-only, requires kernel support, not available on macOS or Windows

**Docker (via `github.com/docker/docker/client`)**
- Available on all platforms via Docker Desktop (macOS, Windows) or package manager (Linux)
- Docker Go SDK provides proper streaming, interrupt handling, and cleanup
- Mounting `-v <dir>:/workspace` gives clean filesystem isolation
- Well-understood by developers; Docker Desktop is widely installed

## Decision

Use Docker as the primary sandbox provider, implemented via the Docker Go SDK (`github.com/docker/docker/client`).

Additionally implement:
- **bubblewrap** for Linux-only environments where Docker is not available (build tag `linux`)
- **none** for development/testing where isolation is not needed (with explicit warning)

The `build.sandbox.type` spec field selects the provider: `docker`, `bubblewrap`, or `none`.

### Why Go SDK over Docker CLI

Using `github.com/docker/docker/client` instead of shelling out to `docker`:
- Proper Go error types (not string parsing)
- Streaming `ContainerLogs` with `follow=true` for real-time output
- Clean `ContainerStop` + `ContainerRemove` on context cancellation (Ctrl+C)
- No dependency on `docker` binary being on PATH (connects directly to Docker socket)

## Consequences

**Easier:**
- Cross-platform sandbox on Linux, macOS, Windows via Docker Desktop
- Users already familiar with Docker don't need new tooling
- Real filesystem isolation: agent cannot write outside `/workspace`

**Harder:**
- Docker Desktop must be installed and running (user setup step)
- Docker SDK adds binary size (~several MB)
- Container startup latency (~1-2s) per build

**Windows Docker Engine (without Docker Desktop):**
Docker Desktop is NOT required. Docker Engine can be installed on Windows via:
- `winget install Docker.DockerCLI` + `winget install Docker.DockerEngine`
- `choco install docker-engine`

This provides the Docker daemon and CLI without the Docker Desktop UI overhead.

**Windows native sandbox (deferred to v2):**
`hcsshim` is the correct solution for native Windows containers without any Docker requirement. Deferred because it requires Windows Container feature setup and adds significant complexity.

**Trade-off accepted:** Docker Engine (not Desktop) is the recommended sandbox on all platforms. The `none` sandbox is available for users who don't want Docker at all.
