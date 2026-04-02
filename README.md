# agent-cli

Go command-line client for AgentLib.

The repository is named `agent-cli`, but the distributed end-user binary is named `agentlib`.

## MVP scope

Current commands:

- `agentlib init`
- `agentlib version`
- `agentlib validate ./agent.yaml`
- `agentlib search reviewer`
- `agentlib show raul/code-reviewer@0.4.0`
- `agentlib install raul/code-reviewer@0.4.0`
- `agentlib remove raul/code-reviewer@0.4.0`

## Registry selection

By default, the CLI targets:

```text
https://agentlib.dev
```

For local development you can point it at a local Worker:

```bash
export AGENTLIB_BASE_URL=http://127.0.0.1:8787
```

## Install modes

`install` and `remove` use the user-global store by default:

```text
~/.agentlib/agents/<namespace>/<name>/<version>/
~/.agentlib/agent.lock.json
```

This makes the CLI predictable no matter which directory you run it from.

For project-local installs:

1. Initialize the project once:

```bash
agentlib init
```

That writes the project marker to:

```text
.agentlib/project.json
```

2. Install locally:

```bash
agentlib install --local raul/code-reviewer@0.4.0
```

That writes artifacts under:

```text
<project>/.agentlib/agents/<namespace>/<name>/<version>/
<project>/.agentlib/agent.lock.json
```

You can also override the local store root when using `--local`:

```bash
agentlib install --local --install-dir vendor/agentlib raul/code-reviewer@0.4.0
```

`--install-dir` requires `--local`. Use `--global` or `-g` to be explicit about the default global mode.

## Development

Use the locally installed Go toolchain:

```bash
export PATH=/home/raul/.local/go/bin:$PATH
export GOCACHE=/tmp/agent-cli-go-build
go test ./...
go build ./cmd/agentlib
```

Or use the Docker build environment:

```bash
docker compose run --rm go-build
```

That command runs `go test ./...` and `go build ./cmd/agentlib` inside the container while reusing Docker volumes for the Go module and build caches.

## Release packaging

Local snapshot packaging uses GoReleaser:

```bash
goreleaser release --snapshot --clean
```

The release config builds `agentlib` for linux, darwin, and windows on `amd64` and `arm64`, then emits per-platform archives plus a checksum file.

## Install

The POSIX installer downloads the matching GitHub Releases archive and checksum, verifies the download, and installs `agentlib` into `~/.agentlib/bin`.

Install a pinned release:

```bash
curl -fsSL https://raw.githubusercontent.com/agentlibdev/agent-cli/v0.1.0/scripts/install.sh | sh -s -- v0.1.0
```

The same pinned install works with `wget`:

```bash
wget -qO- https://raw.githubusercontent.com/agentlibdev/agent-cli/v0.1.0/scripts/install.sh | sh -s -- v0.1.0
```

Replace `v0.1.0` in both places with the release you want to install.

Add the install directory to `PATH` in your shell startup file if needed:

```bash
echo 'export PATH="$HOME/.agentlib/bin:$PATH"' >> ~/.bashrc
export PATH="$HOME/.agentlib/bin:$PATH"
```

Use `~/.zshrc` instead of `~/.bashrc` on `zsh`, or `~/.profile` as a more shell-agnostic fallback. Then open a new shell or source the file you updated.

The Windows installer follows the same release flow, but downloads the `.zip` archive and installs `agentlib.exe` into `%USERPROFILE%\\.agentlib\\bin`.

Install a pinned release from PowerShell:

```powershell
Invoke-WebRequest https://raw.githubusercontent.com/agentlibdev/agent-cli/v0.1.0/scripts/install.ps1 -OutFile install.ps1
powershell.exe -NoProfile -ExecutionPolicy Bypass -File .\install.ps1 -Version v0.1.0
```

Replace `v0.1.0` in both places with the release you want to install.

If `%USERPROFILE%\\.agentlib\\bin` is not already on your user PATH, add it through the Windows Environment Variables UI, then open a new PowerShell session.
