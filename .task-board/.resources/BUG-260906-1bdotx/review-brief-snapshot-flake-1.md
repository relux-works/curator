# Review brief: the Windows snapshot-timestamp flake (cycle 1)

## Subject

- Branch `fix/snapshot-timestamp-flake` at `879b884b` in
  `/Users/iv/Developer/ReluxWorks/.worktrees/curator-snapshot-flake`, three signed commits past
  `origin/main` (`919e2e9c`). PR https://github.com/relux-works/curator/pull/64. Diff:
  `git diff origin/main..HEAD` — 4 files, +200/−5.
- Read `producer-brief-snapshot-flake.md` and `BUG-260906-1bdotx_drafting-report.md`.
- Authority for the gate's purpose: `internal/registry/snapshot.go` and its own documentation of the
  zero-skew reading.
- Change Request revision to accept: the one recorded for BUG-260906-1bdotx.

## What the producer established

Two facts compose. The e2e helpers build `&config.Config{…}` as struct literals and never go through
`parseAudit`, so `SnapshotClockSkewSeconds` is the **zero value, not 300** — the gate is effectively
`CreatedAt.After(now)` with no tolerance. And `now` is fixed as a call argument before the fetch runs,
while the fixture stamped `created_at` inside the HTTP handler at a strictly later instant. RFC3339
truncates downward, so with `Δ` the fetch duration, `floor(t₂) > t₁` exactly when the fetch crosses a
whole-second boundary — giving **P(flake) = Δ per fetch**.

Note that this **contradicts the premise of the brief it was given**, which assumed the 300-second
default applied and asked how anything could be five minutes ahead. The producer said so plainly
rather than bending the evidence to the question. Judge the diagnosis on its merits, not on that.

## Review dimensions

1. **Verify both facts by construction, not by reading.** Print what `newEnv` and `newEnvIn` actually
   hold for `SnapshotClockSkewSeconds` and `SnapshotMaxAgeSeconds`. Then instrument or reason out the
   ordering: is `now` genuinely fixed before the first fetch, and did the handler genuinely stamp
   later? If the ordering claim is right, the probability model follows — check the arithmetic too,
   including the claim that `frac(t₁)` is uniform.

2. **Reproduce the flake deterministically.** The strongest confirmation is making it fire on demand:
   with the *old* fixture, delay the handler by more than one second and show the failure appear on
   any platform, not just Windows. If it reproduces that way, the Windows-specific part of the story
   is explanatory colour rather than the mechanism — say so.

3. **The gate must still catch what it exists for.** This is the line the brief drew: the message is
   a tampering signal, and a snapshot dated in the future is what a rollback or equivocation attack
   looks like. Confirm `clockSkew` is not widened anywhere, the assertion is intact, and no test now
   tolerates a genuinely future-dated snapshot. Apply a narrowing mutant that admits exactly one
   future-dated snapshot and confirm a **named** test fails. The producer added cases pinning the
   bound at every skew a config can carry — check that set is actually every one, including the zero
   the e2e path uses.

4. **The fix's blast radius.** The fixtures now stamp `created_at` once at construction. Does any
   existing test depend on a per-fetch timestamp — freshness, rotation, the `maxAge` path, the
   equivocation checks? Enumerate the tests that consume these fixtures and say which you drove.

5. **The Windows lane row.** A ledger row now makes the Windows lane run the bound by name. Confirm
   it describes what its test asserts, and that the case genuinely runs there rather than skipping.

6. **The honesty of the bounds.** The report labels the Windows-specific explanation an argument, not
   a measurement, and says `ledger-consistency.sh` proves compilation rather than execution. Confirm
   nothing else in the report or the PR body claims a measurement it does not have.

7. **Hosted lanes.** Read `gh pr checks 64`; a red lane is blocking. The Windows lane is the one that
   matters here, and it is slow — wait for it rather than inferring from the others.

## Method

Anchor each `-run` level separately and count `=== RUN` lines: `go test -run '^(Parent/child)$'`
splits on the unbracketed slash, matches nothing and exits 0. Run any conformance lanes sequentially.

## Constraints

Scratch under a throwaway copy or the worktree's `.temp/`; never write into the producer's worktree or
the control root.

## Verdict contract

Attach `BUG-260906-1bdotx_review-findings-1.md`. Blocking or major → set the element to
`development`. Otherwise an explicit ACCEPT at `to-review` with `accept_cr` on the recorded revision,
stating what the hosted lanes reported and whether the Windows lane was green on the exact head you
accepted. Do not mark it done. Then `task-board handoff BUG-260906-1bdotx --role reviewer`.
