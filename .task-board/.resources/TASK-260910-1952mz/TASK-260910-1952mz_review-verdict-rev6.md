# TASK-260910-1952mz — revision 6 independent review

Candidate: `10e805b610552aa72bf4f28de1f719f3f3621289`, base `1de6f8e12f33212843820be756923b1179063d2c`.
All 11 changed worktree files were compared byte-for-byte to the candidate tree and matched. No candidate code was edited. Disposable copies below `.review/rev6` were used for attacks.

## Scope and conformance

| Requirement | Review evidence |
| --- | --- |
| Closed manager-home approval state | `internal/hookapproval/hookapproval.go:1`, `:48`, `:198`, `:237`: documented four-field TSV; manager home is `Config.Home()` (`internal/config/config.go:177`); absent and unreadable reads distinguished. |
| Manager-written digests | `internal/envfiles/envfiles.go:79`, `:100`, `:116`; install publication calls `recordPublishedEnvApprovals` at `internal/install/commit.go:670`, under the existing home lock. Integration tests exercise `install.Project`, including dry-run purity. |
| Both profiles, warning default | `internal/shell/shell.go:61`: A-warning default; `HookWithProfile` selects the closed A/B set. POSIX and PowerShell validate records before project sourcing, with exact diagnostic names and once-per-path warnings. |
| R1 closed-record validation | Generated checks share grammar constants; `TestShellHookRejectsMalformedRecords` independently rerun with real POSIX interpreters and pwsh. Forged-approver narrowing mutant killed. |
| R2 POSIX parsing | `TestGeneratedHookParsesUnderPOSIXShells` and functional sh/dash cases independently rerun. Bash-only prompt integration is guarded and evaluated only by Bash. |
| R3 symlink identity | Go resolves symlinks; POSIX resolves physical parents and links; PowerShell scans path components. Both-profile alias tests independently rerun for both hook families. |
| R4 atomic state replacement | Single rename publication, no destination removal fallback; `TestFailedPublicationPreservesLastValidState` independently rerun. |
| Windows/MSYS identity | POSIX `cygpath -w` conversion, drive normalization and case-folded lookup match native Go records. Missing conversion returns failure; caller warns and refuses in BOTH profiles. Windows cross-spelling proof is committed; native Windows execution accepted from the exact revision's hosted gate, not claimed as locally rerun. |
| Driven vectors | `TestShellHookTrustVectors` reads the external family and reaches production `HookWithProfile`, then real shell activation twice. Cases assert sourced markers, diagnostics, path/approval-command warning count and first/second activation split. Root-content and interpreter capability rows are in `.github/ci/platform-cases.tsv:150`. |
| Release and scope | `CHANGELOG.md:28` names S6/A-warning, migration command and later B flip. No SPEC_PIN change, vendored spec, ax/proposal or sibling command/posture implementation. §8.3 commands and §8.6 posture are explicitly assigned to TASK-260910-3ungjy. |

## Revision 5 to 6

Compared the two published patch resources. Only `internal/shell/shell_hook_trust_test.go` differs: harness-only `SourcedMarker`, conditional expected-marker selection, and `expectA.SourcedMarker = "2"` for changed bytes. Production code and identity mapping are unchanged. The changed-A assertion now proves the new bytes were sourced, instead of expecting stale marker 1.

## Independent validation

Shell: zsh, `set -o pipefail`. Conformance root: `/Users/administrator/Developer/ReluxWorks/curator/curator-spec/conformance/v1`; focused hook run adds `/tmp/pwsh/app` to PATH for real pwsh.

- `go build ./... && go vet ./... && gofmt -l .`: exit 0. The unrestricted gofmt scan names pre-existing board-resource and prior disposable-review files; `gofmt -l internal/ cmd/` separately exits 0 with no source findings.
- `go test -count=1 -json ./internal/shell/... ./internal/hookapproval/... ./internal/envfiles/...`: exit 0. Raw transcript: `focused.json` in the attached evidence archive.
- `go test -count=1 ./internal/install/... -run 'TestProjectInstallRecordsShellHookApprovals|TestProjectDryRunRecordsNoShellHookApprovals'`: exit 0.
- `go test -count=1 ./cmd/curator/... -run TestShellInit`: exit 0.
- `golangci-lint run ./internal/shell/... ./internal/hookapproval/... ./internal/envfiles/... ./internal/install/...`: exit 0, `0 issues.`
- An initial broad invocation including full install and CLI suites was intentionally terminated (exit 143) after recognizing it violated the bounded-check strategy; it is NOT counted as a pass. The completed bounded checks above replace it. Full install/CLI/platform coverage is accepted from the exact candidate hosted evidence.
- Published `TASK-260910-1952mz_change-request_rev6-validation.log`: remote-gate run 35209190206, exit 0, Linux/macOS/Windows Test and applicable Race lanes successful. Hosted test-case coverage is explicitly unknown in that log; this review does not fabricate case counts from job success.

## Negative evidence

Two independently executed narrowing mutants, `go test -count=1`, disposable copy only:

1. Narrow closed approver validation to reject only an empty approver. `TestShellHookRejectsMalformedRecords/forged-approver/posix` fails: forged record sources silently under B. Exit 1.
2. Narrow digest enforcement to record-presence-only (an existing record authorizes changed bytes). `TestShellHookTrustVectors` fails the changed-env-sh cases, including B sourcing forbidden bytes. Exit 1.

Measured mutation coverage: **2/2 killed, 0 survivors**. Bound: POSIX approver and changed-digest clauses; this is not an exhaustive mutation score for every platform/refusal clause.

## Bounds and observations

- The Windows interpreter probe no longer blanket-skips POSIX suites by GOOS. It probes PATH and Git installation roots. Its dedupe uses resolved path strings without Windows case folding, so differently cased spellings can still run the same interpreter twice. This is redundant coverage/time, not a weakened gate; the producer's rev5 failure already showed duplicate names. Non-blocking cleanup.
- Case-insensitive path behavior beyond the tested ASCII/native/MSYS examples is not established by the local platform. Go EqualFold and awk tolower have different Unicode semantics; no Unicode identity completeness claim is made.
- No repository LOGBOOK.md edit: campaign rules forbid it; a task-scoped review logbook resource records review observations instead.

## Verdict

**accepted** — revision 6 is ready for integration through `accept_cr`, not a reviewer `done` transition. No blocking correction found. The explicitly deferred commands/posture remain the sibling task's responsibility.

Additional measured evidence: **14/14 vector cases PASS, 0 vector skips** (8 POSIX cases executed through available sh/dash/bash/zsh, 6 PowerShell cases through real pwsh). The Windows-only cross-spelling and drive tests legitimately skip locally and are covered by the attached exact-candidate hosted gate evidence.

Empty-root vector run independently confirms the `root-content` skip, exit 0. Independent disposable `TestReviewerPowerShellCaseIdentity` verifies with `os.SameFile` that mixed/lower-case paths identify one physical file on this host, then drives the generated PowerShell hook: **2/2 profiles source silently**. The potential case-comparison concern did not reproduce and is not reported as a defect; native Windows case-variant completeness beyond hosted tests remains a stated bound.

Before verdict, `task-board spawn goal "$TASK_BOARD_RUN_ID"` reported this run is not goal-bound. Final candidate comparison again checks all changed files against the published tree. The earlier broad test interruption and shell-start stalls are environmental/operational observations, not green checks or product blockers.

