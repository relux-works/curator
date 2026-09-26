# TASK-260907-2as5sx — rework 2 (orchestrator, binding). THIS IS THE ONLY CURRENT INSTRUCTION.

Revision 2 (gate run 35941410670) fixed rework-1's items but has two new failures:
A. `Test (macos-latest)` gofmt check — run `gofmt -l cmd internal` and format every listed file.
B. `internal/envprofile TestSurfacingEmittedDespitePublicationFailure` (ubuntu/macOS race): `surfacing_order_test.go:271: no §2.3 row
   reached the sink before the failure: manager_state_unreadable: …/profiles/withmcp/source.json: open …: not a directory`.
   The test injects a publication failure and requires the §2.3 surfacing rows to be emitted BEFORE it. Your seam now fails at the
   `source.json` read (ENOTDIR) earlier than before, so no row is emitted. Decide against the spec (environments §2.3 ordering
   and §8.4): if the read of `source.json` in this flow happens before surfacing only because of your refactor, restore the
   prescribed order so the rows are emitted first and the failure is still reported (not swallowed); if the spec truly requires
   the read to precede surfacing, show the clause and change the test's fixture — never delete the assertion. Say which, with the
   clause.
Re-run bounded: `gofmt -l cmd internal` (empty), `go test ./internal/envprofile -run 'Surfacing|Stateread|Absent|Unreadable'`
(+ `-race`), `go test ./internal/crossconformance -run TestDraftSourcesSemanticCases`, AST inventory test, gate scripts. Append
"Revision 3" to results, `task-board resource update` it, then `task-board handoff TASK-260907-2as5sx --role developer`. A
`run_wrote_outside_worktree … policy warn` block is a warning — verify status `to-review`.
