# agent-cli

Go command-line client for AgentLib.

## MVP scope

Current commands:

- `agentlib validate ./agent.yaml`
- `agentlib show raul/code-reviewer@0.4.0`
- `agentlib install raul/code-reviewer@0.4.0`

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

## Development

Use the locally installed Go toolchain:

```bash
export PATH=/home/raul/.local/go/bin:$PATH
export GOCACHE=/tmp/agent-cli-go-build
go test ./...
go build ./cmd/agentlib
```
