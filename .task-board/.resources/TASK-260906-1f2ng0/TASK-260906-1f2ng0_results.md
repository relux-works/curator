# TASK-260906-1f2ng0 results

Developer handoff evidence for `STORY-260905-1n0iy8`. Changes remain uncommitted in the assigned worktree.

## Follow-ups

- **FU-1:** Path installs distinguish `profile_source_path_missing`, `profile_source_path_unreadable`, and `profile_source_invalid`. The production CLI test covers a missing operand, unreadable root, unreadable module directory, and regular file. The verbose run executed and passed all four rows; neither unreadable row skipped.
- **FU-2:** `loadMachinePolicy` uses `Lstat` to select the default policy only for a genuinely absent path. An unreadable config and a dangling config symlink are refused. Tests cover the absent policy through `List` and the loader, and both read-failure shapes.
- **FU-3:** Machine-scope `Use` gathers every scope record equal to the effective profile and passes those removals and the current-file write to one `op.publish` call, producing one transaction plan. A production `Use` test scopes two adapters to the target profile and confirms both records clear while an unrelated `pi=gamma` record remains. A mutant that cleared only the first matching record failed.
- **FU-4:** F16 fixtures now scope both `codex_cli` and `opencode`. A skip-at-most-one mutant failed at the production `Use` test because `opencode` appeared in the machine results.
- **FU-5 (cycle 6):** When every adapter has a scoped current, `profile use` reports `No adapter homes were switched; every registered adapter had a scoped current.` The CLI test confirms success, the message, and preserved scope rows. Disabling the message made its targeted test fail.
- **Cycle 5 partial activation:** The CLI reports a persisted first install when activation is refused before adapter materialization, and preserves `updated profile` wording on partial reinstall. The tests exercise both paths. Requiring a nonempty activation result as a prerequisite for the first-install line made its targeted test fail.
- **Cycle 2 global-skill warning — explicitly declined:** The reviewed branch/local skill omission remains a documented bound. Environments §9.4 defines migrated direct declarations as exact tag or revision pins and specifies `environment_import_lossy` for an unreadable install record; it gives no warning diagnostic or output text for a readable but unrepresentable branch/local declaration. The cycle-2 reviewer called this nonblocking and placed a warning in the later stage that wires live declarations. Per the brief’s explicit decline option, no new warning contract was invented here. The existing package comment records that bound. This disposition is also in the task board notes.

## Validation

Commands were run directly in the assigned worktree. No full landing suite was run manually; handoff runs it once.

| Command | Result |
| --- | --- |
| `set -o pipefail; go test -count=1 ./internal/envprofile -run '^(TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault|TestLoadMachinePolicyRejectsDanglingConfigSymlink|TestMachinePolicyLoadFailureFailsMigration|TestMachineUseSkipsScopedAdapter|TestMachineUseClearsEveryScopeEqualToNewDefault|TestInstallUseSkipsScopedAdapter|TestScopedUseStillSwitchesOnlyThatHome|TestSyncWritesScopedHomeOnce)$'` | exit 0; `ok`, 114.382s |
| `set -o pipefail; go test -count=1 ./cmd/curator -run '^(TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable|TestProfileMachineUseSkipsScopedAdapter|TestProfileUseReportsAllAdaptersAlreadyScoped|TestProfileInstallUsePartialLeavesCurrent|TestProfileInstallFirstActivationFailureReportsPersistedProfile|TestProfileReinstallActivationFailureKeepsUpdatedWording)$'` | exit 0; `ok`, 344.172s. The package process queued behind the shared macOS host-toolchain lock before running. |
| `set -o pipefail; go test -count=1 -v ./cmd/curator -run '^(TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable|TestProfileUseReportsAllAdaptersAlreadyScoped|TestProfileInstallFirstActivationFailureReportsPersistedProfile)$'` with both named CLI narrowing mutants active | exit 1 as expected. Both mutant assertions failed; all four path rows passed, including the two unreadable rows. |
| `go build ./...` | exit 0 |
| `golangci-lint run` | exit 0; 0 issues. It emitted a non-fatal generated-file-filter warning for a deleted concurrent sibling worktree path. |
| `gofmt -d cmd/curator/profile.go cmd/curator/profile_test.go internal/envprofile/envprofile.go internal/envprofile/envprofile_f10f11f12_test.go internal/envprofile/envprofile_f16_test.go internal/envprofile/lock.go internal/envprofile/switch.go` | exit 0; no output |
| `git diff --check` | exit 0 |

Additional narrowing mutants, each expected to fail with exit 1, were run and restored before the green reruns: path permission classification disabled (the unreadable-root row failed); unreadable machine config treated as absent (policy test failed); only the first equal-default scope cleared (`opencode=beta` remained); and at most one scoped adapter skipped (the production `Use` test reported `opencode`). The cycle-5 partial-install and cycle-6 all-scoped output mutants are covered by the combined CLI mutant command above, with their individual targeted test failures visible in its output.

