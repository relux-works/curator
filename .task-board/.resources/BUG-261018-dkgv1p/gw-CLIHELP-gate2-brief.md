# THE ONLY CURRENT INSTRUCTION — BUG-261018-dkgv1p: rerun the full hosted gate again (developer)

Attempt 3 of run 39067488895 failed in the Lint job. The pinned golangci-lint installation hit `Text file busy` (tar exit 2). The cause was in the runners: every self-hosted runner shared one HOME, and golangci-lint-action unpacks its binary into `$HOME`, so two concurrent Lint jobs overwrote each other's binary. The runner operator has fixed this. Each runner now has its own HOME; toolchains and caches are still shared.

## Do this

1. Do not change the candidate unless the gate shows a real failure.
2. Rerun the full hosted gate on your exact signed candidate, 233c8873889a9f25cd0e2d040a39bcf564a6540c. If GitHub refuses a rerun of the old run, push a fresh dispatch of the same SHA through the route you used before.
3. If the gate is green, complete the remaining checklist rows with the run URL and the job results. Then hand off with `task-board handoff BUG-261018-dkgv1p --role developer`.
4. If a real, source-caused failure appears, fix it, re-sign, rerun, and record both runs.

A runner infrastructure failure is still a blocker to report, never a pass.

End your turn after the handoff, or after recording a blocker.
