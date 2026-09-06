# Review brief: stage (c) cycle 2 (rework 1)

## Subject

- Branch `feat/agent-environments-stage-c` at `0bcea201` in
  `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`, 16 signed commits past `origin/main`.
  PR https://github.com/relux-works/curator/pull/61 — pushed, so hosted lanes run on the exact head
  you review. Rework 1 is `git diff 833918d2..HEAD`; the full stage is `git diff origin/main..HEAD`.
- Read `TASK-260906-1uf713_review-findings-stage-c-1.md` (your predecessor's cycle-1 findings) and
  unpack `TASK-260906-1uf713_review-probes-stage-c-1.tgz`; then
  `producer-brief-stage-c-rework-1.md`, `orchestrator-finding-stage-c-windows-heuristic.md`, and
  `TASK-260906-1uf713_rework-report-1.md`.
- Authority: curator-spec `550579d` — `protocol/environments.md` §1, §6, §6.1, §9.1, §9.5, §9.6,
  §9.7, §12.1, §12.2; `profiles/manager.md` §1 and §12; `cli/curator.md`.
- Change Request revision to accept: the one recorded for TASK-260906-1uf713.

## Method

Drive every claim through `run()`, not through a helper. Apply your own narrowing mutants; a mutant
that deletes a gate proves less than one that weakens it to admit exactly one member. Your predecessor
found four blocking, three major and four minor this way, including one surviving mutant of its own.

## Review dimensions

1. **B1 — the locked `require_current_profile`.** The gate moved to the §9.2 activation seam and its
   comment claims every machine-scope switch funnels through it: `profile use`, `install --use`,
   first-install auto-activation, `import --use`, resync. Verify that enumeration is true by finding
   every path that reaches the seam, and drive each. A second bypass is a repeat finding. Confirm a
   *scoped* switch is unaffected, as the comment claims.

2. **B2 — the copied surface and the foreign symlink.** This was the most serious cycle-1 finding: a
   `--takeover` wrote through a foreign-manager symlink into a file outside every managed home. The
   fix removes the target first. Verify with the cycle-1 fixture: after a takeover the surface must
   no longer be a symlink **and** the foreign file's bytes must be unchanged. Then look for any other
   write in the stage that can land on a path the manager does not own — the symlink-fallback copy,
   the backup writer, the seed writer, the import reassembly.

3. **B3, B4 — import correctness.** Managed root-context files must no longer be detected as native
   (verify the marker check cannot be defeated by a home the marker does not cover). For the
   divergent same-named skills, judge whether the chosen outcome — two entries, or one plus a
   recorded loss — is the one §9.6 requires, and whether the producer named the sentence that decides
   it. Drive the two-adapter fixture from cycle 1.

4. **M2, M3, m1–m4.** The `stateForPath` wrapper is gone, so `profile_source_path_unreadable` must
   lead and the doubled prefixes must be gone. Critically: cycle-1's declared *survivor* mutant rested
   on the claim that the store branch is unreachable because `pathManifestDiag` precedes it — the
   wrapper was why nothing failed. The rework was asked to re-run that exact mutant. Verify it did,
   and that the reported outcome is true. `env status` must report the locked requirement. The
   `parseSystemEnvironments` allowed-knob gate must now be class-wide — re-apply your predecessor's
   surviving mutant (`&& key != "forms"`) and confirm it dies.

5. **The Windows fixture, and the pattern behind it.** The Windows lane failed on
   `TestTakeoverWarnsDotfileHeuristic` after the rework: the fixture set `$HOME` only, while the §9.5
   heuristic resolves the operator home through `os.UserHomeDir`, which reads `%USERPROFILE%` on
   Windows. The **orchestrator** fixed that one directly (`0bcea201`) with a shared
   `pinOperatorHome` helper, because this is the third consecutive stage to ship a Windows-only
   fixture defect. Review that commit as you would any other, and then sweep for the same class
   yourself — the producer's own sweep reported "nothing new", so judge whether the sweep was
   competent: a POSIX literal fed to a `filepath` operation, a hardcoded `/` in a comparison, a
   `~/`-relative or `$HOME`-relative lookup, a well-known-location list that is POSIX-only.

6. **M1 — the `path` overlay, carried as a bound.** The spec contradiction behind it is being fixed
   separately (curator-spec PR #47, TASK-260906-3x0w4y): the form requirement becomes git-source-only.
   Stage (c) was told to stop reporting the row as driven, carry an explicit bound naming that task,
   and fix the ledger row that stated the opposite of what the test asserts. Verify all three. Do not
   require the feature to work here; do require the bound to be honest and the ledger truthful.

7. **The attestation standard.** Cycle 1 found the increment-2 report claiming an AC row as *driven*
   when it rested on a manufactured `Policy`, and claiming 15 of 15 while silently dropping a bound
   increment 1 had declared. The rework brief made that the first thing this revision is judged on.
   Read the rework report against it: is every "driven" row driven from the production entry point,
   is every declared bound either fixed or still declared, and is the gate table honest?

8. **The hosted lanes.** Read `gh pr checks 61`. For any lane not green, extract failing cases from
   that run's uploaded evidence artifacts (`test-evidence-<os>`, `race-evidence-<os>`), not from
   `--log-failed`, which returns nothing useful here. A red lane on this head is blocking.
   **Run the two `test-gate.sh` lanes sequentially, never concurrently and never alongside a `-race`
   suite** — increment 2 lost two attempts to contention over the machine-wide Go test lock, and the
   resulting red is indistinguishable from a real regression.

## Constraints

Scratch under a throwaway copy or the worktree's `.temp/`; never write into the producer's worktree or
the control root.

## Verdict contract

Attach `TASK-260906-1uf713_review-findings-stage-c-2.md` with a `repeat-of:` line naming any cycle-1
class that recurs. Blocking or major → set the task to `development`. Otherwise an explicit ACCEPT at
`to-review` with `accept_cr` on the recorded revision, and state plainly whether every hosted lane was
green on the exact head you accepted. Do not mark the task done. Then
`task-board handoff TASK-260906-1uf713 --role reviewer`.
