# Framework Package Export Design

## Goal

Extend the new `package-export/generate` family beyond `openclaw` so AgentLib packages can be projected into framework-oriented environments such as `crewai` and `langchain`.

## Core Idea

Unlike TUI skill clients, frameworks do not usually consume a directory of markdown skills directly. They need either:

- a generated package export directory
- framework-specific metadata
- or a starter integration module that points back to the exported package

The right move is to reuse the same family introduced for `openclaw`:

- `format: package-export`
- `mode: generate`

but let each framework define a slightly different generated layout inside that family.

## Recommended Progression

Do not implement `crewai` and `langchain` together in one step. The lower-risk order is:

1. `crewai`
2. `langchain`

`crewai` is a cleaner first consumer because the initial integration can stay thin:

- export canonical package
- emit `agentlib-export.json`
- emit a small `crewai-agent.py` or equivalent starter module

That proves the framework-target path without committing to a full runtime integration contract.

`langchain` should then reuse the same pattern:

- export canonical package
- emit `agentlib-export.json`
- emit a small `langchain-agent.py` or equivalent starter module

## Adapter Families

After `openclaw`, the target system now has two families:

1. `skill-dir`
2. `package-export`

Keep those families explicit. Do not collapse them into one “do everything” mode.

For framework targets, the package export family should support:

- recursive copy of canonical package
- export metadata
- optional generated entry files

That is enough for both `crewai` and `langchain` v1.

## Built-In Targets

Recommended built-ins:

- `crewai`
  - `format: package-export`
  - `mode: generate`
  - `relativePath: .crewai/agents`
  - `installRoot: ~/.crewai/agents`

- `langchain`
  - `format: package-export`
  - `mode: generate`
  - `relativePath: .langchain/agents`
  - `installRoot: ~/.langchain/agents`

## Scope Boundaries

Do not attempt in v1:

- automatic Python environment management
- package publishing to pip
- runtime execution inside the framework
- bidirectional sync from generated files back into AgentLib
- speculative abstractions for every future framework

The purpose of this phase is only to prove that AgentLib can export framework-ready packages in a stable, inspectable layout.
