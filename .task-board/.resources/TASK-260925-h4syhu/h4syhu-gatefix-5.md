# TASK-260925-h4syhu — last Windows assertion + CHANGELOG revert (THE ONLY CURRENT INSTRUCTION)

Revision 6 is right in production (lockedNetworkRepository returns the typed manager_state_unreadable, environments §8.4.1). Two issues left:
1. windows-latest: draftsources_test.go:91 lstat_failure_stops_fallback — the error IS manager_state_unreadable, but the assertion demands an
   "underlying Lstat failure"; on Windows Go's Lstat reports op GetFileAttributesEx (ERROR_PATH_NOT_FOUND under a regular-file parent).
   Make the assertion platform-correct: errors.As to the typed stateread error + the checkout path + "not absence / no fallback"; accept the
   OS-specific underlying *fs.PathError op (lstat on POSIX, GetFileAttributesEx on Windows) — do NOT drop the underlying-cause check, only
   stop hard-coding the op name. M1 must still be killed.
2. CHANGELOG.md is in the candidate (the replayed 187z6x checkpoint carries its old entry): make CHANGELOG.md equal trunk's bytes; keep the
   entry text in results under "## CHANGELOG entry (for release prep)" (it is probably already there — verify).
3. `GOOS=windows go vet ./internal/install`; bounded darwin run of the three subtests + M1 kill. Real exit codes.
4. `task-board m 'set_status(TASK-260925-h4syhu, status=development)'` first; `git fetch origin main` — if trunk moved (2gt5f6 landing,
   rc.2 prep), combine (keep both sides) and `task-board worktree refresh-candidate TASK-260925-h4syhu`. Append "Revision 7 — platform-correct
   assertion", `resource update`, `task-board handoff TASK-260925-h4syhu --role developer`; stay in the turn. If the loop detector refuses,
   stop and report. No LOGBOOK edit.
