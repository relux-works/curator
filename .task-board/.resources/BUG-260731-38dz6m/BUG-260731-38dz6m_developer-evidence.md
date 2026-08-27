# BUG-260731-38dz6m developer evidence

## Scope and six native Windows residuals

The combined Windows raw `go test -json` evidence named these six failing
top-level cases:

1. `internal/transaction.TestCommitAndRollbackRestoreALinkExactly`
2. `internal/transaction.TestRecoveryFinishesAPreparedLinkTransaction`
3. `internal/transaction.TestEntryRemovalRestoresTheExactLink`
4. `internal/transaction.TestEntryTargetsDoNotAliasTheirDestination`
5. `internal/transaction.TestNamespaceIdentityIsReadOnceWithinOneValidationPass`
6. `internal/godriver.TestFingerprintReportsUnreadableDirectoryIdentically`

The first PR 15 native run, 30643616475, proved the five transaction cases
green and left only the fingerprint case red. Artifact 8799245786 records the
fingerprint failure as an empty diagnostic code where
`toolchain_unreadable` was required.

## Findings and changes

- Link destinations are host-native filesystem text. Four transaction tests
  now capture the destination actually recorded by `os.Symlink` and require
  that exact text through commit/recovery/rollback.
- Windows defers the volume/file-index read used by `os.SameFile`. Namespace
  identity is now completed while the pass snapshot is taken, so later
  comparisons are in-memory and a failed identity read remains fail-closed.
- `os.Chmod(0o000)` does not make a Windows directory unreadable. A deny ACE
  also proved insufficient on the hosted administrative runner because Go
  opens directories with backup intent and `SeBackupPrivilege` can bypass the
  DACL. The Windows test fixture now holds a `GENERIC_READ` directory handle
  with share mode zero, causing later directory opens to fail with a native
  sharing violation regardless of ACL privilege. The test directly asserts
  that `os.ReadDir` is refused before comparing the two fingerprint paths.
- All six top-level cases are explicit in `.github/ci/platform-cases.tsv`.
  No skip, tolerance, platform exclusion, ledger relaxation, or broad
  suppression was added.

## Signed PR

- PR: https://github.com/relux-works/curator/pull/15
- First signed commit: `d07284571f6c6be996e57c09e129cb35d0773c54`
- Privilege-proof fixture commit: `a134fdc422b2fea3c4c1ac09b1f8f7c4afbc75a2`
- `git log --show-signature` reported a good ECDSA signature for both commits.

## Direct validation results

Every command below was run as its own process. Each reported exit code is the
real terminal exit code.

| Command | Exit | Result |
|---|---:|---|
| focused `internal/godriver.TestFingerprintReportsUnreadableDirectoryIdentically` | 0 | pass |
| focused five `internal/transaction` cases | 0 | pass |
| `go test -count=1 ./internal/godriver ./internal/transaction` | 0 | pass |
| Windows amd64 `go test -c` for `./internal/godriver` | 0 | compiled |
| Windows amd64 `go test -c` for `./internal/transaction` | 0 | compiled |
| `go vet ./...` | 0 | pass |
| `go build ./...` | 0 | pass |
| `golangci-lint v2.12.2 run` | 0 | 0 issues |
| `go test -count=1 ./...` | 0 | all packages pass |

The local Go commands used the repository-pinned Go 1.25.5 toolchain with the
stale shell `GOROOT` removed.

## Native CI

- Prior diagnostic run: https://github.com/relux-works/curator/actions/runs/30643616475
- Signed-head validation run: https://github.com/relux-works/curator/actions/runs/30668611796
- Attempt 1 artifact: 8808562274
- Attempt 1 proved all six scoped cases green and the platform-case gate green.
  Its overall Windows job was red only because the PR 13-owned
  `internal/managerlock.TestSubprocessBuildKeyDeduplicationAcrossProjects`
  returned `blocked` instead of `acquired`. That case passed on the immediately
  preceding PR 15 run, and commit `a134fdc` changes no managerlock code.
- Attempt 2 is a failed-job rerun on the identical signed SHA, used to verify
  that no unrelated Windows lane case regressed: exit 0 / success.
- Attempt 2 Windows artifact: 8809078378.
- The attempt 2 observed-case report contains pass rows for all six scoped
  cases, a pass row for the attempt 1 managerlock flake, and no fail rows.
- The attempt 2 platform-case report ends with `platform-case gate: ok`.

The DACL diagnosis is consistent with both implementations involved: Go 1.25's
Windows `Openat` always adds `FILE_OPEN_FOR_BACKUP_INTENT`, and Microsoft
documents that backup intent plus `SeBackupPrivilege` can override file
security checks. The zero-share handle relies on Windows share-access
enforcement instead of access-control permissions.

## Environment note

The configured `ssh win` host was unreachable and is known offline in
Tailscale. Native execution evidence therefore comes from the repository's
`windows-latest` GitHub Actions lane rather than an SSH surrogate.
