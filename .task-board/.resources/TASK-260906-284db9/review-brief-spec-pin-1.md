# Review brief: promote the released suite pin (cycle 1)

## Subject

- Branch `chore/promote-spec-pin` at `04550e28` in
  `/Users/iv/Developer/ReluxWorks/.worktrees/curator-spec-pin`, two signed commits past
  `origin/main`. PR https://github.com/relux-works/curator/pull/66. Diff:
  `git diff origin/main..HEAD` — six files, +243/−85.
- Read `producer-brief-spec-pin.md` and `TASK-260906-284db9_drafting-report.md`.
- Authority: curator-spec `v1.0.0-rc.11` = `87a0d0060bad64ab883d007dcdf35df7485368bf`, a signed
  released tag, and `release/1.0.0-rc.9.json` on that revision.
- Change Request revision to accept: the one recorded for TASK-260906-284db9. Note the delta will
  read empty — the Story workspace is a curator-spec checkout and this leaf delivers into curator.
  That is structural in this epic and is not the finding.

## What this changes and why

`SPEC_PIN` moves from `0ed5c691` (manifest `803918bf…`) to the tagged revision (manifest
`0e195ecd…`), which is the digest `release/1.0.0-rc.9.json` names in
`downstream_consumption.required_manifest_sha256` while `committed_release_pin_advanced` is still
`false`. The consequence is that the environments and schema-2 config families, which until now ran
only in a hand-dispatched candidate lane, run on **every default lane**.

## Review dimensions

1. **Verify the premise, do not take it from the PR body.** Compute all three manifest digests
   yourself — the old pin's root, the tagged root, and what `release/1.0.0-rc.9.json` requires — and
   confirm the tag is signed and its commit is what the pin now names. If the premise is wrong the
   whole leaf is wrong.

2. **The policy in `ci.yml` must still hold.** The pin carries only a qualified released revision and
   a candidate suite never enters through it. Confirm the rewritten comment is true in every clause —
   the previous comment's hashes and reasoning were entirely about the old pin, and a single stale
   sentence here is a durable lie in the file that defines the contract.

3. **The nineteen removed tolerances.** For each row that lost its `root-unset` column, confirm the
   package is genuinely served against the new root and the tolerance therefore could not fire.
   Then confirm the *inverse*: that no row lost a tolerance it still needs. Drive
   `suite-plan.sh` against the new root and against the `SPEC_PIN` root of the previous commit, and
   compare the partitions by name.

4. **The four rows that stayed — judge the scoping call.** `internal/skillspec`, `internal/marker`,
   `internal/moduleroots` and `internal/scriptpolicy` keep `root-unset`, on the reasoning that their
   families were published by the old pin too, so they lost no deferral in this move. That is
   defensible scoping, but notice what it implies: if nothing defers against the new root, those
   tolerances cannot fire either, and they were already dead before this change. This epic has found
   the "ledger row asserting something that cannot happen" class four times. Say whether leaving them
   is right, and if it is, whether the file now explains it well enough that the next reader does not
   have to re-derive it.

5. **The guard from PR #63 must survive.** Dispatch or reproduce locally: a candidate root missing an
   environments family must still fail **closed**, by name, not skip. The pin move must not weaken it.

6. **The measurement that matters.** Read `gh pr checks 66`. The default `Test` and `Race` lanes on
   all three runners must now **serve** the environments families and pass them — not defer them.
   Confirm from the uploaded gate evidence that they were served, not merely that the lane is green:
   a green lane that still defers would look identical from the outside and would mean the leaf did
   nothing.

7. **Everything the move made stale.** A skip reason no longer reachable, a comment describing a
   deferral that no longer happens, a `root-artifacts.tsv` row whose purpose has changed. The producer
   enumerated what it changed; check whether anything it did not change should have been.

## Method

Materialize roots as **plain checkouts** verified against `manifest.json`, never `git archive`. Run
the two `test-gate` lanes **sequentially**, never alongside a `-race` suite. Anchor each `-run` level
separately and count `=== RUN` lines.

## Constraints

Scratch under a throwaway copy or the worktree's `.temp/`; never write into the producer's worktree or
the control root.

## Verdict contract

Attach `TASK-260906-284db9_review-findings-1.md`. Blocking or major → set the task to `development`.
Otherwise an explicit ACCEPT at `to-review` with `accept_cr` on the recorded revision, stating what
the hosted lanes reported and whether the environments families were **served** on them. Do not mark
the task done. Then `task-board handoff TASK-260906-284db9 --role reviewer`.
