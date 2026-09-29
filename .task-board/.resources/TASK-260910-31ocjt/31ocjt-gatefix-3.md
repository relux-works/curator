# TASK-260910-31ocjt — Windows gate fix 3 (THE ONLY CURRENT INSTRUCTION, with 31ocjt-rework-1.md)

Rev5 (tree 4c5d6639) is green everywhere except Test (windows-latest) (run 36517990469). There are two test-side failures, and both are
diagnosed below. This is the LAST revision in the loop budget, so fix exactly these two and nothing else.

1. `TestHTTPSBrokerNamedPipeHasCurrentUserOnlyDACL` (httpsbroker_pipe_windows_test.go:56) fails with
   `named-pipe DACL does not grant only the current user: "D:P(A;;FA;;;LA)"`.
   The DACL is actually correct. The hosted runner account is the built-in local Administrator, and SDDL renders that SID with the alias
   `LA`, so a string comparison against the user's S-1-5-… form fails. Fix the assertion, not the production code:
   - walk the ACL (GetAce / ACCESS_ALLOWED_ACE);
   - assert that the DACL is protected, has exactly one ACCESS_ALLOWED ACE, and that the ACE's SID is EqualSid with the process token user
     SID (windows.GetCurrentProcessToken().GetTokenUser());
   - assert that there are no Everyone/Users/Authenticated Users ACEs.

   Keep the widened-DACL mutant killed: re-run it with the same logic.
2. `TestPrivateHTTPSBrokerAuthenticatesRealGitRepository` (httpsbroker_test.go:305) fails with
   `schannel: SEC_E_UNTRUSTED_ROOT`. Git for Windows uses the schannel backend by default, and schannel ignores http.sslCAInfo, so the
   test's self-signed CA is not trusted. In the TEST ONLY, run the real git with `-c http.sslBackend=openssl -c http.sslCAInfo=<test CA>`
   (or the GIT_CONFIG_COUNT/KEY/VALUE env equivalent) on Windows. Keep the CA scoped to the test; do not disable verification
   (no sslVerify=false).

   If the production fetch path builds the git argv in a way the test cannot pass `-c` to, add a test-only seam and say so.

   This row is the Windows real-git evidence the review asked for. It must actually run and pass on windows-latest: no skip.

Local runs:
- `GOOS=windows go vet ./internal/buildrepo`;
- `GOOS=windows go test -c -o $TMPDIR/b.exe ./internal/buildrepo` (must compile);
- `go test ./internal/buildrepo -run 'Broker|HTTPS'` with real exit codes.

The hosted Windows lane is the arbiter. Update the results, then `task-board handoff TASK-260910-31ocjt --role developer`, then END YOUR
TURN. No CHANGELOG/LOGBOOK edit.
