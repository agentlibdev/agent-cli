# Agent CLI MVP Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Bootstrap the `agent-cli` Go repository with a working MVP that validates local manifests, shows published agent versions, and installs versioned agent packages into a local layout with a lockfile.

**Architecture:** Keep the CLI small and framework-free. Use a thin command dispatcher in `cmd/agentlib`, internal packages for manifest validation, agent reference parsing, registry HTTP access, and local installation, and a simple lockfile stored under `.agentlib`. For the first slice, install exact versions only and avoid dependency resolution.

**Tech Stack:** Go 1.26, standard library HTTP/file APIs, `gopkg.in/yaml.v3`.

### Task 1: Bootstrap the Go module and command entrypoint

**Files:**
- Create: `go.mod`
- Create: `cmd/agentlib/main.go`
- Create: `internal/cli/app.go`
- Create: `README.md`

**Step 1: Add module metadata and a minimal command dispatcher**

Support:
- `validate <path>`
- `show <namespace/name@version>`
- `install <namespace/name@version>`

**Step 2: Add usage/help text**

Keep output concise and deterministic.

### Task 2: Build manifest validation and ref parsing with tests first

**Files:**
- Create: `internal/agentref/ref.go`
- Create: `internal/agentref/ref_test.go`
- Create: `internal/manifest/validate.go`
- Create: `internal/manifest/validate_test.go`

**Step 1: Write failing tests**

Cover:
- exact agent ref parsing
- malformed refs rejected
- local manifest shape validation for required fields

**Step 2: Implement minimal code to pass**

Use strict-enough validation for the MVP:
- `apiVersion`
- `kind`
- `metadata.namespace`
- `metadata.name`
- `metadata.version`
- `metadata.title`
- `metadata.description`

### Task 3: Build registry client and `show`

**Files:**
- Create: `internal/registry/client.go`
- Create: `internal/registry/client_test.go`

**Step 1: Write failing tests**

Cover:
- `show` fetching version metadata
- `show` fetching artifacts list
- upstream errors surfaced cleanly

**Step 2: Implement minimal HTTP client**

Use existing registry routes:
- `GET /api/v1/agents/:namespace/:name/versions/:version`
- `GET /api/v1/agents/:namespace/:name/versions/:version/artifacts`

### Task 4: Build local install and lockfile

**Files:**
- Create: `internal/install/install.go`
- Create: `internal/install/install_test.go`

**Step 1: Write failing tests**

Cover:
- install path layout under `.agentlib/agents/...`
- artifact bytes written locally
- lockfile written to `.agentlib/agent.lock.json`

**Step 2: Implement minimal installer**

Install exact version only.

### Task 5: Wire commands, docs, and verification

**Files:**
- Modify: `cmd/agentlib/main.go`
- Modify: `README.md`

**Step 1: Wire subcommands to internal packages**

**Step 2: Document local usage**

**Step 3: Verify**

Run:
- `PATH=/home/raul/.local/go/bin:$PATH go test ./...`
- `PATH=/home/raul/.local/go/bin:$PATH go build ./cmd/agentlib`
