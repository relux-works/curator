# THE ONLY CURRENT INSTRUCTION — preliminary design research for curator issue #114 (researcher, read-only)

## Context
The owner's idea (2026-10-10): agent instruction files (`AGENTS.md`, `CLAUDE.md` and their environment variants) are not committed as one hand-edited file. They are kept as **chapters**, contributed by the profile, the project layer of CIP-0002, skills and the machine, and assembled at launch into the file each harness reads. This task is a preliminary design study only.

## Read
- relux-works/curator issue #114.
- curator-spec `cips/CIP-0002-project-context-in-managed-launches.md`: the root-context precedence, the project layer, protected copies, and the "Operator input (2026-10-10)" section on PR #136.
- How Curator composes and materialises context today: profile context chapters and overlay weights (curator `internal/envprofile`, `cmd/curator/compose.go`, environments §6), and `internal/contextmaterialize`.
- What each harness discovers by itself at the releases Curator pins: Claude Code, Codex, OpenCode, Pi, Muse. CIP-0002's native inventory is the starting point; verify it against vendor sources.

## Questions
1. Today: how chapters are composed and materialised per environment; which ancestor, nested, imported and override files each harness reads on its own.
2. The chapter model: where chapters live, their identity and pinning, ordering and precedence, conditional or path-scoped chapters.
3. Assembly: when it happens and where the output goes (a managed home or protected copy, never the repository); per-environment output names and dialects; how native discovery of repository files is suppressed or coexists with assembly; nested directories.
4. Coexistence with a project that commits its own `AGENTS.md`: import it as a chapter after approval, or leave it native; migration.
5. Interaction with CIP-0002 admission (operator mode, and agent mode within a granted ceiling).
6. Options with trade-offs, one recommendation, and a sketch of the smallest first slice.

## Rules
- Read-only research. No harness execution, and no tests or builds on this host.
- Deliver the study as a `.research/` file in the Change Request, and attach it as the resource `modular-instructions-design.md`. Never edit `LOGBOOK.md` (campaign rule).
- Cite repository, path and commit, or a documentation URL and version, for every claim. No secrets, personal paths or host names.
- Then `task-board handoff TASK-261010-2uqd3t --role researcher` and END YOUR TURN.
