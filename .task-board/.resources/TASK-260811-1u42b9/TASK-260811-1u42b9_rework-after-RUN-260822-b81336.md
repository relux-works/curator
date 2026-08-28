# Reviewer verdict for TASK-260811-1u42b9

Verdict: **changes requested -> to-dev**

## Scope and goal evidence

- Reviewer run: `RUN-260822-b81336`.
- `task-board spawn goal RUN-260822-b81336`: no active goal; the run is not goal-bound.
- No run directives were recorded at the review checkpoints.
- Reviewed rework outcome: `TASK-260811-1u42b9_rework-evidence_RUN-260822-77c106.md`.
- Review was read-only; no product or test code was modified.

## Accepted rework

1. Materialized external packages are recursively re-admitted, their metadata is reconciled, and their complete owned file inventory is compared with exact tarball extraction evidence. Fresh substituted-source, implicit-binding.gyp, bundled-tree, direct/renamed/nested compiled, and opaque-payload negatives passed.
2. Portable receipts retain `network=not-observed` and omit the lossless-only resolver/cache/lifecycle/process/read/write counters instead of serializing default zero as observed absence.
3. Verified preflight now uses the common nonce-bound provider negotiation seam and the missing/incomplete/incompatible/cross-mode/capability-drift/provider-drift zero-start matrix passes.
4. npm operations now call `closureexec.AssuredOperation.Commit`, `Execute`, and executor-issued receipt verification, and C0 tool-node identity is checked through the common binding.
5. Closed lock parsing, tarball admission/SRI, metadata, lifecycle/bundle/native denial, cache determinism, offline npm, installed-byte reconciliation, and broad repository regressions remain green.

## Required changes

### 1. The canonical permit does not authorize the npm operation that the runner actually performs

The three C5 actions declare only coarse templates and no read/write slots (`internal/npmsource/capture.go:422-425`; `internal/nodesource/nodesource.go:791-796`). `newRunnerSession` resolves an action solely by subtype (`internal/npmsource/materialize.go:371-389`), and `run` checks only the action's `uses_tool` edge. It never reconciles concrete argv, working directory, environment policy, process policy, read set, or write set with the selected `ActionPayload` (`materialize.go:411-514`).

More seriously, `run` discards `invocation.ReadRoots` and `invocation.WriteRoots`. Every permit declares only `capture/input-*` replay reads and the synthetic `receipt.json` write, with `CWD=work` (`materialize.go:443-465`). The actual cache/install/invoke operations instead use absolute tarball, cache, destination, log, entrypoint, and home paths passed in argv or `CURATOR_PROCESS_CWD` (`materialize.go:146-151`, `271-276`, `313-314`). No `WorkCopy` maps the admitted project to the working tree, and the real operation outputs are not typed permit outputs.

The fake and real npm runners make the mismatch executable by interpreting those hidden absolute values directly (`internal/npmsource/conformance_test.go:783-808`, `830-892`). Thus a present C5 subtype authorizes arbitrary concrete arguments and host paths, while the issued portable receipt records the narrower synthetic permit. A lossless verified provider that truthfully reports the real reads/writes/cwd cannot match the permit and must fail with derivation drift.

Required rework: resolve every concrete cache/install/invoke command from the exact C5 action template and policy IDs; use admitted-input mounts/work copies and typed output roots as the paths actually consumed by the process; bind the exact manager/runtime subprocess set; and reject argv, cwd, environment, read, or write substitutions before process start. Add negatives that remove/change `--offline` or `--ignore-scripts`, change cwd/input/cache/output paths, or request an undeclared process, proving common executor rejection rather than runner-authored rejection.

### 2. The portable/exact-executable tests bypass the authority they claim to prove, and verified mode has no positive execution proof

The `PATH substitution` test is self-enforcing: `fakeRunner.Run` returns its own error before incrementing its counter (`conformance_test.go:819-835`). It does not demonstrate that the common executor detects a runner attempting a different executable or environment.

The real npm runner also launches through a custom test seam and appends a `PATH` that is absent from the permit (`conformance_test.go:783-804`). Portable receipts synthesize their environment from the permit (`internal/closureexec/executor.go:343-365`), so the test passes while the actual manager environment differs from the recorded exact invocation. There is no `npmsource` production use of `closureexec.NewManagerProcessRunner`; the only non-test consumer is `rustsource`. The real npm vector materializes but does not invoke Node through that concrete common runner.

Verified tests cover only zero-start failures. The fixture provider's `EnforceAndObserve` always returns `unexpected verified process start` (`conformance_test.go:1083-1085`), so no positive cache derivation, npm ci, or Node invocation proves that a compatible lossless provider can execute the declared permit and return an accepted exact audit.

Required rework: run the portable positive cache/ci/Node path through the concrete shared runner with the exact C0-selected executable and permit-owned environment, and make PATH/environment substitution fail in that common seam. Add a compatible positive verified provider fixture for cache derivation, npm ci, and Node invocation whose exact observed process/read/write/network evidence matches the canonical permits; retain the existing zero-start matrix.

## Fresh validation

| Gate | Exit | Evidence |
| --- | ---: | --- |
| `go test -count=1 ./internal/artifactpolicy ./internal/closureexec ./internal/nodesource ./internal/npmsource` | 0 | `focused-01.log`; artifactpolicy `34.732s`, closureexec `3.762s`, nodesource `1.937s`, npmsource `5.548s` |
| Focused materialized-byte/provider/C5/portable/real-npm vectors | 0 | `review-focus-01.log`; real npm vector ran and did not skip |
| `go test -count=1 -race ./internal/npmsource` | 0 | `race-01.log`; `21.354s` |
| `go test -count=1 -cover ./internal/npmsource` | 0 | `coverage-01.log`; `80.1%` statements |
| `go vet ./internal/npmsource ./internal/nodesource ./internal/closureexec ./internal/artifactpolicy` | 0 | `vet-01.log` |
| `golangci-lint run ./internal/npmsource ./internal/nodesource ./internal/closureexec ./internal/artifactpolicy` | 0 | `lint-01.log`; `0 issues` |
| `go build ./...` | 0 | `build-01.log` |
| `git diff --check` | 0 | `diff-check-01.log` |
| `task-board validate` | 0 | `board-validate-01.log`; board valid |
| `go test -count=1 ./...` | 0 | `repository-suite-01.log`; `cmd/curator 394.214s`, `artifactpolicy 137.916s`, `rustsource 146.591s`, `npmsource 24.628s` |

The green suite confirms the implementation's current expectations. It does not close the findings because the custom runners themselves implement the missing authority and the verified fixture has no successful process path.
