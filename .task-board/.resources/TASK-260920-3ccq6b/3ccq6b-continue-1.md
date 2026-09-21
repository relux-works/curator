# TASK-260920-3ccq6b continuation 1 (orchestrator, binding)

Your previous run (RUN-260920-d8a908) stopped on the host exec hang (results.md §8). The host
is healthy again: at 2026-09-20T17:24Z a fresh shebang script and a fresh `go test` binary
executed normally (0.6 s). The Story workspace still holds your uncommitted work — do NOT
recreate it, do NOT `git checkout`/`git clean`/stash anything; continue from the tree as it is
(`git status` first and confirm the 8 paths from results.md §9 are present).

Finish exactly what §3/§4/§8 of your results.md left compile-only or predicted:
1. Run the blocked rows on this host: `go test ./internal/crossconformance/ -run
   'TestDraftLiteral|TestDraftSSH' -count=1 -v`, `go test ./cmd/curator/ -run
   TestDraftDocsPinExamples -count=1`, `go test ./internal/gitops/ -count=1`, plus
   `TestDraftTransportLegacyGolden`, `TestProjectResolveLegacyUntouched`,
   `TestDraftLiteralRefreshIgnoresUserConfig`, `TestDraftSourcesSemanticCoverage` and the
   `insteadof` subtest. Fix what fails; record real exit codes.
2. Execute the §4 mutant table for real (each mutant applied, test run, reverted with the exact
   same bytes restored — verify with `git diff --stat` that only your intended paths remain
   changed), replacing "PREDICTED" with observed outcomes.
3. Update results.md (§3 verdicts, §4 observed, §8 closed with the healthy-host evidence: the
   commands and timings) and keep §1/§2 rulings and bounds as they stand.
4. Publish the Change Request and hand off (`task-board handoff TASK-260920-3ccq6b --role
   developer`) only when the configured gate is green; if the host stalls again, wait it out
   with bounded retries rather than blocking (the stall windows pass within minutes).
Rulings R1–R4 of 3ccq6b-brief.md remain in force; no new scope.
