# Activation State And Deactivate Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Persist runtime activation state inside the AgentLib store and add a `deactivate` command that removes a package from a target runtime and clears the persisted activation entry.

**Architecture:** Keep the current target materialization model and add a small persisted activation index under the install store root. Reuse the existing `enable`/install activation flow to write activation records after successful target materialization. Add a `Disable` primitive in `internal/targets` that removes the target path, then wire a new `deactivate` CLI command that resolves the target, removes the materialized package, and updates the activation state.

**Tech Stack:** Go, standard library JSON/file I/O, existing `internal/cli`, `internal/install`, and `internal/targets` packages, Go test runner

### Task 1: Add failing tests for activation persistence and deactivate

**Files:**
- Modify: `internal/cli/app_test.go`
- Create: `internal/targets/activation_state_test.go`

**Step 1: Write the failing tests**

Cover:

- `enable` persists an activation record under the install store config file
- install-time activation persists the same record through the shared path
- `deactivate --target <id> <ref>` removes the runtime materialization and clears the stored activation

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/cli ./internal/targets
```

Expected: fail because no persisted activation state or deactivate command exists yet.

### Task 2: Add activation state storage and target disable primitive

**Files:**
- Create: `internal/targets/activation_state.go`
- Modify: `internal/targets/enable.go`
- Test: `internal/targets/activation_state_test.go`

**Step 1: Implement minimal state handling**

Add:

- `config.json` path under the selected store root
- activation records keyed by `targetId` + package ref
- helpers to upsert and remove activation records
- `Disable` to remove the target path safely

### Task 3: Wire CLI persistence and deactivate

**Files:**
- Modify: `internal/cli/app.go`
- Modify: `internal/cli/install_activation.go`
- Modify: `internal/cli/app_test.go`

**Step 1: Implement command wiring**

Add:

- `deactivate [--local|--global|-g] [--install-dir <dir>] --target <id> <namespace/name@version>`
- persistence after successful `enable`
- persistence after successful post-install activation
- cleanup of activation state after successful `deactivate`

### Task 4: Document and verify

**Files:**
- Modify: `README.md`

**Step 1: Update CLI usage**

Document:

- persisted activation state in `config.json`
- `deactivate`

**Step 2: Run verification**

Run:

```bash
go test ./...
go build ./cmd/agentlib
```

Expected: tests pass and the CLI builds cleanly.
