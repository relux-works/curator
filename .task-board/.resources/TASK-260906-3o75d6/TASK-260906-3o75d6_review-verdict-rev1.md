# TASK-260906-3o75d6 review verdict rev1 — CHANGES REQUESTED (to-dev)

Candidate tree 30faf643 verified identical to the worktree (git diff empty).

## Blocking: normative content lost from the CLI guide (brief item 1, review-note check 1)
Before, each carrier row said: the flag takes over the files the op would write "(section 9.5 notice,
section 8.3 backup); without it the write fails with `environment_surface_unmanaged_conflict`".
The new note (cli/curator.md:54-60) copies only protocol/environments.md:2623-2629 up to "on no other
operation." and stops. Word-by-word comparison: the notice + backup behaviour and the
`environment_surface_unmanaged_conflict` failure without the flag are now stated NOWHERE in
cli/curator.md (`grep -n unmanaged_conflict cli/curator.md` -> no hits). The brief requires the note
to state "what the flag covers" completely; the review note requires "nothing lost". Both fail.
Also lost: env resolve's "`[--takeover]` applies only with `--repair`" row wording (note still names
`env resolve --repair`, acceptable, but mention it in the note/pointer if kept).

Fix: extend the note with the next §9.5 sentence(s) (environments.md:2629-2632: same notice and
backup as onboarding when the flag is given; without it §8.3 applies and the op fails with
`environment_surface_unmanaged_conflict`), and extend `takeover_cli_clause` regex in
tools/validate.py so the pin covers that sentence too (add a negative test dropping it).

## Minor
cli/curator.md:57-58 "named above as onboarding triggers" is a dangling reference copied from §9.5;
in the CLI guide nothing above names onboarding triggers. Say "named in environments section 9.5 as
onboarding triggers" (validator equality will need the same adaptation) or equivalent.

## Otherwise fine
Example moved into profile use group; import/global rows without --takeover; env resolve without
--repair not invented; validator pins look stricter than before (row counts, carrier mapping, pointer,
once-only clause), CHANGELOG entry present.

## Not rerun
Local python3 lacks jsonschema, so I could not rerun make validate / the 1xbrz6 negatives or my
mutations (drop carrier from note; add --takeover to import row). Next cycle reviewer must run them.
