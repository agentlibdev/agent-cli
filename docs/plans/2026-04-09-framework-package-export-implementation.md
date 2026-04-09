# Framework Package Export Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Extend the `package-export/generate` target family so `crewai` and then `langchain` become built-in framework targets with generated starter files.

**Architecture:** Reuse the `package-export` family introduced for `openclaw`. Add built-in targets for `crewai` and `langchain`, then extend the generation step so each target can emit a thin framework-specific starter file alongside the exported canonical package and `agentlib-export.json`.

**Tech Stack:** Go, standard library, existing `internal/targets`, existing CLI `enable` flow.

### Task 1: Add `crewai` built-in target metadata

**Files:**
- Modify: `internal/targets/config.go`
- Modify: `internal/targets/config_test.go`

**Step 1: Write the failing test**

Cover:
- built-in `crewai` exists
- built-in `crewai` resolves to `~/.crewai/agents`
- built-in `crewai` uses `format: package-export`
- built-in `crewai` uses `mode: generate`

**Step 2: Run test to verify it fails**

Run: `go test ./internal/targets`

Expected: FAIL until the built-in exists.

**Step 3: Write minimal implementation**

Add the built-in metadata only.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/targets`

Expected: PASS.

### Task 2: Generate `crewai` starter file

**Files:**
- Modify: `internal/targets/enable.go`
- Modify: `internal/targets/enable_test.go`

**Step 1: Write the failing test**

Cover:
- `Enable` with `crewai` writes exported canonical package
- `Enable` writes `agentlib-export.json`
- `Enable` writes a thin `crewai-agent.py` starter file

**Step 2: Run test to verify it fails**

Run: `go test ./internal/targets`

Expected: FAIL until the target-specific generator exists.

**Step 3: Write minimal implementation**

Generate:
- exported package copy
- `agentlib-export.json`
- `crewai-agent.py` with source ref and export directory comment/header

**Step 4: Run test to verify it passes**

Run: `go test ./internal/targets`

Expected: PASS.

### Task 3: Expose `crewai` through CLI UX

**Files:**
- Modify: `internal/cli/app_test.go`
- Modify: `README.md`

**Step 1: Write the failing test**

Cover:
- `agentlib enable --target crewai <ref>` succeeds without custom config
- generated files land under `~/.crewai/agents/...`

**Step 2: Run test to verify it fails**

Run: `go test ./internal/cli`

Expected: FAIL until the built-in and generator are wired through.

**Step 3: Write minimal implementation**

Reuse the existing CLI path; no new top-level command should be required.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/cli`

Expected: PASS.

### Task 4: Repeat the same pattern for `langchain`

**Files:**
- Modify: `internal/targets/config.go`
- Modify: `internal/targets/config_test.go`
- Modify: `internal/targets/enable.go`
- Modify: `internal/targets/enable_test.go`
- Modify: `internal/cli/app_test.go`
- Modify: `README.md`

**Step 1: Write the failing test**

Cover:
- built-in `langchain` exists
- `agentlib enable --target langchain <ref>` succeeds
- generated files include `langchain-agent.py`

**Step 2: Run test to verify it fails**

Run:
- `go test ./internal/targets`
- `go test ./internal/cli`

Expected: FAIL until the second framework target is added.

**Step 3: Write minimal implementation**

Mirror the `crewai` approach with a `langchain-agent.py` starter file.

**Step 4: Run test to verify it passes**

Run:
- `go test ./internal/targets`
- `go test ./internal/cli`

Expected: PASS.

### Task 5: Verify repo

**Files:**
- Modify: `README.md`

**Step 1: Update docs**

Document:
- `crewai`
- `langchain`
- generated starter files
- export locations

**Step 2: Run verification**

Run:
- `go test ./...`
- `go build ./cmd/agentlib`

Expected: PASS.
