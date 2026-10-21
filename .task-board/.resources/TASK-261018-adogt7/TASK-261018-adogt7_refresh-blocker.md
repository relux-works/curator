# TASK-261018-adogt7: refresh blocked by external hosted cleanup

The CHANGELOG combination is corrected and every local required check is green. The only repository edit in this round removes the duplicate rc.7 heading, placing the existing entry in Unreleased without changing trunk release bytes.

Candidate: `e57894d642993b9fce7ea23f7ccfdb1995c2ac38`, signature G; tree `cd1a903247725014af5726b39bee9b1e90482959`; uncommitted Story checkpoint `9c1605a74436d2448f759bd4b66ae6685237d83b`.

| Hosted evidence | Initial job 117654505603 | Retry job 117666342555 |
| --- | --- | --- |
| Repository race/property step | exit 0 | exit 0 |
| Platform-case gate | exit 0 | exit 0 |
| Administrator `Complete runner` cleanup | exit 1, Permission denied | exit 1, same persisted protected pnpm tree |
| `gh run watch 39120346079 --exit-status` | exit 1 | exit 1 |

[Real hosted workflow](https://github.com/relux-works/curator/actions/runs/39120346079): final failure, 19 successful jobs, one infrastructure failure, two declared policy skips, zero pending jobs. `gh run rerun 39120346079 --failed` returned 0 and ran the real replacement job; it did not resolve the external constraint.

The configured job-completed hook belongs to the runner administrator, outside repository workflow steps. This role cannot repair its host permissions or stale runner-owned temporary tree. No repository implementation, CI policy, gate, assertion or test permission was weakened. No local mirror substitutes for a hosted result.

Recommended action: runner administrator repairs cleanup and the stale temporary tree, then reruns the failed job on the exact retained candidate. An alternative is to quarantine the unhealthy runner and assign a healthy equivalent runner; the original runner still needs repair. Required external input is that repair/assignment and a legitimate green hosted gate. Repeated blind retries on the unchanged unhealthy runner add no evidence.

Local receipt: `TASK-261018-adogt7_refresh-validation.json`; actual path-redacted Go events: `TASK-261018-adogt7_refresh-local-events.jsonl`; hosted failure/attempt metadata: `TASK-261018-adogt7_refresh-hosted-gate.json`. Full local coverage: 85 packages, CLI 425/425 pass, install 371/371 and envprofile 372/372 observed with classified skips; 10/10 narrowing mutants killed. AC coverage is 9/9 rows driven, but AC 9 hosted-green acceptance remains unsatisfied. No AC row is declared outside contract; missing brief surface table remains an explicit input gap.

Do not hand off to review while hosted green is false. The candidate remains uncommitted and the real index / root README remain unchanged. The gate branch is retained for administrator recovery. No local process or hosted job is left running.
