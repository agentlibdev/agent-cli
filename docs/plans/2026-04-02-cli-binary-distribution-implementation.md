# CLI Binary Distribution Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add release-time binary packaging and first-party shell/PowerShell installers so `agentlib` can be installed without a Go toolchain.

**Architecture:** Keep the existing Go CLI unchanged and package it as native binaries via GoReleaser. Add thin install scripts that download versioned release artifacts into a project-owned bin directory and print PATH instructions. Defer npm to a follow-up slice so the first milestone stays focused on the canonical release path.

**Tech Stack:** Go, GoReleaser, GitHub Releases, POSIX shell, PowerShell, Markdown docs

### Task 1: Add a stable CLI version surface

**Files:**
- Modify: `cmd/agentlib/main.go`
- Modify: `internal/cli/app.go`
- Create: `internal/version/version.go`
- Test: `internal/cli/app_test.go`

**Step 1: Write the failing test**

Add a CLI test for:

- `agentlib version`
- expected output includes the semantic version string

**Step 2: Run test to verify it fails**

Run: `go test ./internal/cli -run Version`
Expected: FAIL because the command does not exist yet

**Step 3: Write minimal implementation**

Implement:

- a tiny version package with defaults such as:
  - version = `dev`
  - commit = `none`
  - date = `unknown`
- a `version` command in the CLI dispatcher
- a single-line default output such as `agentlib dev`

**Step 4: Run test to verify it passes**

Run: `go test ./internal/cli -run Version`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/agentlib/main.go internal/cli/app.go internal/cli/app_test.go internal/version/version.go
git commit -m "feat: add cli version command"
```

### Task 2: Add release packaging with GoReleaser

**Files:**
- Create: `.goreleaser.yml`
- Modify: `README.md`
- Optionally create: `.github/workflows/release.yml`

**Step 1: Write a minimal verification target**

Document and prepare a local dry-run command:

```bash
goreleaser release --snapshot --clean
```

**Step 2: Add minimal GoReleaser config**

Include:

- binary name `agentlib`
- target matrix for:
  - linux amd64/arm64
  - darwin amd64/arm64
  - windows amd64/arm64
- archive naming rules
- checksum generation
- ldflags wiring version metadata into the binary

**Step 3: Verify snapshot packaging**

Run: `goreleaser release --snapshot --clean`
Expected: archives and checksums generated locally without publishing

**Step 4: Commit**

```bash
git add .goreleaser.yml README.md .github/workflows/release.yml
git commit -m "build: add cli release packaging"
```

### Task 3: Add POSIX installer

**Files:**
- Create: `scripts/install.sh`
- Modify: `README.md`

**Step 1: Write a script-level verification checklist**

The script must:

- accept an optional version override
- detect OS/arch
- download archive and checksum from GitHub Releases
- verify checksum
- install into `~/.agentlib/bin`
- print PATH guidance

**Step 2: Implement the minimal shell installer**

Keep it small and explicit:

- support `curl` or `wget`
- use `tar` for unix archives
- refuse unsupported platforms loudly

**Step 3: Verify against a snapshot or known release**

Run locally against a prepared release URL or snapshot fixture.
Expected: `~/.agentlib/bin/agentlib` exists and `agentlib version` runs.

**Step 4: Commit**

```bash
git add scripts/install.sh README.md
git commit -m "feat: add shell installer"
```

### Task 4: Add Windows installer

**Files:**
- Create: `scripts/install.ps1`
- Modify: `README.md`

**Step 1: Write a script-level verification checklist**

The PowerShell installer must:

- accept optional version override
- detect architecture
- download zip and checksum
- verify checksum
- install into `%USERPROFILE%\\.agentlib\\bin`
- print PATH guidance

**Step 2: Implement the minimal PowerShell installer**

Prefer:

- `Invoke-WebRequest`
- built-in archive extraction
- explicit checksum validation

**Step 3: Verify on Windows**

Run the script on a Windows environment with a known release.
Expected: `%USERPROFILE%\\.agentlib\\bin\\agentlib.exe` exists and `agentlib version` runs.

**Step 4: Commit**

```bash
git add scripts/install.ps1 README.md
git commit -m "feat: add powershell installer"
```

### Task 5: Document official install flow and defer npm cleanly

**Files:**
- Modify: `README.md`

**Step 1: Update install section**

Document:

- official install command for Linux/macOS
- official install command for Windows
- custom bin path location
- how to add that path to `PATH`
- npm as secondary follow-up channel, not primary

**Step 2: Verify docs match shipped assets**

Check:

- script paths exist
- binary name matches archives
- install location matches scripts

**Step 3: Run full verification**

Run:

```bash
go test ./...
go build ./cmd/agentlib
```

If available, also run:

```bash
goreleaser release --snapshot --clean
```

Expected:

- tests pass
- binary builds
- release snapshot succeeds

**Step 4: Commit**

```bash
git add README.md
git commit -m "docs: document cli installation"
```

### Task 6: Follow-up slice for npm wrapper

**Files:**
- Future repo or package path to be decided

**Step 1: Keep this explicitly out of milestone 1**

Do not implement npm in the first distribution slice.

**Step 2: Capture follow-up requirements**

The npm wrapper should later:

- resolve platform
- download the same GitHub Release binary
- expose `agentlib` on install
- avoid reimplementing CLI logic in Node

**Step 3: No commit in this slice**

This is a deferred follow-up, not part of the first execution batch.
