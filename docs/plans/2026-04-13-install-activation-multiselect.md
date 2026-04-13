# Install Activation Multiselect Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make `agentlib install` offer post-install runtime activation with a multiselect prompt over detected runtimes, starting with the existing built-in target adapters and real activation for Codex.

**Architecture:** Reuse the current install store and target enable path instead of inventing a second activation system. Extend `install` with a small activation orchestration layer: detect runtimes, resolve explicit runtime flags or interactive selection, then call the existing target enabler for each selected runtime. Keep this first slice stateless beyond the existing materialized target directories; persistent activation metadata can come in a later step.

**Tech Stack:** Go, standard library flag parsing and terminal I/O, existing `internal/install` and `internal/targets` packages, Go test runner

### Task 1: Add failing CLI tests for post-install activation

**Files:**
- Modify: `internal/cli/app_test.go`

**Step 1: Write the failing tests**

Cover:

- interactive install with detected runtimes prompts for multiselect and enables selected targets
- `--no-activate` skips the prompt and target activation
- non-interactive install skips prompting and only installs
- explicit `--runtime <id>` activates without prompting

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/cli
```

Expected: fail because install does not yet prompt or activate targets.

### Task 2: Add install-time activation orchestration

**Files:**
- Modify: `internal/cli/app.go`
- Create: `internal/cli/install_activation.go`
- Create: `internal/cli/install_activation_test.go`

**Step 1: Implement the minimal orchestration**

Add:

- install flags: `--runtime`, `--all-detected`, `--no-activate`
- interactive detection based on stdin being a TTY
- multiselect prompt over detected enabled runtimes
- activation loop that reuses `targets.Enable`

**Step 2: Run focused tests**

Run:

```bash
go test ./internal/cli
```

Expected: install activation tests pass.

### Task 3: Document the new install flow

**Files:**
- Modify: `README.md`

**Step 1: Update usage**

Document:

- interactive activation after install
- `--runtime`
- `--all-detected`
- `--no-activate`

### Task 4: Verify the full repository

**Files:**
- Modify: `README.md` only if verification reveals missing usage details

**Step 1: Run verification**

Run:

```bash
go test ./...
go build ./cmd/agentlib
```

Expected: all tests pass and the binary builds cleanly.
