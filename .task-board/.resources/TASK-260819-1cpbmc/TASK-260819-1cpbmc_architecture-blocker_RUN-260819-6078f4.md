# TASK-260819-1cpbmc architecture blocker

## Constraint

The accepted assurance operation cannot authoritatively dominate the current production Go compiler dispatch without changing one of two explicitly preserved contracts.

- `closureexec.EnforceObserveProvider` is a one-shot `EnforceAndObserve(context.Context, ExecutionRequest) (Audit, error)` boundary. Its request contains one closed `DerivationPermit` and immutable replay inputs.
- `DerivationPermit` is explicitly a pre-C5 derivation permit. Its invocation kind is closed to manifest, vendor, mirror, and metadata; its read roots must exactly equal admitted replay mounts.
- Production Go compilation uses an interactive authenticated manager/worker protocol. The manager starts the worker, opens stdin/stdout pipes, completes identity and nonce handshakes, sends list/build permits, receives results, and performs shutdown. The worker and compiler also read the trusted Go toolchain.
- The task boundary says to preserve the provider interface and accepted closure substrate. The latest reviewer requires the same `AssuredOperation` to dominate production cache lookup and actual compiler dispatch, with portable production execution using `NewManagerProcessRunner`.

## Evidence

- `internal/closureexec/executor.go:17-23` defines the one-shot provider interface.
- `internal/closureexec/executor.go:53-57` binds that interface to `ExecutionRequest{Permit DerivationPermit, Inputs []ReplayInput}`.
- `internal/closureexec/models.go:194-228` defines the closed pre-C5 permit.
- `internal/closureexec/models.go:263-280` requires every read root to be exactly an admitted replay mount.
- `internal/closureexec/portable_runner.go:45-129` starts a non-interactive command and owns its stdout/stderr; it has no stdin/session transport.
- `internal/godriver/workerclient.go:108-180` starts the production worker, opens stdin/stdout pipes, and performs the authenticated session exchange.
- `internal/buildcache/cache.go:43-46` and `internal/buildrepo/pipeline.go:104-135` show the production cache/request types currently have no assurance operation.

No product code was changed in this run. The dirty worktree and accepted prior task changes were preserved.

## Failed assumptions and rejected forced fits

1. Passing only `AssuranceBinding` into production caches would close cache aliasing but would still let compiler dispatch bypass the exact preflight object and provider.
2. Wrapping `godriver.Build` in an authorization callback would not make `ManagerProcessRunner` or the verified provider start/enforce the actual process tree.
3. Inserting a second one-shot Curator helper under `ManagerProcessRunner` would require a new build permit/receipt kind, a source/toolchain replay model, and either copying the full GOROOT into admitted intake or weakening the exact-read-root contract. That is a new execution architecture, not ordinary wiring.
4. Extending `EnforceObserveProvider` with interactive session transport would solve the dispatch mismatch but violates the explicit instruction to preserve the provider interface and needs a normative provider-contract revision.

## Viable options

### Option A — define a build-specific assured session contract (recommended)

Add a platform-neutral interactive provider/session API and build-specific permit/receipt schemas that compose with the existing `godriver` handshake. Bind the resulting operation evidence into both production cache families. This preserves the mature manager/worker controls and avoids copying the toolchain, but requires an approved provider-contract/spec revision before implementation.

### Option B — define an outer one-shot assured build helper

Add a separate build execution request/permit/receipt model, run a hidden Curator helper through the existing one-shot provider/runner, and have that helper invoke `godriver`. This preserves the current provider method but adds another process layer and requires an explicit decision for trusted toolchain mapping/read evidence; copying GOROOT per operation is costly, while ambient reads conflict with the current exact replay-root rule.

### Option C — narrow production integration to preflight and cache identity

Carry the preflight binding into `buildcache` and `buildrepo` but leave `godriver` dispatch on its current path. This is the smallest change, but it does not satisfy reviewer R1/R3 and does not provide authoritative verified dispatch.

## Required decision

Architecture/spec ownership must choose Option A or Option B and approve the corresponding contract change. Recommendation: Option A, because it composes with the existing authenticated production worker instead of adding a second worker and toolchain-copy boundary. Once that decision is recorded, implementation can proceed without fabricating assurance evidence or weakening the accepted closure model.
