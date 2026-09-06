# Review brief: path overlay declarability, cycle 4 (final)

## Subject

- Branch `feat/path-overlay-declarable` at `18dca85`, three signed commits past curator-spec main.
  PR https://github.com/relux-works/curator-spec/pull/47, all eight checks green. The cycle-3 delta
  alone is `git diff 2f2dfa4..18dca85`.
- Read `TASK-260906-2x4s7i_review-findings-3.md` (your predecessor's cycle-3 findings F10–F15) and
  the earlier verdicts and mutant sweeps attached to this element.
- Authority: `protocol/environments.md` §1, §6, §12.1; `protocol/core.md` §6.1; `profiles/manager.md`;
  `schemas/v1/manager-config-v2.schema.json`.

## Standing, and why this cycle is narrow

Cycle 3 verified **all five** cycle-2 findings genuinely fixed and killed 24 of 30 mutants on a named
case, with two of the six survivors proven semantic no-ops. Its remaining findings were: a 5.4 MB
compiled `generate-vectors` tracked at the repository root (F10, blocking), a stale PR description
(F11), four unpinned classification decisions (F12), a provably dead carve-out in arm 3 (F13), an
over-long manager line (F14), and bounds stated nowhere (F15).

The **orchestrator** made all of those repairs, as it did the cycle-2 ones. Review them as
adversarially as a producer's work — the author is the same party that wrote this brief.

Do not re-derive the cycle-2 fixes your predecessor already drove. Attack the cycle-3 delta and
whatever it disturbed.

## Review dimensions

1. **F10, and whether the class is actually closed.** `generate-vectors` is untracked and
   `/generate-vectors` is ignored root-anchored. Verify the tracked file set is clean —
   `git ls-files` against a fresh clone of the head, not against a dirty worktree — and that no other
   build artefact rides along. Then ask the harder question: `.gitignore` now names two artefacts
   discovered by two separate accidents, and **no lane inspects the tracked file set**, which is why
   nine green checks passed with a 5.4 MB binary present. Say whether that gap deserves a gate of its
   own, and if so what it would assert.

2. **F13 — the arm 3 simplification.** The drive carve-out was removed from arm 3 on the reasoning
   that the Windows-drive pattern demands a slash or backslash after the colon while the SCP pattern
   forbids both, so their intersection is empty. Verify that reasoning is exactly true rather than
   nearly true, and confirm arm 2's carve-out is still load-bearing — mutate it away and show a named
   case dying. A "provably dead" removal that turns out to be live is a silent behaviour change.

3. **F12 — the two newly pinned decisions.** `packages/team:context` is now a valid path and
   `github.com:\example\x` is refused. Verify both are what §1 and core §6.1 require, not merely what
   the current pattern happens to do, and re-run the two mutants (unanchored colon test, backslash
   admitted after the SCP colon) plus your predecessor's remaining survivors. F12 listed five named
   flips; two are now pinned — say what happened to the other three and whether leaving them unpinned
   is defensible.

4. **F11 and F15 — the description and the bounds.** The PR body was rewritten and now carries a
   stated-bounds paragraph. Check every number in it against the head (vector count, overlay case
   count, spellings) and check the bounds are the real ones: that the schema decides the kind by
   spelling because §1 supplies no marker, that `c:example/x` is git on §6.1's one-character host,
   that a `file:` URL is refused as neither kind, and that the schema layer decides admissibility and
   not which §1.1 diagnostic an implementation reports. Flag anything overstated.

5. **F14, and the whole delta's mechanics.** Three signed commits past `origin/main`, human identity,
   no stray file. Re-run `make validate` and `make regenerate-check` yourself. Confirm from the pinned
   manager's source that nothing it consumes gained an unpassable case.

6. **The landing question.** Say plainly whether PR #47 is safe to land, and if it is not, name the
   one thing that must change.

## Constraints

Read-only outside your scratch. Never write into the control root, and do not push or modify the
branch.

## Verdict contract

Attach `TASK-260906-2x4s7i_review-findings-4.md` with a `repeat-of:` line naming any earlier class
that recurs. Blocking or major → set the task to `development`. Otherwise an explicit ACCEPT at
`to-review`. Do not mark the task done. Then `task-board handoff TASK-260906-2x4s7i --role reviewer`.