Important decision and the shared-lock test delay were recorded in task-board notes. Campaign rules prohibit editing `LOGBOOK.md`.

## Revision 2 — Windows platform-case gate rework

The revision-1 Windows gate rejected mode-000 skip wording that was not in
`skip-classes.tsv`. The path-read probes now use a Windows DACL that denies
`GENERIC_READ`, while Unix keeps mode-000 probes and emits the registered
`host-capability` reason only when the runner can still read the path. The
absent-config and unreadable-config checks are separate subtests. Six rows in
`platform-cases.tsv` require the missing, unreadable, regular-file, absent
config, and unreadable config cases on Linux, macOS, and Windows. The two
permission-capability rows allow a skip on Linux/macOS only; Windows must run
the DACL cases.

Windows runtime execution is unverified on this Darwin host. Both Windows test
binaries cross-compiled successfully, and `ledger-consistency.sh` confirms the
named parent tests are present in the Windows build. The ledger requires their
Windows cases to pass without a skip in the hosted candidate lane.

### Revision 2 validation

Commands ran directly in the assigned worktree. The full landing suite was
not run manually; handoff runs it once.

| Command | Result |
| --- | --- |
| `GOOS=windows GOARCH=amd64 go test -c -o .temp/TASK-260906-1f2ng0-windows-build/cmd-curator.test.exe ./cmd/curator` | exit 0; Windows test binary compiled |
| `GOOS=windows GOARCH=amd64 go test -c -o .temp/TASK-260906-1f2ng0-windows-build/envprofile.test.exe ./internal/envprofile` | exit 0; Windows test binary compiled |
| `go test -count=1 ./internal/envprofile` | exit 0; `ok`, 339.181s |
| `go test -count=1 ./cmd/curator` | exit 1; test timeout after 10m0s (`600.746s`), with `TestEnvStatusUnreadableApprovalStateSurfaced` still active. This is a broad package-suite timeout; the focused changed-scope run below passed. |
| `go test -count=1 -timeout=8m ./cmd/curator -run '^(TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable|TestProfileMachineUseSkipsScopedAdapter|TestProfileUseReportsAllAdaptersAlreadyScoped|TestProfileInstallUsePartialLeavesCurrent|TestProfileInstallFirstActivationFailureReportsPersistedProfile|TestProfileReinstallActivationFailureKeepsUpdatedWording)$'` | exit 0; `ok`, 24.461s |
| `go test -count=1 -v ./cmd/curator -run '^TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable$'` | exit 0; all four rows passed, including both unreadable cases without skips |
| `go test -count=1 -v ./internal/envprofile -run '^TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault$'` | exit 0; absent and unreadable subtests passed |
| `sh .github/ci/gate-selftest.sh` | exit 0; 185 passed, 0 failed |
| `sh .github/ci/ledger-consistency.sh .temp/TASK-260906-1f2ng0-ledger-evidence` | exit 0; 247 rows checked across linux, darwin, windows, including all six new platform cases |
| `sh .github/ci/no-broad-suppression.sh` | exit 0; `ok` |
| `go build ./cmd/curator ./internal/envprofile` | exit 0 |
| `golangci-lint run ./cmd/curator/... ./internal/envprofile/...` | exit 0; 0 issues |
| `gofmt -l cmd internal` | exit 0; no files listed |
| `git diff --check` | exit 0 |

The host Go tool reported `go1.26.0 darwin/amd64`; Windows was compile-checked but
not executed locally. The revision-1 narrowing-mutant results above remain
applicable: production behavior did not change in this revision, and the
affected assertions still pass in the focused runs. The platform gate now
requires Windows execution in hosted CI instead of accepting a Windows skip.

### Revision 2 handoff recheck

Reran the focused evidence and validation directly in the assigned worktree:

