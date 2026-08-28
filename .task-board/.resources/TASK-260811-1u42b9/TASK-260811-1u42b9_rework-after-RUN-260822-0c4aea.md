# Reviewer verdict for TASK-260811-1u42b9

Verdict: **changes requested -> to-dev**

## Scope and goal evidence

- Reviewer run: `RUN-260822-0c4aea`
- `task-board spawn goal RUN-260822-0c4aea`: no active goal; the run is not goal-bound.
- No run directives were recorded at the review checkpoints.
- Reviewed implementation outcome: `TASK-260811-1u42b9_implementation-evidence.md`.
- Review was read-only; no product or test code was modified.

## Required changes

### 1. Materialized dependency bytes are not reconciled with admitted tarballs

`Materialize` checks the unchanged lock and then calls `validateInstalledTree`
(`internal/npmsource/materialize.go:283-295`). That validator records only
package directory/symlink paths (`materialize.go:371-440`); it does not inspect,
hash, re-admit, or compare any installed file or embedded `package.json` with
the admitted tarball manifest.

The positive fake-runner test makes the gap executable: for every selected
external package it creates only a directory and a synthetic minimal
`package.json` (`internal/npmsource/conformance_test.go:696-720`), yet the
materialization is accepted (`conformance_test.go:343-350`). The admitted
tarballs contain additional source files. The same seam can add or substitute
a `.node`, Wasm/V8 cache, opaque payload, `binding.gyp`, lifecycle metadata, or
arbitrary JavaScript without detection as long as package paths remain equal.

Required rework: reconcile the complete materialized package contents and
metadata against exact admitted inputs/derived extraction evidence, and run the
shared deny-dominant classifier on the materialized dependency tree before
invocation/publication. Add negative tests for substituted source bytes and
direct/renamed/nested compiled or opaque materialized payloads. A fake install
containing only package names must not satisfy the positive closure vector.

### 2. npm/Node execution bypasses the exact common C0/permit/tool binding

The npm runner contract carries only `Mode`, `ProviderID`, and `Lossless`
(`internal/npmsource/materialize.go:39-49`). npm and Node are invoked by symbolic
tool names (`materialize.go:148`, `275`, and `315`) without consuming the common
Node `RuntimeBinding`, C0 checkpoint, C5 action/build plan, derivation permit,
toolchain fingerprint, executable digest/path, or immediate recheck.

Verified preflight accepts any non-empty provider ID with `Lossless=true`
(`materialize.go:326-337`). It does not bind provider contract/version/binary/
capability receipt or exact compatible manager/runtime identity. The binding is
also queried again after each process returns (`materialize.go:153`, `280`, and
`320`) instead of comparing the observation with one immutable pre-start
binding, so a runner can drift modes/provider state across `Run` without a
zero-start rejection. The current test proves only the missing-provider branch
(`conformance_test.go:463-490`), not incomplete identity, incompatible provider,
cross-mode pre-start, or drifted binding.

Required rework: route cache derivation, `npm ci`, and Node invocation through
the accepted common checkpoint/executor contract with exact C0-bound npm/Node
and provider identities, precommitted permits/plans, immediate rechecks, and a
single immutable assurance binding captured before process start. Add explicit
zero-start negatives for missing, incomplete, incompatible, cross-mode, and
drifted verified bindings.

### 3. Portable audit counters still encode unobserved absence as zero

`Audit` exposes lossless-style integer counters without observation
availability (`materialize.go:28-37`), and `validateAudit` uses zero as the
acceptance value for ambient-cache, lifecycle, process, read, and write events
in both assurance modes (`materialize.go:340-368`). The portable success fixture
sets only mode/boundary/network and leaves those counters at their default zero
(`conformance_test.go:738-740`); that value is therefore indistinguishable from
an observed zero despite portable mode being unable to observe those classes.

Required rework: make portable evidence explicitly unobserved/not-applicable
for lossless-only fields and ensure no default zero is consumed or serialized
as proof of absence. Keep portable acceptance grounded in admitted inputs,
fresh private roots, exact manager invocation/config, frozen lock, cache
receipt, and complete installed-content reconciliation. Apply lossless counters
only to an exact verified provider observation.

## Positive findings

- Supported lock names and v2/v3 schema checks are closed and duplicate JSON is
  rejected.
- Raw tarballs are SHA-512 SRI checked, captured, recursively admitted, and
  reconciled for dependency/platform/lifecycle/bundle/native metadata before
  npm starts.
- Private cache receipts are deterministic and replay uses `npm ci --offline
  --ignore-scripts` with private home/config/cache roots.
- Focused conformance coverage exercises stale locks, integrity, workspaces,
  optional selection, lifecycle, bundled dependencies, implicit node-gyp,
  compiled tarball bytes, missing cache input, and ambient-state discard.

## Fresh validation

| Gate | Exit | Evidence |
| --- | ---: | --- |
| `go test -count=1 ./internal/artifactpolicy ./internal/nodesource ./internal/npmsource` | 0 | `focused-01.log`; artifactpolicy `43.915s`, nodesource `1.302s`, npmsource `3.838s` |
| `go test -count=1 -race ./internal/npmsource` | 0 | `race-01.log`; `4.270s` |
| `go test -count=1 -cover ./internal/npmsource` | 0 | `coverage-01.log`; `80.1%` statements |
| `go vet ./internal/npmsource ./internal/nodesource ./internal/artifactpolicy` | 0 | `vet-01.log` |
| `golangci-lint run ./internal/npmsource ./internal/nodesource ./internal/artifactpolicy` | 0 | `lint-01.log`; `0 issues` |
| `go build ./...` | 0 | `build-01.log` |
| `git diff --check` | 0 | `diff-check-01.log` |
| `task-board validate` | 0 | `board-validate-01.log`; board valid |
| `go test -count=1 ./...` | 0 | `repository-suite-01.log`; cmd/curator `485.517s`, artifactpolicy `178.400s`, install `152.105s`, atomicity `147.599s`, npmsource `15.373s` |

Green tests do not override the acceptance failures above because the positive
fake runner currently demonstrates that incomplete/substituted installed bytes
are accepted and the verified-mode tests do not exercise the required exact
binding/drift branches.
