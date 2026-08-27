# BUG-260731-3a5q1p — curator-install-dryrun-rc6-multiproject-binding

Implementation and evidence. Base: Curator `main` @ `cfffd7c` (`Add declarative compiled skill builds`).
Working branch: `task/BUG-260731-3a5q1p-multiproject-dryrun` (worktree `.temp/BUG-260731-3a5q1p/worktree`).

## 1. Reproduction against the authoritative rc.6 root

Authoritative rc.6 root: `relux-works/curator-spec` @ `b07ef1d51dc6ebcc04cf59a20375d191bf82f6bb`
(`origin/release/v1.0.0-rc.6`), `conformance/v1/manifest.json` `protocol_version: 1.0.0-rc.6`,
manifest sha256 `72c5d717027ca096b14bc32f5d60bb740676974e9429f3d09b730897e5fba89b`.
Materialised at `.temp/BUG-260731-3a5q1p/spec-rc6`.

```
CURATOR_CONFORMANCE_ROOT=<rc.6>/conformance/v1 \
  go test -count=1 -run TestAuthoritativeDryRunCasesMutateNothingPersistent ./internal/install
--- FAIL: .../compiled-cache-miss-is-read-only
    dryrun_conformance_test.go:189: published dry-run scope "multi-project" has no executable binding
exit 1
```

Confirmed pre-existing, exactly as both reviewers of BUG-260731-3gm8kc recorded.

The gap is larger than the scope name. rc.6 adds a third authoritative dry-run case,
`compiled-cache-miss-is-read-only` / `scope: multi-project`, which publishes **28 forbidden
persistent effects** (up from 9) plus five fields the older cases never carried:
`allowed_go_commands`, `forbidden_go_commands`, `artifact_executed`, `logical_cache_key`,
`operation_private_state_after`, `reported_build_outcomes`. Of the 28 effects, **18 had no
binding either**, so `assertNoEffect`'s own `default:` branch was the next failure after the
scope switch.

## 2. What was implemented

One file of product-adjacent test code: `internal/install/dryrun_conformance_test.go`
(+727/-46), plus one line of `.github/ci/root-artifacts.tsv`.

### 2.1 The multi-project binding

Curator's multi-project operation is `install --all` / `upgrade --all`
(`cmd/curator/main.go:520`, `:635`): every selected project target planned in one run against
one manager home, one skills root and one shared `FetchedRepos` deduplication set.
`profiles/manager.md §2.5` takes those targets in the unsigned byte order of their canonical
project identities — reproduced through `managerlock.CanonicalProjects`. A dry run acquires no
lock at all (`managerlock.ErrDryRun`), so canonical order is the only ordering obligation left.

Both projects declare the same compiled command, so both derive **one shared logical cache key**
and both miss the same protected entry. That shared miss is the surface the case exists to
protect: two projects looking at one absent entry must still leave the whole machine alone.

Reachability is real, not simulated:

| boundary | used |
|---|---|
| trusted Go toolchain | **real** `goToolchain` over a test-owned `privateRoot`. The real probe runs exactly `go telemetry off`, `go version`, `go env -json` — the published `allowed_go_commands` verbatim. |
| protected build cache | **real** `buildcache.Store` via zero-value `BuildDeps.Cache`. The boundary is really inspected for the shared key. |
| marker generation | **real** (`markerGeneration`) |
| source-aware builder | replaced by `refusingBuilder`, which records and refuses. It is the only thing in an installation that runs `go list` / `go build` — the published `forbidden_go_commands`. |

Published fields bound:

- `reported_build_outcomes` — each project's reported outcome must be in the published set, and
  against an empty protected cache must be `would-preflight-and-build` (§2.4).
- `logical_cache_key` — the case key must equal the published `compiled_build_fixture.cache_key`,
  the fixture's `execution_policy` must equal `buildmeta.ExecutionPolicy`, and the two projects
  must derive one identical non-empty key. The *absolute* published value is derived from the
  published toolchain/target, which no host has; `internal/buildmeta` binds that derivation
  against `expected/build-driver/cache-key.txt`, and the code says so.
- `artifact_executed: false` — nothing staged, no artifact path reported for a miss.
- `operation_private_state_after: absent` — no toolchain base survives, and once the test-owned
  private root is removed nothing named `curator-install-private-*` is left in the isolated
  temporary directory, so every root the run allocated for itself was released too.

Both `default:` branches are untouched: an unbound scope and an unbound effect are still fatal
(proved in §4, mutations 5 and 6).

### 2.2 The 18 new effect bindings

Every one names a path Curator actually uses. Where several effects share one root, that is
Curator's layout, and the code says why.

