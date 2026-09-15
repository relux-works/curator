# Producer brief: interop coverage rework 1

## Verdict

Review cycle 1 returned CHANGES REQUESTED — one major, one minor, one nit. Read
`TASK-260906-2cfxfv_review-findings-1.md`.

The substance holds and the reviewer verified it first-hand rather than from your logs: the hole was
real at base, all six negative runs reproduce from scratch, the positive run is green, the roots were
plain checkouts verified against `manifest.json`, the split loses nothing from the default lane, the
class move `root-content` → `root-unset` is right, and the six conformance ledger notes accurately
describe what their tests assert. Do not disturb any of that.

What follows is about the recurrence guard, and it matters more than its severity suggests: this leaf
exists to close the silent-skip class, and the guard meant to keep it closed can be satisfied while a
silent skip exists.

## F1 (major) — the contract test can be satisfied while a silent skip exists

`contract_test.go:63` (`skipCall`) matches a skip only when the call's receiver is a bare
`*ast.Ident`. Two receiver forms evade the scan entirely, and the reviewer injected each into
`TestConformanceContextDetectors`, compiled it, and confirmed the case really skips on a fully
serving root:

| injected skip | gate |
| --- | --- |
| `t.Skipf(…)` | killed |
| `guard := t; guard.Skipf(…)` | killed |
| `h := holder{t: t}; h.t.Skipf(…)` — selector chain | **survived** |
| `skip := t.Skipf; skip(…)` — method value | **survived** |

The committed claim is wrong in the same place. `platform-cases.tsv`'s note for this row says the
package's syntax tree contains exactly one Skip call *"under any receiver name"*, and the drafting
report §8 repeats it. The gate does not see two of them, and the DoD row requires ledger rows to
describe what their tests actually assert.

Your mutant M4 is a good mutant, but it exercised only the `guard := t` rename — the shape the gate
*does* catch — so it is evidence for a narrower claim than the one it is cited for. That is the
lesson worth carrying: a mutant proves the claim it exercises, not the claim it is filed under.

**Fix.** In `skipCall`, stop requiring `sel.X` to be an `*ast.Ident`; accept any expression as the
receiver and report it by position rather than by name. Additionally flag `Skip`/`Skipf`/`SkipNow`
used as a **value** — assigned, passed or returned — rather than called. Then correct the ledger note
and report §8 to say what the scan actually covers, and add the two surviving shapes as mutants that
must fail.

## F1b — the second layer does not close the gap either; decide it, do not imply it

The reviewer drove the real `platform-case-gate.sh` over a real `go test -json` stream with the
package **served**:

| skip reason | class | verdict |
| --- | --- | --- |
| `…publishes no vectors/context-detectors.json` | `root-content` | killed |
| `detectors temporarily disabled` | UNCLASSIFIED | killed |
| `CURATOR_CONFORMANCE_ROOT is not set` | `root-unset` | **tolerated — not fatal** |

Cause: the `deferred-only` policy check sits in the `else` branch *after* `if (tol != "")`, so for a
case the ledger lists with a tolerated skip the policy is never reached. A `root-unset`-worded skip in
a package the lane serves is therefore tolerated.

`platform-case-gate.sh` is not touched by your diff and no production path reaches that state today,
so this is defence-depth, not a live regression. **The orchestrator's decision: fix it here.** It is
an ordering fix in the one gate that enforces skip discipline across the whole repository, we are
immediately before an rc, and F1 + F1b compose into precisely the class this leaf exists to close. If
on inspection the fix is not small — if it changes the verdict of any existing tolerated row — stop,
leave it, and record it as an explicit stated bound with the evidence, and say which rows would move.
Do not leave the ledger note claiming coverage the code does not have either way.

## F2 (minor) — one unguarded root read is undeclared

`snapshot_acquisition_test.go:75` reads `tc.Expected` through `readRootFile`, an unguarded
`os.ReadFile` + `t.Fatal`, and on the real candidate root every `tc.Expected` resolves to
`expected/byte-exact-snapshot_sha256.txt` — which is in neither the `root-artifacts.tsv` row, nor
`crossReferencedTrees`, nor `gate-selftest.sh`'s `ENV_REQUIRED`. The reviewer enumerated all 26
root-relative paths the four declared vectors name: 24 under `expected/environments`, one under
`fixtures/byte-exact`, and this one uncovered.

The hole did not move — a root dropping it is served and the case fails red with the open() error —
but it fails mid-case rather than by name at the plan, which is inconsistent with the two trees that
are declared *precisely* so they do not fail that way.

Note the trap the reviewer measured: adding the path to the row alone makes `TestEveryFamily…` fail
with "declared … but nothing here reads it". The row, `crossReferencedTrees` and `ENV_REQUIRED` move
together — or, better and preferred, **derive** the cross-referenced set from the vectors' own path
fields, or at minimum also scan `readRootFile(t, root, …)` literals, so a future unguarded root read
cannot drift undeclared the same way. Prefer the derivation; a third hand-maintained list is a fourth
place to forget.

## F3 (nit)

The row note says every case in the package reads one of the declared artefacts without a per-family
guard. Three of the nine — the contract cases — read no root artefact at all, and the ledger correctly
gives them `skip=- class=-`. Reword.

## Delivery

Small signed commits on `31106aa8` in
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-interop-coverage`. Human identity. **Do not write
`LOGBOOK.md`.** Do not push and do not touch PR #63. **Do not stage with `git add -A`**, and commit
nothing under `.temp/`.

Gates, each a standalone process with its observed exit code: `go build ./...`, `go vet ./...`,
`gofmt -l cmd internal`, `golangci-lint run ./...`, `bash .github/ci/gate-selftest.sh`,
`bash .github/ci/no-broad-suppression.sh`, `bash .github/ci/ledger-consistency.sh`. Re-run **one**
negative candidate case and the positive one; roots as plain checkouts verified against
`manifest.json`, never `git archive`; lanes sequential. If you touch `platform-case-gate.sh`, re-run
`gate-selftest.sh` and say explicitly whether any existing tolerated row changed verdict.

Attach `TASK-260906-2cfxfv_rework-report-1.md`: the four receiver shapes with the two formerly
surviving ones now killed; your decision on F1b with its evidence; the F2 derivation with the trap
addressed; the corrected notes; and the gate table. Then
`task-board handoff TASK-260906-2cfxfv --role developer`.
