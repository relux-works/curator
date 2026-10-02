# TASK-261002-9w4wy3 — windows-exec-hardlink-origin-checks

Production path: `deriveProfileForPlatform` -> `resolveExecForPlatform` -> `readExecIdentityAt`; `VerifyExec` repeats the same checks immediately before launch. The open executable's owner must be TrustedInstaller or LocalSystem. Native Windows code reads its owner descriptor and 128-bit file ID, enumerates all hard-link names with FindFirstFileNameW/FindNextFileNameW, checks every alias's file ID, and requires enumeration count to match stable handle link counts. Every additional physical name must be under captured SystemRoot/WinSxS. Unreadable, partial, duplicated, foreign or changed evidence refuses. The check runs before and after hashing. No production metadata can supply the OS observation seam.

The published dcc7f015e2d97edf2d52928afb6fd79ec8129e8b fixture matches the eight embedded cases exactly. The production family tally is 8 driven / 8 published, 0 known gaps, 0 bounds, 0 skips. The component-store positive, uncaptured-SystemRoot negative, explicit-search negative, outside-System32 negative, and both interpreter hard-link negatives are preserved.

Exactly two gap rows were removed for the selected rc.13 suite identity be11bb1e4c46f21fb5684d586f9c2a8b0d59f3b437bc7ea7aa5aa530fe4d47ca. Rows for other suite identities were left intact, as the brief explicitly limits removal to two rows. The case-count pins are unchanged. This task qualifies the selected rc.13 suite; it makes no claim that the two other suites' remaining ledger rows are suitable for a new candidate-suite run.

Validation commands run directly (no tee), with observed real exit codes:

| Command | Exit | Evidence |
| --- | ---: | --- |
| Baseline `GOOS=windows go vet ./...` before changes | 0 | captured tool output, empty diagnostics |
| Focused resolver tests before ledger removal | 1 | Expected failure: both formerly known-gap cases passed, requiring ledger removal |
| Focused resolver tests after ledger removal | 0 | 8/8 family, origin negatives, preserved positive and uncaptured root |
| rc.13 `go test -count=1 -timeout 6m ./internal/scriptworker ./internal/conformancecoverage` | 0 | packages.log |
| rc.13 `go test -count=1 -timeout 3m ./internal/scriptworker -run '^TestExecutableIdentityCasesAtProductionEntry$' -v` | 0 | family.log; 8/8 |
| `go build ./...` | 0 | captured tool output |
| `GOOS=windows go build ./...` | 0 | windows-build.log |
| Post-change `GOOS=windows go vet ./...` | 0 | windows-vet.log; unchanged from baseline |
| `go vet ./...` | 0 | host-vet.log |
| `golangci-lint run ./internal/scriptworker/...` | 0 | 0 issues after correcting initial capitalized error-string diagnostic (initial exit 1) |
| `GOOS=windows golangci-lint run --new-from-rev HEAD ./internal/scriptworker/...` | 0 | windows-lint-delta.log; 0 new issues; an earlier concurrent attempt exited 3 for the linter process lock |
| Unrestricted Windows package lint during development | 1 | New unsafe-call/buffer-conversion diagnostics corrected; unrelated existing errcheck, gosec, revive, unused diagnostics remain outside this delta |
| `bash .github/ci/ledger-consistency.sh .temp/TASK-261002-9w4wy3/ledger` | 0 | 495 rows across linux/darwin/windows; required native case registered |
| `git diff --check` | 0 | captured tool output |
| Published-fixture comparison and hosted-snapshot byte comparison | 0 | all 8 pinned cases match dcc7f015; all changed/new candidate files match hosted snapshot |

Mutation commands ran with `-count=1` and sources were restored byte-for-byte from a private copy, never from Git:

