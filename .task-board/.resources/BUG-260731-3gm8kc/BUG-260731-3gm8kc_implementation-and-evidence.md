# BUG-260731-3gm8kc — curator-interop-lifecycle-vector-gate

**Role:** developer · **Run:** `RUN-260731-7271a1` · **Date:** 2026-07-31
**Board status at handoff:** `to-review`

---

## 1. What was wrong

`internal/interop/golden_test.go` accepted the manager-lifecycle vector only at
exactly 2 launcher, 3 bootstrap, 3 upgrade and 2 dry-run cases:

```go
if len(vector.LauncherCases) != 2 || len(vector.BootstrapCases) != 3 ||
   len(vector.UpgradeCases) != 3 || len(vector.DryRunCases) != 2 {
```

Protocol 1.0.0-rc.6 (`curator-spec` `671888e`) added a third dry-run case,
`compiled-cache-miss-is-read-only`, so every root at or after rc.6 failed that
gate outright. That is the first step of `Implementations` on curator-spec
PR 14 and PR 15, before any other package runs, which is why all three cells
were red on both PRs while `Specification`, `Formatting` and `Links` passed.

Counting entries was also the *weaker* check: it rejected a root that merely
added a case while accepting one that renamed a case the manager depends on.
Section 5 proves both halves of that claim.

## 2. What changed

### Curator — `internal/interop/golden_test.go`

`TestManagerLifecycleVectors` now names the behaviours it requires
(`requireLifecycleCases`) and validates every case the root actually publishes:

| Group | Required by name |
| --- | --- |
| launcher | `skill-command-without-shell-activation`, `declared-system-command-without-profile` |
| bootstrap | `missing-config-if-missing`, `existing-config-if-missing`, `if-missing-with-force` |
| upgrade | `selected-project-closure`, `all-projects-deduplicate`, `global-closure` |
| dry-run | `project-upgrade`, `global-upgrade` |

Every launcher case still has to carry both platforms, three path roles and the
three preservation flags; bootstrap outcomes are now checked for all three cases
(`created` / `unchanged-success` / `usage-error`, previously only the last two);
the deduplication requirement and the ≥8 forbidden-surface floor per dry-run
case are unchanged.

`TestManagerCompiledCacheMissDryRunVector` is the focused regression for the
rc.6 case. It requires `artifact_executed = false`,
`operation_private_state_after = absent`, the six persistent surfaces only a
compiled build can touch (`toolchain-probe-memo`, `module-cache`,
`go-build-cache`, `compiled-artifact-cache`, `cache-build-lock`,
`manager-home-lock`), a non-empty forbidden Go-command list disjoint from the
allowed one, and — the part that is not a tautology — that every value in
`reported_build_outcomes` is a `install.BuildOutcome` the planner can actually
produce. Against a root that predates rc.6 it skips with a reason the CI gate
classifies as `root-content` and records by name (§4).

### Curator — `.github/ci/toolchain-identity.sh`, `.github/ci/gate-selftest.sh`

Not the lifecycle vector, but required to see this branch's CI at all: the
identity gate asserted `go env GOENV = off`. That command reports the per-user
env *file*, and go 1.25 — the version `go.mod` pins — prints nothing when
`GOENV=off`. Every Go-consuming job died at that step, which is why Curator CI
has been red on `main` since `cfffd7cd`. The stub in `gate-selftest.sh` echoed
`off`, so the self-test never saw the real spelling. The gate now accepts the
empty spelling, still fails closed on any value that names a file, and
self-tests both.

### curator-spec — `.github/workflows/implementations.yml`

The Go manager pin advances `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8` →
`bd6ba08acda3dc801512c408c759ac0ac6f79f26`, on both PR branches.

## 3. Published

| Repo | Branch | Commit | Signature | PR |
| --- | --- | --- | --- | --- |
| curator | `task/BUG-260731-3gm8kc-lifecycle-vector-gate` | `fee35c87886d6cb2737b3cad2fc077ab005e79f3` (gate) → `bd6ba08acda3dc801512c408c759ac0ac6f79f26` (identity) | `G` (both) | **9** (new, → `main`) |
| curator-spec | `release/v1.0.0-rc.6` | `5a06a66` → `b07ef1d51dc6ebcc04cf59a20375d191bf82f6bb` | `G` | **14** |
| curator-spec | `task/BUG-260731-2rhy74-marker-v2-fixture` | `bc83dd1` → `2629aecff19a33e8cd1b5ebcfd898894ff1eeae0` | `G` | **15** |

