# Brief M1: rc.7 code wave, M1 — Emit Muse v4 from verified Curator state (TASK-261016-ont5q1)

Implement the attached plan leaf file(s) exactly: `rc7-leaf-muse-v4-producer.md` 
Read rc7-code-common.md and rc7-decisions.md (attached) first.

- The leaf is the contract: its normative clauses, acceptance criteria, mandatory paired tests, plant controls, dependencies and paths. Read the cited normative sections on curator-spec main yourself; where the leaf and the normative text disagree, the normative text wins: stop and report.
- Stay inside the leaf's attack surface and paths. If the leaf is too big for one reviewable change, implement its first coherent part, say exactly what remains, and do not fake the rest.
- No runtime qualification is claimed unless you executed it. Record skipped or unavailable rows honestly.
- Leave the change uncommitted for the Change Request flow.

## Start from what is already on main (2026-10-17)

TASK-261016-wh6v9a (L15-1 Muse slice) landed at 3113bad. It already delivers, through the shared operation-local credential-store policy projection: the Muse file-backed store literal, shared/isolated admission, the v4 launch-environment fragment composition, and the verified root-context fact; capture and status still never run the vendor tool. The consumer remainder (readiness/login, Codex v5 application, hosted override plant, adopted external sources) is TASK-261017-0pawkw, not yours.

So begin with a gap analysis: for each acceptance row and each mandatory test row of `rc7-leaf-muse-v4-producer.md`, name the code and the tests on main that already satisfy it, or state what is missing. Implement only the missing rows (likely candidates: the status distinctions in acceptance 4, the fifth-renderer-value parity plant, and any root-context negative row not yet driven through public resolve/status). Do not build a second resolver and do not duplicate existing tests. If every row is already satisfied, hand off with the mapping and an explicit justification for an empty or test-only delta. Attach the gap analysis as `TASK-261016-ont5q1_gap-analysis.md`.