| Command | Result |
| --- | --- |
| `sh .github/ci/gate-selftest.sh` | exit 0; 185 passed, 0 failed |
| `sh .github/ci/no-broad-suppression.sh` | exit 0; `ok` |
| `sh .github/ci/ledger-consistency.sh .temp/TASK-260906-1f2ng0-ledger-recheck` | exit 0; 247 rows checked across linux, darwin, windows |
| `go test -count=1 ./internal/envprofile -run '^(TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault|TestLoadMachinePolicyRejectsDanglingConfigSymlink|TestMachineUseClearsEveryScopeEqualToNewDefault|TestMachineUseSkipsScopedAdapter|TestInstallUseSkipsScopedAdapter|TestScopedUseStillSwitchesOnlyThatHome|TestSyncWritesScopedHomeOnce)$'` | exit 0; `ok`, 39.084s |
| `go test -count=1 -timeout=8m -v ./cmd/curator -run '^(TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable|TestProfileMachineUseSkipsScopedAdapter|TestProfileUseReportsAllAdaptersAlreadyScoped|TestProfileInstallUsePartialLeavesCurrent|TestProfileInstallFirstActivationFailureReportsPersistedProfile|TestProfileReinstallActivationFailureKeepsUpdatedWording)$'` | exit 0; `ok`, 33.923s. All four path subtests passed on Darwin, including both unreadable cases without skips. |
| `GOOS=windows GOARCH=amd64 go test -c -o .temp/TASK-260906-1f2ng0-windows-build/cmd-curator.test.exe ./cmd/curator` | exit 0 |
| `GOOS=windows GOARCH=amd64 go test -c -o .temp/TASK-260906-1f2ng0-windows-build/envprofile.test.exe ./internal/envprofile` | exit 0 |
| `go build ./cmd/curator ./internal/envprofile` | exit 0 |
| `golangci-lint run ./cmd/curator/... ./internal/envprofile/...` | exit 0; 0 issues |
| `gofmt -l cmd internal` | exit 0; no files listed |
| `git diff --check` | exit 0 |

Windows test binaries compile locally, but DACL behavior remains runtime-unverified
on this Darwin host. The platform ledger requires these six cases to execute on
Windows without skips; the handoff candidate suite is the required runtime
verification. I did not run the full landing suite manually because the
campaign handoff runs it once. The earlier broad `go test -count=1 ./cmd/curator`
attempt remains recorded above as exit 1 after its 10-minute timeout; the
focused changed-scope tests passed here.

### Revision 3 — Windows unreadable cases are required and executable

For each Windows unreadable subtest, selected the preferred execute-on-Windows
disposition:

- `cmd/curator` `unreadable_root`
- `cmd/curator` `unreadable_module_directory`
- `internal/envprofile` `unreadable_config`

The existing godriver Windows helper does not use a DACL: it opens the path
with `CreateFile` and share mode 0. Its source explains that a DACL deny can be
bypassed for Go's backup-intent directory opens. I extracted that working
fixture as `internal/testsupport.DenyPathRead`, reused it from godriver and
both follow-up test packages, and made fixture setup fail (not skip) if the
exclusive handle does not block `ReadDir`/`ReadFile`. The ledger still requires
all three cases on Windows; its host-capability skips remain limited to
Linux/Darwin mode-bit fixtures.

#### Revision 3 validation

Commands ran directly in the assigned worktree unless wrapped in Docker for
Linux execution. Windows test binaries compile here, but Windows runtime is
not available on this Darwin host; the hosted Windows test lane must execute
the three required cases. The Linux Docker runs dropped `DAC_OVERRIDE` and
`DAC_READ_SEARCH`, so mode-000 fixtures were actually unreadable, not skipped.

| Command | Result |
| --- | --- |
| `docker run --rm --cap-drop=DAC_OVERRIDE --cap-drop=DAC_READ_SEARCH --volume curator-task-260906-1f2ng0-go-cache:/cache --volume /Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260905-1n0iy8/worktree:/workspace --workdir /workspace --env GOMODCACHE=/cache/mod --env GOCACHE=/cache/build golang:1.26 go test -count=1 -timeout=8m -v ./internal/envprofile -run '^TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault$'` | exit 0; `absent_config` and `unreadable_config` both passed on Linux |
| `docker run --rm --cap-drop=DAC_OVERRIDE --cap-drop=DAC_READ_SEARCH --volume curator-task-260906-1f2ng0-go-cache:/cache --volume /Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260905-1n0iy8/worktree:/workspace --workdir /workspace --env GOMODCACHE=/cache/mod --env GOCACHE=/cache/build golang:1.26 go test -count=1 -timeout=8m -v ./cmd/curator -run '^TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable$'` | exit 0; missing, unreadable root, unreadable module directory, and regular file all passed on Linux, with no skips |
| `go test -count=1 -timeout=8m -v ./internal/envprofile -run '^TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault$'` | exit 0; `absent_config` and `unreadable_config` passed on Darwin |
| `go test -count=1 -timeout=8m -v ./cmd/curator -run '^TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable$'` | exit 0; all four path subtests passed on Darwin without skips |
| `go test -count=1 -timeout=8m ./internal/godriver -run '^TestFingerprintReportsUnreadableDirectoryIdentically$'` | exit 0; passed on Darwin |
| `GOOS=windows GOARCH=amd64 go test -c -o .temp/TASK-260906-1f2ng0-windows-build/cmd-curator.test.exe ./cmd/curator` | exit 0; compiled |
| `GOOS=windows GOARCH=amd64 go test -c -o .temp/TASK-260906-1f2ng0-windows-build/envprofile.test.exe ./internal/envprofile` | exit 0; compiled |
| `GOOS=windows GOARCH=amd64 go test -c -o .temp/TASK-260906-1f2ng0-windows-build/godriver.test.exe ./internal/godriver` | exit 0; compiled |
| `sh .github/ci/ledger-consistency.sh` | exit 2; this checkout requires an evidence-directory argument, so no checks ran |
| `sh .github/ci/ledger-consistency.sh .temp/TASK-260906-1f2ng0-ledger-revision3` | exit 0; 247 rows checked across Linux, Darwin, and Windows |
| `sh .github/ci/gate-selftest.sh` | exit 0; 185 passed, 0 failed |
| `golangci-lint run ./cmd/curator/... ./internal/envprofile/... ./internal/godriver/...` | exit 0; 0 issues |
| `GOOS=windows GOARCH=amd64 golangci-lint run ./cmd/curator/... ./internal/envprofile/... ./internal/godriver/...` | exit 1; four existing findings in `internal/godriver` (`CloseHandle` errcheck in `platform_windows.go` and `process_alive_windows_test.go`; two `unsafe.Pointer` gosec findings in `controls_windows.go`); none are in changed lines |
| `GOOS=windows GOARCH=amd64 golangci-lint run --new-from-rev=HEAD ./cmd/curator/... ./internal/envprofile/... ./internal/godriver/...` | exit 0; 0 new issues |
| `go build ./...` | exit 0 |
| `gofmt -l cmd internal` | exit 0; no files listed |
| `git diff --check` | exit 0 |