All three branch heads verified against `git ls-remote`; all worktrees clean.
Nothing was tagged or released. The pin is a full immutable commit ID as
`conformance/README.md` §4 requires; §7 records the one open question about it.

## 4. Evidence — commands and real exit codes

Every command below was run directly, not through a pipe, and the exit code is
the command's own.

### Curator, at the published `bd6ba08`, worktree `.temp/BUG-260731-3gm8kc/curator`

| Command | Conformance root | Exit |
| --- | --- | ---: |
| `go test -count=1 ./internal/interop ./internal/closure ./internal/skillspec` — the Implementations step, verbatim | PR 15 `0c81c1f8` | **0** |
| same | PR 14 / rc.6 `671888e` | **0** |
| `go test -count=1 -timeout 30m ./internal/interop/ -v` — the Interop job, verbatim | `SPEC_PIN` rc.3 `00b1688` | **0** (17 passed, 1 skip) |
| `make check-ci` — gofmt, `go vet ./...`, ledger-consistency, test-gate + platform-case gate | `SPEC_PIN` rc.3 `00b1688` | **0** |
| `bash .github/ci/gate-selftest.sh` | — | **0** — 75 passed, 0 failed (was 74; one new case) |
| `bash .github/ci/no-broad-suppression.sh` | — | **0** |
| `golangci-lint run` (v2.12.2, the exact version CI pins) | — | **0**, `0 issues` |
| `gofmt -l cmd internal` | — | **0**, no output |
| `git diff --check` | — | **0** |

`make check-ci` on the committed tree reports `test-gate: go test exit=0,
platform-case gate exit=0`, `suite-plan: served=33 deferred=7 excluded=0`. The
new skip is in the gate's own evidence, classified rather than hidden:

```
internal/interop	TestManagerCompiledCacheMissDryRunVector	root-content	allowed-root-content	this conformance root publishes no compiled-cache-miss-is-read-only dry-run case
```

### The failure this fixes, reproduced before the change

| Command | Root | Exit |
| --- | --- | ---: |
| `go test -count=1 -run TestManagerLifecycleVectors ./internal/interop` at `cfffd7c` | PR 15 `0c81c1f8` | **1** — expected red: `manager lifecycle vector is incomplete` |

Full pre-change consumer sweep against the rc.6 root (`cmd/curator`,
`internal/closure`, `internal/runtimestore`, `internal/buildcache`,
`internal/interop`, `internal/skillspec`, `internal/install`): exit **1**, with
exactly two failures — `internal/interop` (fixed here) and `internal/install`
(§7). Everything else was already green.

### The identity gate, against a real go1.25.5 rather than the stub

| Script | Environment | Exit |
| --- | --- | ---: |
| this branch | `GOTOOLCHAIN=local GOENV=off` | **0** — prints `toolchain: GOENV=<none>` |
| `main`'s (`cfffd7cd`) | same | **1** — expected red: `GOENV=, not off`, the exact CI message |
| this branch | `GOENV` unset | **1** — expected red: names the per-user env file, still fails closed |

### curator-spec, at the published `2629aec` (PR 15) and `b07ef1d` (PR 14)

| Command | Where | Exit |
| --- | --- | ---: |
| `make release-check VERSION=1.0.0-rc.6` | PR 15 `2629aec` | **0** — `release gate passed for 1.0.0-rc.6 at 2629aec…` |
| `make release-check VERSION=1.0.0-rc.6` | PR 15 `bc83dd1` (intermediate) | **0** |
| `make validate` (schemas, 60 unittests, `go test ./tools/...`) | PR 14 `b07ef1d` | **0** |
| `make regenerate-check` (regenerate + `git diff --exit-code`) | PR 14 `5a06a66` | **0** — regeneration is idempotent |
| `git diff --check`, `git status --porcelain` | both | **0**, clean |

## 5. Negative controls — the new gate has teeth

Each mutation was applied to a copy of a real conformance root and run against
the published tree.

| Root | Mutation | Result | Exit |
| --- | --- | --- | ---: |
| rc.6 | `project-upgrade` renamed | `dry-run cases omit [project-upgrade]; this root publishes [project-upgrade-renamed global-upgrade compiled-cache-miss-is-read-only]` | **1** |
| rc.6 | `compiled-artifact-cache` removed from the rc.6 case | `compiled-cache-miss-is-read-only does not forbid compiled-artifact-cache` | **1** |
| rc.6 | `would-silently-succeed` added to `reported_build_outcomes` | `reports build outcome "would-silently-succeed", which the planner cannot produce` | **1** |
| rc.3 | `project-upgrade` renamed | **old gate at `cfffd7c`: exit 0** — the rename passed unnoticed | **0** |
| rc.3 | same | **new gate: exit 1** — `dry-run cases omit [project-upgrade]` | **1** |

