# Review note — TASK-260910-31ocjt rev6 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Re-review rev6 (base 3f60f7f0, tree ef3cc7af, 13 paths, gate green on every lane — run 36523156113; on windows-latest
TestPrivateHTTPSBrokerAuthenticatesRealGitRepository, TestHTTPSBrokerNamedPipeHasCurrentUserOnlyDACL, …ServesOneClientAndThenRejectsAnother and
…ServerStopsWhenFetchExitsWithoutAClient PASS, not skipped) against your rev4 verdict (F1, F2), `31ocjt-rework-1.md` and `31ocjt-gatefix-3.md`.
Verify:
1. F1: the empty/drained-pipe row exists; M2 (drop `len(secret) == 0`) is killed — re-run yourself, real exit code.
2. F2 production code: Windows named pipe — crypto-random name, FILE_FLAG_FIRST_PIPE_INSTANCE, single instance, protected DACL with one
   ACCESS_ALLOWED ACE for the token user, serves one client then closes, server ends when the fetch child exits; only the NAME is in env;
   broker opens once, fails closed on any error. Unix keeps the inherited fd. No secret in argv/env/disk on either.
3. Test fixes are test-only: the DACL assertion walks ACEs with EqualSid against the token user (no string compare) and still kills a
   widened-DACL mutant (reason from the code; run it if you can build a Windows test binary — GOOS=windows go test -c — otherwise state
   that it ran on the hosted lane only); the real-git row uses http.sslBackend=openssl + a test-scoped CA, never sslVerify=false, and no
   production code path gained a test-only knob that could weaken TLS in production.
4. Diff rev4→rev6 contains only F1/F2 work; no CHANGELOG/LOGBOOK; no stray files; no new module dependency unless stated.
accept_cr or changes requested with file:line. No LOGBOOK.md. (This is the last revision in the loop budget: if a Windows-only item
remains, say whether it can be split into a separate leaf.)
