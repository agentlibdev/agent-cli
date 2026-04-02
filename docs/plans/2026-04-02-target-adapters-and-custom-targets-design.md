# Target Adapters And Custom Targets Design

## Goal

Let AgentLib-managed skills be usable across multiple agent clients without forcing a fake universal runtime. AgentLib should own the canonical package and project it into each target client's expected format through adapters.

## Core Model

The design has four layers:

1. Canonical package in AgentLib
2. Neutral compatibility model
3. Target adapters
4. Local activation layer

The canonical package stays vendor-neutral. It contains the published files, provenance, and package metadata. On top of that, AgentLib introduces a compatibility model describing which targets a package is built for, tested with, or can degrade into. That model is then consumed by target adapters, each of which knows how to materialize the package for a concrete client such as Codex, Claude, Gemini CLI, or OpenClaw.

## Target Definitions

The CLI needs both built-in and custom targets.

Built-in targets:

- `codex`
- `claude`
- `gemini-cli`
- `openclaw`

Custom targets are defined declaratively in JSON. The first version should support:

- a global file at `~/.agentlib/targets.json`
- an optional project override at `.agentlib/targets.json`

Those files should not be arbitrary scripts. They should select from a small set of adapter formats and declare installation/materialization paths.

## Proposed Config Shape

```json
{
  "targets": [
    {
      "id": "codex",
      "type": "built-in",
      "enabled": true
    },
    {
      "id": "openclaw-local",
      "type": "custom",
      "format": "markdown-skill-dir",
      "installRoot": "/home/raul/.openclaw/skills",
      "manifestPath": "/home/raul/.openclaw/config/skills.json",
      "mode": "symlink",
      "enabled": true
    }
  ]
}
```

Fields intentionally kept in v1:

- `id`
- `type`
- `format`
- `installRoot`
- `manifestPath`
- `mode`
- `enabled`

No script hooks yet.

## CLI Direction

The CLI should grow a `targets` command group.

First useful commands:

- `agentlib targets list`
- `agentlib targets detect`
- `agentlib enable <ref> --target <id>`

The first implementation slice should only ship `targets list`. That gives a stable foundation for config parsing, built-in registry of supported targets, and future UX.

## Web Follow-Up

The registry should later expose compatibility badges on the web:

- `Built for`
- `Tested with`
- `Adapter available`

That should not block the CLI target system. The web can consume compatibility metadata after the CLI model exists and the API contract is added in `agentlib`.
