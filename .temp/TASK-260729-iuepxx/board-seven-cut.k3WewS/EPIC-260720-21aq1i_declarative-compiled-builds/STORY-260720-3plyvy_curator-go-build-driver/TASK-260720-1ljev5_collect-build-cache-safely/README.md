# Collect compiled artifacts safely

## Description
Extend maintenance garbage collection to immutable build cache entries while honoring marker v2 references, in-flight transactions, protected-state boundaries, grace periods, and the manager-home lock.

## Scope
Own internal/scopes GC extensions and focused manager-state or buildcache helpers. Under the manager-home lock, mark logical build keys referenced by every valid marker v2 in live project, global, and hybrid scopes plus every recoverable transaction journal. Sweep only unreferenced protected entries older than the existing grace policy and retain transaction-owned or uncertain candidates conservatively. Revalidate the cache boundary before traversal, never execute or adopt cache content, and prune consumers only inside the same serialized maintenance transaction. Keep runtime and snapshot GC behavior compatible. Do not implement status presentation or rebuilds here.

## Acceptance Criteria
Referenced build entries and all keys named by incomplete journals survive; unreferenced protected entries younger than grace survive and older entries are removed atomically; invalid markers, corrupt receipts, untrusted roots, reparse or symlink escapes, and ownership or DACL failures never cause unsafe traversal or deletion and produce actionable maintenance warnings; concurrent install, rollback, recovery, and standalone curator gc serialize on the home lock with no lost consumer update; deleting one cache entry cannot escape the Curator cache root; runtime GC regressions remain green and Unix, Windows, fault, and race tests pass.