The Linux test cache volume was removed after the runs. No full landing suite
was run manually; Windows runtime remains for the hosted candidate lane.

### Revision 4 — Windows read-denial ACL fixture

Replaced the exclusive-handle fixture with a DACL deny for the current user:
`FILE_READ_DATA` for files and `FILE_LIST_DIRECTORY` for directories. The
shared Windows test helper captures the original DACL, applies the deny, and
registers cleanup that restores the original DACL and protected/unprotected
inheritance state before the temporary root is removed. The helper probes the
real read/list operation and accepts only `ERROR_ACCESS_DENIED`, so an unrelated
read failure cannot masquerade as an unreadable fixture. The production CLI,
machine-policy, and godriver unreadable-directory tests use this helper. The
platform ledger now describes those DACL controls instead of the exclusive
handle.

Windows runtime was unavailable on this Darwin host. The three Windows test
binaries compile, and the hosted Windows candidate lane remains responsible for
runtime verification. No Windows platform-control carve-out was added because
this revision has not established that the DACL fixture is unusable there.

#### Revision 4 validation

Commands were run directly in the assigned worktree. Darwin ran the changed
behavior cases; Windows was cross-compiled but not executed locally.

| Command | Result |
| --- | --- |
| Initial `GOOS=windows GOARCH=amd64 go test -c -o .temp/TASK-260906-1f2ng0-windows-revision4/cmd-curator.test.exe ./cmd/curator` | exit 1; two Windows API arguments needed their named `ACCESS_MASK` and `SECURITY_INFORMATION` types; fixed and reran below |
| `GOOS=windows GOARCH=amd64 go test -c -o .temp/TASK-260906-1f2ng0-windows-revision4/cmd-curator.test.exe ./cmd/curator` | exit 0 |
| `GOOS=windows GOARCH=amd64 go test -c -o .temp/TASK-260906-1f2ng0-windows-revision4/envprofile.test.exe ./internal/envprofile` | exit 0 |
| `GOOS=windows GOARCH=amd64 go test -c -o .temp/TASK-260906-1f2ng0-windows-revision4/godriver.test.exe ./internal/godriver` | exit 0 |
| `go test -count=1 -timeout=8m -v ./cmd/curator -run '^TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable$'` | exit 0; all four path cases passed on Darwin |
| `go test -count=1 -timeout=8m -v ./internal/envprofile -run '^TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault$'` | exit 0; absent and unreadable configuration subtests passed on Darwin |
| `go test -count=1 -timeout=8m -v ./internal/godriver -run '^TestFingerprintReportsUnreadableDirectoryIdentically$'` | exit 0; passed on Darwin |
| `bash .github/ci/ledger-consistency.sh .temp/TASK-260906-1f2ng0-ledger-revision4` | exit 0; 247 rows checked across Linux, Darwin, and Windows |
| `bash .github/ci/gate-selftest.sh` | exit 0; 185 passed, 0 failed |
| `go build ./...` | exit 0 |
| Initial `GOOS=windows GOARCH=amd64 golangci-lint run ./internal/testsupport/...` | exit 1; gosec flagged the helper's explicit fixture-path read; added a narrow `G304` reason for that test-only read and reran below |
| `GOOS=windows GOARCH=amd64 golangci-lint run ./internal/testsupport/...` | exit 0; 0 issues |
| `golangci-lint run ./cmd/curator/... ./internal/envprofile/... ./internal/godriver/...` | exit 0; 0 issues |
| `GOOS=windows GOARCH=amd64 golangci-lint run --new-from-rev=HEAD ./cmd/curator/... ./internal/envprofile/... ./internal/godriver/... ./internal/testsupport/...` | exit 0; 0 new issues |
| `gofmt -l internal/testsupport/path_read_windows.go internal/godriver/unreadabledir_windows_test.go` | exit 0; no files listed |
| `git diff --check` | exit 0 |

