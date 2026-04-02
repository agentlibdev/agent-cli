# CLI Binary Distribution Design

**Goal:** Ship the first real end-user installation path for `agentlib` via prebuilt binaries and thin install scripts, with `npm` explicitly treated as a secondary distribution channel.

## Scope

This slice does not change the CLI command surface. The product goal is distribution, not feature expansion. The installer should make it easy for a user to get a working `agentlib` binary onto their machine without requiring Go, Git, or a local build toolchain.

Primary install channels:

- Linux/macOS: `curl ... | sh`
- Windows: `iwr ... | iex`

Secondary channel:

- `npm i -g @agentlibdev/cli`

The canonical artifact source should be GitHub Releases in `agent-cli`, not npm tarballs and not source builds performed during install.

## Recommended approach

Publish native binaries for a small target matrix using GoReleaser, then build thin installer scripts around those release assets.

Install location:

- Linux/macOS: `~/.agentlib/bin/agentlib`
- Windows: `%USERPROFILE%\\.agentlib\\bin\\agentlib.exe`

The shell and PowerShell installers should:

1. detect OS and architecture
2. resolve the matching release asset
3. download the archive plus checksum
4. verify the checksum before unpacking
5. place the binary in the AgentLib-owned bin directory
6. print PATH guidance if that directory is not already reachable

This keeps the installer logic small and makes release assets the single source of truth.

## Why this shape

Using prebuilt binaries matches the actual implementation language: the CLI is already Go. That means installation should optimize for downloading a tested artifact, not rebuilding from source under user-specific environments.

Using a project-owned install directory avoids guessing user shell setup and avoids silently writing into platform-specific shared locations. It also keeps uninstall and future self-update behavior straightforward.

Making npm secondary is correct here. npm can improve discoverability and give a familiar install command, but it should still download the same release binary rather than turning Node into a required runtime dependency for the CLI itself.

## Milestone boundaries

Included in this milestone:

- GoReleaser config
- release archives and checksums
- `install.sh`
- `install.ps1`
- README install docs

Explicitly deferred:

- npm wrapper package
- Homebrew tap/formula
- self-update command
- auth or publish commands in the CLI
- shell completion

## Success criteria

After this slice:

- maintainers can cut a release and obtain versioned binaries automatically
- a Linux/macOS user can install with `curl ... | sh`
- a Windows user can install with `iwr ... | iex`
- the README documents the official installation path clearly
