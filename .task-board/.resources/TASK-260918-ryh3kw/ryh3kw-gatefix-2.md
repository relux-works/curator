# TASK-260918-ryh3kw — gate fix 2 (THE ONLY CURRENT INSTRUCTION, with ryh3kw-sec-brief.md, ryh3kw-decision-1.md)

Good progress: revision 3 is green everywhere except ONE Windows case (run 36303430458, test-evidence-windows-latest go-test.json):
internal/envprofile TestEnvironmentsReadFailureVectors/passthrough-parent-not-directory-unreadable — "unreadable status findings =
[environment_passthrough_detached: passthrough entry auth.json is detached]". On Windows a path whose parent component is a regular file
returns ERROR_PATH_NOT_FOUND (maps to not-exist), so it is reported as absent/detached instead of unreadable. Fix it portably in
internal/stateread: when the leaf read reports not-exist, Lstat the nearest existing ancestor; if an ancestor component EXISTS AND IS NOT A
DIRECTORY (and is not a link to a directory), classify unreadable (not absent). Only genuinely missing components remain absence
(ERROR_FILE_NOT_FOUND / ERROR_PATH_NOT_FOUND with every existing ancestor a directory). Keep the ENOTDIR fast path on unix. Add a unit row in
internal/stateread for this on all platforms and kill a mutant (ancestor walk removed → the Windows lane row must fail; show the unix
unit-row failure locally with a real exit code). The published case must be driven, not ledgered.
Trunk is d41da0fb (your base is still 0ffe2e1d — the previous brief asked for this): `git fetch origin main`, combine keeping both sides;
the 2 restore vectors parked for TASK-260927-1wc76r (landed in 38c68570) must be driven now or re-attributed to their real blocker. VERIFY
`git diff --name-only origin/main -- . ':!.task-board'` lists only your paths. Set status development first; append "Revision 4 — Windows
blocked-parent is unreadable"; handoff and WAIT for the gate (never interrupt); hand off only green. No CHANGELOG/LOGBOOK edit.