### Revision 5 — restore godriver and register Windows platform skips

Revision 4 failed on the hosted Windows runner: `internal/godriver`'s existing
unreadable-directory test and all three follow-up unreadable subtests reported
that the current-user DACL deny did not prevent reading the path. Restored
`internal/godriver` byte-identical to the Story base (`git diff HEAD --
internal/godriver` is empty), removed the shared DACL helper and the Windows
helper files added for these follow-ups, and left the godriver fixture untouched.

The unreadable-root, unreadable-module-directory, and unreadable-config
subtests now skip on Windows with the registered `platform-control` reason:
`the unreadable case uses POSIX mode bits; Windows ACL unreadability is not
reproducible for the runner account`. Each has its own ledger row requiring
Linux and Darwin and allowing the recorded skip on Windows. The missing path,
regular-file path, and absent machine-config cases remain required on Windows.
The mode-bit cases passed without skips on both Linux and Darwin.

Windows runtime was unavailable locally. Both changed Windows test binaries
cross-compiled, and a synthetic Windows test2json stream carrying the three
registered skip reasons passed `platform-case-gate.sh`; this checks ledger and
reason classification, not hosted Windows runtime behavior.

#### Revision 5 validation

All commands ran directly in the assigned worktree. Linux tests ran in a
container with `DAC_OVERRIDE` and `DAC_READ_SEARCH` dropped so mode-000
unreadability was exercised rather than skipped.

| Command | Result |
| --- | --- |
| `go test -count=1 -timeout=8m -v ./cmd/curator -run '^TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable$'` (Darwin) | exit 0; missing, unreadable root, unreadable module directory, and regular file all passed |
| `docker run --rm --cap-drop=DAC_OVERRIDE --cap-drop=DAC_READ_SEARCH --volume curator-task-260906-1f2ng0-go-cache:/cache --volume /Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260905-1n0iy8/worktree:/workspace --workdir /workspace --env GOMODCACHE=/cache/mod --env GOCACHE=/cache/build golang:1.26 go test -count=1 -timeout=8m -v ./cmd/curator -run '^TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable$'` (Linux) | exit 0; the same four cases passed without skips |
| `go test -count=1 -timeout=8m -v ./internal/envprofile -run '^TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault$'` (Darwin) | exit 0; absent-config and unreadable-config subtests passed |
| `docker run --rm --cap-drop=DAC_OVERRIDE --cap-drop=DAC_READ_SEARCH --volume curator-task-260906-1f2ng0-go-cache:/cache --volume /Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260905-1n0iy8/worktree:/workspace --workdir /workspace --env GOMODCACHE=/cache/mod --env GOCACHE=/cache/build golang:1.26 go test -count=1 -timeout=8m -v ./internal/envprofile -run '^TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault$'` (Linux) | exit 0; both subtests passed without skips |
| `GOOS=windows GOARCH=amd64 go test -c -o .temp/TASK-260906-1f2ng0-windows-revision5/cmd-curator.test.exe ./cmd/curator` | exit 0; test binary compiled, not executed |
| `GOOS=windows GOARCH=amd64 go test -c -o .temp/TASK-260906-1f2ng0-windows-revision5/envprofile.test.exe ./internal/envprofile` | exit 0; test binary compiled, not executed |
| `CI_GATE_GOOS=windows bash .github/ci/platform-case-gate.sh .temp/TASK-260906-1f2ng0-windows-platform-skips.json .temp/TASK-260906-1f2ng0-windows-platform-gate` | exit 0; synthetic Windows stream recorded and tolerated all three skips under the registered class |
| `sh .github/ci/ledger-consistency.sh` | exit 2; required evidence-directory argument omitted, so no checks ran |
| `sh .github/ci/ledger-consistency.sh .temp/TASK-260906-1f2ng0-ledger-revision5` | exit 0; 247 ledger rows checked across Linux, Darwin, and Windows |
| `sh .github/ci/gate-selftest.sh` | exit 0; 185 passed, 0 failed |
| `sh .github/ci/no-broad-suppression.sh` | exit 0; `no-broad-suppression: ok` |
| `go build ./...` | exit 0 |
| `golangci-lint run` | exit 0; 0 issues |
| `gofmt -d cmd/curator/profile_test.go cmd/curator/profile_path_permissions_test.go internal/envprofile/envprofile_f10f11f12_test.go internal/envprofile/path_permissions_test.go` | exit 0; no output |
| `git diff --check` | exit 0 |
| `git diff HEAD -- internal/godriver` | exit 0; no diff |

