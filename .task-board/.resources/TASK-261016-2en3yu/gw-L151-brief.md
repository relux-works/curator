# Brief L151: rc.7 code wave, L15-1 — Resolve the machine-owned credential-store policy once (TASK-261016-2en3yu)

Implement the attached plan leaf file(s) exactly: `rc7-leaf-credential-store-policy.md` 
Read rc7-code-common.md and rc7-decisions.md (attached) first.

- The leaf is the contract: its normative clauses, acceptance criteria, mandatory paired tests, plant controls, dependencies and paths. Read the cited normative sections on curator-spec main yourself; where the leaf and the normative text disagree, the normative text wins: stop and report.
- Stay inside the leaf's attack surface and paths. If the leaf is too big for one reviewable change, implement its first coherent part, say exactly what remains, and do not fake the rest.
- No runtime qualification is claimed unless you executed it. Record skipped or unavailable rows honestly.
- Leave the change uncommitted for the Change Request flow.

## Scope decision (convener, 2026-10-16)

This task is narrowed to slice 1, exactly as described in the task description: config grammar, CLI, typed transport into resolve, and fail-closed unsupported selections. The remainder moves to TASK-261016-hzbo83 (L15-1b). Finish slice 1: check every DoD item against the slice, run the full remote gate to a green verdict, and hand off for review.
