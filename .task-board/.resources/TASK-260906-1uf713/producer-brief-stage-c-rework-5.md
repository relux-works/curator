# Producer brief: stage (c) rework 5

## Verdict

Review cycle 5 returned CHANGES REQUESTED — one major, four minor, no blocking. Read
`TASK-260906-1uf713_review-findings-stage-c-5.md`.

Rework 4 landed clean: **all four cycle-4 findings are fixed and verified**, including both halves of
§1 driven together, all three `KindPath` refusals reachable from the CLI with mutants P2/P3/P4 killed,
the skip registry widened rather than the skip text bent, and the builtin `default` refused under a
lock at both levels. And this is the **first fully green head of the stage** — all eleven hosted lanes
on `71e6baec`, Windows included. The reviewer also judged your rework-4 report against the attestation
standard and found it accurate, including reporting the candidate lane red rather than omitting it.
That is the standard now; keep it.

## The orchestrator's decision on C5-m3 — read this before anything else

The candidate lane is red against curator-spec main, on all three runners, only in `internal/config`,
on exactly the 30 subcases the `path`-overlay reconciliation added after this task's authority
`550579d` was frozen. Against `550579d` — the authority named in the task description, in every brief
and as the CR base — the same families run and pass; against `SPEC_PIN` the three registered
root-artifact rows defer as designed.

That gap is work the orchestrator created by landing curator-spec PR #47 mid-flight, and consuming it
means porting the spelling discriminator into the Go reader and reaching the feature from all three
surfaces. **It is not this leaf's work.** It is filed as **TASK-260906-19gjyw** and stage (c) will not
be reported complete against curator-spec main until that lands.

So: keep the M1 bound in this rework, but **retire and reissue it**. It is now half stale — the spec
contradiction it was raised for is gone, and what survives is an implementation gap. Reword the bound
and the ledger rows to say exactly that, naming TASK-260906-19gjyw rather than the spec task. Do not
implement the consumption here.

## C5-M1 (major) — the retry the §9.5 stop invites is a silent no-op that reports success

`profile install <same path> --as <same name> --use --takeover` drops **both** flags and exits 0
saying `updated profile <name>`.

This is not a corner. It is the exact retry §9.5's own stop invites: `install --use` fails with
`environment_surface_unmanaged_conflict` or `environment_foreign_manager_detected`, the profile is
already installed by then, so the operator's retry lands in `installLocked`'s `prior == source` branch
and into `reinstallPathLocked`, which never reads `options.Use` or `policy.Takeover` and returns before
`resyncCurrentScopes`. No takeover, no backup, no activation, and the machine still has no current
profile — while the command says it succeeded. Only `profile use <name> --takeover` recovers, which
the operator has no reason to know.

`reinstallPathLocked`'s doc comment (`envprofile.go:715-726`) asserts that a `--use` switch of a
non-current profile takes the fresh-install path. That is false, and it is `repeat-of` cycle-1 M1 and
cycle-2 C2-m2 — a comment claiming a production path the code does not have. Two cycles have now found
this class; the comment is part of the fix, not decoration.

Decide from §9.1 and §9.5 what the reinstall must honour, and drive every combination through `run()`:
`--use` alone, `--takeover` alone, both, and neither, on a `path` root that is current and one that is
not.

## Minors

- **C5-m1** — the C4-B1 regression test pins only the `--as` addressing mode; a mutant restricted to
  `options.As != ""` survives. Pin the bare `profile install <same path>` form too. This is the same
  shape as every blocking finding in this stage: the untested addressing mode.
- **C5-m2** — two stage-(c) skip reasons are still unregistered. C4-B2 was one word in one reason;
  find the rest by enumerating every skip this stage introduces against `skip-classes.tsv`, not by
  waiting for a lane to go red.
- **C5-m4** — `reinstallPathLocked`'s copied blocking-audit gate has no test: a mutant admitting
  exactly a blocking context member survives, and so does the `updateLocked` twin it was copied from.
  The reviewer reports exploitability as **unknown** rather than inferring it, because a second
  independent gate remains in the same loop. Give both gates a narrowing mutant and a named test, and
  say in the report whether a blocking-but-not-strict finding is reachable at all — if it is not, that
  is a fine answer, stated.

## Delivery

Small signed commits on `71e6baec` in `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`.
Human identity, nothing into the control root. Do not push and do not touch PR #61.

**Do not stage with `git add -A`.** Stage named paths and read `git status --short`.

Gates, each a standalone process with its exit code: `go build ./...`, `go vet ./...`,
`gofmt -l cmd internal`, `golangci-lint run ./...`, `gate-selftest.sh`, `no-broad-suppression.sh`,
`ledger-consistency.sh`, `test-gate.sh` against **both** roots, `go test -count=1 -timeout 30m
./cmd/curator`. **Run the two lanes sequentially.** For the candidate root use the task authority
`550579d`, and report the `87a0d00` result separately as the known consumption gap.

Attach `TASK-260906-1uf713_rework-report-5.md`: finding → resolution for all five; the four
`--use`/`--takeover` combinations driven through `run()` with the corrected doc comment quoted; the
bare-install regression pin; the skip enumeration with its result; both audit-gate mutants and your
answer on reachability; the reworded bound naming TASK-260906-19gjyw; the two-root gate table. Then
`task-board handoff TASK-260906-1uf713 --role developer`.
