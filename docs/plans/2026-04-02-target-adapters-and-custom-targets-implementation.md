# Target Adapters And Custom Targets Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add the first CLI slice for multi-target interoperability: built-in target definitions, global/project target config loading, and `agentlib targets list`.

**Architecture:** Create a dedicated `internal/targets` package that owns built-in target definitions and JSON config loading from `~/.agentlib/targets.json` and `.agentlib/targets.json`. The CLI gets a new `targets list` command that merges built-ins with config-defined custom targets and prints them in a predictable text format.

**Tech Stack:** Go, standard library, existing CLI command dispatch.

### Task 1: Add target model and config loading

**Files:**
- Create: `internal/targets/config.go`
- Create: `internal/targets/config_test.go`

**Step 1: Write the failing test**

Cover:
- loading built-in targets when no config exists
- loading global custom targets from `~/.agentlib/targets.json`
- loading project custom targets from `.agentlib/targets.json`
- project targets augment, not replace, built-ins
- invalid JSON returns a clear error

**Step 2: Run test to verify it fails**

Run: `go test ./internal/targets`

Expected: FAIL because the package does not exist yet.

**Step 3: Write minimal implementation**

Implement:
- built-in target table
- config types
- loaders for global and project files
- merge behavior

**Step 4: Run test to verify it passes**

Run: `go test ./internal/targets`

Expected: PASS.

### Task 2: Expose `agentlib targets list`

**Files:**
- Modify: `internal/cli/app.go`
- Modify: `internal/cli/app_test.go`

**Step 1: Write the failing test**

Cover:
- `agentlib targets list` prints built-ins
- custom targets appear in the output
- disabled targets are still visible but marked disabled

**Step 2: Run test to verify it fails**

Run: `go test ./internal/cli`

Expected: FAIL because the command does not exist yet.

**Step 3: Write minimal implementation**

Add:
- `targets` command group
- `targets list` subcommand
- deterministic output with id, type, format, mode, and enabled state

**Step 4: Run test to verify it passes**

Run: `go test ./internal/cli`

Expected: PASS.

### Task 3: Update docs and verify

**Files:**
- Modify: `README.md`

**Step 1: Update docs**

Document:
- built-in targets
- `~/.agentlib/targets.json`
- `.agentlib/targets.json`
- `agentlib targets list`

**Step 2: Run verification**

Run:
- `go test ./...`
- `go build ./cmd/agentlib`

Expected: PASS.

### Task 4: Record follow-ups

**Files:**
- Create: `../agentlib/docs/plans/2026-04-02-web-compatibility-badges-followup.md`

**Step 1: Record the next registry slices**

Document:
- API compatibility metadata contract
- storage location in `agentlib`
- `Built for` / `Tested with` / `Adapter available` badges in the web UI

**Step 2: Stop after documentation**

Do not change `agentlib` code in this slice. Keep the first implementation isolated to `agent-cli`.
