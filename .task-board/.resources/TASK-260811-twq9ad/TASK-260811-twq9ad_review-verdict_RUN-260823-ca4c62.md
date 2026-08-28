# Reviewer verdict for TASK-260811-twq9ad

Verdict: **changes requested -> to-dev**

## Review authority and scope

- Reviewer run: `RUN-260823-ca4c62`
- `task-board spawn goal RUN-260823-ca4c62`: `Active Goal: none (run is not goal-bound)`
- Reviewed producer outcome: `TASK-260811-twq9ad_implementation-results.md`
- Reviewed implementation: `internal/yarnclassicsource/`
- No product code was modified by this reviewer.

## Blocking findings

### 1. High: Yarn's default rc chain remains executable configuration authority

`Materialize` invokes Yarn with `--use-yarnrc` but without
`--no-default-rc` (`internal/yarnclassicsource/materialize.go:300`). Yarn
Classic 1.22.22 merges the supplied rc with its ordinary project, ancestor,
home, system, and environment-discovered rc chain. A direct probe against the
exact pinned Yarn showed an omitted project `.yarnrc` setting
`--install.modules-folder ambient_modules` in `config-with-default.log`; the
same probe with `--no-default-rc` excluded it in `config-closed.log`.

This is not merely a missing parser check. `Parse` validates only the caller's
`Configuration` map (`lock.go:527-573`), while capture copies the complete
project and verifies only config entries already present in that map
(`capture.go:117-137`). Therefore a project `.yarnrc` or `.npmrc` omitted from
the request can affect the manager before reconciliation. The real-Yarn test
repeats the same open argv at `conformance_test.go:518`.

Required rework:

- disable Yarn's default rc discovery and supply only the task-private,
  admitted configuration (`--no-default-rc` plus the private rc), or prove an
  equivalently closed mechanism;
- enumerate and bijectively reconcile all project/workspace Yarn/npm config
  files against the parsed configuration, rejecting omitted files before any
  manager process starts;
- add a real regression with poisoned project, ancestor, home, system/env
  config where the configured layout and closure identity remain unchanged.

### 2. High: workspace manifest enumeration is caller-selected rather than complete

`reconcileWorkspaces` derives the workspace set solely from the supplied
`Manifests` map (`lock.go:458-488`). Capture copies the whole project but only
checks those already-listed manifest paths (`capture.go:117-131`). With a root
`packages/*` declaration, an additional on-disk `packages/extra/package.json`
omitted by the caller enters the admitted project tree and is visible to Yarn,
although it is absent from the closed graph. Post-install extra-package
reconciliation is too late: the unrepresented workspace has already influenced
manager resolution/layout and execution.

Required rework:

- discover root/workspace manifests from the captured immutable tree using the
  closed workspace grammar;
- require an exact bijection with the parsed manifest set and fail
  `closure_lock_stale`/`closure_graph_incomplete` before manager execution;
- add missing/extra/glob-expanded workspace manifest vectors and assert zero
  Yarn starts.

### 3. Medium: peer semver reconciliation can bind an invalid peer

The local `versionSatisfies` implementation accepts every same-major version
for a caret range (`lock.go:632-655`). That makes `0.9.0` satisfy `^0.2.3`,
which is contrary to Node/Yarn semver caret semantics. It also strips
prerelease suffixes before comparison (`lock.go:668-670`). Since this helper
selects peer instances, it can produce a canonical graph different from
Yarn's actual peer/layout decision.

Required rework: use a pinned, tested Yarn-compatible semver evaluator, or
narrow the accepted grammar and reject ranges whose semantics are not exactly
implemented. Add zero-major, prerelease, compound/unsupported, ambiguous-peer,
and target-context cases.

### 4. High: required real offline/ambient-cache proof is absent

The only real Yarn test runs `exec.Command` directly and inherits
`os.Environ()` (`conformance_test.go:486-520`). It has no OS-level network deny,
does not exercise `Materialize` through a real protected executor, and does not
make ambient manager config/cache inaccessible. The fake runner checks permit
fields and simulates an install, but cannot prove S03/S08/N12 or the acceptance
criterion that replay is frozen/offline from empty ambient state.

Required rework: execute the exact pinned Yarn through the real protected
executor/network-denial harness with a minimal explicit environment, empty
ordinary cache/home/config, and poisoned inaccessible ambient cache/config.
Assert the same graph/tree receipts, audited network `none`, missing-artifact
failure, and zero later action/publication for negative branches.

## Verification

The existing focused gates are green but do not cover the findings above:

- `go test -count=1 ./internal/yarnclassicsource` — exit 0
- `go test -race -count=1 ./internal/yarnclassicsource` — exit 0
- `go vet ./internal/yarnclassicsource` — exit 0
- `golangci-lint run ./internal/yarnclassicsource/...` — exit 0
- exact Yarn 1.22.22 rc probe — reproduced default project rc ingestion; the
  `--no-default-rc` control excluded it

Because the runtime configuration/workspace graph is not closed before Yarn
starts and the mandatory real offline proof is missing, the task does not yet
satisfy its acceptance criteria. This is ordinary implementation rework, not a
stop-the-line blocker.
