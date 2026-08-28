# TASK-260811-1u42b9 developer rework evidence

Run: `RUN-260822-a5f43f`

Status target: `to-review`

## Rework outcome

The npm cache, install, and Node invocation paths now execute the concrete
commands authorized by the common C0/C5 and `closureexec` contracts.

- npm runtime actions contain closed full-command templates, including frozen
  offline/script-disabled flags, private cache/config/log locations, working
  directory, environment policy, process policy, and Node invocation slots.
- `npmsource` resolves those templates to logical paths and rejects argv, cwd,
  environment, input mount, cache/output root, or action substitutions before
  the process-start seam.
- Canonical permits use exact admitted-input mounts and typed work copies.
  Successful retained work copies are the actual cache/materialization outputs;
  immutable replay mounts and non-retained work copies are removed by the
  shared portable runner.
- Private npm cache bytes are derived from one exact admitted tgz per operation,
  reconciled, receipted, protected, and rebound as derived cache input for
  `npm ci`. The materialized tree is fully re-admitted/reconciled, normalized
  from contained workspace links into a link-free replay snapshot, and rebound
  before Node invocation.
- Portable integration stages and fingerprints the exact npm distribution and
  Node runtime selected at C0, then runs cache derivation, `npm ci --offline
  --ignore-scripts`, and Node invocation through
  `closureexec.ManagerProcessRunner`; it performs no PATH lookup for the
  selected executable.
- Verified positive coverage negotiates the common nonce-bound provider,
  executes cache derivation, npm ci, and Node invocation, and returns exact
  process/read/write/network evidence. An extra observed process is rejected by
  the common executor. Existing missing, incomplete, incompatible, cross-mode,
  nonce/capability-drift, and provider-identity-drift cases remain zero-start.
- Materialized source substitution, direct/renamed/nested compiled payloads,
  opaque payloads, bundled dependencies, implicit `binding.gyp`, lifecycle,
  stale lock, integrity, and ambient-cache vectors remain fail-closed.

## Changed authority paths

- `internal/closureexec/intake.go`
- `internal/closureexec/portable_runner.go`
- `internal/closureexec/portable_runner_other_test.go`
- `internal/npmsource/capture.go`
- `internal/npmsource/materialize.go`
- `internal/npmsource/conformance_test.go`
- `README.md`

## Fresh validation

Every final gate below ran as a standalone process without a pipe.

| Gate | Exit | Evidence |
| --- | ---: | --- |
| `go test -count=1 ./internal/artifactpolicy ./internal/closureexec ./internal/nodesource ./internal/npmsource` | 0 | `focused-final.log`; artifactpolicy `34.898s`, closureexec `3.322s`, nodesource `1.278s`, npmsource `7.835s` |
| Focused materialized-byte, exact-C5, real portable npm/Node, verified positive/negative vectors | 0 | `review-focus-final.log`; real npm was run, not skipped |
| `go test -count=1 -race ./internal/npmsource` | 0 | `race-final.log`; `29.141s` |
| `go test -count=1 -cover ./internal/npmsource` | 0 | `coverage-final.log`; `80.4%` statements |
| `go vet ./internal/npmsource ./internal/nodesource ./internal/closureexec ./internal/artifactpolicy` | 0 | `vet-final.log` |
| `golangci-lint run ./internal/npmsource ./internal/nodesource ./internal/closureexec ./internal/artifactpolicy` | 0 | `lint-final.log`; zero issues |
| `go build ./...` | 0 | `build-final.log` |
| `git diff --check` | 0 | `diff-check-final.log` |
| `task-board validate` | 0 | `board-validate-final.log`; board valid |
| `go test -count=1 ./...` | 0 | `repository-suite-final.log`; npmsource `46.808s`, rustsource `145.294s`, install `114.830s`, atomicity `114.579s` |

The first lint probe during implementation exited 1 with three mechanical
findings (one `gosec` annotation, one ineffectual initialization, and one
De Morgan simplification). They were corrected before the final lint and all
subsequent gates above; the failed probe is not represented as passing.

## Review focus mapping

1. Portable is still the functional default and the real test covers exact
   private-cache derivation, offline frozen npm ci, graph reconciliation, and
   Node invocation without a provider.
2. Portable receipts retain `network=not-observed` and do not claim lossless
   process/read/write absence.
3. Verified execution uses fresh common provider negotiation and exact audit;
   invalid authority remains zero-start.
4. Lossless observations are an additional verified gate, not a portable
   prerequisite.
5. Shared binary/opaque denial and npm lifecycle/bundle/node-gyp rules are
   enforced both before npm and against the complete materialized packages.
6. Focused, race, coverage, vet, lint, build, diff, board, real npm, verified,
   and uncached repository-wide gates were rerun after the authority correction.

