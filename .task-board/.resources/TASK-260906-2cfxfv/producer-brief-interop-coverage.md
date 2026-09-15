# Producer brief: make a missing environments vector family fail the candidate lane closed

## Where and what

- Repository `~/Developer/ReluxWorks/curator`. Branch and worktree named at spawn; base is curator
  main after PR #62 lands — the orchestrator confirms the exact OID. First run
  `git submodule update --init --recursive`.
- Files in scope: `.github/ci/root-artifacts.tsv`, `.github/ci/platform-cases.tsv`,
  `internal/interop/` (possibly a package split), and whatever `suite-plan.sh` needs to keep telling
  the truth. Authority: curator-spec main.

## The hole

`internal/interop` reads five environments vector families with a guarded skip:

```
context_detectors_test.go, context_resolution_test.go, context_materialization_test.go,
context_versions_test.go, snapshot_acquisition_test.go
```

each `t.Skipf("… publishes no vectors/… (pre-environments suite; root-content)")`, and the six
ledger rows tolerate a `root-content` skip on all three platforms.

`internal/interop` declares **none** of those families in `.github/ci/root-artifacts.tsv`. So
`suite-plan.sh` never defers the package, `CI_REQUIRE_FULL_ROOT=1` never fires for it, and
`skip-classes.tsv` marks `root-content` policy **allow in every lane, the candidate lane included**.

The consequence, measured on this repository's own mechanism: a candidate root that stopped
publishing `vectors/environments.json` would pass the candidate lane **green**, with all twenty-five
environments cases silently skipped. That is the exact failure class stage (a) review finding F9
named — a gate installed but never run — and stage (b) closed it for `internal/envfragment` and
`internal/envmarker` by registering their families, which the cycle-3 reviewer then verified fails
closed: `suite-plan` exits 1 in the candidate lane when either family is missing.

Nothing here is broken today. The families are published and the cases run. This closes the hole
before an rc, so that a root regression cannot pass unnoticed.

## Why it is not one line

Registering `internal/interop` wholesale would defer the **whole package** on the `SPEC_PIN` root,
and that package also carries the pre-environments conformance cases the default lane depends on —
dropping them from every default lane is a worse outcome than the hole.

So the shape is a decision, not a transcription. Weigh at least these, and say in the report which
you chose and why:

1. **Split the package.** Move the five environments conformance tests into their own package (for
   example `internal/interop/environments`), register its families in `root-artifacts.tsv`, and leave
   the rest of `internal/interop` served as it is today. Then the `SPEC_PIN` root defers only the new
   package, and the candidate lane fails closed on a missing family.
2. **A per-family assertion in the served lane.** Keep the package whole and add a check that, when
   the root is served and `CI_REQUIRE_FULL_ROOT=1`, a missing environments family is fatal rather
   than a tolerated skip.
3. Anything better you can defend from the gate scripts as they are.

Whatever you choose, the ledger rows must end up describing what their tests actually assert, and the
default lane must keep every non-environments `internal/interop` case it runs today.

## Acceptance, and how to prove it

The proof is a **negative run**, not a green one: materialize a candidate root, remove one
environments vector family from it, and show the candidate lane failing **by name** — not skipping.
Do that for at least `vectors/environments.json` and one other family. Then show the same lane green
against the unmodified root, and show the `SPEC_PIN` root still deferring exactly what it should.

Materialize every root as a **plain checkout** verified file-by-file against the published
`manifest.json`. Do **not** use `git archive`: the repository's own `.gitattributes` filters rewrite
`conformance/v1/fixtures/byte-exact/subst.txt` from 40 bytes to 65 and corrupt the root silently.

`bash .github/ci/gate-selftest.sh` and `bash .github/ci/ledger-consistency.sh` must be green, and
`suite-plan.sh` must still report a truthful partition on both roots.

## Method

- Drive the gate scripts themselves, not a description of them.
- When a mutant harness is involved, anchor each `-run` level separately and count `=== RUN` lines:
  `go test -run '^(Parent/child)$'` splits on the slash, matches nothing and **exits 0**, which reads
  as a survivor.
- A green suite is not evidence. This whole task exists because a lane can be green while the cases
  it is supposed to run are skipped.

## Delivery

Small signed commits, human identity. **Do not write `LOGBOOK.md`.** Do not push, do not open a PR.
**Do not stage with `git add -A`** — the gate run produces artefacts; stage named paths and read
`git status --short`. Run the two `test-gate` lanes **sequentially**, never alongside a `-race` suite.

Attach `TASK-260906-2cfxfv_drafting-report.md`: the shape you chose with the alternatives weighed; the
negative runs showing the lane failing by name on a removed family; the two-root partition; the ledger
rows before and after; and the gate table with observed exit codes. Then
`task-board handoff TASK-260906-2cfxfv --role developer`.
