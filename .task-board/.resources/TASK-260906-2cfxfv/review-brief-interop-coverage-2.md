# Review brief: interop coverage, cycle 2 (landing decision)

## Subject

- Branch `feat/interop-root-artifacts` at `e3a97d7b` in
  `/Users/iv/Developer/ReluxWorks/.worktrees/curator-interop-coverage`, four signed commits past
  `origin/main` (`fb916acd`). PR https://github.com/relux-works/curator/pull/63. The rework is
  `git diff 31106aa8..HEAD`.
- Read `TASK-260906-2cfxfv_review-findings-1.md` (your predecessor's F1, F1b, F2, F3),
  `producer-brief-interop-rework-1.md`, and the rework report.
- Authority: curator-spec `87a0d006`; the gate scripts are the contract.
- Change Request revision to accept: the one recorded for TASK-260906-2cfxfv.

## Standing

Cycle 1 verified the substance first-hand: the hole was real at base, all six negative runs reproduce,
the positive is green, the split loses nothing from the default lane, the class move
`root-content` → `root-unset` is right, and the six conformance ledger notes describe what their tests
assert. Do not re-derive that. Its findings were a skip scan that missed two receiver shapes, a gate
that tolerated a served `root-unset` skip, one undeclared unguarded read, and an overstated note.

## Review dimensions

1. **F1 — all four receiver shapes.** Re-inject each of the four skips your predecessor used —
   plain `t.Skipf`, the `guard := t` rename, the `h.t.Skipf` selector chain, and the
   `skip := t.Skipf` method value — into a real case, compile it, and confirm the scan now catches
   all four. Then invent shapes nobody has tried: a skip behind an interface, a skip in a helper in
   another file of the same package, `testing.TB` as the receiver type, a skip reached through a
   closure. Report anything that still evades. The claim in the ledger note must match exactly what
   the scan covers — no more.

2. **F1b — the gate change, and its blast radius.** `platform-case-gate.sh` now runs the
   `deferred-only` policy check inside the tolerated branch too. The producer answers "no existing
   row changes verdict" three ways: a static count of 70 tolerated rows with only `root-unset` being
   `deferred-only`; both recorded lane streams replayed through old and new gates with
   `skips-observed.tsv` byte-identical; and a fresh `SPEC_PIN` lane end to end at exit 0 with 33
   skips. **Re-measure at least the replay yourself** — this is a change to the one gate that
   enforces skip discipline across the whole repository, immediately before an rc. Then attack it:
   construct a ledger row and stream where the new branch fires and confirm it is right to fire, and
   one where it must not fire and confirm it does not.

3. **The three new `gate-selftest.sh` cases.** Read what each asserts and mutate the thing it
   protects to confirm each can fail. A self-test that cannot fail is the defect this leaf exists to
   close, one level up — and `gate-selftest.sh` has now grown twice in this leaf.

4. **F2 — the derivation.** The artefact declaration is now derived from the vectors' own path fields
   rather than hand-maintained. Verify it actually covers
   `expected/byte-exact-snapshot_sha256.txt` and the other 25 paths your predecessor enumerated, and
   that the derivation cannot silently return an empty or short set — if the vectors change shape,
   does it fail loudly or quietly cover nothing? Check the row, `crossReferencedTrees` and
   `ENV_REQUIRED` no longer drift independently.

5. **Regression surface.** Confirm the six negative runs still fail by name and the positive is still
   green after the rework, on roots materialized as plain checkouts verified against `manifest.json`.
   Confirm `golden_test.go` still serves what the pinned Go manager consumes.

6. **The landing question.** Say plainly whether PR #63 is safe to land. If it is, say so without
   hedging.

## Method

Anchor each `-run` level separately and count `=== RUN` lines — `go test -run '^(Parent/child)$'`
splits on the slash, matches nothing and exits 0, which reads as a survivor. Run the two `test-gate`
lanes sequentially, never alongside a `-race` suite. Never materialize a root with `git archive`.

## Constraints

Scratch under a throwaway copy or the worktree's `.temp/`; never write into the producer's worktree or
the control root.

## Verdict contract

Attach `TASK-260906-2cfxfv_review-findings-2.md` with a `repeat-of:` line naming any class that
recurs. Blocking or major → set the task to `development`. Otherwise an explicit ACCEPT at
`to-review` with `accept_cr` on the recorded revision, stating what the hosted lanes reported on the
exact head you accepted. Do not mark the task done. Then
`task-board handoff TASK-260906-2cfxfv --role reviewer`.
