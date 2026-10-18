# BUG-261018-bisu6n: pnpm-conformance-test-leaves-readonly-tmp

## Description
TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization (internal/pnpmsource/conformance_test.go:1127) leaves read-only pnpm store files under the runner tmp (capture-store/trees, store/v10/index). On the self-hosted runner relux-rise1-curator-4 the next job cleanup fails with Permission denied, so every later job on that runner fails (PR #140 run 39119857144, twice). Fix: the test restores write permission on everything it created (t.Cleanup chmod -R u+w before removal), plus a test that the tmp tree is removable after the run.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
The pnpm conformance test and swept helpers leave no read-only tree; Linux self-hosted CI jobs end with a guarded always() step that makes their own tmp root writable and never deletes; regression witness, script refusal tests and plant control pass; full local suite and hosted gate green with runners recorded.
