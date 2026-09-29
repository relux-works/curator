# TASK-260910-31ocjt — rework 1 (THE ONLY CURRENT INSTRUCTION, with 31ocjt-brief.md)

Review of rev4 (TASK-260910-31ocjt_review-verdict-rev4.md) requested two changes. Read that verdict first.

F1: add a row that invokes the broker Password prompt twice on one pipe, or once on an empty or already-drained pipe. Assert that the
second or empty call exits 1 with empty stdout. Show with real exit codes that mutant M2 (drop the `len(secret) == 0` fail-closed check in
internal/buildrepo/httpsbroker.go) survives before the row and is killed after it.

F2 (Windows): Git for Windows restricts handle inheritance to stdio (core.restrictInheritedHandles, PROC_THREAD_ATTRIBUTE_HANDLE_LIST).
An inherited pipe handle most likely never reaches git-remote-https → askpass. Your current "every platform" row simulates the one step real
git does not perform, so it is not evidence. Implement the brief's Windows option:
- a named pipe with a private DACL (current user only; no Everyone, Users or Authenticated Users);
- an unguessable name (crypto random), created with FILE_FLAG_FIRST_PIPE_INSTANCE and a single instance;
- serve the secret to exactly one client connection, then close. The server goroutine ends when the fetch child exits, even if nobody
  connected;
- only the pipe NAME travels in the environment, never the secret. The broker opens it, reads once, and fails closed on any error.

Use golang.org/x/sys/windows if it is already a dependency; do not add a new module dependency without saying so in the results.
Unix keeps the inherited-fd design.

Rows:
- Windows: a real-git row that fetches or ls-remotes an authenticated HTTPS repository through the real git.exe → askpass chain on the
  hosted windows lane. Reuse TestPrivateHTTPSBrokerAuthenticatesRealGitRepository and remove its Windows skip. If the hosted Windows git
  cannot run it, register the skip in skip-classes.tsv + platform-cases.tsv with the exact technical reason, and state Windows as
  unverified in the results. Do not claim cross-platform.
- A second client connecting to the pipe gets nothing.
- The pipe DACL has no world/user-group ACE. Check it with GetSecurityInfo on the created handle.

Mutants with real exit codes: DACL widened to Everyone; pipe serves twice. Each must be killed on the lane where its row runs. Report which
lane.

Also add a one-line "remediated by TASK-260910-31ocjt" note next to docs/security-audit-2026-09.md:196, if that file keeps per-finding status
lines.

Bounded local runs (go test ./internal/buildrepo ./internal/crossconformance; GOOS=windows go vet ./internal/buildrepo). The hosted gate is
the arbiter. No CHANGELOG/LOGBOOK edit. Set status development, update results, `task-board handoff TASK-260910-31ocjt --role developer`,
END YOUR TURN. The runner publishes the CR and runs the gate; do not wait for it.
