# TASK-260924-20o9dk — gate fix (THE ONLY CURRENT INSTRUCTION)

Revisions 1–3 failed the hosted gate on the SAME three tests (latest run 36222502043, every lane) — the autonomous successors did not fix
them. Your conformance work otherwise stands (105/105 semantic, 3/3 snapshot; local-snapshot bounds per `20o9dk-decision-1.md`).
1. internal/crossconformance TestDraftLiteralKeepsAgentAndAskpass (draftsources_literal_isolation_test.go:629) and
   TestDraftLiteralRefreshIgnoresUserConfig (draftsources_semantic_v2_test.go:569): refresh now resolves the CURRENT endpoint plan (#90
   rule you implemented), so their fixture endpoint `https://fixture.test/kit` is attempted for real → repository_endpoint_unavailable.
   Fix the FIXTURES, not the rule: give each test a current machine endpoint plan that resolves to its local fixture repository (the same
   way the new v2-refresh-current-endpoint rows do), so each still asserts what it was written for (agent/askpass preserved; user git config
   ignored) — and add an assertion that the stored remote.origin.url is NOT what selected the endpoint.
2. TestIntegrationSurfaceStartsNoProcessOutsideTheSharedSeams (guard_test.go:50): your new draftsources_gitshim_test.go starts processes
   outside the instrumented process boundary. Route it through the shared seam other test shims use (look at how existing fixtures start
   git) — do not add a blanket allow-list entry; if an allow-list row is truly the project's pattern for test shims, cite the precedent.
3. Run locally, bounded and split: `go test ./internal/crossconformance -run 'TestDraftLiteralKeepsAgentAndAskpass|TestDraftLiteralRefreshIgnoresUserConfig|TestIntegrationSurfaceStartsNoProcessOutsideTheSharedSeams' -count=1`
   and `-race` for the first two; plus your new #90 rows. Real exit codes.
4. `task-board m 'set_status(TASK-260924-20o9dk, status=development)'` first; append "Revision 4 — gate fix" to results, `resource update`,
   `task-board handoff TASK-260924-20o9dk --role developer`; stay in the turn while the gate runs. If the loop detector refuses the spawn or
   handoff, stop and report — do not work around it. A write-boundary `policy warn` block is a warning.
