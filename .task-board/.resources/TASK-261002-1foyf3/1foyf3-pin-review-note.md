# Review note — TASK-261002-1foyf3 pin-only (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Same-provider review (operator rule, tb-R164). Scope per `cutover-rescope.md` (operator decision: rc.3 WITHOUT the v2 writer).

Verify, with real exit codes:
1. SPEC_PIN = 43bf0a2506d5c354a73bbc3ea4623d4653db10c7 (the FINAL re-tagged curator-spec v1.0.0-rc.14; tag object 661bead0). No daf15ec8 pin remains anywhere in the delta. docs/ci-gates.md agrees. Manifest sha256 6f832d81… verified against the release asset or the tagged tree.
2. `internal/hashing` EnableV2Writers is false; no writer-flip leftovers (no marker-test adaptations for the flip, no `internal/envprofile/rc14_cutover_test.go`).
3. Gap ledger: exactly one rc.14 known gap `snapshot-acquisition/cases/byte-exact-snapshot`, owner TASK-261003-1uzji7, the agreed reason; exact counts consistent.
4. No CHANGELOG.md or LOGBOOK.md edits.
5. The producer reports two initial exit-1 runs: a stale-cache lint (rerun with a fresh cache exit 0) and rc.13 archive-materialization checks. Determine whether the rc.13 archive-materialization failure is pre-existing on main 68210ecc (run the same command on main) or introduced by this delta. Introduced = defect.
6. The CR validation gate (remote-gate) result for this revision: cite it. If the hosted rc.14 matrix is not part of it, say so; the orchestrator dispatches the candidate lane separately.

accept_cr if all hold; otherwise changes requested with concrete findings.
