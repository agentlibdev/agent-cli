# Codex First Built-In Target Adapter Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Register the minimum built-in TUI client catalog and make `Codex` the first built-in target that can be enabled without any custom `targets.json` entry.

**Architecture:** Extend the built-in target table with user-home-relative target paths for the minimum TUI clients, but only make `Codex` materially enable-able in this slice. Built-in targets gain a `relativePath` concept that resolves against `HOME`, allowing `agentlib enable --target codex <ref>` to materialize an installed package into `~/.agents/skills/...` without custom configuration.

**Tech Stack:** Go, standard library, existing `internal/targets`, `internal/cli`, and install target resolution.

### Task 1: Add built-in target metadata for the minimum client catalog

**Files:**
- Modify: `internal/targets/config.go`
- Modify: `internal/targets/config_test.go`
- Create: `docs/plans/2026-04-02-codex-first-built-in-targets.md`

**Step 1: Write the failing test**

Cover:
- built-in targets now include the minimum catalog:
  - `antigravity`
  - `claude-code`
  - `cursor`
  - `codex`
  - `gemini-cli`
  - `github-copilot`
  - `opencode`
  - `windsurf`
- built-in `codex` resolves `installRoot` to `$HOME/.agents/skills`
- built-in targets keep deterministic ordering

**Step 2: Run test to verify it fails**

Run: `go test ./internal/targets`

Expected: FAIL until built-in metadata is extended.

**Step 3: Write minimal implementation**

Add a built-in target table with:
- stable `id`
- human-readable `name`
- `relativePath`
- resolved `installRoot` based on `HOME`

**Step 4: Run test to verify it passes**

Run: `go test ./internal/targets`

Expected: PASS.

### Task 2: Make built-in Codex enable-able without custom config

**Files:**
- Modify: `internal/targets/enable.go`
- Modify: `internal/targets/enable_test.go`
- Modify: `internal/cli/app.go`
- Modify: `internal/cli/app_test.go`

**Step 1: Write the failing test**

Cover:
- `agentlib enable --target codex <ref>` works with no custom target config
- the package is materialized into `$HOME/.agents/skills/<namespace>/<name>/<version>`
- built-in Codex defaults to `symlink`

**Step 2: Run test to verify it fails**

Run: `go test ./internal/targets ./internal/cli`

Expected: FAIL until built-in Codex has a resolved install root.

**Step 3: Write minimal implementation**

Use built-in metadata from the target registry so `enable` can materialize `codex` directly.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/targets ./internal/cli`

Expected: PASS.

### Task 3: Update docs and verify repo

**Files:**
- Modify: `README.md`

**Step 1: Update docs**

Document:
- minimum built-in client catalog
- Codex as the first built-in ready-to-enable target
- home-relative built-in skill directories

**Step 2: Run verification**

Run:
- `go test ./...`
- `go build ./cmd/agentlib`

Expected: PASS.
