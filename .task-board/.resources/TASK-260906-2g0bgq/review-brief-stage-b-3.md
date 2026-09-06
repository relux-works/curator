# Review brief: stage (b) cycle 3 (rework 2)

## Subject

- Repository `~/Developer/ReluxWorks/curator`, worktree
  `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-b`, branch
  `feat/agent-environments-stage-b` at `1a936e77`, 22 signed commits past `origin/main`. PR
  https://github.com/relux-works/curator/pull/60 — the head is now pushed, so hosted lanes are running
  on the revision you are reviewing. Rework 2 is the five commits `7bca4d4b..1a936e77`; the full stage
  is `git diff origin/main..HEAD`.
- Read `TASK-260906-2g0bgq_review-findings-stage-b-2.md` (your predecessor's cycle-2 findings),
  `producer-brief-stage-b-rework-2.md` (the orchestrator's decision, including where it overrode your
  predecessor's proposed fix and why), and `TASK-260906-2g0bgq_rework-report-2.md`.
- Authority: curator-spec `f39f4a9` (`protocol/environments.md` revision 1.1 §5.3, §5.5, §5.8, §7,
  §8, §9.5, §10, §12; the fragment and marker schemas; `conformance/v1/expected/environments/*`).
  Change Request revision to accept: the one recorded for TASK-260906-2g0bgq.

## Standing

Cycle 2 re-drove F1–F6 and F8–F11 with its own narrowing mutants and found each fixed. **Do not
re-derive that work.** Your subject is rework 2 — B1, B1b, M1, m2, m3, m4 — plus whatever the changed
lines break elsewhere.

## Review dimensions

1. **B1, and the orchestrator's override.** Your predecessor proposed a `root-content` tolerated skip;
   the orchestrator instead required registration in `.github/ci/root-artifacts.tsv` with a
   `root-unset` ledger class, on the reasoning that `root-content` is policy `allow` in every lane —
   the candidate lane included — so it would let a candidate root that dropped a family pass with the
   case silently skipped, whereas `CI_REQUIRE_FULL_ROOT=1` makes a deferral fatal. **Test that
   reasoning rather than accept it.** Materialize the `SPEC_PIN` root read-only and run
   `test-gate.sh` against it; then run it against curator-spec main with `CI_REQUIRE_FULL_ROOT=1`.
   Verify the three drivers are deferred-and-skipped in the first and served-and-passing in the
   second. Then attack it: remove one of the two families from a scratch copy of the main root and
   confirm the candidate-lane invocation fails closed rather than skipping. If it does not, the
   override is wrong and that is a blocking finding against the orchestrator's decision, not against
   the producer.

2. **B1b.** The boundary fixture now builds on `t.TempDir()` with `filepath.Join`. Confirm the
   assertions still assert what they did — in particular that the traversal case still contains a real
   `..` segment and that the out-of-root case is genuinely outside. Apply the narrowing mutant your
   predecessor used (`root := filepath.Dir(filepath.Clean(envRoot))`) and confirm it still kills.
   Separately judge the producer's decision to give `fragment_schema_test.go:checkPath` *slash*
   semantics because the published vectors carry POSIX paths: is that right, or has it just moved the
   platform assumption somewhere a Windows runner will not catch? Say which, with the reason. You
   cannot run a Windows runner; the hosted lane can — read it (see 6).

3. **M1.** `useLocked` now calls `TargetFor` when `--env` is present and `TargetByID` when it is not.
   Drive it through `run()`, not the helper: `--env pi`, `--env opencode`, `--env claude_code`,
   `--env codex_cli`, each with the declared target, plus a target with no `--env` at all, plus an
   undeclared target. Apply the producer's own mutant and confirm the kill. Then ask whether the
   `--env`-absent branch is right: with no `--env`, any declared target is admitted for any adapter —
   is that what §7.6 says, or a remaining hole?

4. **m2, m3.** The status structs gained `json` tags; check every nested struct got them and that the
   asserted member list is the complete top-level set, not a sample. The write-through-link loop
   gained a non-vacuity guard; confirm the guard fires when the loop is emptied (mutate `storeDocPath`
   or the match predicate).

5. **m4, and the meta-finding.** Cycle 2 marked B1 `repeat-of` cycle 1's F3 — a gate attested but
   never read — and said the next step is a gate on how gate lines are produced. Judge this report's
   gate table against that: every row a standalone command with an observed exit code, root-sensitive
   gates run against both roots, the CI claim verbatim and correctly labelled as the *old* head's
   state. Re-run enough of the table yourself to know whether it is true, and say which rows you
   re-ran and which you took on the evidence.

6. **The hosted lanes.** This head is pushed and CI is running on it. Read `gh pr checks 60` and, for
   any lane that is not green, extract the failing cases from that run's uploaded evidence artifacts
   (`test-evidence-<os>`, `race-evidence-<os>`) rather than from `--log-failed`, which returns
   nothing useful here. A red lane on this head is blocking regardless of what any local run shows.
   The Windows lane is the one that matters most for B1b.

## Constraints

Scratch under the story worktree's `.temp/` or a throwaway copy; never write into the producer's
worktree or the control root.

## Verdict contract

Attach `TASK-260906-2g0bgq_review-findings-stage-b-3.md` with a `repeat-of:` line naming any earlier
finding whose class recurs. Blocking or major → `development`. Otherwise an explicit ACCEPT at
`to-review` with `accept_cr` on the recorded revision, and state plainly whether every hosted lane was
green on the exact head you accepted. Do not mark the task done. Then
`task-board handoff TASK-260906-2g0bgq --role reviewer`.
