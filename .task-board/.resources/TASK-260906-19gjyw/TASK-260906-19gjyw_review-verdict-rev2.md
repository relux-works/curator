# TASK-260906-19gjyw — review verdict, `CR-TASK-260906-19gjyw-2` revision 2

**ACCEPT.** repeat-of: none.

Reviewer run `RUN-260907-b8a04e`, cycle 2. Full evidence:
`TASK-260906-19gjyw_review-findings-2.md`.

## What is accepted

The deliverable is **curator `feat/consume-overlay-rule@e4ddca19`**, PR
https://github.com/relux-works/curator/pull/62, four signed commits past
merge-base `7c74a492`. Authority: curator-spec `87a0d0060bad`.

## Why an empty repository delta is the right outcome for this leaf

The story workspace this Change Request snapshots is a checkout of
**curator-spec** whose `HEAD` is `87a0d0060bad…`, the authority revision itself.
I confirmed `HEAD^{tree}` = base tree = candidate tree =
`2e6ca472ae8542a95e446ed4afb11cb75820536d`, `git diff <base> <candidate>` over
zero paths, and `git status --short` clean. This leaf's scope is the **curator**
repository, so its deliverable lies outside the snapshot by construction and the
producer could not have written into this tree without writing into the wrong
repository. An empty tree is the truthful record. Integrating this CR moves no
curator code; **landing PR #62 is the act that delivers the leaf.**

## What the hosted lanes report on the exact head accepted

* PR #62 head `e4ddca19` — `MERGEABLE`, `CLEAN`, **11 checks SUCCESS**;
  `Candidate suite` `SKIPPED`, which is correct for a `pull_request` event since
  that lane is `workflow_dispatch`-only.
* Run **34076889889** at `e4ddca19`: the full 225-row platform-case gate ran on
  ubuntu, macos and windows and observed **both edited ledger rows passing by
  name**; `Ledger consistency` reports `[must=linux,darwin,windows skip=-]` for
  each; the `SPEC_PIN` root still defers `internal/config`,
  `internal/envfragment`, `internal/envmarker` as designed.
* Candidate lane, run **34071813375** at `38702164`: all three `Candidate suite`
  jobs `success` — ubuntu 3m38s, macos 8m7s, windows 32m17s — each job's own log
  showing `CANDIDATE_REF: 87a0d0060bad64ab883d007dcdf35df7485368bf`,
  `CI_REQUIRE_FULL_ROOT: 1`, `served=71 deferred=0 excluded=1`,
  `test-gate: stage served exit=0`, and `ok internal/config ::
  TestManagerConfigV2SchemaCases`. The thirty subcases are closed.

## Cycle 1's three findings, each closed by measurement

* **F1 (major).** M10 re-applied here: `TestInstallNeverDemotesANetworkIdentityToAPath`
  now exits **1** (it exited 0 at `38702164`), naming all three bare canonical
  operands, logging `checked 8 network-identity operands, 3 of them bare
  canonical`. The new `bare == 0` guard fires when those operands are removed,
  and it cannot be satisfied vacuously — every spelling `ValidCanonical` accepts
  is classified `path` by the discriminator, so the widening decides each one.
  Test comment, drafting report and ledger row 312 all corrected.
* **F2 (minor).** Re-derived independently over **82 operands** driven through
  stage (c)'s *and* the head's complete install compositions (both
  `isPathOperand` and the old `identity.Parse` carve-out restored): 61
  unchanged, 21 changed. The refused→path class reproduces and is pinned; M19
  survives at exit 0 on the reviewed matrix and is killed by the added operand.
  Of the 23 operands that reached the network before this change, 22 stay `git`
  and one becomes an explicit refusal — **zero become a local path.**
* **F3 (trivial).** Doc comment moved. A `go/ast` pass over the six touched
  production files finds only two doc/symbol mismatches, both pre-existing on
  `origin/main` and neither introduced here.

## Regression surface

`envprofile.go` at `e4ddca19` is **AST-identical** to `38702164` with comments
stripped and declarations sorted. No product behaviour moved; the rework is
tests, a ledger description column, and comments.

## Gates re-run in this review

`go build ./...` 0 · `go vet ./...` 0 · `gofmt -l cmd internal` 0 lines ·
`no-broad-suppression.sh` ok · `ledger-consistency.sh` 225 rows ok ·
`go test ./internal/envprofile ./internal/identity ./internal/config` ok ·
`go test ./cmd/curator` ok (298.0s). Run sequentially, each a standalone
process, on a clean copy.

## Stated bound

The candidate lane has not been dispatched at `e4ddca19`; its standing
measurement is at `38702164`. The argument that this does not matter was judged
rather than accepted and holds: the non-test delta is AST-identical,
`internal/envprofile` reads no conformance root (0 files match
`CURATOR_CONFORMANCE_ROOT`), the ledger edit touches only the description
column, and the changed tests were observed `ok` on all three runners at
`e4ddca19`. Only the conformance-root dimension is unmeasured at this head, and
nothing in the delta reads it. Reported as a reasoned bound, not as a
measurement.

## Landing

**PR #62 is safe to land.**