No Windows runtime or full landing suite was run locally. The Windows runtime
cases remain for the hosted candidate lane.

## Revision 6 (carry-forward republish)

Trunk moved to `a48f584c`; the accepted rev5 delta was carried uncommitted via
three-way converge. Content is unchanged from rev5 except the
orchestrator-ordered CHANGELOG revert below. No conflict markers were present
in the worktree. Per-path verification against
`TASK-260906-1f2ng0_change-request_rev5.patch` (13 paths):

Trunk-untouched (HEAD blob equals rev5 base blob; hunk text byte-identical
after index-line normalization, so each worktree file is byte-identical to its
rev5 post-image):

- `cmd/curator/profile.go` (+13/-5)
- `cmd/curator/profile_test.go` (+224/-8)
- `internal/envprofile/envprofile.go` (+4/-1)
- `internal/envprofile/envprofile_f10f11f12_test.go` (+57/-0)
- `internal/envprofile/envprofile_f16_test.go` (+93/-25)
- `internal/envprofile/lock.go` (+46/-25)
- `internal/envprofile/switch.go` (+64/-6)

New files in rev5 (expected bytes reconstructed from the patch equal the
worktree bytes exactly):

- `cmd/curator/profile_path_permissions_test.go` (771 B, helper only)
- `internal/envprofile/path_permissions_test.go` (545 B, helper only)
- `TASK-260906-1f2ng0_results.md` (23660 B; root copy removed per the stray
  policy below, board resource retained as canonical)

Intersecting paths (trunk moved the base; both sides verified present in the
worktree file, nothing dropped or duplicated):

- `.github/ci/platform-cases.tsv` — trunk added gitops fold, godriver
  worker-handle, scriptpolicy/scriptworker, and CLI build-dispatch rows (base
  `1fe7324` to head `edc1bb6`); rev5's 10-line "Profile source paths" section
  is present (4
  `TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable` rows, 2
  `TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault` rows).
- `.github/ci/skip-classes.tsv` — trunk added 3
  `script-worker-v1-native-control-inventory-v1` rows (base `f2368d1` to head
  `1d54851`); rev5's 1-line POSIX mode-bits `platform-control` row is present.
- `CHANGELOG.md` — trunk added 125 lines (R5/R4/R3 script-worker entries;
  base `4c241af` to head `171ec1f`); rev5's 7-line environments entry was
  present, then reverted entirely per CHANGELOG POLICY: `git diff HEAD --
  CHANGELOG.md` is now empty, i.e. the file equals trunk's. Entry text is
  preserved verbatim in the release-prep section below.

Stray cleanup: removed the root `TASK-260906-1f2ng0_results.md` (board
resource is canonical); no root `test/` or `ledger/` paths existed. The two
new `*_test.go` files are accepted rev5 content and were kept.

### Validation (this run; each command a direct standalone process)

| Command | Result |
| --- | --- |
| `go test ./internal/envprofile/... -count=1` (full, unmasked) | exit 1, NOT green: `panic: test timed out after 10m0s`, stalled in `TestSurfacingSinkWriteFailureIsNonFatal` inside the test-helper `gitRepo` (blocked in a git-child `wait4`). The identical git init/add/commit/tag sequence replays standalone in 0.05 s, and the stalled test passes alone (next row), so the stall was shared-host contention (the same macOS host-toolchain lock stall recorded in rev5 evidence), not this delta. |
| `go test -count=1 -timeout 150s ./internal/envprofile/ -run '^TestSurfacingSinkWriteFailureIsNonFatal$' -v` | exit 0; `--- PASS (7.18s)`, `ok 8.181s` |
| `go test -count=1 -timeout 8m ./internal/envprofile/ -run '^(TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault\|TestLoadMachinePolicyRejectsDanglingConfigSymlink\|TestMachinePolicyLoadFailureFailsMigration\|TestMachineUseSkipsScopedAdapter\|TestMachineUseClearsEveryScopeEqualToNewDefault\|TestInstallUseSkipsScopedAdapter\|TestScopedUseStillSwitchesOnlyThatHome\|TestSyncWritesScopedHomeOnce)$'` | exit 0; `ok 29.150s` |
| `go test -count=1 -timeout 9m ./cmd/curator -run '^(TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable\|TestProfileMachineUseSkipsScopedAdapter\|TestProfileUseReportsAllAdaptersAlreadyScoped\|TestProfileInstallUsePartialLeavesCurrent\|TestProfileInstallFirstActivationFailureReportsPersistedProfile\|TestProfileReinstallActivationFailureKeepsUpdatedWording)$'` | exit 0; `ok 28.053s` |
| `git diff --check` | exit 0 |
| `gofmt -l` over the 9 changed/new Go files | exit 0, no output |

