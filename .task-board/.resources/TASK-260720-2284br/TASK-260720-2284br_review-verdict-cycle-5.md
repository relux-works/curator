# TASK-260720-2284br review verdict — cycle 5

## Verdict

Accepted. Route to `done`.

R5 closes the cycle-4 acceptance blocker. Declaration observations are now
bound to the exact immutable bytes consumed by closure parsing, so a supported
A → B → A rewrite cannot pass revalidation while committing the transient B
closure.

## R5 closure

- `internal/install/generation.go` opens each declaration document once, reads
  one payload, and derives the recorded content generation from that payload.
- `manifest.ParseBytes`, `devsub.ParseBytes`, and
  `scopes.ParseHybridDecls` parse the returned payload; their existing `Load`
  entry points delegate to the same parsers and retain their prior contracts.
- Project manifest, global manifest, project substitutions, and hybrid
  activation all enter the observation set through `observeDocument`.
- `documentKey` prevents those four inputs from being passed to the separate
  path-digest observation API. Marker observations retain their prior path
  semantics.
- `runCommit` rechecks the observations under the manager-home lock after
  journal recovery and before cache publication or target staging. A mismatch
  routes to `restartClosure`.

The design handles both relevant writer shapes: in-place rewrite makes the open
handle consume the new bytes, while rename replacement leaves it on the old
inode. In both cases the generation follows the bytes parsed, and the later
path re-read detects a settled replacement when necessary. Absent,
unreadable, content-identical, and mode-only cases have explicit coverage.

## Acceptance and architecture assessment

The R1–R4 closures remain intact by source inspection, permanent regressions,
and the previously accepted reviewer evidence:

- target classes sort deterministically, with the consumer ledger last;
- the transaction engine journals before mutation, recovers before new
  mutation, and rolls back targets in exact reverse order;
- project/global/hybrid context, runtime, canonical and forwarding shims, env
  files, adapter and mirror ledgers, removals, and consumer state are included
  in the fault sweep;
- cache publication and all shared target mutation occur under the home lock
  after optimistic revalidation;
- concurrent project consumer merging and rollback isolation remain covered;
- post-commit GC failures remain warnings and do not roll back a durable
  installation.

This fits the project architecture: declaration packages own parsing, install
owns planning/revalidation, staging owns deterministic target classes, and
transaction owns durable publication, recovery, and reverse rollback.

## Evidence checked first-hand

Submitted archive:

- `TASK-260720-2284br_gate-evidence-rework-4.tar.gz`
- SHA-256:
  `9d4f1d8a7efd9dd9de094f97ef2cc0e62423a104a32cca67267bc6c254abc829`
- driver and durable exits postdate the final product sources;
- build, vet, format, diff check, full install/atomicity/rest package chunks,
  R5/R4/reviewer-overlay regressions, five race gates, and pinned repo-wide
  golangci-lint all record exit 0;
- both negative controls record exit 1 for the intended binding failures, and
  the before-read variant demonstrates all four transient declarations
  committing when the fix is removed.

Independent final-tree checks:

- `gofmt -l .`, `git diff --check`, `go build ./...`, and `go vet ./...`:
  exit 0;
- pinned golangci-lint v2.4.0 `run ./...`: exit 0, `0 issues`;
- combined R5, R4, recovery, stale-generation, concurrent-project, and rollback
  isolation regression run: exit 0 in 77.629s;
- manifest, devsub, scopes, transaction, staging, and adapters suites:
  exit 0;
- focused R5 race run: exit 0 in 72.134s with no data race.

## Godriver red-gate assessment

The submitted default-parallel godriver red log is honest non-task contention,
not an acceptance blocker. An independent default-parallel rerun under host
load 7.8–10.5 exited 1 in 376.544s, and every failure was the same hard
15-second `go-v1 process_timeout`; no functional assertion failed. The same
tree passed `go test ./internal/godriver -count=1 -parallel 1` in 29.382s.
`internal/godriver` is sibling-owned and untouched by this task. Raising its
security-relevant probe limit here would be an ownership-boundary forced fit.

## Boundaries

- Validation is Darwin/arm64; the implementation is portable, but no Windows
  runtime evidence is claimed.
- Hardened kernel/container containment remains separate under
  STORY-260728-327soo and is not claimed here.
- No product code was modified, staged, committed, or published during review.
