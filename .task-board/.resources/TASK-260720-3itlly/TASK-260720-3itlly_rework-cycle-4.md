# TASK-260720-3itlly — rework cycle 4

Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1zntv0/worktree`
Base HEAD: `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`. Nothing committed or staged.

Addresses the two cycle-3 reviewer findings.

## R1 — dry-run created a closure scratch root beside the probe root

### What changed

There is now exactly **one operation-private ephemeral root per installation**,
and everything a run needs outside persistent state lives inside it.

New `internal/install/private.go`:

- `operationPrivatePrefix = "curator-install-private-"` — the single prefix
  under which the manager creates ephemeral roots, so a run's complete
  temporary footprint is greppable and assertable.
- `privateRoot` — lazily created on first use, `dir(prefix)` allocates a
  uniquely named subdirectory, `remove()` drops the whole root once and is
  idempotent.
- `releasePrivateRoot` — deferred teardown that reports an unremovable private
  path as a scope-prefixed operator message, never as an install verdict. It
  mirrors the `releasePlan` contract established in cycle 3.

Wiring:

- `internal/install/install.go` and `internal/install/global.go` no longer call
  `os.MkdirTemp("", "curator-dry-run-")` / `"curator-global-dry-run-"`. The
  dry-run closure workspace is `private.dir("closure-")`.
- `BuildDeps.resolve` gained the `private *privateRoot` parameter and hands it
  to `goToolchain`.
- `goToolchain.Probe` allocates `go-probe-base-*` and `goToolchain.Establish`
  allocates `go-build-base-*` **inside** the operation-private root instead of
  directly in `TMPDIR`. Both removal paths are unchanged: the probe removes its
  base on success and on failure, and `goSession.Release` still removes the
  session base through the driver's cleanup-only path.
- Defer order in both scopes is `releasePrivateRoot` registered first, then
  `releasePlan`. LIFO therefore releases the plan (and the staging root it
  owns) before the operation-private root is removed.

Observable effect: a project or global dry run now creates exactly one
temporary root, named `curator-install-private-*`, and `TMPDIR` is empty when
it returns — on the success path and on the failure path alike.

### Documented deviation from the literal requirement

The reviewer asked that "the only temporary filesystem state created by dry-run
belongs to `Toolchain.Probe`". That exact shape is not reachable, and the
reason is a hard constraint rather than an implementation shortcut:

- Dry-run closure resolution must hand a **real directory tree** to
  `skillspec.Load`, `skillcheck.Validate`, `audit.Gate*` (via `Subject.Snapshot`),
  `hashing.ContentSHA256` for registry resolution, and `buildsource.Validate`.
  `buildsource.Validate` in particular exists to detect on-disk mutation and
  holds open directory handles for `Use`/`Recheck`; a virtual filesystem would
  destroy the property it is there to provide.
- The tree can only come from `git archive`, which writes. Writing it to the
  persistent snapshot cache is forbidden — `TestDryRunTouchesNothing` and
  `TestGlobalUpgradeDryRunLeavesPersistentStateUnchanged` assert
  `<home>/cache` stays absent — so it must be ephemeral.
- The probe root cannot host it: `Toolchain.Probe` runs *after* closure
  resolution, and a plan with no build command never probes at all while its
  closure still needs the tree. Making the toolchain own skill snapshots would
  also invert the ownership this task is built around.

So the choice is between "one ephemeral root" and "two ephemeral roots", not
between "one" and "none". This cycle collapses it to one, keeps it inside the
boundary this task owns, and proves the footprint with isolated-`TMPDIR`
regressions. The pre-existing `closure.Options.ScratchRoot` mechanism itself is
untouched and still owned by `internal/closure` (it predates this task at base
HEAD).

### Regressions

All in `internal/install/private_test.go` unless noted. `isolateTempDir`
redirects `TMPDIR`/`TMP`/`TEMP` at a private directory so the complete
temporary-root set of a run is observable; it is called after the fixture has
finished allocating its own temp dirs.

- `TestProjectDryRunKeepsEveryEphemeralPathInOneOperationPrivateRoot` — the
  declared skill is **relocated out of the skills root** and declared by clone
  URL, so resolution must clone and snapshot it. That is precisely the path
  that previously forced a separate scratch workspace. Asserts: mid-run
  `TMPDIR` holds exactly one `curator-install-private-*` root and nothing
  containing `dry-run`; `TMPDIR` is empty afterwards; every persistent path and
  lock is absent; the clone was not persisted into the skills root.
- `TestGlobalDryRunKeepsEveryEphemeralPathInOneOperationPrivateRoot` — the same
  for the machine-wide scope.
- `TestDryRunRemovesItsOperationPrivateRootOnFailure` and
  `TestGlobalDryRunRemovesItsOperationPrivateRootOnFailure` — planning fails on
  a corrupt cache entry; `TMPDIR` is still empty.
- `TestRealRunKeepsEveryEphemeralPathInOneOperationPrivateRoot` — a committed
  staging run keeps its session base inside the same root and leaves `TMPDIR`
  empty.
- `internal/install/stage_test.go`:
  `TestDefaultToolchainProbeRemovesItsProbeRootOnFailure` (renamed from
  `...RemovesItsProbeRoot`), the new
  `TestDefaultToolchainProbeRemovesItsProbeRootOnSuccess`, and
  `TestDefaultToolchainEstablishRemovesItsPrivateRootOnFailure` now run against
  a real `privateRoot` and assert no `go-probe-base-*`/`go-build-base-*`
  survives, on success and on failure. The success case probes a real trusted
  toolchain (probing runs `go version`/`go env` only, so no worker
  re-execution is involved) and skips if none is resolvable.

Discrimination: restoring the two `os.MkdirTemp("", "curator-*dry-run-")` calls
makes both containment tests fail with the reviewer's exact finding —
`temporary roots during the dry run = [curator-dry-run-2191237997]` and
`= [curator-global-dry-run-508338614]`. Verbatim output is in the cycle-4 log.

## R2 — global builds bypassed the MCP and registry gates

### What changed

`Global` now runs the identical pre-build gate set as `Project`, in the same
order relative to build planning:

- **MCP verification** after the collision/system/skill-dependency checks and
  before the migration warnings. `Options.VerifyMcp` still overrides. The
  default is the new shared `mcpVerifier(env, agents, scope)` helper in
  `install.go`, which `Project` now uses as well — one implementation, so the
  two scopes cannot drift. Global's `mcp.Env` is
  `{ProjectRoot: GlobalRoot(home), UserHome: userHome}`: the global scope root
  is its project-level configuration surface, and the `userHome` argument
  `Global` already receives is the user-level one.
- **Registry attestation resolution** after the audit gate and before
  `BuildDeps.resolve`, moved-tag inspection, and `planBuilds`.
  `Options.ResolveAttest` overrides; the default is
  `resolveRegistries(cfg, nodes, "global", !opts.DryRun)`, the same call
  `Project` makes, so revocation, strict-policy unknowns, and tampered
  snapshots fail the global scope identically.
- Global markers changed from `buildMarker(..., nil, nil)` to
  `buildMarker(..., mcpFound[node.Name], attestations[node.Name])`.

Both gates fail closed and sit above the first line that can touch the
toolchain, so a global build can no longer establish a session, inspect the
protected cache, or compile without proven MCP requirements and registry
attestation.

### Regressions

- `TestGlobalMcpFailureBlocksToolchainCacheAndBuild` and
  `TestGlobalRegistryFailureBlocksToolchainCacheAndBuild` — injected failing
  callbacks. Each asserts, through the shared `requireGateBlockedGlobalBuild`
  helper: status `failed` carrying the gate's message, **zero** toolchain
  probes and sessions, **zero** cache inspections, **zero** builder calls,
  **zero** `OnStaged` handoffs, empty `Builds`/`Staged`, and — via
  `liveState.requireUnchanged` — a byte-for-byte unchanged installed project
  tree, runtime store, live build cache, consumer ledger, and global scope.
- `TestGlobalMarkersCarryMcpAndAttestationEvidence` — a successful global
  install records the verified MCP finding and the resolved attestation
  (registry, status, key id) in the global marker.

Discrimination: removing both gates and reverting the marker call makes all
three fail, and the failure output shows the exact defect — a run with a
failing MCP or registry callback reporting `Status:ok` with
`outcome=would-preflight-and-build`, a staged artifact, and
`build-skill tag v1 ... installed`. Verbatim output is in the cycle-4 log.

## Phase order after this cycle

Project: 1–6 manifest/agents/gitignore/devsub/hybrid/locale → 7 closure
(operation-private root opened here) → 8 skill check → 9 collisions → 10
system and skill dependencies → 11 MCP → 12 migration warnings → 13 audit →
14 registry → 15 build boundaries → 16 moved tags → 17 read-only build
planning → 18 dry-run return → 19 private staging + `BuildPlan.Verify` +
`OnStaged` → 20 `scopes.RecordConsumer`, the first persistent write.

Global: identical, with the global manifest, the global scope root, and no
hybrid/devsub phases.

## Verification

Each command run directly as a standalone process, real exit codes:

| Command | Exit |
| --- | --- |
| `gofmt -l internal cmd` | 0 (no output) |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `git diff --check` | 0 |
| `go test -count=1 ./internal/install/` | 0 |
| `go test -count=1 ./internal/godriver/` | 0 |
| `go test -count=1 ./...` | 0 — 36 packages |
| `golangci-lint@v2.1.6 run ./internal/install/...` | 0 — 0 issues |

Expected-red, reported truthfully as failing:
`golangci-lint@v2.1.6 run ./...` exits **1** on 45 pre-existing issues in
`buildcache`, `buildsource`, `gitignore`, `runtimestore`, `scopes`, and
`snapshot`. Zero of them are in `internal/install`, and none of those files
were touched in this cycle.

Full transcript: `TASK-260720-3itlly_verification-cycle-4.log`.

## Files touched this cycle

- added `internal/install/private.go`
- added `internal/install/private_test.go`
- modified `internal/install/builddeps.go`, `internal/install/install.go`,
  `internal/install/global.go`, `internal/install/stage_test.go`

No cross-package product code was modified in this cycle.
