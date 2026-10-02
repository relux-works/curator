# BUG-261002-ot3ea1 — global-upgrade-sweeps-in-use-builds

Verdict: accepted, revision 5. No blocking findings. Reviewer run RUN-261002-e2f897.
Base c085b4d22a0b6a51e3070277e7f3cffeb84e67e7; candidate tree f7dda90c9112460dcc30b1dcf573277fac8366db. Working tree matches the candidate before and after checks (git diff --quiet f7dda90: exit 0). No repository edits, commits, or LOGBOOK writes. Mutants used temporary Go overlays outside the repository.

## Reviewed surfaces

| Surface | Finding and evidence |
| --- | --- |
| Production wiring | global upgrade uses install commit maintenance: internal/install/commit.go:146,748,758 -> scopes.Collect at internal/scopes/gc.go:117 -> Store.Sweep. CLI gc uses cmd/curator/main.go:2290,2313,2328. All share the native process dependency by default. |
| Protected namespaces and deletion | internal/buildcache/collect.go:97-109 takes one process snapshot and skips all protected-build deletion with a warning on enumeration/normalization failure. Lines 171 and 176 protect both regular entries and retired-tree cleanup in both namespaces. Existing reference, journal, grace and protected mutation checks remain intact. |
| Image matching | Absolute paths and canonical symlink aliases; directory separator boundary avoids sibling-prefix matches; Windows case-insensitive comparison. Any image under a build counts, independent of command registration or daemon name. |
| macOS | sysctl process enumeration and saved exec path from kern.procargs2, independent of argv[0]. Read uncertainty fails safe; exit/zombie exclusions require OS evidence. Native current-process query and malformed saved-argument cases pass locally. |
| Linux | /proc enumeration, positive numeric PID validation, exe link and stat fallback. Unreadable/malformed observations fail safe. State-read guard has a function-specific OS-state reason. G304 annotation is limited to the stat read with numeric-PID justification; no blanket nolint. Linux lint exits 0. |
| Windows | Toolhelp snapshot and QueryFullProcessImageName with limited-query access. Any failure retains all protected builds. Local cross-vet exits 0; native tests use hosted evidence. |
| Tests and daemon behavior | 14/14 matrix rows pass (7 scenarios x 2 namespaces), including in-use, other image beneath build, unused, sibling, error/warning, daemon and unused daemon. No preexisting daemon-only rule was removed. Partial-table and symlink tests pass; scopes collector suite passes. |
| Hygiene and integration | All 17 changed paths reviewed. CHANGELOG.md and LOGBOOK.md are byte-identical to base (exit 0), as rev5 explicitly requires. git diff --check exits 0. Platform ledger additions coexist with base rows; hosted gate self-tests pass. |

Design: retain live process builds independently of age, keeping existing publication grace; no extra previous-build grace. This directly covers long-lived processes without a time cutoff.

## Independently executed validation

Every Go command used -work (also GOFLAGS=-work for nested Go commands). Every validation command acquired ~/.mini-build-lock by atomic directory creation and released only its own lock. Syspolicyd crash count rose from 368 to 369 during initial read-only inspection; waited over five minutes before first Go check. It was running at 369 before and after every validation below. A lock-busy attempt exited 75 before launching Go; it is not a test result.

| Command / log | Actual exit | Duration seconds |
| --- | --- | --- |
| go test -work ./internal/buildcache ./internal/scopes -count=1 -timeout=3m (packages.log) | 0 | 13.22 |
| go test -work ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$' -count=1 -v -timeout=3m (audit.log) | 0 | 8.58 |
| go test -work ./cmd/curator -run '^TestGC' -count=1 -v -timeout=3m (gc.log) | 0 | 50.97 |
| Drop liveness matcher with Go overlay; TestSweepLiveProcessBuilds (mutant-drop.log) | 1 expected | 3.13 |
| Narrow matcher to tb-sessiond only with Go overlay; same test (mutant-daemon.log) | 1 expected | 2.24 |
| GOOS=windows go vet -work ./internal/buildcache ./internal/scopes ./internal/envprofile ./cmd/curator (windows-vet.log) | 0 | 0.99 |
| GOOS=linux golangci-lint run ./internal/buildcache/... --timeout=3m (linux-lint.log) | 0, zero issues | 0.94 |
| Baseline liveness matrix, partial table, symlink and native Darwin query, -count=1 -v (baseline.log) | 0 | 1.91 |

Audit coverage is 476/476 readers (369 seam, 107 reviewed allowlist; 447 production files scanned). CLI GC mask passes 6/6 tests. Drop mutant fails 6/14 rows: in-use, other-file and daemon in each namespace. Daemon-only mutant fails 4/14 rows: ordinary in-use and other-file in each namespace; daemon rows pass. Both mutants killed (2/2), with baseline 14/14 green. Raw logs, exact overlay payloads and check runner are attached as BUG-261002-ot3ea1_review-validation-rev5.tar.gz.

## Reused evidence and bounds

Read attached BUG-261002-ot3ea1_change-request_rev5-validation.log: remote-gate run 36999725635 reports success, command exit 0, 1/1 required command shard green, including native Linux/macOS/Windows test lanes, race lanes, lint and three gate self-tests. This is accepted hosted evidence, not rerun by this reviewer. The log declares test_case_coverage=unknown; no full-matrix case ratio is inferred. Full CLI/envprofile suites and native Linux/Windows executions were not repeated locally.

The real cached-child CLI test passed through fail-safe retention: macOS returned invalid argument for one process image. Complete successful native process-table enumeration on this host remains unknown; the deterministic seam proves selective retention/removal. A single snapshot does not observe processes started after it. This change covers protected go-v1 and go-v1-receipt-3 cache sweep; separate external buildrepo artifact GC is unchanged. Existing root-dependent conformance cases need their external fixture; hosted evidence supplies the conformance lane.

No active goal is bound to this run (spawn goal queried immediately before verdict). No directives recorded. Acceptance routes to integrating via accept_cr; producer-side integration remains outstanding. LOGBOOK checklist is not applicable under the explicit No LOGBOOK instruction; findings are recorded here.
