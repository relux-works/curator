# TASK-260916-2ok97n (R5) — rework 5 (orchestrator, binding). THIS IS THE ONLY CURRENT INSTRUCTION. HIGHEST PRIORITY.

Revision 5's Windows fix (bind worker evidence to the manager's private Job Object via a duplicated query handle; missing/foreign
handle refuse) is correct — KEEP it. Gate run 35944695862 failed only on the ledger:
    ledger case skipped where the ledger does not tolerate it (linux): cmd/curator :: TestCLICompatibleVerifiedProviderOwnsBuildDispatchAndReceipt
    main_test.go:242: rc5-native-control-inventory-v1 defines no record for host linux; the portable execution policy is specified for macOS and Windows only
The skip on Linux is by design. Make this test's `.github/ci/platform-cases.tsv` row `must_run_on darwin,windows`, `skip_allowed_on
linux`, class `platform-control` (the existing skip class "rc5-native-control-inventory-v1 defines no record for host" already covers
the reason) — mirror how other inventory-bound rows are declared. Nothing else changes. Run `sh .github/ci/ledger-consistency.sh`,
`sh .github/ci/gate-selftest.sh`, and the test locally on darwin. Append "Revision 6" to results, `task-board resource update` it,
then `task-board handoff TASK-260916-2ok97n --role developer`; stay in the turn during the gate. Do NOT run refresh-candidate (trunk
is frozen for R5). A `run_wrote_outside_worktree … policy warn` block is a warning — verify status `to-review` and revision 6 published.
