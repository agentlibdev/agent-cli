# Status And Activations List Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `agentlib status <ref>` and `agentlib activations list` so users can inspect installed packages and runtime activations without reading `config.json` manually.

**Architecture:** Reuse the existing install store layout and persisted activation index. Add small read helpers for installed package paths and activation lookups, then wire two CLI commands that resolve the store root the same way as `install`/`activate`/`deactivate`.

**Tech Stack:** Go, standard library file I/O, existing `internal/install`, `internal/targets`, and `internal/cli` packages, Go test runner

### Task 1: Add failing CLI tests

**Files:**
- Modify: `internal/cli/app_test.go`

**Step 1: Write the failing tests**

Cover:

- `status <ref>` for an installed package with one or more persisted activations
- `status <ref>` for a missing package
- `activations list` showing stored target/ref/path rows

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/cli
```

Expected: fail because the commands do not exist yet.

### Task 2: Add read helpers

**Files:**
- Modify: `internal/install/install.go`
- Modify: `internal/targets/activation_state.go`

**Step 1: Implement minimal helpers**

Add:

- installed package path resolver/check
- activation lookup by ref

### Task 3: Wire CLI commands

**Files:**
- Modify: `internal/cli/app.go`

**Step 1: Implement command handlers**

Add:

- `status [--local|--global|-g] [--install-dir <dir>] <namespace/name@version>`
- `activations list [--local|--global|-g] [--install-dir <dir>]`

### Task 4: Document and verify

**Files:**
- Modify: `README.md`

**Step 1: Update usage**

Document:

- `status`
- `activations list`

**Step 2: Run verification**

Run:

```bash
go test ./...
go build ./cmd/agentlib
```

Expected: tests pass and the CLI builds cleanly.
