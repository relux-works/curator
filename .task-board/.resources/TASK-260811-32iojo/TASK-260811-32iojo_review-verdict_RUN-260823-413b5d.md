# Reviewer verdict for TASK-260811-32iojo

Verdict: **changes requested -> to-dev**

Reviewer run: `RUN-260823-413b5d` (not goal-bound; `task-board spawn goal` reported no active goal).

## Blocking finding

Modern Yarn condition parsing still fails open for optional edges. `conditionsMatch` converts every `evaluateYarnCondition` error into the ordinary prune reason `condition_unsupported` (`internal/yarnmodernsource/lock.go:755-760`), and `markSelection` permits that result for optional and optional-peer edges (`lock.go:546-552`). Consequently malformed or unsupported condition bytes are accepted into a nonempty graph/lock identity and silently prune a dependency instead of rejecting the lock/profile before capture.

The executable reviewer probe used a root `optionalDependencies` edge and showed:

- valid `os=linux`: accepted and selected;
- malformed `os=linux && cpu=x64`: accepted, package pruned as `condition_unsupported`, nonempty `LockDigest`;
- `!!os=linux` and `!!!os=darwin`: accepted and pruned as `condition_unsupported`, each with a nonempty `LockDigest`, although the adjacent implementation contract says the pinned grammar normalizes even leading `!` counts and treats odd counts as negation (`lock.go:768-772`).

This violates the task's exact condition binding and fail-closed requirements and the shared N10 target-condition vector. Evaluator errors are unsupported grammar, not target non-selection evidence. The same boundary can accept malformed conditions on unreachable lock entries without evaluating them.

## Required rework

1. Validate every admitted lock condition with the pinned condition grammar before graph identity or capture is emitted, including unreachable, optional, and optional-peer package entries.
2. Propagate malformed/unsupported evaluator input as a deterministic fail-closed error; never translate an evaluator error into an admissible prune reason.
3. Reconcile the implementation and its stated multi-`!` contract: either implement the exact pinned normalization or explicitly reject it as unsupported before graph emission. Do not silently prune it.
4. Add regression tests for malformed optional, optional-peer, and unreachable conditions proving an empty returned graph/config identity and zero capture/manager/build/publication starts.

## Prior rework independently verified

- The previous wrong-type `dependencies: []` and trailing-lock-document probes now return `closure_lock_format_unsupported` with empty graph/config identities and no canonical alias.
- A 26-case additional grammar probe covered lock/rc scalar, map, sequence, nested-type, duplicate, unknown-field, and multi-document variants; all 26 rejected before identity emission.
- The prior peer overlay exploit now fails with `closure_metadata_mismatch`. Permanent positive/negative tests verify exact `peerDependencies` and `peerDependenciesMeta` agreement, so artifact metadata cannot widen or overwrite lock authority.
- Missing/duplicate/drifted workspace entries, workspace dependency scopes, patch byte identity/bijection, required peer handling, and the nonfunctional/preseeded PnP boundaries remain fixed in focused tests.
- A real pinned Yarn 4.9.2 PnP install plus Node dependency invocation passed through the verified executor under macOS `sandbox-exec` with `(deny network*)`.

## Validation evidence

| Command | Exit | Result |
| --- | ---: | --- |
| `go run ./.temp/TASK-260811-32iojo-review-4/probe.go` | 0 | Both prior lock exploits reject; no digest alias. |
| Prior peer overlay test through `-overlay` | 1 | Expected-red old exploit: now receives `closure_metadata_mismatch`. |
| `go run ./.temp/TASK-260811-32iojo-review-5/grammar_probe.go` | 0 | 26/26 added malformed lock/rc variants reject with empty identities. |
| `go run ./.temp/TASK-260811-32iojo-review-5/condition_probe.go` | 0 | Reproduces accepted malformed/multi-negation optional conditions and false pruning. |
| Focused regressions plus real Yarn PnP with `CURATOR_TEST_YARN_MODERN_JS=.../4.9.2/.../yarn.js` | 0 | Prior fixes and functional OS-denied replay pass. |
| `go test -count=1 -race ./internal/yarnmodernsource` | 0 | Focused race gate passes. |
| `golangci-lint run ./internal/yarnmodernsource` | 0 | `0 issues.` |
| `go vet ./internal/yarnmodernsource` | 0 | Vet passes. |
| `go build ./internal/yarnmodernsource` | 0 | Build passes. |
| `gofmt -l internal/yarnmodernsource` | 0 | No output. |
| `git diff --check` | 0 | Whitespace check passes. |
| `go test -count=1 ./...` | 0 | Full uncached repository suite passes; `cmd/curator` 491.863s and `internal/yarnmodernsource` 10.895s. |
| `task-board validate` | 0 | Board valid. |

Scratch logs and probes are under `.temp/TASK-260811-32iojo-review-5/`. No product code was modified, staged, committed, reset, or cleaned by this reviewer.
