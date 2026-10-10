# THE ONLY CURRENT INSTRUCTION — research for curator issue #113: project surfaces the Skillfile does not cover (researcher, read-only)

## Context
CIP-0002 composes an admitted **project layer** over a profile, outside the repository; the owner confirmed this model on 2026-10-10 (see the "Operator input" section on curator-spec PR #136). The Skillfile declares a project's skills. Rules, knowledge, environment-specific instruction files and native settings have no resolution path yet (relux-works/curator issue #113).

## Questions
1. **Inventory**, per environment (Claude Code, Codex, OpenCode, Pi, Muse), at the releases Curator pins and at the latest release, of every project-level surface:
   - rules, including conditional, path-scoped and execution-policy rules;
   - knowledge and memory;
   - instruction files and overrides;
   - settings, MCP, hooks, plugins and extensions;
   - commands and skills.
2. For each surface:
   - how it is discovered as untrusted data;
   - how it is normalised into a closed declaration;
   - how it is admitted (by the operator, or by an agent within a granted ceiling);
   - its precedence against the profile;
   - how it is materialised as protected copies;
   - which native discovery must be suppressed.
3. **The Skillfile boundary:** what stays declared in the Skillfile and its lock, and what lives in the project layer. How skills that are installed today into the repository under gitignore move into the project layer.
4. **Recommendation:** a new CIP or an amendment to CIP-0002, and the smallest first slice.

## Rules
- Read-only research. No harness execution, and no tests or builds on this host.
- Deliver the study as a `.research/` file in the Change Request, and attach it as the resource `project-surfaces-coverage.md`. Never edit `LOGBOOK.md` (campaign rule).
- Cite repository, path and commit, or a documentation URL and version, for every claim. No secrets, personal paths or host names.
- Then `task-board handoff TASK-261010-1992si --role researcher` and END YOUR TURN.
