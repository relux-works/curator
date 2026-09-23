# TASK-260908-2kqa77 rework 1 → revision 2 (orchestrator brief, binding)

The revision-1 gate FAILED (run 35722093898) — the gate you added does not run on Linux at all:

```
Lint / Verify GoReleaser rc channel values
awk: cmd. line:59: error: Invalid range end: /^[|>][+-0-9]*$/
##[error]Process completed with exit code 1.
```

Root cause: inside a bracket expression `[+-0-9]`, the `+-0` is a RANGE from `+` (0x2B) to `0`
(0x30). BSD awk (your macOS host) tolerates it; gawk (ubuntu-latest) rejects it, and per POSIX a
`-` that is meant literally MUST be first or last in the bracket expression. That single line takes
down the `Lint` job on Linux and the `Gate self-test` job on ubuntu AND windows
(`gate-selftest: 152 passed, 8 failed, 1 skipped`), with these named self-test rows failing:
`the prerelease failure names the observed value`, `the second-entry failure names scoops[1]`,
`a good second entry passes`, `unquoted auto passes (same parsed string)` (plus their siblings).

## Required
1. Fix the bracket expression(s): write the literal `-` first or last (`[-+0-9]`, `[|>][-+0-9]*`),
   and audit EVERY bracket expression in `.github/ci/goreleaser-config-gate.sh` for the same trap,
   plus any other gawk/BSD-awk divergence (POSIX-only constructs: no 3-argument `match`, no `\x`
   escapes, no `gensub`, no `length(array)`, no `asort`, no `delete array` on a whole array in a
   for-loop guard, no `--re-interval`-dependent `{n,m}` unless you verify it, no `\s`/`\d`).
2. Re-run the self-test locally AND prove the script parses under a strict POSIX/gawk-like reading:
   if gawk is not installed on the host, at minimum run `awk --posix` where available, and state
   explicitly in results.md which awk implementations you executed against and which you did not.
   The hosted gate (ubuntu gawk + Windows Git-Bash awk + macOS BSD awk) is the arbiter — publish
   only when all three `Gate self-test` jobs and `Lint` are green.
3. Re-check every one of the 8 failing self-test rows on the hosted run, not only locally: the four
   named above are assertions about message wording and about passing fixtures, so make sure they
   fail for the awk reason and not for a second, independent defect. If any of them survives the
   awk fix, that is a real second bug — fix it and say so.
4. Everything the revision-1 results claimed still holds otherwise: keep the mutant table, add the
   awk-portability row (a bracket-range regression must fail the self-test on Linux), and keep the
   `ci.yml` wiring pin.

Continue from the revision-1 tree (no checkout/clean/stash), append a "Revision 2" section to
`TASK-260908-2kqa77_results.md` naming the defect, the fix and the awk implementations you
exercised, then `task-board handoff TASK-260908-2kqa77 --role developer`. Republish only on a green
gate.
