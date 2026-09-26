# TASK-260907-2as5sx — rework 1 (orchestrator, binding). THIS IS THE ONLY CURRENT INSTRUCTION.

Revision 1 (refreshed onto 1511b345, gate run 35932571195) has two real failures:

A. REGRESSION (ubuntu + macOS race and test lanes): `internal/crossconformance TestDraftSourcesSemanticCases/external-evidence-mismatch-execution_policy`
   now fails: `review: manager_state_unreadable: …/.agents/skills/review/.csk-install.json: install marker is not valid for its
   schema … want re-derivation success`. Before your seam, a schema-INVALID marker in this path led to re-derivation; the
   semantic case (the spec's normative expectation) still requires that. §8.4 separates ABSENT from FAILED READ — it does not
   say a successfully read but invalid/stale marker is "unreadable". Keep three outcomes distinct where the spec does:
   absent / read-failed (unreadable) / read-OK-but-invalid (whatever the spec prescribes for that reader — here: re-derive).
   Find every reader where you mapped decode/schema-invalid to `manager_state_unreadable`, check the spec's rule for that
   reader (environments §8.4/§8.4.1, the marker sections), and restore the prescribed behaviour; add a row per such reader for
   the invalid-marker case, and keep the absent-vs-unreadable rows.
B. Windows platform-case gate: `skip with an unrecognised reason on windows` for `internal/marker TestCurrentDistinguishesAbsentAndUnreadable`,
   `TestReadStateDistinguishesAbsentAndUnreadable`, `internal/ui TestSkillsUnderDistinguishesAbsentAndUnreadableMarkers`. Give each
   its own ledger row `linux,darwin  windows  platform-control <POSIX mode-bit unreadability>` like `internal/gitops
   TestExtractPreservesExecutableBit` (row ~86) with the matching registered skip reason. Do NOT touch shared DACL helpers.
Run bounded: `go test ./internal/crossconformance -run TestDraftSourcesSemanticCases` (+ `-race`), your focused rows, the AST
inventory test, ledger-consistency + gate-selftest. Append "Revision 2" to results, `task-board resource update` it, then
`task-board handoff TASK-260907-2as5sx --role developer`. A `run_wrote_outside_worktree … policy warn` block is a warning —
verify status `to-review` and that revision 2 was published.
