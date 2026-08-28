# TASK-260827-18tswm rework outcome — RUN-260828-adfe6e

## Candidate reconstruction

The task-owned delivery was reconstructed from the exact PR #47 CI head
`c2215f9b929e11a32d75bff1205d296c135ddd7f`. The 68-path scope is the union of
the original CI-green delivery and its task-owned follow-up commits: the closed
Rust exception, pinned Cargo host classification, dry-run diagnostic, host
GOROOT isolation, Windows port/hardening, hosted-Windows npm corrections, and
the final Windows skip classifications. Board paths are excluded.

`c221-content-verification-01.log` proves that 64 delivered paths remain
byte-for-byte equal to `c2215f9b`. The four delivered paths intentionally
changed by this rework are `LOGBOOK.md`, the two Swift integration call sites,
and `internal/testtoolchain/lock.go` (package documentation only). The two new
paths are the shared Swift classifier and its regression test. No changed path
exists outside that task-owned union.

This directly ties the fresh candidate to the tested delivery content without
copying the independent TASK-260825-1yzubs parallel-refactor delta.

## Finding 3 correction

Both Swift integration call sites now use the shared pure predicate
`testtoolchain.SwiftManifestLinkerUnavailable`. Its table regression proves:

- the exact `Invalid manifest` plus clang `posix_spawn` conjunction classifies;
- `Invalid manifest` alone remains fatal;
- the `posix_spawn` diagnostic alone remains fatal;
- an unrelated Swift derivation-permit failure remains fatal.

The existing fatal assertions remain in place after the conditional skip. No
assertion was weakened, no toolchain identity was added, and release-pin
promotion was not touched.

## CI evidence

GitHub Actions run
https://github.com/relux-works/curator/actions/runs/33130874599 is green on
`c2215f9b`: Test on Ubuntu/macOS/Windows, Race on Ubuntu/macOS, platform-case
gate, Lint, Naming, Interop, and all three Gate self-tests. Windows Race is
deliberately absent from the landed workflow because the Go race detector needs
a C toolchain there. Reviewer resource
`TASK-260827-18tswm_review-verdict_RUN-260828-8a2060.md` records the downloaded
test-evidence inspection: zero `UNCLASSIFIED`/`FATAL` rows and the expected
classified pnpm/Yarn rows. Original checklist item 6 remains unchecked and is
superseded by the amended criterion.

## Local validation

Every listed command ran directly without a pipe. Logs are task-scoped under
`.temp/TASK-260827-18tswm/`.

| Command | Exit | Evidence |
| --- | ---: | --- |
| `go test -count=1 -run '^TestSwiftManifestLinkerUnavailableClassifiesOnlyExactConjunction$' ./internal/testtoolchain` | 0 | `go-test-swift-classifier-01.log` |
| `go test -count=1 ./internal/testtoolchain ./internal/swiftpmsource ./internal/swiftpmbuild` | 0 | `go-test-affected-02.log` |
| `go build ./...` | 0 | `go-build-all-01.log` |
| `go vet ./...` | 0 | `go-vet-all-01.log` |
| `golangci-lint run` (v2.12.2) | 0 | `golangci-lint-01.log` |
| `bash .github/ci/gate-selftest.sh` | 0; 81 passed, 0 failed | `gate-selftest-01.log` |
| `bash .github/ci/ledger-consistency.sh .temp/TASK-260827-18tswm/ledger` | 0; 80 rows | `ledger-consistency-01.log` |
| `bash .github/ci/no-broad-suppression.sh` | 0 | `no-broad-suppression-01.log` |
| `gofmt -l cmd internal` | 0; empty output | `gofmt-check-01.log` |
| `git diff --check` | 0 | `git-diff-check-01.log` |
| task-owned content verifier | 0; 64 exact paths, zero unexpected paths | `c221-content-verification-01.log` |

The first attempted affected-package wrapper ended with exit 1 because it used
zsh's read-only variable name `status`; that invocation is not counted as test
evidence. The exact test command was rerun as `go-test-affected-02.log` and
exited 0.

## Handoff blocker

`task-board handoff TASK-260827-18tswm --role developer` ran after the outcome
and amended checklist item were attached and exited 1. It refused solely on
legacy checklist item 6, which requires a Windows Race lane that the landed
workflow deliberately does not define. Item 6 is explicitly superseded by the
checked amended item 17 and must remain unchecked under the evidence-honesty
contract: no exact Windows Race command ran green.

The installed `task-board` exposes `check_item`/`uncheck_item` but no mutation
to supersede, replace, or remove a checklist row. Its handoff contract requires
every row checked and does not interpret the amendment. There is also no
standalone Change Request publication command; publication is coupled to a
successful producer handoff. Therefore neither a truthful `to-review`
transition nor fresh revision 2 can be produced with the current board model.

- Constraint: the amended acceptance contract and the handoff validator are
  mutually incompatible.
- Failed approach: ordinary developer handoff; exact refusal was `unchecked
  checklist items [6] ... handoff evidence missing`.
- Rejected workaround: checking item 6 administratively. That would claim an
  unrun Windows Race gate passed and violate the explicit evidence contract.
- Viable option A (recommended): add first-class checklist supersession to the
  `skill-project-management` source repo, install that task-board version, mark
  item 6 superseded by item 17, then resume this run and execute handoff.
- Viable option B: migrate this task's checklist through an approved board
  operation that replaces item 6 with item 17 while preserving history, then
  resume and hand off.
- Exact external input needed: an operator-supported board mutation/tool version
  that makes item 6 non-required without marking its unrun command green.
