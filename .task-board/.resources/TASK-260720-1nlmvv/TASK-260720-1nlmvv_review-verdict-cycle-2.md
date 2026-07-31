# TASK-260720-1nlmvv review verdict — cycle 2

## Verdict: CHANGES REQUESTED

Route: to-dev. The cycle-2 candidate is mechanically green but still fails the explicit repair atomicity and diagnostic contracts. No producer code was modified during review.

## Blocking findings

1. P1 — corrupt and untrusted cache repair is not rollback-safe after publication. runCommit publishes build winners before stageTargets, journal planning, preparation, or commit (internal/install/commit.go:537-585). buildcache.Publish immediately quarantines a corrupt or untrusted live entry and selects the replacement (internal/buildcache/publish.go:148-174). The cache is not a transaction target and existing rollback snapshots omit it. Therefore any later target-stage, journal-prepare, or target-commit fault restores the installed scope but leaves the old referenced cache entry quarantined and the replacement live. The new E2E covers only a toolchain refusal before publication, not injected failures after publication. This violates the cycle-2 requirement to preserve the live referenced cache until atomic commit and roll back both cache and install on every injected failure.

2. P1 — build-related Result.Errors still bypass bounded path redaction. Planning and staging use failBuild, but commit/publication errors use raw failf in Project and Global (internal/install/install.go:565-571; internal/install/global.go:328-334). buildcache publication errors can include operation-private artifact or cache paths, so these errors can leak absolute paths and unbounded detail. This violates the explicit cycle-2 requirement to apply the same redaction before every Result.Errors surface.

3. P2 — missing-toolchain rows omit identities already known. toolchainInventory populates skill, command, build root, source directory, outcome, reason, and diagnostic, but leaves driver and build-source identity empty (internal/install/plan.go:407-426), even though source validation completed and the closed driver is go-v1 (internal/install/plan.go:352-359). Human output consequently prints driver= and JSON carries an empty build_source object. This does not satisfy the per-active-command driver and build-source digest contract. Existing tests assert only command/source_dir and omit these fields.

4. P1 — concurrent cache changes are not covered by build-state-changed. status fingerprints only install markers before and after the dry-run plan (cmd/curator/builds.go:568-597). Cache receipt/artifact evidence is inspected once during the unlocked plan and is not revalidated before the final status --check verdict. A same-marker cache removal, corruption, or replacement between inspection and output can therefore publish a stale current verdict or stale drift verdict without build-state-changed. This leaves the concurrent-state acceptance criterion and the requested concurrent GC/currentness repair coverage open.

## Required rework

Make corrupt/untrusted cache replacement transaction-owned or otherwise reversibly staged through every post-publication failure; add injected failures after publication for both install and upgrade and assert exact cache/install restoration. Route every build-phase and build-publication Result.Errors path through the single bounded redactor. Populate toolchain-refusal rows with every identity known before the toolchain probe, especially go-v1 driver and validated build-source digest. Revalidate or bracket cache evidence through the final status verdict and add a real concurrent cache/currentness regression.

## Independent validation

PASS: focused internal/install tests including repair, refusal, redaction, and publication preservation (5.174s); focused real cmd/curator integrations for compiled currentness, corrupt repair, transitive status, marker refusal, and missing toolchain (354.202s); full diagnostic-focused install group (2.632s); go vet ./cmd/curator ./internal/install ./internal/godriver; task-file gofmt; git diff --check 17804ce. The green gates confirm semantic rework rather than a mechanical test failure.