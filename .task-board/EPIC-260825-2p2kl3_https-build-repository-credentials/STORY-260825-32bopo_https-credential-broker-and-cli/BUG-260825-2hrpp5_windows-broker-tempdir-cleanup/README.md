# BUG-260825-2hrpp5: windows-broker-tempdir-cleanup

## Description
On windows-latest two broker tests fail in cleanup, not in assertion: TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts and TestHTTPSBrokerStateContainsHostAndUsernameOnly both report TempDir RemoveAll cleanup unlinkat ...manager-wrappers/curator-build-https-askpass.exe: Access is denied. Windows refuses to unlink an executable while a handle to it is still open, so the broker copy the test materializes and runs is still held when t.TempDir cleanup runs. The assertions themselves pass; the run fails on the deferred cleanup. Fix the lifecycle so the executable is fully released before cleanup, rather than suppressing the cleanup error.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
Both tests pass on windows-latest in CI; the fix releases the executable rather than ignoring the cleanup failure; macOS and Linux stay green.
