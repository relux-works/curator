# Review brief: cli takeover and import rows, cycle 2 (rework 1)

## Subject

- Story branch `task-board/story/STORY-260905-2z9pw4` at `e01de3f`, one signed commit past
  curator-spec main `f39f4a9309f41a9208da817eba9129cf5a9f8dc0` (the cycle-1 commit `f013e0c` was
  squashed away, as required). Diff: `git diff f39f4a9..e01de3f` — 4 files, +49/−13. Change Request
  revision to accept: the one recorded for TASK-260906-1hn93j (`task-board worktree status`).
- Read `TASK-260906-1hn93j_review-findings-cli-1.md` (your predecessor's cycle-1 findings),
  `producer-brief-cli-takeover-rework-1.md` (the orchestrator's decision and widened scope), and
  `TASK-260906-1hn93j_rework-report-1.md`.
- Authority: `protocol/environments.md` and `profiles/manager.md` **as amended by this commit** for
  the published rows, and at `f39f4a9` for judging whether the amendment is faithful.

## What changed and why

Cycle 1 established that the standalone `curator env takeover --takeover` shape is excluded by §9.5,
not merely unsourced, and that fixing the CLI index alone would still publish an enumeration the spec
does not state. The orchestrator therefore widened the scope: close the gap in `environments.md` §9.5
and the mirrored sentence in `profiles/manager.md`, then publish the rows on the closed sentence.

This is now a **normative amendment**, so review it as one.

## Review dimensions

1. **Is the §9.5 amendment faithful and minimal?** It must fix exactly three things — the takeover is
   a flag carried by a mutating operation and never an operation of its own; the flag is accepted on
   exactly the five operations §9.5 already names as onboarding triggers and on no other; the takeover
   covers only the unmanaged files the carrying operation would write and never selects a scope. Check
   it fixes all three and nothing more. In particular verify that after the amendment §8.3's "a second
   takeover of the same path" and the surviving "a specific unmanaged file" reading both still read
   true, that `environment_surface_unmanaged_conflict` is the only diagnostic named and no new one was
   invented, that no other section's meaning shifted, and that nothing in §9.6, §8.3, §7.6 or Decision
   0010 now contradicts the amended text. Re-run your predecessor's `grep -rn takeover` over
   `protocol/`, `profiles/`, `decisions/` and confirm every remaining hit is consistent with the new
   shape.

2. **Does the manager sentence agree without restating?** The manager profile states manager-side
   obligations and cites environments sections; it is not a second normative source. Judge whether the
   amended manager sentence agrees with §9.5 exactly and whether its voice is right.

3. **The five carrying rows.** The producer published `[--takeover]` on `profile install`,
   `profile use`, `profile use --clear`, `profile update`, `profile sync` and `env resolve`. That is
   **six** rows for five enumerated operations, because `profile use` has two published forms. Judge
   whether `profile use --clear` can meet unmanaged files at all — it re-materializes the scope from
   the machine default, so it writes — and whether the amended sentence covers it. If it does not,
   that is a finding. Check `env resolve`'s clause states the `--repair` relationship exactly as the
   amendment does, and that no row published the flag that the enumeration does not cover.

4. **The import row.** `[--use]` was added per cycle-1 finding 3 with the §9.1 activation clause.
   Verify the clause matches what the install row already publishes and what §9.1 states. Then ask the
   question the enumeration raises: `profile import` installs through the §9.1 `path` pipeline, so can
   it meet unmanaged files, and if so why does it carry no `[--takeover]`? §9.6 says "The import writes
   nothing into any native home by itself" — decide whether that sentence settles it, and say so
   either way.

5. **Surface-index discipline, again.** Six rows now repeat a near-identical takeover clause in a table
   whose convention is one line per command. Judge whether the repetition states anything
   `environments.md` does not, and whether it is the right shape for this document or should be one
   note. This is a judgement call — make it and justify it; do not simply pass it.

6. **Mechanics.** Exactly one signed commit past `f39f4a9`, human identity, no `LOGBOOK.md` or stray
   file (`git show --stat e01de3f`). `CHANGELOG.md` places the normative amendment and the CLI-surface
   change under the headings the repository's convention gives each. `make validate` and
   `make regenerate-check` — **re-run them yourself** (venv under `.temp/venv`), do not trust the
   report's tails. Confirm the producer's statement about whether the amendment touches any schema,
   vector, diagnostic table or conformance case, by checking rather than by reading its claim.

7. **The cycle-1 minor.** Your predecessor found the producer's verification bound was proxy-derived
   and one sub-claim false. Judge whether this report's bounds are established by methods that
   actually establish them.

## Constraints

Read-only outside your scratch (`.temp/` inside the workspace). Never write into the control root.

## Verdict contract

Attach `TASK-260906-1hn93j_review-findings-cli-2.md` with a `repeat-of:` line naming any cycle-1
finding whose class recurs. Blocking or major → set the task to `development`. Otherwise an explicit
ACCEPT at `to-review` with `accept_cr` on the recorded revision. Do not mark the task done. Then
`task-board handoff TASK-260906-1hn93j --role reviewer`.
