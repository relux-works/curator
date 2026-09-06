# Review brief: stage (c) cycle 4 (rework 3, landing decision)

## Subject

- Branch `feat/agent-environments-stage-c` at `4a8a1d42` in
  `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`, 25 signed commits past `origin/main`.
  PR https://github.com/relux-works/curator/pull/61 — pushed, so hosted lanes run this exact tree.
  Rework 3 is `git diff 4df4d507..HEAD`.
- Read `TASK-260906-1uf713_review-findings-stage-c-3.md` and unpack
  `TASK-260906-1uf713_review-probes-stage-c-3.tgz` — both cycle-3 blockers are reproducible from it.
  Then `producer-brief-stage-c-rework-3.md` and `TASK-260906-1uf713_rework-report-3.md`.
- Authority: curator-spec `550579d` — §1, §6, §8.4, §9.1, §9.2, §9.5, §9.6, §12.1, §12.2;
  `profiles/manager.md`; `cli/curator.md`.
- Change Request revision to accept: the one recorded for TASK-260906-1uf713.

## Standing

Cycle 3 attacked the composition weight rules, both precedence primitives against materialised bytes,
the import's loss list and reassembly, and all six cycle-2 findings, and found them correct. Its
eight mutants on `parseSystemEnvironments` all died. Do not re-derive that.

Three cycles have each found a defect living in an addressing mode nobody tested — stage (a) F1,
cycle-1 B3, cycle-2 C2-B1, cycle-3 C3-B2. Spend your effort there rather than on the well-covered
paths.

## Review dimensions

1. **C3-B2 — the `path` snapshot is immutable again.** `updateLocked`'s `KindPath` arm should resolve
   the root from the store snapshot under the old lock's `state_sha256`, never from `source.Path`.
   Drive it: install a `path` profile, edit the source, then `update`, `sync` and `use`, and assert
   the pin across all three; then `profile update --all` with an imported profile present, which
   cycle 3 showed bricked the whole command. Confirm `profile install <same path>` still re-reads —
   that is the reinstall §1 sanctions — and that overlays still re-resolve under a `path` root, which
   the brief warned "return the old lock" would break. Apply your own narrowing mutant.

2. **C3-B1 — the seam, made structural.** The gate moved out of the `else` branch to a single
   construction point covering both branches, keyed on the resolved profile and preceding the
   installed check, and `--clear` with an operand is now refused. Verify the claim in its comment that
   `CurrentFile` has exactly one production writer below the gate, by finding every writer yourself.
   Then enumerate every caller that can reach the seam in machine scope — `profile use` both forms,
   `install --use`, first-install auto-activation, `import --use`, resync — and drive each through
   `run()` under a lock. A third bypass is the finding this cycle exists to prevent, so hunt for a
   path the enumeration misses rather than confirming the ones it names.

3. **C3-M1 — the transcribed pin.** The test should now transcribe §12.2's six keys independently
   rather than reading `LockableEnvKeys`. Verify it does not read that map even indirectly, then prove
   it both ways: **widen** the map by one key and **narrow** it by one, and confirm each fails. This
   is the fourth cycle on this class; if the pin still derives from the thing it protects, say so
   plainly and stop treating it as fixable by another test.

4. **C3-m1, C3-m2.** `compose add` under a locked `overlays_allowed: false` should now warn or be
   explained. The four-rule weight test should drive all four §6 rules disagreeing at once — cycle 3
   built that case and got weight 40, `overlay: true`, `required_by: [mid1, mid2]`; confirm the
   committed test asserts the same and would fail if any rule's precedence changed.

5. **The whole stage, and the landing question.** This is the fourth cycle. Beyond the six findings,
   look where nothing has looked yet, and say plainly whether stage (c) is safe to land. If it is,
   say so without hedging.

6. **The hosted lanes.** Read `gh pr checks 61`. For any lane not green, extract failing cases from
   that run's uploaded evidence artifacts, not from `--log-failed`. A red lane on this head is
   blocking. **Run the two `test-gate.sh` lanes sequentially**, never concurrently and never alongside
   a `-race` suite.

## Constraints

Scratch under a throwaway copy or the worktree's `.temp/`; never write into the producer's worktree or
the control root.

## Verdict contract

Attach `TASK-260906-1uf713_review-findings-stage-c-4.md` with a `repeat-of:` line naming any class
that recurs. Blocking or major → set the task to `development`. Otherwise an explicit ACCEPT at
`to-review` with `accept_cr` on the recorded revision, and state plainly whether every hosted lane was
green on the exact head you accepted. Do not mark the task done. Then
`task-board handoff TASK-260906-1uf713 --role reviewer`.
