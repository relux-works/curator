# TASK-260910-31ocjt review verdict — rev4 (tree 9874753f, base 3f60f7f0): CHANGES REQUESTED

Reviewer: claude-opus-5-5 low. Worktree bytes == candidate tree (git diff 9874753f empty). Shell zsh, set -o pipefail, go1.26.0 darwin/amd64.

## Reran (real exit codes)
- go vet ./internal/buildrepo ./internal/crossconformance ./internal/testcli → 0
- go test -count=1 ./internal/buildrepo -run 'Broker|Askpass|HTTPS' → 0
- go test -count=1 ./internal/crossconformance -run 'Broker|Askpass|HTTPS|StartsNoProcess' → 0 (guard passes; e2e goes through testcli, no exemption)

## Mutants
- M1 secret back in fetch env (admission.go acquireNetworkFormat appends CURATOR_BUILD_HTTPS_ASKPASS_SECRET=secret): KILLED, buildrepo exit 1 (TestSelectedHTTPSFetchSecretUsesPipeAndNeverEntersTheProcessEnvironment).
- M2 fail-closed removed (httpsbroker.go readHTTPSBrokerSecret: drop `len(secret) == 0`): SURVIVES — buildrepo exit 0, crossconformance exit 0. An empty/already-drained pipe (second Password prompt after a 401 retry, or a helper invoked twice) answers an EMPTY password with exit 0. The code is currently correct, but no row pins it; review note item 2 requires it.

## Findings (must fix)
F1 (item 2, test gap): add a row that invokes the broker Password prompt twice on one pipe (or with an empty pipe) and asserts exit 1 / empty stdout on the second; M2 must be killed with a real exit code. Anchor: internal/buildrepo/httpsbroker.go:124 (`len(secret) == 0`), rows in httpsbroker_test.go TestHTTPSCredentialBrokerAnswersOnlyPinnedGitPrompts and/or crossconformance TestDraftSourcesBrokerAskpassDispatch.

F2 (Windows, production path unverified and likely broken): internal/buildrepo/httpsbroker_pipe_windows.go:13-22 relies on inheriting the pipe handle through git.exe -> git-remote-https.exe -> askpass. Git for Windows restricts child handle inheritance to stdio by default (core.restrictInheritedHandles, PROC_THREAD_ATTRIBUTE_HANDLE_LIST), so the askpass broker most likely receives an invalid handle and every authenticated HTTPS fetch on Windows fails closed (regression vs the env hand-off). Also note Go's os.Pipe on Windows creates inheritable handles (makeInheritSa) including the writer — only safe because Go uses a handle list. The only "every platform" row (TestHTTPSBrokerPipeSurvivesFetchAndAskpassExecOnEveryPlatform, httpsbroker_test.go:220) uses a Go fake fetch child that explicitly re-inherits the handle (runHTTPSBrokerFetchChild) — it simulates exactly the step real Git for Windows does not perform, so it is a bound, not evidence. The real-git row TestPrivateHTTPSBrokerAuthenticatesRealGitRepository skips on Windows (httpsbroker_test.go:268). Required: either (a) the brief's Windows option — a named pipe with a private (current-user-only) ACL whose name travels in env and which the broker opens once (FILE_FLAG_FIRST_PIPE_INSTANCE, single instance, server closes after one read), with a real-git Windows row, or (b) hosted windows-lane evidence of a real `git ls-remote`/fetch through the handle chain. If neither is feasible, state Windows as unverified and do not claim cross-platform.

## Verified OK
- CURATOR_BUILD_HTTPS_ASKPASS_SECRET removed from production; remaining hits are test mutant-controls, LOGBOOK history and docs/security-audit-2026-09.md:196 (audit finding text — acceptable as history, but consider a "remediated by 31ocjt" note).
- Unix: reader closed in parent after Start; writer written after Start then closed; helper reads to EOF and closes; handle value only (not secret) in env; grandchild env row present.
- No CHANGELOG/LOGBOOK edits, no stray files, no new stateread/managed-write sites.
