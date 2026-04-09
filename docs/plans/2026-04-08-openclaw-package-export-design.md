# OpenClaw Package Export Design

## Goal

Add the first non-`skill-dir` target family to `agent-cli`, starting with `openclaw`.

The current target system works well for TUI clients that consume a directory of skills through symlinks or copies. `openclaw`, `crewai`, and `langchain` are different: they need exported package layouts or generated integration files rather than a plain skill directory. `openclaw` should be the first adapter that proves that second family.

## Recommended Approach

The best next step is not to generalize all frameworks at once. It is to introduce one new adapter family:

- `format: package-export`
- `mode: generate`

This family means:

1. Read the canonical AgentLib package from the installed store
2. Generate a target-specific export under the target root
3. Optionally emit a small manifest/index file for discovery

For `openclaw`, the export should stay intentionally thin in v1:

- copy the canonical package into a generated target directory
- emit a small `agentlib-export.json` file describing source ref, export time, and adapter id
- leave room for a future OpenClaw-native manifest once that contract is stable

This avoids inventing a fake OpenClaw runtime contract too early while still proving the `generate` pipeline.

## Why This Approach

Three options exist:

1. Treat `openclaw` like another `skill-dir`
2. Add a new generic `package-export` family
3. Jump straight to framework-specific manifests for every non-TUI target

Option 2 is the right move.

Option 1 is too narrow and will break as soon as `openclaw` needs metadata beyond files-on-disk. Option 3 is too broad for one slice and will turn into speculative architecture. `package-export` gives a stable middle layer that `openclaw` can use first, and `crewai` / `langchain` can later share.

## Proposed Data Model

Extend `Target` with the minimum metadata needed for generated exports:

- `format`
- `mode`
- `installRoot`
- optional `exportLayout`

For `openclaw` built-in:

- `id: openclaw`
- `format: package-export`
- `mode: generate`
- `relativePath: .openclaw/agents`
- `installRoot: ~/.openclaw/agents`

That keeps target selection consistent with the rest of the CLI while clearly separating this from `skill-dir` mode.

## Export Layout

Recommended v1 export root:

```text
~/.openclaw/agents/<namespace>/<name>/<version>/
```

Generated contents:

- exported package files
- `agentlib-export.json`

The metadata file should include:

- `targetId`
- `sourceRef`
- `sourceStorePath`
- `exportedAt`
- `formatVersion`

## Scope Boundaries

This slice should not attempt:

- OpenClaw execution/runtime integration
- bidirectional sync
- generated Python code
- framework support for `crewai` or `langchain`

Those come later, after the `generate` adapter path is proven with `openclaw`.
