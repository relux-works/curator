# TASK-260720-12r55p stop-the-line packet

## Constraint and evidence

The task simultaneously requires:

1. consumption of the exact signed `1.0.0-rc.5` suite at curator-spec commit
   `f5d7673039226ab81de2f4f87e2155ae995c4df3`, whose
   `conformance/v1/manifest.json` hashes to
   `sha256:b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`; and
2. manager-lifecycle coverage for ordering, commit, rollback, recovery, status,
   repair, and GC.

Those requirements cannot be satisfied from the same root. The signed rc.5
root's `vectors/manager-lifecycle.json` hashes to
`sha256:2ddbd2665a63f44dc0e03e060f4cd34bfde219a56b3192511fe1ef81047feedf`
and contains exactly ten cases in four clusters:

- `bootstrap_cases`: 3
- `dry_run_cases`: 2
- `launcher_cases`: 2
- `upgrade_cases`: 3

It contains no ordering, commit, rollback, recovery, status, repair, or GC
clusters. Searching the exact manifest-covered rc.5 suite finds those later
surfaces only in `vectors/external-repository-lifecycle.json`, which the task
explicitly assigns to the out-of-scope `go-repository-v1` line.

The hard prerequisite `TASK-260720-3ag6pi` is now board-status `done`, but its
accepted cycle-4 verdict verifies a subsequent rc.6 candidate: it reports 32
manager-lifecycle cases and explicitly says rc.6 was not committed, signed,
tagged, or published. It does not change the immutable rc.5 root or provide an
authorized rc.5 replacement root for this task.

The rc.5 release record itself is internally consistent: its
`downstream_consumption.required_manifest_sha256` is the required `b6f56aac...`
digest. The annotated `v1.0.0-rc.5` tag verifies successfully and peels to the
commit above.

## Attempts and observations

- Verified all three stored hard dependencies are accepted `done` and the
  board currently computes `isBlocked=false`.
- Fast-forwarded the clean CocoaSkills main clone and recorded the signed base
  `870daa30aea0ed4dc5554ac5dcd0c671f8d04e09`.
- Created isolated task worktrees for CocoaSkills and the signed rc.5 spec so
  concurrent curator-spec regeneration cannot change the oracle.
- Ran the pre-change rc.5 conformance command directly. It truthfully exits 1:
  the current harness asks for the later rc.6-only
  `expected/marker-v2.json`. This mismatch is ordinary implementable rework;
  the absent lifecycle vectors are the stop-the-line issue.
- No tracked CocoaSkills product or test file was changed.

## Viable options

1. **Retarget to an immutable reviewed rc.6-or-later candidate (recommended if
   the expanded lifecycle AC is required).** Supply its exact revision,
   manifest digest, and release/candidate metadata, update this task's rc.5
   pin/version assertions, and relink/review the verification prerequisite.
   This preserves root-driven independent consumption but changes the task's
   protocol target.
2. **Keep exact rc.5 and narrow the lifecycle AC to the ten published cases.**
   This preserves `b6f56aac...` and the signed rc.5 evidence, but drops the
   required ordering/transaction/recovery/status/repair/GC coverage.
3. **Split the lines.** Authorize this task to consume only the exact rc.5
   surfaces, and create/relink a successor task for the expanded lifecycle
   suite once an immutable reviewed root exists. This preserves both evidence
   histories but requires an explicit AC/dependency edit.

Combining rc.5 vectors with an rc.6 lifecycle file, copying the out-of-scope
external-repository vector, or hard-coding the missing cases would violate the
single-root manifest contract and the explicit no-fabrication/no-duplication
constraints.

## Exact input required

Board/product authority must choose one option and update the task's scope and
acceptance criteria accordingly. If option 1 or 3 is chosen, provide the exact
immutable candidate root revision and full manifest digest and relink the
reviewed verification gate before implementation resumes.
