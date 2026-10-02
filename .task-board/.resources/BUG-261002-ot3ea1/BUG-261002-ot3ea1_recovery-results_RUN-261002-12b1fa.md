# BUG-261002-ot3ea1 — global-upgrade-sweeps-in-use-builds

Ready for review. Recovery run RUN-261002-12b1fa retains the existing uncommitted implementation and repairs the remote validation integration failures. No commits, branch changes or LOGBOOK writes were made.

## Result and design

`internal/buildcache.Store.Sweep` retains a build when any executable in the process snapshot lies beneath its directory, including canonical symlink aliases and Windows case-insensitive paths. Both go-v1 and go-v1-receipt-3 namespaces, including retired-tree cleanup, use the rule. Enumeration or normalization errors retain all protected builds and emit a warning. Existing marker/journal retention, publication grace and protected deletion boundaries stay intact.

Native enumeration uses macOS sysctl saved executable paths, Linux /proc/PID/exe, and Windows Toolhelp plus QueryFullProcessImageName. The injected Store.ExecutablePaths seam supplies deterministic process observations in tests. The chosen design is liveness retention with existing publication grace; no extra previous-build grace was added because process lifetimes are not age-bounded. There was no daemon-specific liveness rule in the original sweep; daemon retention now uses the same generic rule.

Production call sites: global upgrade → install.Global → runCommit → collectAfterCommit → scopes.Collect → Store.Sweep. Standalone gc → collectUnderLock → scopes.Collect → Store.Sweep is exercised by a real cached child executable.

## Recovery findings and fixes

The previous remote gate (`sh scripts/remote-gate.sh`, exit 1, run 36967359936) failed the manager-read audit on all five test/race lanes, the unrecognized Windows-only skip on Unix lanes, and Linux lint. Raw remote evidence was retrieved and inspected; no other named test failed in its five streams (the Ubuntu race merged stream contains three non-JSON lines, so its overall completeness remains bounded by that observation).

- Documented the OS-owned Linux process-table reader in the audit's function-specific allowlist. Missing images still require stat evidence of exit/zombie/kernel state, and read errors still retain builds. This is outside manager-owned state, so the manager-state reader seam is inappropriate.
- Moved the Windows case comparison test to process_windows_test.go to compile only on Windows, removing its artificial Unix skip.
- Added a G304-only annotation to the fixed proc-root/numeric-PID stat read, with a recorded reason. No broad security exemption was added.
- Registered all new platform-specific and liveness cases in the platform ledger. The CHANGELOG remains one behavior line.

## Verification executed by this recovery run

All Go, lint and build commands ran as standalone child processes under `bash .temp/BUG-261002-ot3ea1/with-build-lock.sh`; output redirection preserved their real exit codes. No tee pipelines were used. The host syspolicy service was running at all observed checkpoints (successive crashes=358).

| Command | Real exit | Evidence |
| --- | --- | --- |
| go test ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$' -count=1 -v -timeout=3m | 0 | recovery-audit.log: 476/476 readers covered |
| go test ./internal/buildcache ./internal/scopes -count=1 -v -timeout=3m | 0 | recovery-packages.log: full two-package suites |
| go test ./cmd/curator -run '^TestGC' -count=1 -v -timeout=4m | 0 | recovery-cli.log: 6/6 GC tests |
| GOOS=windows go vet ./internal/buildcache ./internal/scopes ./internal/envprofile ./cmd/curator | 0 | recovery-windows-vet.log |
| GOOS=linux golangci-lint run ./internal/buildcache/... ./internal/scopes/... ./internal/envprofile/... ./cmd/curator/... --timeout=4m | 0 | recovery-linux-lint.log: 0 issues |
| golangci-lint run ./internal/buildcache/... ./internal/scopes/... ./internal/envprofile/... ./cmd/curator/... --timeout=4m | 0 | recovery-native-lint.log: 0 issues |
| bash .github/ci/ledger-consistency.sh .temp/BUG-261002-ot3ea1/recovery-ledger | 0 | recovery-ledger.log: 501 rows across Linux, Darwin, Windows |
| go build -o .temp/BUG-261002-ot3ea1/recovery-curator ./cmd/curator | 0 | recovery-build.log |
| bash .github/ci/no-broad-suppression.sh | 0 | recovery-suppression.log |
| git diff --check | 0 | recovery-diff-check.log |
| cmp internal/buildcache/collect.go .temp/BUG-261002-ot3ea1/recovery-collect.original | 0 | exact mutant restoration |
| go test ./internal/buildcache -run '^TestSweepLiveProcessBuilds$' -count=1 -v -timeout=2m after restoration | 0 | recovery-restored.log: 14/14 rows |

The matrix covers in-use, another executable file beneath the build, unused, sibling-prefix, enumeration error, daemon and unused-daemon across both namespaces (14/14). The real scopes collector covers retention/removal and enumeration-error warning propagation (2/2).

Mutants were rerun here, not accepted from earlier evidence. Both execute `go test ./internal/buildcache -run '^TestSweepLiveProcessBuilds$' -count=1 -v -timeout=2m` directly. Removing the retention guard exits **1**, with 6 in-use/other-file/daemon rows failing. Narrowing it to daemon image paths exits **1**, with 4 ordinary runner/other-file rows failing while daemon rows pass. These are expected-red failing gates, not passing suites. recovery-mutants.py restores the saved bytes in finally; its harness exits 0 only after observing the expected failures and exact restoration. Raw logs and exit-code JSON are attached.

Diagnostic: invoking the non-executable lock wrapper directly exited 126 before running go test. The audit command was then run through bash and exited 0. No expected failure is reported as a passing test.

## Bounds and remaining validation

No prior green result is substituted for the commands above. Historical implementation notes remain in the earlier developer outcome. Full CLI and envprofile suites were not rerun: the CLI mask exercises the changed GC entry point, and the envprofile mask exercises the changed source audit. Five root-dependent conformance cases skipped because CURATOR_CONFORMANCE_ROOT is unset. Native Linux/Windows tests were not run on this macOS host; Windows vet and Linux lint compile those paths.

The native CLI test exercised fail-safe retention: macOS returned invalid argument for one process image, so successful full native enumeration is unknown. Deterministic process-table tests prove selective retention and unused-build removal. One process snapshot per sweep does not observe processes started afterwards. External buildrepo artifact GC remains outside this change.

The full remote matrix has not been rerun manually in this recovery session; it is the configured Change Request validation triggered by the role handoff. Its earlier exit 1 is preserved, and no green remote qualification is claimed. All local evidence is attached before handoff. The explicit sweep-brief.md instruction No LOGBOOK makes checklist item 7 not applicable; findings are in these outcomes and board notes.
