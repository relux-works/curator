# TASK-260925-h4syhu — gate fix after the cww1ov merge (THE ONLY CURRENT INSTRUCTION)

Revisions 4 and 5 (identical tree d7d2b7f8) fail ONLY on windows-latest:
  internal/install TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts/lstat_failure_stops_fallback (draftsources_test.go:90):
  lockedNetworkRepository error = "source_snapshot_unavailable: cannot inspect repository checkout for blocked\child: manager_state_unreadable: …"
The Lstat failure IS now classified as manager_state_unreadable on Windows, but the merged code (0017/cww1ov wrapping in
internal/install/draftsources.go) re-labels it as source_snapshot_unavailable. Operator addendum (lock replay): `source_snapshot_unavailable`
is ONLY for an unreachable source; an unreadable local checkout is manager state that cannot be read — it must surface as the typed
manager_state_unreadable (no fallback), not as source_snapshot_unavailable.
1. Decide from the spec text (core §8.4 absence vs read failure; skillfile-sources §3 replay codes) which outer code a failed Lstat of the
   checkout must carry; fix PRODUCTION so the typed manager_state_unreadable is what callers see (errors.As works through any wrapping),
   on all OSes; keep cww1ov's other wrapping behaviour intact. State the spec clause you relied on.
2. Keep the test's intent (typed unreadable error, no fallback) — do not weaken the assertion to accept source_snapshot_unavailable.
3. Re-kill M1 (continue first in the Lstat error branch) on darwin; `GOOS=windows go vet ./internal/install`; run the three subtests + the
   guard test locally (bounded). Real exit codes.
4. `task-board m 'set_status(TASK-260925-h4syhu, status=development)'` first; before handoff `git fetch origin main` — if trunk moved
   (2elcdc, 20o9dk, 306v4m landed), combine that diff too (keep both sides) and `task-board worktree refresh-candidate TASK-260925-h4syhu`.
   Append "Revision 6 — unreadable checkout code after cww1ov merge", `resource update`, `task-board handoff TASK-260925-h4syhu --role developer`;
   stay in the turn while the gate runs. If the loop detector refuses, stop and report. No CHANGELOG/LOGBOOK edit.
