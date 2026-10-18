# Round 3 validation evidence

Candidate: `3e00fcc8abe75e411a88773c65c4042ae2448eb0`, signature G.
Tree: `2fe068da665b203ed48e52cfcb7fba74de96e37e`.
Story branch remains uncommitted at checkpoint
`5d8e6b07f7f485557d9bbb065c9692348ca065c2`; the detached signed
snapshot is only the exact gate candidate, not a producer commit on the Story.

| Validation | Real exit | Evidence / bound |
|---|---:|---|
| Full local package coverage merge | 0 | 85 served packages, 28 CLI/envprofile groups, 3,494 unique top-level test names |
| Production platform-case gate | 0 | 36 classified skips, no failing/unclassified cases |
| New takeover behavioral tests + list-only narrowing control | 0 | Child expected-red exit 1; both named list regressions fail on the mutant |
| Existing native narrowing controls | 0 | All twelve child gates exit 1 with their named behavioral failures |
| Lint, private task cache | 0 | 0 issues; shared-cache failed attempt retained separately |
| go vet | 0 | Complete retained log, no diagnostics |
| go build | 0 | Complete retained log, binary built |
| git diff --check | 0 | No diagnostics |
| gofmt -l on the six rework files | 0 | Empty output |
| Integrity / signature / remote head / fresh trunk overlap | 0 | Worktree tree equals candidate; G; no path overlap; root README unchanged |
| Missing-group merger control, expected failure | 1 | Missing group evidence rejected; not a passing suite |
| Redaction control | 0 | Five planted private-path/token/email shapes redacted, SHA retained |

Fresh replay versus evidence reuse: all fifteen CLI groups and six envprofile
groups ran on the current identity. Seven envprofile groups reuse green exact
evidence from the preceding snapshot because internal source, dependencies,
config and environment are identical; they do not run the changed CLI mutation
child or source scanner. Both sensitive groups were freshly executed. Eighty-three
unchanged unrelated packages reuse the round-2 full stream; the current merged
stream verifies every package and every requested test name. The raw ledger
records actual Go commands and exits, plus the prior/current identity boundary.

Historical red diagnostic attempts remain red: shared lint-cache contamination;
old activation expectations; duplicate-guard survivor; and non-applied mutation
anchor. None is counted as a successful mutant kill or hosted pass. All were
resolved and the affected suites executed green afterward.

Rework versus the accepted round-2 signed snapshot is six files under
`cmd/curator/`, 211 insertions and 10 deletions. Rows 1–11 and the existing
reviewer tests remain present. The coverage map reports 12 of 12 surface rows
and 5 of 5 AC rows, with explicit bounds for help controls and no excluded
registered command. Muse/opencode/managed-home layout and root README are unchanged.

Hosted gate: **green on attempt 2**, exact candidate above.
`gh run watch 39117060646 --exit-status` exited **0**. Provider status reports
20 successful jobs and two conditional skips.
https://github.com/relux-works/curator/actions/runs/39117060646

Attempt 1 exited **1**: eighteen successful jobs, two Complete runner failures
after successful repository steps. Logs prove cleanup Permission denied in a
shared Linux temporary fixture store. In-progress retry requests exited **1**;
the accepted `gh run rerun 39117060646 --failed` exited **0**. Attempt 2 repeated
only those two jobs, both passing including cleanup. Queue diagnostics showed
four matching online/busy Linux runners. No hosted status was fabricated or
overridden, and no local mirror substituted for the required gate.

The hosted Windows artifact was downloaded and checked (exit **0**): all 67
unrelated registry commands, all four new tests, all twelve retained narrowing
controls, and TestNativeSkillsCLIOperations passed. Artifact/source identity and
stream hashes are recorded in r3-hosted-windows-summary.json.

All 3,490 prior top-level names are present; the only additions are the four
new round-3 tests. Fresh final integrity is green after the hosted result.
The worktree remains uncommitted for the native handoff snapshot.

Evidence packaging: the board refused binary archive upload with
`resource_binary_refused` (CLI exit 1). The same 121 sanitized payloads are
attached as `BUG-261018-uoisfe_validation-logs-r3.txt`, with a per-file hash
manifest and bundle index. Artifact validation exited 0: every payload hash
matched, and public notes passed the private-path scan. The private ignored
archive remains available at the project-relative path recorded in the index;
no binary-policy bypass or base64 substitute was used.