The last two rows are the point: the replacement is strictly stronger than the
length check, not merely more permissive.

## 6. CI — the acceptance criterion

### curator-spec `Implementations`, the AC's target

| PR | Before (pin `17804ce`) | After (pin `bd6ba08`) |
| --- | --- | --- |
| **14** ubuntu / macOS / windows | fail / fail / fail | **pass / pass / pass** |
| **15** ubuntu / macOS / windows | fail / fail / fail | **pass / pass / pass** |

The pre-change failure signature, from PR 15's ubuntu job `91044069441`:
`golden_test.go:473: manager lifecycle vector is incomplete`, `--- FAIL:
TestManagerLifecycleVectors`.

At the intermediate pin `fee35c87`, **all six cells passed** (PR 14 run
`30615543452`; PR 15 run `30615573475`), with the new tests visible by name in
the job logs on both ubuntu and windows:

```
=== RUN   TestManagerLifecycleVectors
--- PASS: TestManagerLifecycleVectors (0.00s)
=== RUN   TestManagerCompiledCacheMissDryRunVector
--- PASS: TestManagerCompiledCacheMissDryRunVector (0.00s)
```

### 6a. Final pin runs — every check on both PRs is green

PR 14 at `b07ef1d` (run `30616098340`) and PR 15 at `2629aec` (run
`30616107892`), read back with `gh pr checks`:

| Check | PR 14 | PR 15 |
| --- | --- | --- |
| Implementations (ubuntu-latest) | **pass** | **pass** |
| Implementations (macos-latest) | **pass** | **pass** |
| Implementations (windows-latest) | **pass** | **pass** |
| Specification (ubuntu / macOS / windows) | pass / pass / pass | pass / pass / pass |
| Formatting, Links | pass, pass | pass, pass |

Eight of eight checks pass on each PR. The acceptance criterion — Implementations
green on ubuntu, macOS and Windows for both PR 15 and PR 14 — is met.

### 6b. Curator PR 9's own CI, against the control branch

Run `30615765014` (PR 9, `bd6ba08`) versus run `30616027892`
(`ci/goenv-control-BUG-260731-3gm8kc` = `main` + only the identity repair).
Settled job for settled job, the two are identical:

| Job | PR 9 | Control | Reading |
| --- | --- | --- | --- |
| Interop conformance gate | **pass** | **pass** | was failing at the identity gate before |
| Gate self-test — ubuntu / macOS / windows | pass / pass / pass | pass / pass / pass | 75 assertions, incl. the new GOENV case |
| Naming gate | **pass** | **pass** | |
| Test (macos-latest) | **pass** | **pass** | the lane the control inventory covers |
| Race (macos-latest) | **pass** | still running at handoff | the control's slowest job; not needed for the comparison |
| Lint | fail | fail | same two `unused` findings — §7a |
| Test (ubuntu-latest) | fail | fail | same six `cmd/curator` compiled cases — §7a |
| Race (ubuntu-latest) | fail | fail | same six cases under `-race` |
| Test (windows-latest) | fail | fail | same `undefined: decodeHelperOutput` vet error — §7a |
| Candidate suite | skipping | skipping | dispatch-only job, no candidate supplied |

Every job this change could affect passes. Every failing job fails identically
on a branch that does not contain this change.

## 7. Three findings this task did not fix, and why

Each is pre-existing, in a different subsystem, and would have to be repaired by
absorbing another item's scope. All three are raised on the board.

### 7a. Curator CI is red on `main`, in two independent ways — `BUG-260731-lepevi`, `BUG-260731-11bpa4`

`main` `cfffd7cd` has never passed CI: every Go job died at the identity gate
(§2), which this PR repairs. With that gate honest, two further pre-existing
failures become visible:

* **Lint** — `internal/godriver/controls_other.go:35 func (*controlDomain).destroy
  is unused` and `internal/transaction/namespace.go:310 func
  existingNamespaceAncestor is unused`. Both are linux-build-only, which is why
  `golangci-lint run` on darwin reports 0 issues.
