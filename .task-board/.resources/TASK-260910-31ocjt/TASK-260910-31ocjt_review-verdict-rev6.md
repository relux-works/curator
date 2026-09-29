# TASK-260910-31ocjt review verdict — rev6 (tree ef3cc7af, base 3f60f7f0): ACCEPTED

Worktree re-verified byte-identical to candidate tree ef3cc7af after mutants restored. Shell zsh, pipefail, real exit codes.

## Rev4 findings
- F1 fixed: TestHTTPSCredentialBrokerRejectsEmptyPasswordPipe (httpsbroker_test.go:150) drives RunHTTPSCredentialBroker with an empty pipe → exit 1, empty stdout. M2 (drop `len(secret) == 0`, httpsbroker.go:93): with row → FAIL exit 1 (killed); row removed → ok exit 0 (survives). Confirmed.
- F2 fixed (production): httpsbroker_pipe_windows.go — 32-byte crypto/rand hex name, FILE_FLAG_FIRST_PIPE_INSTANCE, max instances 1, PIPE_REJECT_REMOTE_CLIENTS, SDDL `D:P(A;;GA;;;<token user SID>)`, non-inheritable handle, ConnectNamedPipe once then write+close, Serve cancelled when cmd.Wait returns (admission.go runCommandWithHTTPSSecret), only the name in env; broker validates name grammar, opens once, fails closed on any error. Unix keeps inherited fd via ExtraFiles, parent reader closed after Start (ChildStarted); drained/second read → empty → fail closed.
- DACL test (pipe_windows_test.go:30) walks ACE via GetAce + SID.Equals against token user, checks protected + exactly one ACCESS_ALLOWED ACE + forbidden Everyone/Users/AuthUsers; a widened DACL (Everyone ACE) fails AceCount!=1 — reasoned from code; Windows execution on hosted lane only (run 36523156113 PASS). Windows test binary compiles locally (GOOS=windows go test -c exit 0; go vet exit 0).
- Real-git row uses `-c http.sslVerify=true` + `http.sslBackend=openssl` (Windows only) + test-scoped sslCAInfo; test-only, no production TLS knob.

## Reruns (local darwin)
- go test ./internal/buildrepo -run 'Broker|HTTPS|Askpass' → exit 0
- go test ./internal/crossconformance -run Broker → exit 0
- M-env (secret re-added to fetchEnv in admission.go) → TestSelectedHTTPSFetchSecretUsesPipeAndNeverEntersTheProcessEnvironment FAIL, exit 1 (killed)

## Hygiene
- Env var constant removed; remaining mentions are negative assertions in tests, the audit finding text + "Remediated by" line, and a historical LOGBOOK entry (not edited).
- No go.mod change (x/sys/windows already a dependency). No CHANGELOG/LOGBOOK edits. 13 paths, all product/test/docs.

## Bounds / residuals
- Unix: fd is inherited by every git descendant until the helper drains it; a sibling descendant reading first could obtain it (one-shot, then empty). Stated bound, not a blocker.
- Windows pipe name is visible in env to descendants; protected by DACL (same user) + single-instance/one-shot. Same-user processes remain in trust boundary.