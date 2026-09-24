# TASK-260916-2ok97n (R5) — rework 4 (orchestrator, binding). THIS IS THE ONLY CURRENT INSTRUCTION. HIGHEST PRIORITY.

Revision 4 (refreshed onto trunk 1511b345, gate run 35936900617) is green on ubuntu/macOS; Windows fails ONE test that passes on
the same trunk in other Stories' gates (e.g. BUG-260922-6chzf9 rev3 on 1511b345), so treat it as a regression from this Story:

    cmd/curator TestCLICompatibleVerifiedProviderOwnsBuildDispatchAndReceipt:
      go-v1 build_execution_capability_evidence_invalid: worker job limit flags are 0x0, want at least 0x2308

The build worker's Windows job-object limits (`internal/godriver/controls_windows.go`) are observed as 0x0. This Story's delta
touches `internal/godriver/identity.go` and several `internal/scriptworker/*_windows.go` (apply/exec/interpreter/inventory/
teardown). Find how the delta changes the go-v1 build worker path on Windows (shared primitive, job assignment, identity
re-check order, evidence collection, a helper moved between packages…), fix the cause in production code, and keep every
script-worker behaviour of revisions 3-4 (System32 identity bound, unresolved exec report-only).
Evidence: `gh run download 35936900617 -n test-evidence-windows-latest` (from your worktree) → `test/go-test.json`.
Add a cross-platform row if the cause is expressible off-Windows; cross-compile the Windows test binaries; bounded runs of
`./internal/godriver/... ./internal/scriptworker/...` and the cmd/curator build-dispatch tests. Append "Revision 5" to results,
`task-board resource update` it, then `task-board handoff TASK-260916-2ok97n --role developer`; stay in the turn while the gate
runs. A `run_wrote_outside_worktree … policy warn` block is a warning — verify status `to-review` and revision 5 published.
Do NOT run refresh-candidate: the base is current (trunk frozen for R5).
