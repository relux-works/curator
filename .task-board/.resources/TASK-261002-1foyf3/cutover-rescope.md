# THE ONLY CURRENT INSTRUCTION — TASK-261002-1foyf3 RE-SCOPED (operator decision, binding, 2026-10-03 ~18:45Z)

Ivan decided: **rc.3 ships WITHOUT the v2 writer.** Your stop-line was correct. The atomic v1→v2 profile hash migration and the writer flip move to TASK-261003-1uzji7 (rc.4), which has your blocker evidence attached.

New scope of this task, the rc.14 pin only:
1. **Revert in your draft:**
   - `internal/hashing/hashing.go` EnableV2Writers back to **false**;
   - the legacy marker-test adaptations that existed only because of the flip;
   - the new `internal/envprofile/rc14_cutover_test.go`. Its regression belongs to TASK-261003-1uzji7, and the orchestrator keeps it via your blocker evidence.
2. **SPEC_PIN** = `43bf0a2506d5c354a73bbc3ea4623d4653db10c7`, the FINAL rc.14 peeled commit (signed tag v1.0.0-rc.14 re-created on it after a test-only spec fix, curator-spec PR #124; core manifest digest unchanged 6f832d81…). Replace every daf15ec8 pin in your draft with it. Update docs/ci-gates.md accordingly.
3. **Gap ledger.** KEEP the rc.14 `snapshot-acquisition/cases/byte-exact-snapshot` known-gap row and change its owner to **TASK-261003-1uzji7**, with the reason "rc.14 expects curator-content-v2 writes; writer enabled after the atomic v1→v2 profile hash migration". Exact counts.
4. **CHANGELOG: do NOT edit it.** The rc.3 release-notes step writes all pre-rc.3 entries centrally (CHANGELOG is a conflict hotspot). Revert any CHANGELOG change in your draft.
5. **Checks.** Run, with real exit codes and following host-rules (-work), `go test ./internal/hashing ./internal/marker ./internal/conformancecoverage ./internal/interop/... -count=1`. The hosted gate, at the default pin = rc.14, is the arbiter.

Then run `task-board handoff TASK-261002-1foyf3 --role developer` and END YOUR TURN. Never edit LOGBOOK.md.