* **Test (ubuntu)** — six `cmd/curator` compiled-build cases fail with
  `rc5-native-control-inventory-v1 defines no record for host linux`. Linux is
  the platform the inventory deliberately does not cover; the spec's own
  `conformance/README.md` §5 records that linux qualification is a separate open
  item.
* **Test (windows)** — `go vet` rejects the package before any test runs:
  `internal\runtimestore\targets_windows_test.go:97:14: undefined:
  decodeHelperOutput`.

**Proven pre-existing, not inherited from this change.** A control branch,
`ci/goenv-control-BUG-260731-3gm8kc` = `main` + *only* the two-file identity
repair, was dispatched through the same workflow (run `30616027892`). It
reproduces all three failures with identical signatures — same two `unused`
findings, same `decodeHelperOutput` vet error — while `Interop conformance
gate`, `Naming gate` and `Gate self-test` on all three runners pass. The
lifecycle-vector change is not in that branch at all. `git diff --stat
cfffd7c..bd6ba08` touches exactly three files: `internal/interop/golden_test.go`,
`.github/ci/toolchain-identity.sh`, `.github/ci/gate-selftest.sh` — none of them
in `cmd/curator`, `internal/godriver`, `internal/transaction` or
`internal/runtimestore`.

### 7b. `internal/install` cannot consume an rc.6 root either

`TestAuthoritativeDryRunCasesMutateNothingPersistent` binds every published
dry-run case to a real planning run selected by its `scope`, and fails closed on
a scope it cannot execute:

```
dryrun_conformance_test.go:189: published dry-run scope "multi-project" has no executable binding
```

The rc.6 case publishes `scope: multi-project`. Fixing it means building a real
multi-project dry-run machine in `internal/install` — product work, not a gate
fix, and explicitly outside this task's scope ("Curator interop lifecycle-vector
consumer plus the immutable Curator pin"). It affects neither gate in the AC:
`internal/install` is not in the Implementations package set, and Curator CI
pins rc.3, where the case does not exist. Weakening that `default:` branch to
skip unknown scopes was rejected — it would silently unassert a published case,
which is exactly what this task is fixing elsewhere.

## 8. Open question for review

The pin now names `bd6ba08`, the head of an **unmerged** Curator PR.
`conformance/README.md` §4 requires a full immutable commit ID — satisfied — and
adds that "a pin may advance only after that implementation has passed the same
released suite in its own required CI". Curator's required CI is red at that
commit for the three unrelated reasons in §7a, and there is no Curator commit
that both accepts the rc.6 vector and passes CI, because `main` itself does not.
The alternative is leaving `Implementations` red on both PRs. Recommended
sequence: land PR 9, then advance the pin to the merged `main` commit — a
one-line follow-up — and land `BUG-260731-lepevi` / `BUG-260731-11bpa4` to make
Curator CI honest again.

## 9. What is deliberately left unchecked

Two checklist items are **not** ticked, because the exact gate they name did not
come back green:

* *"Attach evidence, hand off to independent Opus review, and require Curator
  plus PR 15 CI green."* — PR 15's CI is green (8/8), evidence is attached, the
  handoff is made. **Curator's own CI is not green**: `Lint`, `Test (ubuntu)`,
  `Race (ubuntu)` and `Test (windows)` fail for the three pre-existing reasons
  in §7a, proven identical on a branch that does not contain this change. Making
  them green means absorbing `BUG-260731-lepevi` and `BUG-260731-11bpa4`, one of
  which touches the linux qualification the spec itself defers.
* *"Lint clean."* — `golangci-lint run` at the pinned v2.12.2 exits **0** with
  `0 issues` on darwin, and reports nothing in any file this change touches. The
  CI `Lint` job runs on linux and fails on two `unused` findings in
  `internal/godriver` and `internal/transaction`. The repo's gate is the CI job,
  so the item stays unchecked.

The evidence branch `ci/goenv-control-BUG-260731-3gm8kc` is kept until review so
the comparison can be re-run; it should be deleted afterwards.

## 10. Artifacts

Local logs, mutated roots and CI artifacts under
`.temp/BUG-260731-3gm8kc/` (gitignored):
`logs/01-baseline-rc6-consumers.log`, `logs/02-check-ci-rc3-pin.log`,
`logs/04-gate-selftest.log`, `logs/06-golangci-lint.log`,
`logs/07-interop-job-rc3.log`, `logs/08-check-ci-committed-tree.log`,
`logs/09..14` (spec gates), `negative/` (four mutated conformance roots),
`ci-artifacts/` (downloaded CI evidence).
