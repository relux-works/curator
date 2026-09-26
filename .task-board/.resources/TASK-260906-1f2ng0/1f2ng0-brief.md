# TASK-260906-1f2ng0 — stage (a) follow-ups from reviews (curator)

Control root: curator; work only in your Story worktree (STORY-260905-1n0iy8). Read `campaign-producer-rules.md`.
Landing gate = hosted CI, once, at handoff. Source: `.task-board/.resources/TASK-260905-30zs8t/TASK-260905-30zs8t_review-findings-stage-a-5.md`
("Follow-up, not a cycle") and this task's description (FU-1..FU-4).

For EACH follow-up, either implement it with tests or decline it with a recorded reason — no silent skips:
- **FU-1** `profile_source_path_missing` / `profile_source_path_unreadable` (environments §1.1): the diagnostics
  now exist in `internal/envprofile/envprofile.go` (`DiagPathMissing`, `DiagPathUnreadable`, `pathManifestDiag`).
  VERIFY at the production CLI entry that a missing path operand yields `…_missing`, a chmod-000 directory yields
  `…_unreadable`, and neither fires on the other case (four rows including the regular-file case). If already
  covered by committed tests, cite them; otherwise add the rows.
- **FU-2** `loadMachinePolicy` (`envprofile.go` ~1616): judge absence vs read failure against §8.4 — absent file
  ⇒ default policy is correct; any other read error must refuse. Prove both with tests (fix only if wrong).
- **FU-3** operator warning when a global skill does not migrate into the default profile lock (cycle-2 note):
  implement the warning with the exact wording the spec gives (search environments §9.4 for the migration rule),
  or record why the spec does not ask for it.
- **FU-4 / cycle-5 FU-2** a partial install whose activation fails before `materializeScope` prints no
  "installed profile" line (`cmd/curator/profile.go` install branch). NOTE: TASK-260907-187z6x (in flight) is
  changing the SAME branch so a partial reinstall prints `updated profile …`; keep your change compatible
  (print the line for a partial first install; do not undo the reinstall wording) and say how.
Tests at the production entry, one narrowing mutant per implemented behaviour, CHANGELOG. Attach
`TASK-260906-1f2ng0_results.md` and hand off with `task-board handoff TASK-260906-1f2ng0 --role developer`.
