# Global And Local Install Modes Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make `agentlib install` use a user-global store by default while supporting explicit project-local installs with `--local`, `--global`/`-g`, `--install-dir` for local overrides, and `agentlib init` to mark a directory as a local AgentLib project.

**Architecture:** The CLI will stop deriving install targets directly from the current working directory. Instead, a small install-target resolver will select either the user home store (`~/.agentlib`) or a project-local store rooted at a detected or initialized project marker. `install`, `remove`, and the new `init` command will share this resolver so behavior stays consistent and testable.

**Tech Stack:** Go, standard library, existing `internal/cli` and `internal/install` packages.

### Task 1: Document and encode target resolution rules

**Files:**
- Create: `docs/plans/2026-04-02-global-local-install-modes.md`
- Create: `internal/install/target.go`
- Test: `internal/install/target_test.go`

**Step 1: Write the failing test**
Create tests covering:
- default target resolves to `$HOME/.agentlib`
- `--global`/`-g` resolves to `$HOME/.agentlib`
- `--local` resolves to `<project>/.agentlib` when `project.json` exists
- `--local --install-dir <dir>` resolves to `<dir>`
- `--local` without initialized project returns a clear error
- conflicting flags `--global` and `--local` return a clear error

**Step 2: Run test to verify it fails**
Run: `go test ./internal/install ./internal/cli`
Expected: FAIL because target resolver/types do not exist yet.

**Step 3: Write minimal implementation**
Add a small target resolver that:
- reads `HOME`
- checks `.agentlib/project.json`
- returns a typed result with `Root`, `Mode`, and `LockfilePath`

**Step 4: Run test to verify it passes**
Run: `go test ./internal/install ./internal/cli`
Expected: PASS for new target-resolution tests.

**Step 5: Commit**
```bash
git add docs/plans/2026-04-02-global-local-install-modes.md internal/install/target.go internal/install/target_test.go
git commit -m "feat: add install target resolution"
```

### Task 2: Update install/remove CLI parsing

**Files:**
- Modify: `internal/cli/app.go`
- Modify: `internal/cli/app_test.go`
- Test: `internal/cli/app_test.go`

**Step 1: Write the failing test**
Add tests covering:
- `agentlib install ref` uses global target by default
- `agentlib install --local ref` rejects non-initialized directories
- `agentlib install --local --install-dir custom ref` accepts custom local root
- `agentlib remove` defaults to global target
- `agentlib init` creates `.agentlib/project.json`

**Step 2: Run test to verify it fails**
Run: `go test ./internal/cli`
Expected: FAIL because flags/command are not parsed yet.

**Step 3: Write minimal implementation**
Use `flag.FlagSet` per command to parse:
- `install`: `--local`, `--global`, `-g`, `--install-dir`
- `remove`: same target selectors
- `init`: optional path later, for now current directory only

**Step 4: Run test to verify it passes**
Run: `go test ./internal/cli`
Expected: PASS.

**Step 5: Commit**
```bash
git add internal/cli/app.go internal/cli/app_test.go
git commit -m "feat: add global and local install flags"
```

### Task 3: Adapt installer implementation to resolved roots

**Files:**
- Modify: `internal/install/install.go`
- Modify: `internal/install/install_test.go`
- Test: `internal/install/install_test.go`

**Step 1: Write the failing test**
Update install/remove tests to operate on explicit resolved roots instead of implicit cwd-derived roots.

**Step 2: Run test to verify it fails**
Run: `go test ./internal/install`
Expected: FAIL until install/remove accept resolved targets.

**Step 3: Write minimal implementation**
Refactor `Run` and `Remove` to accept a resolved target root and keep on-disk layout identical under that root.

**Step 4: Run test to verify it passes**
Run: `go test ./internal/install`
Expected: PASS.

**Step 5: Commit**
```bash
git add internal/install/install.go internal/install/install_test.go
git commit -m "refactor: install agents into resolved targets"
```

### Task 4: Update docs and verify end-to-end

**Files:**
- Modify: `README.md`

**Step 1: Update docs**
Document:
- default global behavior
- `agentlib init`
- `--local`
- `--global` / `-g`
- `--install-dir` requiring `--local`

**Step 2: Run verification**
Run:
- `go test ./...`
- `go build ./cmd/agentlib`

Expected: PASS.

**Step 3: Commit**
```bash
git add README.md
git commit -m "docs: document global and local install modes"
```
