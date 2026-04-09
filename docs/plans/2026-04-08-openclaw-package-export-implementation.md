# OpenClaw Package Export Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `openclaw` as the first built-in target that uses generated package export instead of `skill-dir` symlink/copy.

**Architecture:** Extend the target materialization layer so `Enable` supports a second family: `format: package-export` with `mode: generate`. The first concrete built-in using that path is `openclaw`, exporting the canonical package into `~/.openclaw/agents/<namespace>/<name>/<version>/` plus a small `agentlib-export.json` metadata file.

**Tech Stack:** Go, standard library, existing `internal/targets`, existing install store layout.

### Task 1: Add `openclaw` built-in metadata

**Files:**
- Modify: `internal/targets/config.go`
- Modify: `internal/targets/config_test.go`

**Step 1: Write the failing test**

Cover:
- built-in `openclaw` exists
- built-in `openclaw` resolves to `~/.openclaw/agents`
- built-in `openclaw` uses `format: package-export`
- built-in `openclaw` uses `mode: generate`

**Step 2: Run test to verify it fails**

Run: `go test ./internal/targets`

Expected: FAIL until the built-in is added.

**Step 3: Write minimal implementation**

Add `openclaw` to the built-in table with explicit export metadata.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/targets`

Expected: PASS.

### Task 2: Add generated export path to `Enable`

**Files:**
- Modify: `internal/targets/enable.go`
- Modify: `internal/targets/enable_test.go`

**Step 1: Write the failing test**

Cover:
- `Enable` with `openclaw` copies package contents into the export path
- `Enable` writes `agentlib-export.json`
- the metadata file contains source ref and target id

**Step 2: Run test to verify it fails**

Run: `go test ./internal/targets`

Expected: FAIL until `generate` is implemented.

**Step 3: Write minimal implementation**

Implement:
- `mode: generate` for `format: package-export`
- recursive copy of canonical package into target root
- emission of `agentlib-export.json`

**Step 4: Run test to verify it passes**

Run: `go test ./internal/targets`

Expected: PASS.

### Task 3: Expose `openclaw` through CLI UX

**Files:**
- Modify: `internal/cli/app_test.go`
- Modify: `README.md`

**Step 1: Write the failing test**

Cover:
- `agentlib enable --target openclaw <ref>` succeeds without custom config
- exported files land under `~/.openclaw/agents/...`

**Step 2: Run test to verify it fails**

Run: `go test ./internal/cli`

Expected: FAIL until built-in config and generate mode are wired through.

**Step 3: Write minimal implementation**

Reuse existing CLI path; only update docs if the command already works after the lower layers are added.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/cli`

Expected: PASS.

### Task 4: Verify repo

**Files:**
- Modify: `README.md`

**Step 1: Update docs**

Document:
- `openclaw` as the first generated package-export target
- export root
- metadata file

**Step 2: Run verification**

Run:
- `go test ./...`
- `go build ./cmd/agentlib`

Expected: PASS.
