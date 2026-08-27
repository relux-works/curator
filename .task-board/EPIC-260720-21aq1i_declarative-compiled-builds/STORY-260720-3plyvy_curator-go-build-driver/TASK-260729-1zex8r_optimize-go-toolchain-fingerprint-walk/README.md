# TASK-260729-1zex8r: optimize-go-toolchain-fingerprint-walk

## Description
Optimize the Go driver toolchain fingerprint traversal that dominates compiled-build status/install runtime. Preserve the exact canonical record set, ordering, byte stream and SHA-256 identity while replacing repeated os.Root component resolution with directory-scoped traversal. Treat internal/godriver/fingerprint.go as a trust-boundary change: add equivalence and adversarial regressions, re-run go-v1 conformance/vector evidence, and provide measured before/after performance. This task may consume the digest-identical prototype and evidence from TASK-260729-2kaopg but must use its own worktree. No timeout increase, assertion weakening, cache clearing, host installation, staging, commit, publication or pin changes.

## Scope
internal/godriver fingerprint traversal and directly related tests/conformance evidence only; no global-status behavior changes

## Acceptance Criteria
1. Old and new traversal produce byte-for-byte identical canonical fingerprint records and digest on representative toolchains, symlinks, errors and mutation/fail-closed cases. 2. No trust, path-containment, race, cancellation or error-redaction guarantee regresses. 3. Measured fingerprint latency materially improves and supports cmd/curator default package completion below 480 seconds on a quiet clean -count=1 run. 4. Focused godriver tests, accepted go-v1 conformance/vector gates, build, vet, gofmt and diff checks pass. 5. TASK-260729-2kaopg can then run two consecutive foreground default-timeout go test -count=1 ./... gates without exclusions or timeout overrides.