| Mutation | Command scope | Exit | Observed failure |
| --- | --- | ---: | --- |
| Drop owner check | `TestExecutableIdentityCasesAtProductionEntry` | 1 | windows-exec-unowned-file-hardlinks accepted incorrectly; mutant-drop-owner.log |
| Drop component-store check | Same family | 1 | windows-exec-noncomponent-store-hardlinks accepted incorrectly; mutant-drop-store.log |
| Narrow owner check to empty SID only | Same family | 1 | Nonempty ordinary SID admitted; mutant-owner-empty-only.log |
| Narrow link traversal to the first two names | `TestWindowsExecHardlinkOriginEvidence/third-link-outside-store` | 1 | Third foreign alias admitted; mutant-first-two-links.log |

Hosted Windows qualification: https://github.com/relux-works/curator/actions/runs/36956099103

Both `Native origins (windows-2022)` and `Native origins (windows-latest)` jobs concluded success. Each ran `go vet ./...`, `go build ./...`, and rc.13 `go test -count=1 -timeout 8m -v ./internal/scriptworker ./internal/conformancecoverage`. The native test never replaces the observation seam: it creates actual NTFS hard links, changes owners through Windows security APIs, exercises both origin refusals and both allowed owner SIDs, adds an outside alias before VerifyExec, and accepts the runner's real System32 cmd.exe. Existing real Python/Node launch tests also run in the package suite.

This is a focused hosted two-Windows-version qualification, not a full repository CI matrix replay. Temporary hosted verification workflow is included only in the remote snapshot, not the candidate worktree. Initial workflow_dispatch attempt exited 1 (HTTP 404 because the temporary workflow was absent from the default branch); push-triggered qualification then succeeded.

Candidate HEAD: source tree snapshot recorded in hosted-snapshot.json. Hosted snapshot commit: 494088911f5d76fc1084f530ab6afda10dc08078; source tree: 91fbebfbc4f368b876cafe200844e8edc5d94a23. Worktree and HEAD remain uncommitted and unchanged by the remote snapshot operation.

One CHANGELOG line was added. No LOGBOOK change: explicitly prohibited by the current task brief. Important implementation and evidence decisions are recorded here instead.

The extra restored-source race command (`CURATOR_CONFORMANCE_ROOT=/tmp/TASK-261002-9w4wy3-spec/conformance/v1 go test -race -count=1 -timeout 6m ./internal/scriptworker ./internal/conformancecoverage`) returned exit 1 while the host was exhibiting exec stalls. Its redirected log cannot yet be read, so the failure cause is unknown and this run is not counted green. The requested syspolicyd inspection reported running / 349 successive crashes before validation; later inspections themselves stalled. The headless turn remains active awaiting host recovery, artifact collection and board attachment. No unknown outcome is classified as passing.

Host recovery and race follow-up: syspolicyd eventually reported running with 351 successive crashes (previously 349). The full race log shows four 60-second worker-ready failures followed by the package's 6-minute timeout in TestScriptWorkerHappyPath, not a data-race report. I reran the changed resolver tests and all those affected worker-start cases myself with `-race -count=1 -timeout 4m`: exit 0. Exact `-run` mask and verbose output are in recovery-race.log. The earlier full non-race local package results and both hosted full package results remain accepted against byte-identical restored source; I did not replay the full repository matrix or the entire failed extra race suite.

Downloaded hosted logs were inspected: both platforms explicitly record 8/8 driven, all five native origin subtests pass, and real Python/Node declared cmd.exe launch tests pass. Hosted job/step JSON confirms both native lanes' vet, build and package-test steps succeeded (exit 0). Download and metadata/log capture commands exited 0.

The obsolete inherited LOGBOOK checklist item was replaced with task-outcome recording because the current brief explicitly prohibits LOGBOOK. No LOGBOOK file was touched. The initial checklist-edit call without the CLI's destructive-operation confirmation flag was refused without applying that mutation; it was repeated with `--confirm` within the brief's authorized scope.

Hosted TASK-261002-9w4wy3-windows-2022 tests.log SHA-256: 9114a8ad3c6bd8b20ba9e75478cfb201ed0d0d771c092bdfa817a660a263f9f7.

Hosted TASK-261002-9w4wy3-windows-latest tests.log SHA-256: 34532f3bca97ca493ba27352565ee250e94251dbc8cb328f13796cbe47406ce2.
