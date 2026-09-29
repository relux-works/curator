# TASK-260910-31ocjt — rework 1 results

## Outcome

HTTPS askpass now receives the selected secret from a fetch-scoped transport. On Unix, the parent passes a pipe reader as an extra fd, closes the parent's reader after the fetch starts, and writes the secret to the pipe. The broker reads once and closes its fd. On Windows, the fetch environment carries only a cryptographically random named-pipe name. The pipe is single-instance, created with `FILE_FLAG_FIRST_PIPE_INSTANCE`, has a protected DACL granting access to the current user, serves one client, and closes when the fetch exits, including when no client connects. The secret is not placed in argv, environment variables, broker state, or a file.

The old `CURATOR_BUILD_HTTPS_ASKPASS_SECRET` production handoff is removed. The broker reads the transport only for the exact pinned password prompt and fails closed on transport errors or empty data. It clears the received secret after use. Crossconformance starts the broker via `testcli.RunWithHTTPSBrokerSecret`, using the shared process seam; no guard exemption was added. The audit finding now has a `Remediated by TASK-260910-31ocjt.` note.

## Specification anchor

Checked curator-spec v1.0.0-rc.13 at SPEC_PIN `23435129ebc4c29e5b7f75ec72a0aa0cd3f16065`. `protocol/repository-transport.md` §§3 and 7 constrain resolved credential handling and prohibit exporting or persisting secrets through manifests, locks, receipts, logs, command arguments, or compiler-worker environments. The external-repository acquisition vector pins the askpass broker as `core.askPass`. I found no clause or vector specifying pipe/fd delivery or the fetch/grandchild environment rows, so the new production-entry rows exercise those task-specific rules.

## Production-entry rows

- `TestSelectedHTTPSFetchSecretUsesPipeAndNeverEntersTheProcessEnvironment` drives `AcquireNetwork`; it verifies the fetch child environment has no secret, the broker receives the secret, and the askpass helper and its grandchild do not have the legacy secret variable.
- `TestPrivateHTTPSBrokerAuthenticatesRealGitRepository` drives real Git through the askpass broker and asserts successful authenticated HTTPS access, secret-free fetch environment, and helper/grandchild environment absence. Its Windows skip was removed.
- `TestHTTPSCredentialBrokerRejectsEmptyPasswordPipe` verifies an empty pipe returns exit 1 and empty stdout, protecting the fail-closed check.
- Windows-only rows inspect the created pipe DACL with `GetSecurityInfo`, verify that a second client cannot connect after the first, and verify that server cancellation ends a fetch with no client.
- The crossconformance broker dispatch row uses the instrumented `testcli` process seam and checks the broker child environment for the old secret variable.

## Mutants

- M1, append the secret back to the fetch environment: killed. `go test ./internal/buildrepo -run '^TestSelectedHTTPSFetchSecretUsesPipeAndNeverEntersTheProcessEnvironment$'` exited **1**; the regression row observed the secret in the fetch, helper, and grandchild environments. The mutant was reverted.
- M2, remove the broker's `len(secret) == 0` fail-closed check: the old `TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts` rows exited **0** with the mutant (survived). The new `TestHTTPSCredentialBrokerRejectsEmptyPasswordPipe` row exited **1** with the mutant (killed). The guard was restored; clean focused tests pass.
- The Windows DACL-widened and second-client/serve-twice mutants were **not run locally** because this host is Darwin. Their Windows runtime outcomes remain unknown pending the hosted Windows lane; the Windows tests are present and no Windows skip was added.

## Verification

Commands were run as standalone processes. Exit codes are the actual command results.

