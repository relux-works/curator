# Producer brief: stage (c) rework 4

## Verdict

Review cycle 4 returned CHANGES REQUESTED — two blocking, one major, two minor. Read
`TASK-260906-1uf713_review-findings-stage-c-4.md` in full; every finding carries the command that
produced it and C4-B1 is bisected to a commit.

Rework 3's substance held: the reviewer confirmed the C3-B1 seam is structural and single, the C3-M1
pin now transcribes §12.2 and dies under both a widening and a narrowing of the map, and the C3-m2
four-rule weight test asserts what cycle 3 measured by hand. Do not disturb that.

Two of the four findings are consequences of the C3-B2 fix. That is not a criticism of the fix — the
class it closed was real and its own mutant dies — but the arm it rewrote now needs the rest of its
coverage.

## C4-B1 (blocking) — the §1 reinstall is dead, and reports success anyway

§1 is one sentence with two halves: the snapshot is immutable **and never re-read**, *and* later edits
change nothing **until the operator reinstalls**. Rework 3 fixed the first half and broke the second.

`installLocked` routes a same-name, same-source install into `updateLocked`, and `updateLocked`'s
`KindPath` arm now resolves from the store snapshot and never from `source.Path`. The `stateForPath`
call eleven lines above has already read and stored the edited tree; the delegation throws that
snapshot away and re-pins the old one. Driven: install, edit the source, reinstall — exit 0,
`updated profile pk`, the same lock hash, the pin unmoved, and the edit absent from the materialized
surface. A control install under a fresh name proves the edit was real. Bisected: `4df4d507` re-reads,
`7fcf9c1b` and `4a8a1d42` do not.

So no operator path refreshes a `path` profile in place any more, and the command that is supposed to
do it prints success. For a `git` source the delegation is right — a reinstall of a range *is* an
update — so the fix is the `path` case, not the delegation in general.

Whatever shape you choose, the acceptance evidence is the pair: `update` must **not** move the pin
after a source edit, and `install <same path> --as <same name>` **must**, with the new bytes reaching
the materialized surface. Drive both through `run()` in one test so they cannot drift apart again.

## C4-B2 (blocking) — the Windows lane is red on an unregistered skip *reason*

`Test (windows-latest)` fails the platform-case gate:

```
FAIL  ledger case skipped for the wrong reason on windows:
      internal/envprofile :: TestImportUnreadableMarkerIsLoss
```

The ledger row is fine — `platform-cases.tsv:355` already carries `host-capability`. What is missing
is the **reason regex**. `.github/ci/skip-classes.tsv` registers
`this environment can (inspect|read) a mode-000 directory`, and rework 2's skip
(`import_test.go:785`) says **file**, correctly, because the artefact is `.agent-environment.json`.
Three sibling skips say "directory" and classify; this one word turns the lane red.

Widen the registry pattern to cover the file case, or split it into its own row. **Do not** silence
the case and **do not** change the skip text to say "directory" when the artefact is a file — the gate
is doing exactly its job by refusing an unclassified reason through.

Two things about this one matter beyond the fix. It has been red since rework 2 and no local lane can
see it: on a non-root Unix runner `chmod 0o000` works, the test does not skip, and the classifier
never sees the message. And it is the orchestrator's earlier Windows finding one layer up — there the
*fixture* was POSIX-only, here the *skip-reason registry* is. When you add a skip, the reason text is
part of the contract, not prose.

The orchestrator's own share of this: rework 2 and rework 3 were pushed without reading the Windows
lane before the review was spawned, so two cycles ran against a head that was already red. That
changes now — but you should still assume, every time, that a lane you did not read is red.

## C4-M1 (major) — the three refusals C3-B2 introduced are undriven, and one is the §8.4 class again

The new `KindPath` arm adds three refusals and no test drives any of them. The reviewer's mutants each
weaken exactly one member and each survives the full `./internal/envprofile/` and `./cmd/curator/`
suites:

- dropping `|| rootMember.StateHash == ""` from the no-state-pin refusal — survives;
- deleting the snapshot-name check — survives;
- on the snapshot read failure, returning the old lock as `unchanged` with a `nil` error — survives.

The third is the one that matters: it is §8.4 verbatim, *a failed read of manager-owned state turned
into "nothing to do"*, the same class as `readSkillsLedger` and `readRootSurface`. With that mutant
built, `profile update` against a removed store entry reports `unchanged` at exit 0 forever while the
profile is unrepairable. The stock code is correct — refusal (2) is reachable with no code change at
all, which is how the reviewer produced the comparison.

**This is the third distinct site of one class in this stage.** Drive all three refusals, and then say
in the report whether anything structural would stop a fourth — a shared helper, a linter rule, a test
that enumerates the read sites — or whether the answer is vigilance. Either answer is acceptable;
silence is not. The stage's own DoD says every new refusal is proven by a narrowing mutant that kills
a named test, and three of three are unproven.

## Minors

- **C4-m1** — `Policy.CheckMachineUse` narrowed with `|| name == DefaultProfile` survives both suites.
  Production is correct — `profile use default` under a lock exits 1 naming the knob — but every
  committed refusal row uses an ad-hoc name, and `default` is the builtin an operator under a lock
  would try first. One more driven row.
- **C4-m2** — `profile update` on a machine that already violates the lock publishes the new lock and
  *then* resyncs, so the lock moves and the surfaces stay stale at exit 1. The reviewer records this
  with evidence rather than as a defect, and notes you may reasonably answer that a machine in a
  configuration-error state should fail loudly. If that is the answer, say so in the report; do not
  leave it undocumented.

## Delivery

Small signed commits on `4a8a1d42` in `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`.
Human identity, nothing written into the control root. Do not push and do not touch PR #61.

**Do not stage with `git add -A`.** Stage named paths and read `git status --short` before committing.

Gates, each a standalone process with its observed exit code: `go build ./...`, `go vet ./...`,
`gofmt -l cmd internal`, `golangci-lint run ./...`, `gate-selftest.sh`, `no-broad-suppression.sh`,
`ledger-consistency.sh`, `test-gate.sh` against **both** roots, `go test -count=1 -timeout 30m
./cmd/curator`. **Run the two lanes sequentially.**

For C4-B2 specifically, reproduce the classifier locally: feed the observed skip reason through
`platform-case-gate.sh`'s matching against your amended `skip-classes.tsv` and show it classifying,
since no Unix lane will exercise the skip itself.

Attach `TASK-260906-1uf713_rework-report-4.md`: finding → resolution for all five; the C4-B1 paired
evidence (`update` does not move the pin, `install` does, both through `run()`); the three refusal
mutants now dying with named tests; the local classifier reproduction for C4-B2; your answer on
whether anything structural prevents a fourth §8.4 site; the two-root gate table; and honest bounds.
Then `task-board handoff TASK-260906-1uf713 --role developer`.
