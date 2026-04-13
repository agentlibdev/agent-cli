# agent-cli

Go command-line client for AgentLib.

The repository is named `agent-cli`, but the distributed end-user binary is named `agentlib`.

## MVP scope

Current commands:

- `agentlib init`
- `agentlib targets list`
- `agentlib targets detect`
- `agentlib enable --target codex-local raul/code-reviewer@0.4.0`
- `agentlib version`
- `agentlib validate ./agent.yaml`
- `agentlib search reviewer`
- `agentlib show raul/code-reviewer@0.4.0`
- `agentlib install raul/code-reviewer@0.4.0`
- `agentlib deactivate --target codex raul/code-reviewer@0.4.0`
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

After a successful install, the CLI can also activate the package into detected runtimes.

Default behavior:

- if the process is interactive and compatible runtimes are detected, `install` shows a multiselect prompt
- if the process is non-interactive, `install` only installs unless you pass activation flags explicitly

Examples:

```bash
agentlib install raul/code-reviewer@0.4.0
agentlib install --runtime codex raul/code-reviewer@0.4.0
agentlib install --runtime codex --runtime claude raul/code-reviewer@0.4.0
agentlib install --all-detected raul/code-reviewer@0.4.0
agentlib install --no-activate raul/code-reviewer@0.4.0
```

Activation during install currently reuses the same target materialization path as `enable`.

Activation state is persisted in the selected AgentLib store root:

- global: `~/.agentlib/config.json`
- local: `<project>/.agentlib/config.json`

Each activation record tracks the runtime target id, the package ref, the materialized path, and activation time.

## Target adapters

The CLI has built-in target definitions for:

- `antigravity`
- `claude-code`
- `crewai`
- `cursor`
- `codex`
- `gemini-cli`
- `github-copilot`
- `langchain`
- `openclaw`
- `opencode`
- `vscode`
- `windsurf`

List the currently known targets with:

```bash
agentlib targets list
```

Detect which of those targets appear to be present on the current machine with:

```bash
agentlib targets detect
```

The first detection pass is intentionally simple:

- built-in targets are detected from known commands in `PATH`
- if no command is found, built-ins can still be detected from an existing default skill directory
- custom targets are detected from an existing `installRoot` or `manifestPath`

Built-in targets resolve their default skill directory from `HOME`. For example:

- `codex` -> `~/.agents/skills`
- `claude-code` -> `~/.claude/skills`
- `cursor` -> `~/.cursor/skills`
- `gemini-cli` -> `~/.gemini/skills`
- `antigravity` -> `~/.gemini/antigravity/skills`
- `crewai` -> `~/.crewai/agents`
- `langchain` -> `~/.langchain/agents`
- `openclaw` -> `~/.openclaw/agents`
- `opencode` -> `~/.config/opencode/skills`
- `vscode` -> `~/.vscode/agentlib/skills`

Enable an already installed package into a configured target with:

```bash
agentlib enable --target openclaw-local raul/code-reviewer@0.4.0
```

Disable an activated package from a target with:

```bash
agentlib deactivate --target codex raul/code-reviewer@0.4.0
```

Codex is the first built-in target that works without any custom `targets.json` entry:

```bash
agentlib enable --target codex raul/code-reviewer@0.4.0
```

OpenClaw is the first built-in target that uses generated package export without custom config:

```bash
agentlib enable --target openclaw raul/code-reviewer@0.4.0
```

Some built-ins also accept short aliases:

- `claude` -> `claude-code`
- `gemini` -> `gemini-cli`
- `copilot` -> `github-copilot`

These built-ins are also ready to use without custom config:

```bash
agentlib enable --target gemini raul/code-reviewer@0.4.0
agentlib enable --target opencode raul/code-reviewer@0.4.0
agentlib enable --target claude raul/code-reviewer@0.4.0
agentlib enable --target cursor raul/code-reviewer@0.4.0
agentlib enable --target antigravity raul/code-reviewer@0.4.0
agentlib enable --target vscode raul/code-reviewer@0.4.0
```

`enable` reads the package from the AgentLib store:

- global by default: `~/.agentlib/agents/...`
- local with `--local`

and materializes it into the target `installRoot` using the target mode:

- `symlink`
- `copy`
- `generate` for package-export targets such as `openclaw`

Generated package-export targets currently include:

- `openclaw`: exports into `~/.openclaw/agents/<namespace>/<name>/<version>/` and writes `agentlib-export.json`
- `crewai`: exports into `~/.crewai/agents/<namespace>/<name>/<version>/` and writes `agentlib-export.json` plus `crewai-agent.py`
- `langchain`: exports into `~/.langchain/agents/<namespace>/<name>/<version>/` and writes `agentlib-export.json` plus `langchain-agent.py`

The CLI can also load custom target definitions from:

- global: `~/.agentlib/targets.json`
- project: `.agentlib/targets.json`

Minimal example:

```json
{
  "targets": [
    {
      "id": "openclaw-local",
      "type": "custom",
      "format": "markdown-skill-dir",
      "installRoot": "/home/raul/.openclaw/skills",
      "manifestPath": "/home/raul/.openclaw/config/skills.json",
      "mode": "symlink",
      "enabled": true
    }
  ]
}
```

Project targets augment the built-in and global target list; they do not replace it.

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

To build distributable test binaries into `dist/`:

```bash
docker compose run --rm go-dist
```

That command:

- runs `go test ./...`
- writes `dist/agentlib-linux-amd64`
- writes `dist/agentlib-linux-arm64`
- writes `dist/agentlib-windows-amd64.exe`
- writes `dist/agentlib-windows-arm64.exe`

For WSL, use the Linux binary. WSL does not need a separate build target.

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