| effect | surface |
|---|---|
| `source-checkout` | `<skillsRoot>/clonable`, `<home>/dev` (rc.3's `source-clone` renamed) |
| `toolchain-probe-memo` | no `go-probe-base-*` anywhere; Curator memoises no toolchain identity |
| `module-cache` | no `gopath` / `gomodcache` anywhere (godriver session layout) |
| `go-build-cache` | no `gocache`, no `go-build-base-*` anywhere |
| `compiled-artifact-cache` | `<home>/cache/build` |
| `quarantine` | `<home>/cache/build` + no `.quarantine-*` (buildcache's own prefix) |
| `permission-repair` | `<home>/cache/build` — the only boundary whose modes a manager establishes or repairs; an untrusted entry is quarantined, never permission-repaired |
| `revocation-state` | `<home>/state` — Curator keeps no separate revocation store; a revocation is decided from signed registry records and configuration, so registry rollback state below the state root is the only surface it could leave |
| `project-lock` | `<home>/state/locks/v1/projects` + no `*.lock` under any project |
| `cache-build-lock` | `<home>/state/locks/v1/build` |
| `manager-home-lock` | `<home>/state/locks` (the home lock file sits directly in the versioned root) |
| `journal` | `<home>/state/transactions` |
| `backup` | no `.curator-txn-*` sidecar beside any live target |
| `runtime-tree` | `<home>/runtime` |
| `context-tree` | `<project>/.agents/skills` + every declared agent's `adapters.AgentPaths` root |
| `environment-file` | `envfiles.ShellName` / `PowerShellName` in every project and global env dir |
| `install-marker` | no `marker.Name` (`.csk-install.json`) anywhere |
| `command-shim` | `<project>/.agents/bin`, `<home>/global/bin` |
| `adapter-ledger` | no `adapters.LedgerName` (`.csk-managed.json`) anywhere |
| `adapter-mirror` | `<home>/global/skills`, `scopes.HybridSkillsRoot(home)` |
| `consumer-ledger` | `<home>/`+`scopes.ConsumersName` |
| `gc-metadata` | consumer ledger + runtime store + protected cache; collection keeps no metadata of its own |

`project-artifacts` / `context-tree` / `environment-file` / `command-shim` now assert over **every**
project the operation planned, not only the first.

### 2.3 Bindings are probes, and a second test runs them backwards

`assertNoEffect` now delegates to `effectSurfaces`, which *reports* where a surface is visible
instead of asserting it away. That makes the binding runnable in both directions, and
`TestDryRunEffectBindingsSeeWhatARealOperationWrites` runs it the other way: a **real** project
install, a **real** global install and a **real** non-dry `managerlock` operation produce these
surfaces in their production locations, and every binding must then see them.

A binding that cannot see its own surface would make its absence after a dry run prove nothing.
That test now fails instead. **Completeness is part of it**: every effect any published case
names must be witnessed, so a future effect cannot gain a binding without gaining a witness.

Surfaces the real operation reaches in production locations (21 of 27): `source-fetch`
(`FETCH_HEAD`), `source-checkout`, `snapshot-cache`, `response-cache`, `audit-state`,
`registry-state`, `revocation-state`, `project-lock`, `cache-build-lock`, `manager-home-lock`,
`journal`, `runtime`/`runtime-tree`, `context-tree`, `environment-file`, `install-marker`,
`command-shim`, `adapter-ledger`, `adapter-mirror`, `consumer-ledger`, `gc-metadata`,
`project-artifacts`, `global-artifacts`.

Surfaces produced explicitly, with the reason recorded beside each in the code:

- `configuration` — a real install of an already-registered project rewrites nothing.
- `compiled-artifact-cache`, `quarantine`, `permission-repair` — publishing a real entry needs a
  compiled artifact and a real receipt; `internal/buildcache` owns and asserts those paths.
- `backup` — a committed transaction removes its own sidecars, so no completed run can leave one;
  `internal/transaction` asserts the mid-commit state.
- `toolchain-probe-memo`, `module-cache`, `go-build-cache` — allocated through the **same**
  `privateRoot` allocator production uses; a dry run establishes no session, so the session base
  and its Go caches have no compiler-free real path in this package at all.

### 2.4 The suite is no longer parallel — deliberate, and not a weakening

Each case now calls the package's existing `isolateTempDir(t)` so the operation-private roots
under test are the only ones in `TMPDIR`/`TMP`/`TEMP`. `t.Setenv` forbids a parallel ancestor, so
`t.Parallel()` had to go. This is a scheduling change, not an assertion change — it is what makes
the published `operation_private_state_after` contract bindable for real, and it strengthens the
older `project` and `global` cases too. Cost: ~0.2 s at rc.3 for two sequential cases.

### 2.5 `root-artifacts.tsv`

`internal/install` declared only `vectors/build-drivers.json`, yet this test has always read
`vectors/manager-lifecycle.json` unguarded. Added it, so a root without it **defers**
`internal/install` (per the table's own contract) instead of failing it on an `open` error.
No partition change for rc.3 or rc.6 — both publish it (`served=40 deferred=0` on rc.6).

## 3. Evidence — every command, real exit code

All on macOS (darwin/arm64), Go per `go.mod`, in the task worktree. Logs in
`.temp/BUG-260731-3a5q1p/logs/`.

| # | command | exit | note |
|---|---|---|---|
| 1 | `CURATOR_CONFORMANCE_ROOT=<rc.6> go test -count=1 -run TestAuthoritativeDryRunCasesMutateNothingPersistent ./internal/install` **before the fix** | **1** | expected-red: the reproduction (`scope "multi-project" has no executable binding`) |
| 2 | `CURATOR_CONFORMANCE_ROOT=<rc.6> go test -count=1 -v -run 'TestAuthoritativeDryRunCasesMutateNothingPersistent\|TestDryRunEffectBindingsSeeWhatARealOperationWrites' ./internal/install` | **0** | all 3 cases PASS; multi-project 5.19 s (two real toolchain probes) |
| 3 | `CURATOR_CONFORMANCE_ROOT=<rc.6> go test -count=1 ./internal/install/...` | **0** | `install-rc6-final.log`; install 128.19 s, atomicity 115.63 s — **the acceptance criterion** |
| 4 | `CURATOR_CONFORMANCE_ROOT=<rc.3-pin> go test -count=1 -v -run '<both tests>' ./internal/install` | **0** | `install-rc3-dryrun.log`; rc.3 publishes 2 cases, both still PASS — no regression |
| 5 | `CURATOR_CONFORMANCE_ROOT=<rc.6> go test -race -count=1 -run '<both tests>' ./internal/install` | **0** | `install-race-rc6.log`, 37.90 s |
| 6 | `CURATOR_CONFORMANCE_ROOT=<rc.6> CI_REQUIRE_FULL_ROOT=1 bash .github/ci/test-gate.sh` on **this branch** | **1** | `test-gate-rc6.log`. `go test exit=1`, platform-case gate exit=0. Sole failure: `internal/interop TestManagerLifecycleVectors` at `golden_test.go:488` — `len(vector.DryRunCases) != 2`, the **pre-existing** hardcoded length gate on `main`. Not in my diff. This is BUG-260731-3gm8kc / PR 9's scope (fixed there by `fee35c8 Gate the lifecycle vector by name`). Gate evidence records `internal/install` and all three dry-run cases as `pass`. |
| 7 | same gate on **`main` + PR 9 head `bd6ba08` + this patch** | **0** | `test-gate-rc6-combined.log`. `go test exit=0`, platform-case gate exit=0, `served=40 deferred=0 excluded=0`, `CI_REQUIRE_FULL_ROOT=1`. **This is the proof that advancing `SPEC_PIN` to rc.6 does not turn CI red on this test.** |
| 8 | `golangci-lint run` (v2.12.2, the CI-pinned version) | **0** | `lint-final.log`, `0 issues` |
| 9 | `gofmt -l cmd internal` | **0** | no output |
| 10 | `go vet ./...` | **0** | `vet.log` |
| 11 | `bash .github/ci/no-broad-suppression.sh` | **0** | `no-broad-suppression: ok` |
| 12 | `bash .github/ci/ledger-consistency.sh` | **0** | `49 rows checked across linux darwin windows`; no ledger row needed — no build-tagged file added |

Gate evidence (`test-gate-rc6/observed-cases.tsv`) records, with no skip:

```
pass internal/install TestAuthoritativeDryRunCasesMutateNothingPersistent
pass internal/install TestAuthoritativeDryRunCasesMutateNothingPersistent/compiled-cache-miss-is-read-only
pass internal/install TestAuthoritativeDryRunCasesMutateNothingPersistent/global-upgrade
pass internal/install TestAuthoritativeDryRunCasesMutateNothingPersistent/project-upgrade
pass internal/install TestDryRunEffectBindingsSeeWhatARealOperationWrites
```

### Not run

- **Linux and Windows CI.** Requires pushing the branch; see §5. No local Linux runner exists on
  this host (no docker/podman; lima cannot boot a vz VM here).
- **`go test` with the rc.3 root *exported* over the whole package** fails on
  `TestAuthoritativeCacheOutcomesDriveInstallation` — rc.3 publishes no `vectors/build-drivers.json`.
  Pre-existing and expected: `root-artifacts.tsv` exists precisely so `suite-plan.sh` **defers**
  `internal/install` under such a root. It is not how CI runs rc.3.

## 4. The bindings were proved falsifiable — six mutations, each reverted

| # | mutation | result |
|---|---|---|
| 1 | bind `compiled-artifact-cache` to `<home>/no-such-surface` | witness test FAILS: *"the binding for `compiled-artifact-cache` reported nothing after a real operation produced it, so its absence after a dry run proves nothing"* |
| 2 | plan only one project | FAILS: *"the multi-project binding planned 1 projects, want at least two"* |
| 3 | second project declares a different compiled skill | FAILS: *"the projects derived the different logical cache keys [sha256:4443f4… sha256:b751be…], so no shared entry was under test"* |
| 4 | create `<home>/cache/build/go-v1` right after the dry runs | FAILS: *"the dry run produced the forbidden persistent effect `snapshot-cache` at …/cache"* |
| 5 | rename the published scope to `fleet-wide` in a copy of the root | FAILS: *"published dry-run scope `fleet-wide` has no executable binding"* — the `default:` branch is intact |
| 6 | append effect `telemetry-ledger` to the published case | **both** tests FAIL: *"published forbidden effect `telemetry-ledger` has no executable binding"* |

## 5. Not done — one DoD item is externally blocked

**"Publish a signed Curator PR targeting main"** could not be done from this session.

`commit.gpgsign=true` with `gpg.format=ssh` and `user.signingkey=~/.ssh/ivanopcode`. That key is
passphrase-protected and the passphrase is not obtainable here:

```
$ eval "$(ssh-agent -s)"; ssh-add --apple-load-keychain; ssh-add -l
The agent has no identities.
$ security find-generic-password -s "SSH: /Users/iv/.ssh/ivanopcode"
security: SecKeychainSearchCopyNext: The specified item could not be found in the keychain.
$ printf probe | ssh-keygen -Y sign -n git -f ~/.ssh/ivanopcode -P ''
Load key "…/ivanopcode": incorrect passphrase supplied to decrypt private key
$ ssh -o BatchMode=yes -T git@github.com
git@github.com: Permission denied (publickey).
```

So neither a signed commit nor a push to `origin` is possible, and therefore no PR and no
Linux/Windows CI run. The work is complete and verified locally; it needs a human with an
unlocked agent for exactly these steps.

### Exact handoff commands

The change is saved as `BUG-260731-3a5q1p_multiproject-dryrun-binding.patch`
(also present uncommitted on branch `task/BUG-260731-3a5q1p-multiproject-dryrun` in
`.temp/BUG-260731-3a5q1p/worktree`, verified byte-identical to the patch).

```bash
eval "$(ssh-agent -s)" && ssh-add ~/.ssh/ivanopcode      # interactive: enter the passphrase
cd .temp/BUG-260731-3a5q1p/worktree
git add internal/install/dryrun_conformance_test.go .github/ci/root-artifacts.tsv
git commit                                               # message drafted in §6
git push -u origin task/BUG-260731-3a5q1p-multiproject-dryrun
gh pr create --base main --head task/BUG-260731-3a5q1p-multiproject-dryrun \
  --title 'Bind the multi-project dry-run case'
```

Then the three-platform rc.6 proof. Mirroring what BUG-260731-3gm8kc did with
`origin/ci/goenv-control-BUG-260731-3gm8kc`, push a throwaway CI branch carrying this patch
**and** PR 9's interop fix (the tree already validated green in evidence row 7), and dispatch the
non-default candidate lane — it runs the full gate with `CI_REQUIRE_FULL_ROOT=1` on
`ubuntu-latest`, `macos-latest` and `windows-latest`:

```bash
git push origin HEAD:ci/rc6-candidate-BUG-260731-3a5q1p   # from .temp/BUG-260731-3a5q1p/combined
gh workflow run ci.yml --ref ci/rc6-candidate-BUG-260731-3a5q1p \
  -f candidate_ref=b07ef1d51dc6ebcc04cf59a20375d191bf82f6bb \
  -f candidate_manifest_sha256=72c5d717027ca096b14bc32f5d60bb740676974e9429f3d09b730897e5fba89b
```

Dispatching on the PR branch alone will be red on `internal/interop` until PR 9 merges — that is
evidence row 6, not a regression from this change.

## 6. Commit message drafted for the handoff

```
Bind the multi-project dry-run case

The rc.6 conformance suite publishes a third authoritative dry-run case,
`compiled-cache-miss-is-read-only`, in scope `multi-project`.
`TestAuthoritativeDryRunCasesMutateNothingPersistent` had no executable
binding for that scope and no binding for the twenty-eight persistent
effects it forbids, so against an rc.6 root the test failed on its own
`default:` branch. Both branches stay exactly as strict as they were:
an unbound scope and an unbound effect are still fatal.

The binding is Curator's own multi-project operation. `install --all`
and `upgrade --all` plan every selected project target in one run --
one manager home, one skills root, one shared fetch-deduplication set
-- and profiles/manager.md 2.5 takes those targets in the unsigned byte
order of their canonical project identities. Both projects declare the
same compiled command, so both derive one shared logical cache key and
both miss the same protected entry: the shared surface the case exists
to protect. The trusted toolchain and the protected build cache are the
real ones, so the compiled surfaces the case forbids are genuinely
reachable rather than vacuously absent -- the real probe runs the three
package-independent commands the case allows, and the real boundary is
really inspected for the shared key. Only the source-aware builder is
replaced, by one that records and refuses every call a dry run must
never make.

Each effect binding is now a probe that reports where its surface is
visible rather than an assertion that fires, so it can be run in both
directions, and `TestDryRunEffectBindingsSeeWhatARealOperationWrites`
runs it the other way: a real project install, a real global install
and a real locked operation produce these surfaces in their production
locations, and every binding must then see them. A binding that cannot
see its own surface would make its absence after a dry run prove
nothing, and now fails instead. Completeness is part of that test, so a
future effect cannot gain a binding without gaining a witness.

The suite is no longer parallel: each case isolates the process
temporary directory so the operation-private roots under test are the
only ones in it, and `t.Setenv` forbids a parallel ancestor. That
isolation is what binds the published `operation_private_state_after`
contract for real.

`root-artifacts.tsv` now declares the manager-lifecycle vector the
package has always read unguarded, so a root without it defers
`internal/install` instead of failing it.
```

## 7. Review notes — the three judgement calls worth arguing with

1. **Two sequential dry runs is what "multi-project" means for a dry run.** Curator has no
   single-call multi-project planner; `--all` is the operation, and §2.5's only multi-project
   obligation for planning is canonical-identity order — which is reproduced. A dry run acquires
   no lock at all, so nothing else in §2.5 applies to it. If a reviewer thinks the case demands a
   single-call API, that is a product change, not a test change, and should be raised as such.
2. **Several effects share one root** (`permission-repair`/`quarantine`/`compiled-artifact-cache`
   → `<home>/cache/build`; `revocation-state` → `<home>/state`). That is Curator's layout, not
   binding laziness: it keeps no separate revocation store, and the protected cache boundary is
   the only surface whose modes it establishes or repairs. Each case says so in a comment.
3. **Six of 27 witnesses are produced rather than reached** (§2.3), because no compiler-free
   operation in this package can reach them. Each names the package that does assert it. The
   alternative — a real `go build` under the manager-worker policy inside a unit test — is a much
   heavier dependency for a strictly weaker claim about paths.

## 8. Tool readiness

`golangci-lint` was absent on this host. Installed task-locally at the exact version
`.github/workflows/ci.yml` pins for `golangci/golangci-lint-action@v7`, so the local result and
CI's are the same tool:

```bash
GOBIN=.temp/BUG-260731-3a5q1p/bin go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
.temp/BUG-260731-3a5q1p/bin/golangci-lint run   # → 0 issues, exit 0
```

Nothing was installed system-wide and no project file references the task-local binary.

`git worktree` submodules are not populated automatically. Both worktrees needed
`git submodule update --init --recursive`; without it `internal/ui` fails to build and
`golangci-lint` reports a spurious `typecheck` error on the missing `tuitestkit` replacement
directory. Not a code problem — worth knowing before reading such a failure as a regression.

Scratch layout, all under `.temp/BUG-260731-3a5q1p/`:

| path | what |
|---|---|
| `worktree/` | the task branch `task/BUG-260731-3a5q1p-multiproject-dryrun`, change uncommitted |
| `combined/` | `main` + PR 9 head `bd6ba08` + this patch — the tree that produced evidence row 7, and the tree to push as the CI candidate branch |
| `spec-rc6/` | the authoritative rc.6 conformance root (`curator-spec` worktree at `b07ef1d5`) |
| `logs/` | every log named in §3, plus both gate evidence directories |
| `bin/` | the task-local pinned `golangci-lint` |
