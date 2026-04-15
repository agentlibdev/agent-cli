# Kiro Built-In Target

**Goal:** Add `kiro` as a first-class built-in target so AgentLib-managed skills can be materialized directly into Kiro's skills directory.

**Source basis:** Kiro's CLI docs support skills via `skill://.kiro/skills/**/SKILL.md` and use `.kiro/` as the local/global configuration root. For AgentLib's built-in target model, the matching user-global install root is `~/.kiro/skills`.

**Target contract**

- `id`: `kiro`
- `name`: `Kiro`
- `format`: `kiro`
- `relativePath`: `.kiro/skills`
- `installRoot`: `~/.kiro/skills`
- `mode`: `symlink`

**Why this shape**

- Kiro fits the existing `skill-dir` family better than the `package-export` family
- no generated adapter files are needed for the first slice
- the built-in keeps CLI, compatibility metadata, and registry UI aligned on the same `targetId`
