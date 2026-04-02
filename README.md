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
