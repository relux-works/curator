# Review brief: cli/curator.md takeover and onboarding-import rows (cycle 1)

## Subject

- Story branch `task-board/story/STORY-260905-2z9pw4` at `f013e0c`, exactly one signed commit past
  curator-spec main `f39f4a9309f41a9208da817eba9129cf5a9f8dc0`. The spawn puts you in the managed
  workspace. Diff: `git diff f39f4a9..f013e0c`. Change Request revision to accept: the one recorded
  for TASK-260906-1hn93j (`task-board worktree status`).
- Read `producer-brief-cli-takeover-import.md` and `TASK-260906-1hn93j_drafting-report.md` first. The
  report's row → sourcing table and its five open questions Q1–Q5 are the producer's own account of
  what it could and could not source. Judge that account; do not assume it.
- Authority: `protocol/environments.md` at `f39f4a9` — §8.3, §9.1, §9.2, §9.3, §9.5, §9.6, §9.7 — and
  the existing `cli/curator.md` table conventions and flag family.

## Review dimensions

1. **Sourcing, attacked clause by clause.** For every clause of both new rows, find the
   `environments.md` sentence that carries it, or record that none does. The producer claims each
   clause is sourced except five spellings it marks CHOICE. Verify the claim and, just as important,
   verify the marking: a clause presented as sourced whose sentence does not actually say it is a
   worse defect than an honestly-marked choice. Check `environment_surface_unmanaged_conflict` and
   `environment_import_lossy` are spelled as the diagnostics tables spell them.

2. **The `env takeover --takeover` shape — the orchestrator's own concern, test it hard.** The
   published row is a subcommand named `takeover` that additionally requires a flag named
   `--takeover`. Read §9.5 and decide which of these the spec actually describes:

   - (a) a standalone `env takeover` operation that happens to need a confirming flag;
   - (b) a `--takeover` modifier on the mutating operations that already trigger onboarding —
     `profile install`, `profile use`, `profile sync`, `profile update`, `env resolve --repair` — so
     that an operation refused by §8.3 with `environment_surface_unmanaged_conflict` proceeds on a
     re-run under the flag, with the §9.5 notice and backup.

   The orchestrator's reading is (b): §9.5's trigger list names "an explicit takeover" beside those
   operations rather than attaching it to a command of its own; §8.3 frames takeover as what makes an
   otherwise-refused write legal; and (b) dissolves the producer's Q2 entirely, because the scope of
   the takeover is the scope of the operation performing it — no new operand grammar is needed.
   Under (a) the tautology stands and Q2 stays open with no sentence to close it.

   Attack that reading rather than adopt it. If the spec genuinely fixes (a), say so with the
   sentence that fixes it. If it fixes neither, say that too: the answer then is that the CLI index
   cannot publish the row until `environments.md` gains the sentence, and the finding is that a row
   was published on an unsourced structural choice.

3. **The remaining spellings.** `curator profile import`, `--as <name>`, `--allow-lossy`. Judge each
   against the published flag family (`--purge`, `--repair`, `--use`, `--restore-backups`, `--as`,
   `--allow <hash>`) and against §9.6. `--as` reusing the install row's spelling and `imported` as the
   default name are directly supported; say whether `--allow-lossy` and the `profile` placement are
   the best available spellings or merely acceptable ones.

4. **Surface-index discipline.** `cli/curator.md` indexes operator surface; it is not a second
   normative source. Confirm neither row introduces a rule, default, or behaviour that
   `environments.md` does not already state — in particular that the takeover row's silence on a
   default scope, and the import row's account of the consent gate, add nothing new. Check the
   producer's "deliberately not published" list is genuinely out of surface scope rather than a
   convenient omission.

5. **Mechanics.** Exactly one signed commit past `f39f4a9`, human identity, no `LOGBOOK.md` or stray
   file (`git show --stat f013e0c`). `CHANGELOG.md` entry follows the convention the previous CLI
   batch (`f61ee9a`) used. `make validate` and `make regenerate-check` green — re-run them yourself
   (venv under `.temp/venv`), do not trust the report's tails. The examples block lines match the
   rows they illustrate.

## Constraints

Read-only outside your scratch (`.temp/` inside the workspace). Never write into the control root.

## Verdict contract

Attach `TASK-260906-1hn93j_review-findings-cli-1.md` — severity, file/section, the quoted row clause,
what is wrong, the fix, and a clause → source-sentence table covering both rows in full. Blocking or
major → set the task to `development`. Otherwise an explicit ACCEPT at `to-review` with `accept_cr` on
the recorded revision. Do not mark the task done. Then `task-board handoff TASK-260906-1hn93j --role
reviewer`.
