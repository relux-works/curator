# Review brief: consume the landed overlay rule, cycle 2 (landing decision)

## Subject

- Branch `feat/consume-overlay-rule` at `e4ddca19` in
  `/Users/iv/Developer/ReluxWorks/.worktrees/curator-overlay-consume`, four signed commits past
  `origin/main`. PR https://github.com/relux-works/curator/pull/62, **all eleven hosted lanes green on
  this head**. The rework is `git diff 38702164..HEAD` — three files, +43/−20.
- Read `TASK-260906-19gjyw_review-findings-1.md` (your predecessor's F1–F3),
  `producer-brief-consume-rework-1.md`, and the rework report.
- Authority: curator-spec `87a0d006`.
- Change Request revision to accept: the one recorded for TASK-260906-19gjyw.

## Standing

Cycle 1 verified the substance and found no product defect. It confirmed by construction that
`canonicalGit` returns a bare `host/path` verbatim, that `ensureRepo` clones it over https, and that
the planted-directory attack reproduces — under a mutant dropping the widening, installing
`github.com/evil-org/pkg` with a planted local directory and an excluding allowlist **succeeds**. It
matched the discriminator against the committed schema across 121 spellings in three engines with zero
disagreements, and it re-ran 16 of 17 mutants. Do not re-derive that.

Its three findings were: a property test that excluded its own subject, a "complete" table missing one
class, and a doc comment on the wrong function. All three are addressed here.

## Review dimensions

1. **F1, and whether the repair is real.** `TestInstallNeverDemotesANetworkIdentityToAPath` used to
   survive the mutant that deletes the widening while logging `checked 5`. Apply that mutant yourself
   — delete the widening in `installOperandKind`, leaving `return identity.ClassifySource(trimmed)` —
   and confirm the test now **fails**. Then check the guard: the producer added a `bare` counter that
   fails if the matrix ever loses the bare-canonical spelling. Verify that guard actually fires by
   removing those operands from the matrix, and that it cannot be satisfied vacuously.

2. **The three artifacts that carried the false claim.** The test comment, the report's M10 row, and
   the `.github/ci/platform-cases.tsv` row that asserted a property no test asserted. Confirm each now
   says what is true. The ledger row matters most — AC row 6 requires the ledger rows to be truthful,
   and a durable ledger asserting an unheld property is worse than a stale comment.

3. **F2 — the added class.** `packages/team:context` and `a/b:c` were git→refused at `canonicalGit`
   before and are `path` now. Confirm the row is in the table, that `installOperandCases` pins the
   operand, and — the part worth attacking — that the enumeration is now genuinely complete. Your
   predecessor found the gap by restoring stage (c)'s `isPathOperand` beside the new helper and
   driving 57 operands through both. Do the same rather than reading the table, and say how many
   operands you drove.

4. **F3, and its neighbourhood.** The doc comment is moved. Check no other comment in the two
   implementation commits documents the wrong symbol — this epic has found a comment claiming a
   production path the code does not have three separate times.

5. **Regression surface.** The rework touches tests and comments. Confirm no product behaviour moved:
   the discriminator's classification is unchanged, and the install kind decisions are unchanged.
   `git diff 38702164..HEAD` should contain no change to a non-test code path except the doc comment;
   verify that rather than assume it.

6. **The landing question.** Say plainly whether PR #62 is safe to land. If it is, say so without
   hedging — cycle 1 already found the substance correct, and this cycle's subject is three
   corrections to evidence.

## Constraints

Scratch under a throwaway copy; never write into the producer's worktree or the control root. If you
materialize a conformance root, use a **plain checkout** verified against `manifest.json`, never
`git archive`. Anchor each `-run` level separately and count `=== RUN` lines — an anchored
`Parent/child` filter matches nothing and exits 0, which reads as a survivor.

## Verdict contract

Attach `TASK-260906-19gjyw_review-findings-2.md` with a `repeat-of:` line naming any class that
recurs. Blocking or major → set the task to `development`. Otherwise an explicit ACCEPT at
`to-review` with `accept_cr` on the recorded revision, stating what the hosted lanes reported on the
exact head you accepted. Do not mark the task done. Then
`task-board handoff TASK-260906-19gjyw --role reviewer`.
