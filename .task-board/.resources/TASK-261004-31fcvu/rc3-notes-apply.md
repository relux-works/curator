# THE ONLY CURRENT INSTRUCTION — TASK-261004-31fcvu: apply the drafted rc.3 notes on the current base (developer, docs)

The previous run wrote the rc.3 section and full reconciliation (`TASK-261004-31fcvu_results.md`, `TASK-261004-31fcvu_validate.py`), but its Story workspace sat on a stale pre-rewrite base, so no Change Request could be built. The orchestrator discarded that workspace. Your workspace is fresh from main 876127f7.

1. Replace CHANGELOG.md with the attached `rc3-changelog-run3.md` (heading already normalised to the existing style `## 0.15.0-rc.3 - 2026-10-04`).
2. Verify against the CURRENT base: every section from `## 0.15.0-rc.2 - 2026-09-26` down is byte-identical to main's; every line of main's `## Unreleased` is either represented in the rc.3 section or deliberately dropped as a duplicate (list any dropped line with its reason); the history map in the results still holds for `v0.15.0-rc.2..origin/main` (re-run the attached validator, adapting only the heading text if needed).
3. Only CHANGELOG.md changes. Never edit LOGBOOK.md. Keep commands short; on a hang longer than 5 minutes wait instead of retrying (host rules).
4. Update the results resource with the new verification, then `task-board handoff TASK-261004-31fcvu --role developer` and END YOUR TURN.
