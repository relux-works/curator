# BUG-260923-3mazfw review verdict — CR rev2: ACCEPTED

1. Root cause: evidenced in results/notes — 10 hosted macOS artefacts, target test passed 9×, failed run 35855550263 (2026-09-23 11:46:47Z, runner 1000017779) with `fork/exec /opt/homebrew/bin/git: permission denied`. Lower OS cause (xattr/Homebrew update) honestly stated as unknown (bound).
2. Fix at production seam `internal/gitignore/gitignore.go` `runCheckIgnoreWithEACCESRetry` (called from `Missing`): retries only `*os.PathError` Op `fork/exec` + EACCES, once, after 100 ms; second failure wrapped with `%w` (EACCES preserved, fails closed, no .gitignore write). No blanket retry, no timeout widening, no skip. Rows: transient retry (2 calls), git exit status no retry, EPERM/ENOENT/open-EACCES no retry, persistent EACCES fails closed after exactly 2 calls.
3. Mutants (disposable clone, base 48da2690 + rev2 diff; unmutated `go test ./internal/gitignore` ok):
   - M1 retry any error → killed by TestMissingDoesNotRetryGitExitStatus + TestMissingDoesNotRetryOtherSpawnErrors
   - M2 retry up to 5× (≈forever) → killed by TestEnsurePersistentSpawnEACCESFailsClosedAfterOneRetry
   - M3 drop Op=="fork/exec" check → killed by TestMissingDoesNotRetryOtherSpawnErrors/EACCES outside fork-exec
4. Hosted gate rev2: run 35897148993 success, all lanes incl. Test/Race macos-latest and Windows.
Residual (non-blocking): AC3 "repeatedly" rests on runs 35881328824 + 35897148993 macOS green; other git spawn sites in the install path are not covered by this retry (stated scope is the check-ignore site that failed).
