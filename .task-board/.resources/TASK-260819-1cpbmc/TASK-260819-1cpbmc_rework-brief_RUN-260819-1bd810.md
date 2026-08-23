# Reviewer verdict: changes requested

Run: RUN-260819-1bd810
Route: to-dev

## Blocking findings

### R1 — Production execution still bypasses the selected assurance strategy

cmd/curator/assurance.go performs PreflightAssurance and discards the returned AssuranceBinding. In portable mode PreflightAssurance only returns the static portable capability record; it does not construct NewManagerProcessRunner or NewAssuredExecutor. Repository-wide non-test usage has no caller of NewManagerProcessRunner, and the only external NewAssuredExecutor use is the verified-only compatibility constructor inside internal/closureexec.

Actual Go processes still start through internal/godriver/executor.go and internal/godriver/workerclient.go with direct os/exec paths that do not receive an AssuredOperation, permit, or portable receipt. The added CLI test runs install --dry-run against an empty skill list, so it proves parsing and early verified refusal but starts no real portable build and emits no portable execution evidence.

Required rework: make the production build operation own the assurance selection and carry one AssuredOperation through cache lookup, permit commit, dispatch, receipt validation, and publication. Portable production execution must instantiate and use the manager runner. Add a command/integration test with a real compiled command that proves default portable execution and its honest receipt, plus explicit verified preflight failure with zero cache adoption and zero process starts.

### R2 — Production protected caches remain assurance-agnostic

The new AssuredCacheInput and closureexec.ProtectedStore binding is correct in isolated unit tests, but neither has a production caller. Production install planning and commit continue through internal/buildcache Inspect and Publish, while repository-backed builds use internal/buildrepo.DiskProtectedStore LookupArtifact and StoreArtifact. Those APIs receive the historical cache input only and do not bind assurance mode, policy, provider identity, or fresh capability receipt.

Required rework: integrate the typed assurance binding into every production cache address and persisted entry that can satisfy an execution result, or replace those lookup/publication paths with the assured closure store. Demonstrate a real production portable entry cannot satisfy verified lookup, and cross-provider and capability-receipt drift cannot be adopted.

### R3 — Verified preflight does not authoritatively dominate later cache and process operations

The CLI calls preflight before install/status planning, but the negotiated binding is discarded. If the future resolveCLIProvider seam returns a healthy provider, the CLI can pass verified preflight and then perform assurance-agnostic cache lookup and direct godriver process execution outside that provider. Current nil-provider refusal is fail-closed only because no provider is shipped; it does not preserve a safe future verified seam.

Required rework: return an operation object from the production selection boundary and require that exact object at both production cache and process APIs. Tests must inject a compatible provider and prove its receipt is used for the cache key and dispatch, then inject health, identity, and capability drift and prove zero cache adoption and zero process starts.

## Confirmed improvements

- R5 is fixed: every present provider field rejects wrong types and nulls, with closed shape validation.
- The isolated closureexec store now separates mode, provider, and capability receipt identities.
- The concrete portable runner has real-process negative tests and honest non-lossless capability evidence.
- Independent focused checks passed:
  - go test -count=1 ./internal/closureexec ./internal/config
  - go test -count=1 ./cmd/curator -run TestCLIExecutionAssuranceSelectionIsPortableDefaultAndVerifiedFailClosed
  - go test -race -count=1 ./internal/closureexec
  - git diff --check
- Static production-call search confirms no non-test NewManagerProcessRunner or closureexec.ProtectedStore construction.

This is recoverable implementation work, not a human-only or external blocker.