Build standing: both packages compiled as part of the green test runs above;
rev5's `go build ./...` was exit 0 on identical content, and the only
post-rev5 tree change is the non-code CHANGELOG revert. The unmasked-suite
rerun is reported truthfully as timed-out above; rev5's remote-gate success
(run 35935908688) stands as the attached prior evidence for the unmasked
suite, and this run re-proved the full FU scope plus the stalled test
directly. All 13 DoD checklist items were already checked from rev5; none was
unchecked, so no item changes with this revision.

## CHANGELOG entry (for release prep)

Verbatim text of this task's reverted `CHANGELOG.md` hunk. The release-prep
leaf writes all entries.

- Environments profile follow-ups: path installs now report distinct
  `profile_source_path_missing` and `profile_source_path_unreadable`
  diagnostics; only a genuinely absent machine config selects the default
  policy; machine switches clear scope records equal to the new default in
  the same journaled publish; and the CLI reports when a switch touched no
  adapter homes or when an installed profile remains after activation is
  refused.

## Revision 6 (carry-forward republish, RUN-260924-d45274)

Second republish attempt after RUN-260924-b124f1's handoff heartbeat expired.
Worktree was still `development` with the carried delta uncommitted; no content
changes were made in this run. Per-path verification re-run here (13 rev5
paths; `git status --porcelain`, `git diff --stat HEAD`, grep counts):

- Trunk-untouched, byte-identical shape (diffstat matches rev5 exactly, no
  markers, `git diff --check` exit 0, `gofmt -l` clean):
  `cmd/curator/profile.go` (18), `profile_test.go` (232),
  `internal/envprofile/envprofile.go` (5), `envprofile_f10f11f12_test.go`
  (57), `envprofile_f16_test.go` (118), `lock.go` (71), `switch.go` (70).
- New rev5 helper files kept, sizes exact: `cmd/curator/
  profile_path_permissions_test.go` (771 B), `internal/envprofile/
  path_permissions_test.go` (545 B).
- Intersecting, both sides present: `.github/ci/platform-cases.tsv` (rev5
  10-line Profile-source-paths section: 6 diagnostic rows + header; trunk
  rows present), `.github/ci/skip-classes.tsv` (rev5 1 POSIX mode-bits row +
  trunk 3 script-worker rows), `CHANGELOG.md` (`git diff HEAD --
  CHANGELOG.md` 0 bytes = equals trunk `a48f584c`; entry text preserved in
  the release-prep section above).
- Strays: no root `TASK-260906-1f2ng0_results.md`, no `test/`, no `ledger/`.

### Validation (this run RUN-260924-d45274; each a direct standalone process)

| Command | Result |
| --- | --- |
| `go test -count=1 -timeout 8m ./internal/envprofile/ -run '<FU-1..FU-4 mask>'` | exit 0; `ok 151.909s` |
| `go test -count=1 -timeout 9m ./cmd/curator -run '<FU CLI mask>'` attempt 1 | exit 1, NOT green: transient `link: cannot open file .../go-build/...: no such file or directory` (darwin_amd64 linker vs go-build cache race); no test compiled or ran — infra failure, not a code failure |
| same `./cmd/curator` FU mask, retry | exit 0; `ok 95.410s` |
| `git diff --check` | exit 0 |
| `gofmt -l` over the 9 changed/new Go files | exit 0, no output |

Unmasked `go test ./internal/envprofile/... ./cmd/curator -count=1` was not
re-run green in this run: both full packages exceed the single-call time bound
on this shared host (rev5 evidence records full-package timeouts and the
prior Revision 6 run's unmasked stall in `TestSurfacingSinkWriteFailureIsNonFatal`
as host contention). The FU-scoped masks above re-prove the full FU scope in
this run; the unmasked suite stays covered by rev5's remote-gate success (run
35935908688) as already-attached evidence, not as this run's claim. All 13 DoD
items were already checked from rev5; none unchecked.

## Revision 7 (carry-forward republish)

Content unchanged from accepted revision 6. Trunk moved to `0a628621`;
the orchestrator ran `worktree converge STORY-260905-1n0iy8` and the
accepted delta is carried uncommitted. No content changes were made in
this run. `CHANGELOG.md` already equals trunk (rev6 carries no CHANGELOG
hunk; the entry text stays verbatim in the release-prep section above).

