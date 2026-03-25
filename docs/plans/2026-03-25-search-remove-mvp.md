# Agent CLI Search Remove MVP Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extend the `agent-cli` MVP with `search` and `remove`, then verify and publish the CLI as the first usable local client for the AgentLib registry.

**Architecture:** Keep the CLI small and framework-free. Reuse the existing registry client for `search` by fetching the current paginated list endpoint and filtering results locally in the CLI. Add local uninstall behavior to the install package so filesystem cleanup and lockfile updates stay in one place. Keep exact-version semantics for `remove` and avoid introducing semver range logic or backend API changes.

**Tech Stack:** Go 1.26, standard library HTTP/file APIs, existing internal packages, TDD with `go test`.

### Task 1: Add failing tests for registry search and local remove

**Files:**
- Modify: `internal/registry/client_test.go`
- Modify: `internal/install/install_test.go`

**Step 1: Write the failing search test**

Cover:
- `FetchAgents` reads `GET /api/v1/agents`
- agent list shape decoded correctly

**Step 2: Run the targeted registry test and verify it fails**

Run:

```bash
export PATH=/home/raul/.local/go/bin:$PATH
export GOCACHE=/tmp/agent-cli-go-build
go test ./internal/registry -run FetchAgents
```

**Step 3: Write the failing remove test**

Cover:
- installed version directory removed
- unrelated versions preserved
- lockfile removed when it points at the deleted version

**Step 4: Run the targeted install test and verify it fails**

Run:

```bash
export PATH=/home/raul/.local/go/bin:$PATH
export GOCACHE=/tmp/agent-cli-go-build
go test ./internal/install -run Remove
```

### Task 2: Implement minimal registry search and local remove

**Files:**
- Modify: `internal/registry/client.go`
- Modify: `internal/install/install.go`

**Step 1: Add minimal registry list support**

Implement:
- `AgentSummary` type
- `FetchAgents(ctx)` against `GET /api/v1/agents`

**Step 2: Add minimal local remove support**

Implement:
- `Remove(workingDir, ref)`
- delete `.agentlib/agents/<namespace>/<name>/<version>`
- update or delete `.agentlib/agent.lock.json` if it points at that exact version

**Step 3: Re-run targeted tests and make them pass**

### Task 3: Wire CLI commands and cover command behavior

**Files:**
- Modify: `internal/cli/app.go`
- Create: `internal/cli/app_test.go`

**Step 1: Write failing CLI tests**

Cover:
- `search <query>` usage and filtered output
- `remove <namespace/name@version>` usage and success output

**Step 2: Run targeted CLI tests and verify they fail**

Run:

```bash
export PATH=/home/raul/.local/go/bin:$PATH
export GOCACHE=/tmp/agent-cli-go-build
go test ./internal/cli -run 'Search|Remove'
```

**Step 3: Implement minimal command wiring**

Keep:
- `search` requiring one query argument
- `remove` requiring one exact version ref
- human-readable output only

**Step 4: Re-run CLI tests and make them pass**

### Task 4: Update docs, verify end to end, and publish

**Files:**
- Modify: `README.md`

**Step 1: Document the two new commands**

Add:
- `agentlib search review`
- `agentlib remove raul/code-reviewer@0.4.0`

**Step 2: Run full verification**

Run:

```bash
export PATH=/home/raul/.local/go/bin:$PATH
export GOCACHE=/tmp/agent-cli-go-build
go test ./...
go build ./cmd/agentlib
```

**Step 3: Smoke the CLI against local AgentLib**

Run:

```bash
export PATH=/home/raul/.local/go/bin:$PATH
export GOCACHE=/tmp/agent-cli-go-build
AGENTLIB_BASE_URL=http://127.0.0.1:8787 go run ./cmd/agentlib search reviewer
AGENTLIB_BASE_URL=http://127.0.0.1:8787 go run ./cmd/agentlib install raul/code-reviewer@0.3.1
AGENTLIB_BASE_URL=http://127.0.0.1:8787 go run ./cmd/agentlib remove raul/code-reviewer@0.3.1
```

**Step 4: Commit and publish**

Commit the feature branch, merge to `main`, remove the generated binary if present, and push `main` to `origin`.
