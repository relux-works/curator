# BUG-260731-38dz6m independent review verdict

## Verdict

ACCEPTED.

Reviewed Curator PR 15 at the exact current head `a134fdc422b2fea3c4c1ac09b1f8f7c4afbc75a2` against current `main` `ef6ce607f4fb00aa94069d2aad426b08f276fb57`. The GitHub API reports the PR open, clean, and mergeable; the branch and local review worktree are unchanged.

## Six combined-base Windows failures

The current combined Windows evidence names exactly these scoped top-level cases:

1. `internal/transaction.TestCommitAndRollbackRestoreALinkExactly`
2. `internal/transaction.TestRecoveryFinishesAPreparedLinkTransaction`
3. `internal/transaction.TestEntryRemovalRestoresTheExactLink`
4. `internal/transaction.TestEntryTargetsDoNotAliasTheirDestination`
5. `internal/transaction.TestNamespaceIdentityIsReadOnceWithinOneValidationPass`
6. `internal/godriver.TestFingerprintReportsUnreadableDirectoryIdentically`

## Implementation review

- The four link cases were invalid POSIX-literal expectations on Windows. The tests now capture the destination actually written by `os.Symlink` and require that exact host-native text through commit, recovery, and rollback. This preserves the exact round-trip contract rather than introducing a tolerance.
- Windows `os.Stat` / `os.Lstat` defer the volume/file-index read consumed by `os.SameFile`. The namespace pass now completes that identity when it is cached, so later comparisons are in-memory and an identity that cannot be completed fails closed. Non-Windows behavior remains the existing stat identity through a platform-split helper.
- The fingerprint production paths are unchanged. The test fixture is platform-split: Unix retains mode-bit denial, while Windows holds a `GENERIC_READ` directory handle with zero sharing and directly proves `os.ReadDir` is refused before comparing both fingerprint implementations. This avoids the hosted runner's backup-privilege DACL bypass.
- The patch changes only nine scoped files under `internal/transaction`, the single `internal/godriver` fingerprint test seam, and the platform-case ledger. It does not touch PR 13-owned managerlock, buildcache, staging, globalbins, GOROOT, or unrelated CI code.

## Independent evidence

- Both PR commits are cryptographically valid in GitHub verification and report good local ECDSA signatures: `d07284571f6c6be996e57c09e129cb35d0773c54` and `a134fdc422b2fea3c4c1ac09b1f8f7c4afbc75a2`. No tag contains the head.
- CI run `30668611796`, attempt 2, completed successfully on exact head `a134fdc`. Windows Test, Ubuntu Test/Race, macOS Test/Race, lint, all gate self-tests, interop, and naming are green. The candidate job is the workflow's expected skip.
- Reviewer parsing of native Windows artifact `8809078378` finds pass rows for all six scoped cases and the new Windows namespace regression, no failed observed case, no scoped skip, and `platform-case gate: ok`.
- Attempt 1 artifact `8808562274` already passed all six scoped cases. Its sole failure was `internal/managerlock.TestSubprocessBuildKeyDeduplicationAcrossProjects`; the identical-SHA rerun passes it, and the PR has no managerlock diff. The flake classification is therefore supported and no out-of-scope fix was absorbed.
- Reviewer-local checks at exact head: `go test -count=1 ./internal/godriver ./internal/transaction`, relevant and full `go vet`, `go build ./...`, gofmt check, Windows/amd64 test-binary cross-compiles for both packages, `golangci-lint v2.12.2` (0 issues), `.github/ci/no-broad-suppression.sh`, and `git diff --check` all exit 0.
- Diff inspection adds no `t.Skip`, platform exclusion, tolerance, suppression, or ledger relaxation. The six original cases and the Windows-specific regression are new must-run rows with no allowed skip.

## Non-blocking finding

The final platform ledger behavior text for `TestFingerprintReportsUnreadableDirectoryIdentically` still says “Windows DACL”, while the accepted fixture now uses Windows share-access denial. This is descriptive metadata only: the row still binds the exact case as must-run with no tolerance, and the native artifact proves it executes. Correct the wording opportunistically; it does not change this verdict.

## Handoff

This reviewer made no code or PR mutation and supplied no `commit_ack`. PR 15 is accepted for the commit-owning mover to land on `main`; the mover should record the merge and complete the remaining landing checklist evidence.