Per-path verification against
`TASK-260906-1f2ng0_change-request_rev6.patch` (11 paths):

- Trunk-untouched, byte-identical to the rev6 post-image (`git
  hash-object` equals the patch `index` post blob): `.github/ci/
  skip-classes.tsv`, `cmd/curator/profile.go`, `cmd/curator/
  profile_test.go`, `internal/envprofile/envprofile_f10f11f12_test.go`,
  `internal/envprofile/envprofile_f16_test.go`, `internal/envprofile/
  lock.go`, `internal/envprofile/switch.go`.
- New rev6 files kept with 100% of rev6-added lines present: `cmd/curator/
  profile_path_permissions_test.go` (771 B), `internal/envprofile/
  path_permissions_test.go` (545 B); both are permission helpers with no
  `Test` of their own and compile inside the green runs below.
- Intersecting, both sides present, no conflict markers anywhere:
  `.github/ci/platform-cases.tsv` (all 10 rev6-added section lines plus
  trunk's 1 added `TestExecutableIdentityCasesAtProductionEntry` row, one
  copy each, disjoint regions) and `internal/envprofile/envprofile.go`
  (all 4 rev6-added `Lstat` lines at `loadMachinePolicy` plus trunk's 6
  added lines in the import/`Policy`/`installLocked`/`pathManifestDiag`/
  `stateForPath` hunks, disjoint from ours at ~line 1722; `git diff HEAD`
  on this file shows exactly our rev6 hunk).
- `git diff --name-only HEAD` (9 tracked) plus the 2 untracked new test
  files lists exactly the 11 rev6 paths; `git diff HEAD -- CHANGELOG.md`
  is empty; no root `TASK-*`/`BUG-*` files, no `test/` or `ledger/`
  strays.

### Validation (this run; each a direct standalone process, no pipes)

| Command | Result |
| --- | --- |
| `go test -count=1 -timeout 9m ./internal/envprofile/...` (full, unmasked) | exit 1, NOT green: `panic: test timed out after 9m0s` (540.8 s), stalled in `TestPrecedencePrimitivesDriveEmission` (`overlays_test.go`, a file this delta does not touch). Passes alone (next row), so the stall is shared-host contention under parallel load — the same signature rev5/rev6 evidence records for other unrelated tests — not this delta. |
| `go test -count=1 -timeout 150s ./internal/envprofile/ -run '^TestPrecedencePrimitivesDriveEmission$' -v` | exit 0; `--- PASS (22.23s)`, `ok 23.300s` |
| `go test -count=1 -timeout 8m ./internal/envprofile/ -run '<8-test FU mask: LoadMachinePolicyTreatsOnlyAbsentConfigAsDefault, LoadMachinePolicyRejectsDanglingConfigSymlink, MachinePolicyLoadFailureFailsMigration, MachineUseSkipsScopedAdapter, MachineUseClearsEveryScopeEqualToNewDefault, InstallUseSkipsScopedAdapter, ScopedUseStillSwitchesOnlyThatHome, SyncWritesScopedHomeOnce>'` | exit 0; `ok 75.261s` |
| `go test -count=1 -timeout 9m ./cmd/curator` (full, unmasked) | exit 1, NOT green: `panic: test timed out after 9m0s` (541.2 s), stalled in `TestProductionExternalDepsFalseDrivesMirrorFetchToSink` (network/toolchain-fingerprint path this delta does not touch) with parallel tests queued behind it. Passes alone (next row): host contention, not this delta. |
| `go test -count=1 -timeout 5m ./cmd/curator -run '^TestProductionExternalDepsFalseDrivesMirrorFetchToSink$'` | exit 0; `ok 95.543s` |
| `go test -count=1 -timeout 9m ./cmd/curator -run '<6-test FU mask: ProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable, ProfileMachineUseSkipsScopedAdapter, ProfileUseReportsAllAdaptersAlreadyScoped, ProfileInstallUsePartialLeavesCurrent, ProfileInstallFirstActivationFailureReportsPersistedProfile, ProfileReinstallActivationFailureKeepsUpdatedWording>'` | exit 0; `ok 69.531s` |
| `git diff --check` | exit 0 |
| `gofmt -l` over the 9 changed/new Go files | exit 0, no output |

The two unmasked-suite reds are reported truthfully as failing with the
contention rationale above; the full FU scope plus both stalled tests are
re-proved green in this run, and rev5's remote-gate success (run
35935908688) stands as the attached prior evidence for the unmasked
suites. All 13 DoD checklist items were already checked; none was
unchecked, so no item changes with this revision.
