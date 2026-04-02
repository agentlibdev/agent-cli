# agent-cli

Go command-line client for AgentLib.

The repository is named `agent-cli`, but the distributed end-user binary is named `agentlib`.

## MVP scope

Current commands:

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

## Local install layout

`install` writes artifacts into:

```text
.agentlib/agents/<namespace>/<name>/<version>/
```

and writes a lockfile to:

```text
.agentlib/agent.lock.json
```

`remove` deletes one exact installed version from that layout and removes the lockfile if it points at the same exact version.

## Development

Use the locally installed Go toolchain:

```bash
export PATH=/home/raul/.local/go/bin:$PATH
export GOCACHE=/tmp/agent-cli-go-build
go test ./...
go build ./cmd/agentlib
```

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
