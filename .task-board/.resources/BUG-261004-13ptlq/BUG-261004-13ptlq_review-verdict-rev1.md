# BUG-261004-13ptlq — unmanage cannot restore its own symlink backup

Verdict: accepted. CR-BUG-261004-13ptlq-1 revision 1.
Reviewed base e5489b6ba22c9e9cf7a925e03f3653d115188999 and candidate tree 46dfb0de9cdfa3bba1d92279edd36c02a5dc289a. No blocking findings. No code, LOGBOOK.md or CHANGELOG.md edits by reviewer.

## Swept surfaces
| Surface | Review evidence | Result |
| --- | --- | --- |
| Backup entry decoding | Lstat type retained; Readlink records literal text; regular bytes/mode unchanged; unsupported entries refused | Accept |
| Production restore | run -> cmdEnvUnmanage -> Unmanage -> readBackupTree/preflightUnmanagePlan -> applyUnmanagePlan -> atomicManagedLink; existing private staging and entry rename reused | Accept |
| Parent routes | Existing parents inspected before writes; every planned ancestor checked against restored links; removals use removeManagedEntry | Accept |
| Entry and target integrity | 4/4 absolute, relative, dangling and directory-target CLI round trips; 1/1 destination-link replacement case | Accept |
| Refusal coverage | 2/2 existing-parent routes and 2/2 planned-parent routes, including deeper ancestors; native surface and marker unchanged on refusal | Accept |
| N3 compatibility | Existing 0600, 0644, 0751 and 0400 mode round trips rerun; N3 505e1526 ancestor check exit 0 | Accept |
| Gate script | One-line removal of workspace path from public snapshot message; validation semantics unchanged | Accept |

## Independently executed
GOFLAGS=-work go test ./internal/envprofile ./cmd/curator -run '^Test(Unmanage|EnvUnmanage)' -count=1 -timeout=4m: exit 0. Internal package 0.591s, CLI package 55.360s. Existing TestMain host-GOROOT build lock retained. git diff --check: exit 0. Exact byte comparison of all four changed files against candidate blobs: exit 0, 4/4 match.

Fresh origin main advertised and fetched as 54bed271b7609bf206a04369202473c430d0d96a. Upstream changes since CR base affect only internal/audit/audit.go, internal/audit/audit_test.go and internal/install/draftaudit_test.go; no overlap with candidate paths. Integration still owns combined-tree validation and landing.

## Reused evidence, inspected rather than rerun
BUG-261004-13ptlq_results.md and producer spawn log RUN-261004-47061c establish red-first production CLI failure: all 4/4 link shapes failed with environment_backup_record_unreadable, backup entry is not a regular file; process exit 1. The revised-source follow-link mutant exited 1: internal test observed changed destination target bytes, and 4/4 CLI cases rejected a regular-file result. Narrowed existing/planned-parent mutants exited 1; deeper-route tests detected mutation before refusal. Reviewer inspected producer log failure output; reviewer did not modify production code or replay mutants. Producer local lint/build results reused, additionally covered by exact-tree hosted validation.

Authoritative hosted CR receipt: BUG-261004-13ptlq_change-request_rev1-validation.log records sh scripts/remote-gate.sh exit 0, exact command shard coverage 1/1. Independently queried https://github.com/relux-works/curator/actions/runs/37230338773: success, snapshot 37b0152256573c0efc045701ffaee33dd1e84310. git diff --exit-code from snapshot to candidate tree exited 0. Required jobs green 11/11: Linux/macOS/Windows tests and gate self-tests, Linux/macOS race, lint, naming and interop. Optional jobs skipped 2/2, not counted as passes. Earlier standalone runs are not substituted for this CR receipt.

Bounds: new filesystem regressions are Unix-only; Windows existing suite and compilation passed. External target checks prove bytes, inode identity, mode and mtime, not access-time or syscall tracing. Atomic entry replacement relies on the existing managed writer; concurrent hostile filesystem races and Windows ACL recovery are not newly attested here.

Run goal queried: no active goal, run is not goal-bound. Acceptance is recorded with accept_cr revision 1 and routes to integrating; reviewer supplies no commit acknowledgement and does not mark done.
