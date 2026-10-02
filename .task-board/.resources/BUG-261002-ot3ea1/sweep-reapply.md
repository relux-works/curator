# THE ONLY CURRENT INSTRUCTION — BUG-261002-ot3ea1 re-apply onto current trunk → revision 3 (orchestrator, binding)

Rev2 was green, but trunk 3fc3de2d changed `.github/ci/platform-cases.tsv` and CHANGELOG.md, and `worktree converge` refused with a conflict on platform-cases.tsv. The orchestrator discarded the workspace on purpose. Your rev2 delta is `refs/campaign/ot3ea1-rev2-20261002` (d7df3b8f), on base 2cb29dac.

1. In your fresh workspace, run `git diff 2cb29dac d7df3b8f -- . ':!.task-board' | git apply --3way`.
   - Resolve platform-cases.tsv by keeping trunk's rows AND your rows. Check the exact format with `bash .github/ci/gate-selftest.sh` or its ledger subset.
   - CHANGELOG keeps both sides.
2. Every other path must equal rev2. Re-run the buildcache tests, `go test ./internal/envprofile -run TestManagerOwnedAbsenceReadsAreGuarded -count=1` and lint if available, with real exit codes and per host-rules (-work).
3. Add a "Revision 3 (re-apply)" section to the results. Then run `task-board handoff BUG-261002-ot3ea1 --role developer` and END YOUR TURN. Never edit LOGBOOK.md.
