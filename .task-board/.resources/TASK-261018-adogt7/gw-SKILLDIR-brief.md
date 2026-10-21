# TASK-261018-adogt7: implement CIP-0021, `directory` on profile skill requirements (developer)

## Why

Ivan approved this on 2026-10-18; it is needed for the agents-infra exit. 18 of the 40 skills one host needs live in subdirectories of shared repositories, and today a profile can address a skill only at a repository root.

## Normative source

curator-spec PR #162 (CIP-0021, decision 0027): the accepted candidate, head e41f963, landing on main. Read its `protocol/environments.md` §2 and §1, `protocol/rc7-carriers.md`, the schemas and the vectors. Implement exactly that; where code and spec disagree, the spec wins.

## Required

1. **Parse** (`internal/contextpkg`, `parseRequirements`). In context manifest `schema_version: 2`, `requires.skills.<name>` admits `directory` with the context and MCP grammar. In manifest schema 1 it stays refused, as today.
2. **Resolve** (`internal/contextresolve`). Agreement across requirers works as it does for contexts, using the diagnostic the spec names. The skill manifest and `SKILL.md` are read below the directory.
3. **Lock** (`internal/contextlock`). Lock `schema_version: 4` records `directory` on skill members, and lock identity covers it.
4. **Audit.** Findings, capabilities and content identity are taken below the selected directory, through `packageRoot(entry, directory)` as for contexts.
5. **Materialization.** Managed homes link the selected directory as the skill root; in-place surfaces do the same. BUG-261018-uoisfe (NATSKILL) is changing where in-place skill links go. Build on current trunk and converge if it lands first; do not duplicate its work.
6. **Snapshot discipline.** A directory that escapes the snapshot or traverses a symlink is refused with the spec's diagnostic.
7. **Migration is unchanged.**
8. **Tests.**
   - Drive the spec's positive and negative vectors through the production entry points.
   - Add at least one narrowing mutant per gate: the schema 1 refusal, agreement, escape and symlink, and audit below the directory.
   - Use temporary homes only; never touch the real HOME.
9. **Validation.** Run the full local suite and the full hosted gate on the exact signed candidate. A runner infrastructure failure is a blocker to report, never a pass.

## Scope

- In scope: code, tests, a CHANGELOG line under Unreleased, and a docs note under `docs/`.
- Not in scope: the root `README.md`, which trunk changes often and so forces converge rounds.

## Hand-off

1. Attach `TASK-261018-adogt7_gaps.md` with the vector coverage ratio, the mutant table and the production call sites.
2. Run `task-board handoff TASK-261018-adogt7 --role developer`.
3. END YOUR TURN.
