# Review note — TASK-261002-2ipeqa rev3 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Same-provider delta review (operator rule, R138). Your rev2 finding P1 (no candidate lane evidence) is now addressed:
- The orchestrator dispatched curator CI run **37017428885** (workflow_dispatch) on the rev3 gate commit e4baf6fa, with candidate_ref e3a88ced and candidate_manifest_sha256 6f832d81. Conclusion: SUCCESS. Candidate suite is green on ubuntu, macos and windows; only the optional rose-air lane was skipped.
- The earlier run 37005296654, on rev2, failed only on snapshot-acquisition/cases/byte-exact-snapshot (it expects curator-content-v2 writes, and the writer is OFF).

Verify:
1. The rev2 → rev3 delta is ONLY that one rc.14 gap row, owned by TASK-261002-1foyf3 with a truthful reason, plus the exact count adjustment. Nothing else changed.
2. Run 37017428885's head resolves to the rev3 candidate tree, and its Candidate suite jobs truly executed: check `gh run view 37017428885 --json jobs`.
3. Everything you ruled OK for rev2 still holds.

accept_cr, or changes requested with file:line.