- `go test ./internal/buildrepo -run 'TestHTTPSCredentialBroker|TestHTTPSBroker|TestPrivateHTTPSBroker|TestSelectedHTTPSFetchSecret'` — **0** (Darwin; rerun after final Unix transport edit).
- `go test ./internal/buildrepo` — **0** (full package, 153.806s; prior clean run; the final transport edit was covered by the focused and race runs below).
- `go test -race ./internal/buildrepo -run 'TestHTTPSCredentialBroker|TestHTTPSBroker|TestPrivateHTTPSBroker|TestSelectedHTTPSFetchSecret'` — **0** (Darwin; after final Unix transport edit).
- `go test ./internal/crossconformance -run 'TestDraftSourcesBrokerAskpassDispatch|TestIntegrationSurfaceStartsNoProcessOutsideTheSharedSeams|TestIntegrationProductionSourceImportsNoRepositoryPackage'` — **0** (rerun after final Unix transport edit).
- `go vet ./internal/buildrepo ./internal/crossconformance ./internal/testcli` — **0** (after final source edit).
- `golangci-lint run ./internal/buildrepo ./internal/crossconformance ./internal/testcli` — first run **1** (reported the Unix child-reader close result and exported API comments); those issues were fixed. Rerun — **0**, `0 issues`.
- `GOOS=windows go vet ./internal/buildrepo` — **0** (cross-target validation only).
- `GOOS=windows go test -c -o /tmp/curator-buildrepo-windows.test.exe ./internal/buildrepo` — **0** (compiled only, after final source edit).
- `GOOS=windows go test -c -o /tmp/curator-crossconformance-windows.test.exe ./internal/crossconformance` — **0** (compiled only).
- `git diff --check` — **0**.
- Earlier `go test ./internal/buildrepo ./internal/crossconformance` attempt — **1**. It exposed a test-helper fd issue and a Windows wait-status type mismatch, and the full crossconformance suite reached its 10-minute timeout. The fd and type issues were corrected. The focused crossconformance rows above pass; a full crossconformance pass is not claimed.

Windows Git, named-pipe ACL inspection, second-client behavior, and Windows runtime mutants have not been executed on this Darwin host. Windows behavior is unverified locally; hosted Windows evidence is pending after handoff. The Windows test binaries were cross-compiled only.

No new manager-state reads or managed writes were introduced. No `CHANGELOG.md` or `LOGBOOK.md` edits were made; this results resource records the findings and limitations.

## CHANGELOG entry (for release prep)

Pass HTTPS askpass credentials to the trusted fetch child through a pipe or private named pipe instead of a process environment variable. Verify the fetch environment and askpass descendants do not receive the secret.

## Gate fix 3 (rev6) — Windows test-side fixes only
1. `TestHTTPSBrokerNamedPipeHasCurrentUserOnlyDACL`: replaced the SDDL string comparison (failed on the hosted local-Administrator account, rendered `LA`) with an ACL walk: DACL must be SE_DACL_PROTECTED, AceCount==1, ACE 0 (windows.GetAce) must be ACCESS_ALLOWED_ACE_TYPE and its SID `Equals` the process token user SID; the ACE SID must not equal Everyone/Users/Authenticated Users. Production code unchanged. Widened-DACL mutant (Everyone ACE added) is killed by AceCount!=1 / SID mismatch — this row runs only on the hosted windows-latest lane; local mutant execution on Windows is not possible here (macOS host), so the kill is to be read from that lane.
2. `TestPrivateHTTPSBrokerAuthenticatesRealGitRepository`: on Windows the test's git argv adds `-c http.sslBackend=openssl` before `-c http.sslCAInfo=<test CA>` (schannel ignores sslCAInfo). Test-only; sslVerify stays true; no skip. The test builds its own git argv, so no production seam was needed.

Local runs (zsh, macOS):
- `GOOS=windows go vet ./internal/buildrepo` → exit 0
- `GOOS=windows go test -c -o $TMPDIR/b.exe ./internal/buildrepo` → exit 0
- `go test ./internal/buildrepo -run 'Broker|HTTPS'` → exit 0
Windows is UNVERIFIED until the hosted windows-latest lane runs; it is the arbiter.
