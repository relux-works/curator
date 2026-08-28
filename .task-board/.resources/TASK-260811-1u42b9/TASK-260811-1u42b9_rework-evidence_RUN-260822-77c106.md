# npm shared-executor rework evidence

Board task: `TASK-260811-1u42b9`

Run: `RUN-260822-77c106`

## Outcome

The npm cache derivation, frozen offline installation, and Node invocation paths
now use the shared `closureexec.Executor` contract. The adapter no longer owns
an independent runner binding, permit digest, or audit authority.

Each npm operation now:

1. obtains one immutable `closureexec.AssuredOperation` through executor
   preflight before cache/process activity;
2. re-derives and checks the exact common Node C0 checkpoint and assurance-bound
   C5 plan;
3. resolves a predeclared npm cache, npm ci, or Node invocation action from the
   C5 action set and verifies its `uses_tool` edge against the exact C0 tool
   node;
4. commits a canonical `closureexec.DerivationPermit` with every admitted
   project/tarball receipt and replay mount, the exact executable path/digest,
   toolchain fingerprint, host/target, policies, limits, and typed evidence;
5. executes only by committed permit ID through the same assured operation,
   with immediate tool and admitted-input rechecks; and
6. accepts only a causal receipt verified as issued by that executor.

Portable results retain `network=not-observed` and do not serialize synthetic
resolver/cache/lifecycle/process/read/write zero counters. Verified assurance
uses the common nonce-bound provider negotiation and drift checks.

## Code and test scope

- `internal/closureexec/executor.go`
  - added operation-scoped causal-head read, execute, and issued-receipt
    verification methods.
- `internal/nodesource/nodesource.go`
  - added selection-neutral runtime action declarations and exact C0-bound
    manager/Node `uses_tool` bindings for C5.
- `internal/npmsource/capture.go`
  - declares npm cache, npm ci, and Node invocation actions before C4/C5.
- `internal/npmsource/materialize.go`
  - replaced adapter-local Runner/Audit authority with shared executor permits
    and receipts.
- `internal/npmsource/conformance_test.go`
  - portable fakes now implement only the executor process-runner seam;
  - real npm launches the exact C0-selected executable rather than PATH lookup
    at dispatch;
  - added zero-start rejection for absent C5 action and PATH substitution;
  - added common provider zero-start cases for missing provider, incomplete
    identity/capabilities, incompatible identity, cross-mode C5, nonce/
    capability receipt drift, and provider identity drift;
  - retained exact installed-content reconciliation and native/opaque/bundle/
    lifecycle negative coverage.

## Fresh validation

Every command below ran as a standalone process after the last code change.

| Gate | Exit | Evidence |
| --- | ---: | --- |
| `go test -count=1 ./internal/artifactpolicy ./internal/closureexec ./internal/nodesource ./internal/npmsource` | 0 | `.temp/TASK-260811-1u42b9/focused-01.log`; real npm vector ran |
| `go test -count=1 -race ./internal/npmsource` | 0 | `.temp/TASK-260811-1u42b9/race-01.log` |
| `go test -count=1 -cover ./internal/npmsource` | 0 | `.temp/TASK-260811-1u42b9/coverage-01.log`; 80.1% statements |
| `go vet ./internal/npmsource ./internal/nodesource ./internal/closureexec ./internal/artifactpolicy` | 0 | `.temp/TASK-260811-1u42b9/vet-01.log` |
| `golangci-lint run ./internal/npmsource ./internal/nodesource ./internal/closureexec ./internal/artifactpolicy` | 0 | `.temp/TASK-260811-1u42b9/lint-01.log`; no issues |
| `go build ./...` | 0 | `.temp/TASK-260811-1u42b9/build-01.log` |
| `git diff --check` | 0 | `.temp/TASK-260811-1u42b9/diff-check-01.log` |
| `task-board validate` | 0 | `.temp/TASK-260811-1u42b9/board-validate-01.log`; board valid |
| `go test -count=1 ./...` | 0 | `.temp/TASK-260811-1u42b9/repository-suite-01.log`; `cmd/curator 469.051s`, `artifactpolicy 143.553s`, `rustsource 150.217s`, `npmsource 23.506s` |

No validation command was skipped and no expected-red result was reported as
green. No tool or workflow deviation occurred in this rework run.
