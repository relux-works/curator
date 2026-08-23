# Reviewer verdict for TASK-260811-1u42b9

Verdict: **changes requested -> to-dev**

## Scope and goal evidence

- Reviewer run: `RUN-260822-411c9c`.
- `task-board spawn goal RUN-260822-411c9c`: no active goal; the run is not goal-bound.
- No run directives were recorded at either review checkpoint.
- Reviewed rework outcome: `TASK-260811-1u42b9_rework-evidence_RUN-260822-cb137e.md`.
- Review was read-only; no product or test code was modified.

## Accepted rework

1. Materialized external packages are recursively re-admitted, their metadata is reconciled, and their complete owned file inventory is compared with exact extraction evidence derived from admitted tarballs (`internal/npmsource/materialize.go:485-557`). Fresh negative tests passed for substituted source, direct/renamed/nested compiled payloads, opaque payloads, implicit `binding.gyp`, and unadmitted bundled trees.
2. Portable audit serialization now contains only assurance mode and exit code; verified-only lossless observations are nullable and omitted in portable mode (`internal/npmsource/materialize.go:32-50`, `357-371`). The fresh serialization vector passed.
3. Lock, tarball SRI, metadata, lifecycle, bundle, native-payload, private-cache, installed-graph, and real offline npm vectors remain green.

## Required changes

### 1. npm still bypasses the shared protected executor and canonical permit/receipt contract

`npmsource` defines its own `Runner`, `Invocation`, `Audit`, and `LosslessObservation` authority (`internal/npmsource/materialize.go:20-67`). At dispatch it computes an adapter-local `npm-execution-permit-v1` digest and calls `Runner.Run` directly (`materialize.go:452-482`). It never constructs or validates `closureexec.DerivationPermit`, never calls `closureexec.Executor.Commit`/`Execute`, never supplies the exact admitted-input receipt/mount set, and never requires an executor-issued causal `DerivationReceipt`.

This is not equivalent to the accepted common contract. The shared permit validates admitted input receipts, mounts, C0 tool node, executable identity, host/target, process/read/write/network policy, resource limits, evidence schema, and causal predecessor (`internal/closureexec/models.go:194-280`). The shared executor is the atomic single-use commit/start seam and independently rechecks provider, toolchain, and admitted inputs (`internal/closureexec/executor.go:224-340`). The npm-local digest enforces none of those properties.

Required rework: execute cache derivation, `npm ci`, and Node invocation through the real shared `closureexec.Executor` operation. Build canonical permits from exact capture receipt IDs/mounts and the C0-bound tool node, commit before start, execute only by permit ID, and accept only executor-issued causal receipts. Remove the adapter-local authority duplication or make it a non-authoritative projection of the canonical permit/receipt.

### 2. Verified provider evidence is self-assertable and the zero-start tests do not exercise provider negotiation

`Runner.Preflight` returns an arbitrary `RunnerBinding`; npm calls only `AssuranceBinding.Validate()` (`materialize.go:399-435`). That validator checks the shape and fixed capability names, not that a fresh nonce-bound provider receipt was negotiated (`internal/closureexec/models.go:570-588`). Authentic fresh provider authority is created and retained by `closureexec.Executor.Preflight`, which negotiates health/capabilities and later revalidates the same receipt before execution (`internal/closureexec/executor.go:128-178`, `280-295`). npm never uses that path.

The verified fixtures manufacture `ProviderIdentity`, `CapabilityReceiptID`, and capability fields directly, and the fake runner returns a synthetic lossless audit (`internal/npmsource/conformance_test.go:573-610`, `892-918`). Consequently the advertised missing/incomplete/incompatible/cross-mode/drift zero-start cases prove only rejection of mutated structs; they do not prove that a missing, incompatible, or drifted real provider cannot start npm.

Required rework: bind npm to one immutable `closureexec.AssuredOperation` obtained from the configured executor before cache lookup or process start. Add zero-start tests using the common verified-provider negotiation seam for missing provider, incomplete identity/capabilities, incompatible contract, cross-mode authority, nonce/receipt drift, and provider identity drift.

### 3. The real portable npm test does not prove exact C0 tool binding or C5 authorization

The real test's `RecheckTool` simply echoes the expected fixture fingerprints without hashing any executable (`internal/npmsource/conformance_test.go:776-778`). `Run` ignores `Invocation.Tool.ExecutableRelativePath` (`bin/npm`) and instead launches `exec.LookPath("npm")` (`conformance_test.go:780-805`). Thus the passing real `npm ci` vector demonstrates offline materialization, but also demonstrates that a different executable can run than the one recorded in C0.

The C5 ID is likewise only copied into the local invocation. npm capture creates package instances and dependencies but no npm cache/install/invoke action nodes (`internal/npmsource/capture.go:380-415`); dispatch never proves the requested operation occupies a declared C5 action slot. Attaching a rederived plan ID therefore does not authorize the actual command.

Required rework: select and fingerprint the actual npm/Node executable used at dispatch, bind its exact toolchain node into C0, and have the executor launch that relative path after immediate recheck. Represent or otherwise canonically bind each permitted manager/runtime action to the accepted C5 plan, and add a negative test where the runner attempts PATH substitution or an action absent from C5; both must fail before process start.

## Fresh validation

| Gate | Exit | Evidence |
| --- | ---: | --- |
| Focused rework vectors, including real npm | 0 | `rework-focus-01.log`; real `npm ci` ran and passed, not skipped |
| `go test -count=1 ./internal/artifactpolicy ./internal/nodesource ./internal/npmsource` | 0 | `focused-02.log`; artifactpolicy `167.849s`, nodesource `4.382s`, npmsource `24.022s` |
| `go test -count=1 -race ./internal/npmsource` | 0 | `race-01.log`; `13.634s` |
| `go test -count=1 -cover ./internal/npmsource` | 0 | `coverage-01.log`; `80.1%` statements |
| `go vet ./internal/npmsource ./internal/nodesource ./internal/artifactpolicy` | 0 | `vet-01.log` |
| `golangci-lint run ./internal/npmsource ./internal/nodesource ./internal/artifactpolicy` | 0 | `lint-01.log`; `0 issues` |
| `go build ./...` | 0 | `build-01.log` |
| `git diff --check` | 0 | `diff-check-01.log` |
| `task-board validate` | 0 | `board-validate-01.log`; board valid |
| `go test -count=1 ./...` | 0 | `repository-suite-01.log`; `cmd/curator 460.412s`, `artifactpolicy 172.431s`, `rustsource 163.287s`, `npmsource 21.842s` |

One initial focused wrapper completed all three packages successfully but the zsh wrapper then exited 1 because it assigned the read-only shell variable `status`; no `.exit` artifact was produced. This tooling deviation is preserved in `focused-01.log`. The gate was rerun correctly as `focused-02` and exited 0.

Green tests do not override the trust-boundary failures above because the current tests substitute the authority that they are meant to verify.
