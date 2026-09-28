# TASK-260906-3o75d6 — state the takeover clause once in cli/curator.md (curator-spec, editorial)

Control root: curator-spec; work only in your assigned Story worktree. Read `campaign-producer-rules.md`
first. Board validation command (`make validate`) runs once at handoff.

Context: review cycle 2 of TASK-260906-1hn93j left editorial observations, deliberately deferred so
the normative fix stayed minimal. The sibling leaf TASK-260906-1xbrz6 just landed (curator-spec
`eadb1c0`): §9.4/§9.5/§9.6 and manager §12.3 now state that `profile import` activation and the
global operations carry no takeover flag, and `tools/validate.py` pins those exact sentences and the
rule that no `profile import`/`global` CLI row carries `--takeover`.

Deliverable (no normative change):
1. `cli/curator.md`: the ~25-word takeover clause currently repeats across the carrying rows of the
   table whose convention is one line per command. State it ONCE, completely (the five carriers, the
   closure, what the flag covers), in the place the table's conventions call for — a note under the
   table — and reduce each carrying row to the flag plus a pointer. Add no rule that environments.md
   does not state.
2. Move the takeover example into the example group of the operation it illustrates (`profile use`).
3. Do not invent `env resolve --takeover` without `--repair` (it stays unstated, as the review noted).
4. Keep every `tools/validate.py` pin from 1xbrz6 green (the CLI no-takeover rule for import/global
   rows and the carrier list); if a pin reads the per-row clause, adapt the pin to the single note
   WITHOUT weakening what it proves — say exactly how.
5. CHANGELOG (unreleased, editorial).
Attach `TASK-260906-3o75d6_results.md` (before/after, how each pin still holds, `make validate` exit
code) and hand off with `task-board handoff TASK-260906-3o75d6 --role developer`.
