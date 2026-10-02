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


## Revision 3 (re-apply)

RUN-261002-2bca1f re-applies revision 2 onto trunk c085b4d22a0b6a51e3070277e7f3cffeb84e67e7. Repository changes remain uncommitted and are ready for review.

The instructed `git diff 2cb29dac d7df3b8f -- . ':!.task-board' | git apply --3way` exited 1 because the platform ledger conflicted. CHANGELOG applied cleanly. Resolved the ledger by preserving every trunk byte, including its two Windows executable-origin rows, and appending all revision-2 rows. CHANGELOG retains both sides and adds exactly one build-cache behavior line.

Recovery anomaly: the supplied d7df3b8f stash commit contains only the nine previously tracked paths and omits all nine new process implementation/test files. Restored those exact files from the attached BUG-261002-ot3ea1_change-request_rev2.patch. All paths other than the ledger and CHANGELOG equal revision 2 byte-for-byte: **16/16** (rev2-equality.json). There are exactly 18 changed repository paths. LOGBOOK was not edited, per the explicit binding instruction; findings are recorded in these results and board notes.

The revision-2 design remains: Store.Sweep retains any executable beneath a build directory in both namespaces, including daemon images and retired-tree cleanup; enumeration errors retain builds with a warning. Native macOS sysctl, Linux procfs and Windows process-image enumeration are restored unchanged. Existing publication grace is unchanged; no extra previous-build grace was added. The Linux reader keeps the precise OS-owned-state audit allowlist reason and its G304-only numeric-PID annotation. Production reaches the sweep through global upgrade's collectAfterCommit and the standalone gc entry point.

### Checks rerun by this run

Every command below ran here as a standalone child process, without tee or a pipeline. Go/test/build/lint/vet commands were serialized through ~/.mini-build-lock, with **GOFLAGS=-work**. WORK directories and child command output are recorded in the logs. The syspolicy service was running before and after every command; observed successive crashes remained **361 → 361** throughout. Durations exclude build-lock acquisition. No prior green result substitutes for these reruns. The git-only diff check does not acquire the build lock.

| Command | Real exit | Duration | Successive crashes | Evidence |
| --- | --- | --- | --- | --- |
| `go test ./internal/buildcache -count=1 -v -timeout=3m` | 0 | 16.495s | 361 → 361 | buildcache.log |
| `go test ./internal/envprofile -run TestManagerOwnedAbsenceReadsAreGuarded -count=1 -v -timeout=3m` | 0 | 11.627s | 361 → 361 | state-read-audit.log |
| `go test ./internal/scopes -count=1 -v -timeout=3m` | 0 | 1.913s | 361 → 361 | scopes.log |
| `go test ./cmd/curator -run ^TestGC -count=1 -v -timeout=4m` | 0 | 63.188s | 361 → 361 | cli-gc.log |
| `go test ./internal/buildcache -run ^TestSweepLiveProcessBuilds$ -count=1 -v -timeout=2m` | 1 | 2.056s | 361 → 361 | mutant-delete.log |
| `go test ./internal/buildcache -run ^TestSweepLiveProcessBuilds$ -count=1 -v -timeout=2m` | 1 | 2.793s | 361 → 361 | mutant-daemon-only.log |
| `go test ./internal/buildcache -count=1 -v -timeout=3m` | 0 | 12.411s | 361 → 361 | restored-buildcache.log |
| `golangci-lint run ./internal/buildcache/... ./internal/scopes/... ./internal/envprofile/... ./cmd/curator/... --timeout=4m` | 0 | 6.708s | 361 → 361 | native-lint.log |
| `GOOS=linux golangci-lint run ./internal/buildcache/... ./internal/scopes/... ./internal/envprofile/... ./cmd/curator/... --timeout=4m` | 0 | 12.744s | 361 → 361 | linux-lint.log |
| `env GOOS=windows go vet ./internal/buildcache ./internal/scopes ./internal/envprofile ./cmd/curator` | 0 | 2.172s | 361 → 361 | windows-vet.log |
| `bash .github/ci/ledger-consistency.sh .temp/BUG-261002-ot3ea1/revision3/ledger` | 0 | 11.841s | 361 → 361 | ledger-consistency.log |
| `go build -o .temp/BUG-261002-ot3ea1/revision3/curator ./cmd/curator` | 0 | 2.364s | 361 → 361 | build.log |
| `git diff HEAD --check` | 0 | 0.064s | 361 → 361 | diff-check.log |

Mutants are **failing expected-red gates**, not passing suites. Deleting the liveness check failed 6/6 in-use/other-file/daemon rows. Narrowing it to tb-sessiond failed 4/4 ordinary runner/other-file rows while daemon rows passed. The harness restored the original bytes in finally; standalone cmp exited 0 and the final comparison confirms exact restoration. The restored full buildcache suite exited 0.

The liveness matrix passed **14/14** rows; scopes collector rows **2/2**; CLI GC tests **6/6**; manager-read audit **476/476** readers. Both lint runs report 0 issues. The ledger subset verifies **503** rows across linux/darwin/windows, including trunk and restored rows. Native build and Windows vet validate compilation of touched packages.

Diagnostic: the initial lock-wrapped git diff check attempt exited 75 after its bounded lock wait and did not launch git. Evidence preparation then exited 1 because that unexecuted check had no metadata. Ran git diff HEAD --check directly as a standalone process: exit 0, recorded above. No unexecuted or failing gate is reported as passing.

Bounds: the native CLI live-child GC test encountered `invalid argument` reading one macOS process image, proving fail-safe retention through the entry point rather than a successful full native process snapshot. Deterministic process-table tests prove selective retention/removal. The snapshot does not observe processes started later. Four buildcache and one scopes conformance cases skipped because CURATOR_CONFORMANCE_ROOT is unset. Full CLI/envprofile suites, native Linux/Windows runtime tests and the hosted matrix were not rerun manually in this revision-3 developer session; the targeted GC/source-audit tests, native/Linux lint and Windows vet provide the stated evidence. No full conformance or foreign-platform native qualification is claimed.

The attached archive contains raw logs, exit-code/duration/host metadata, validation and mutant harnesses, equality evidence, source hashes, the candidate patch and ledger reports. The text above this revision-3 section remains the historical revision-2 outcome. The conditional rejected-review branch has not occurred for revision 3 and remains for review routing.

Post-validation host anomaly: after the last verification and after both outcomes were attached, new shell launches stalled for several minutes. Existing pending commands were awaited without restarting tests. The service check eventually returned running with successive crashes=362, increased from the checks' stable 361. The pending board-notes mutation then exited 0. No local Go build/test/vet/run was started after this increase; the previously recorded green commands and their 361 → 361 observations are unchanged. This post-validation observation is preserved in post-validation-host.log.
