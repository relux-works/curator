# Add cross-platform manager operation locks

## Description
Provide deterministic cross-process lock primitives for project operations, optional build-key deduplication, and all manager-home mutations on Unix and Windows.

## Scope
Own a new internal/managerlock package with build-tagged POSIX and Windows implementations plus subprocess tests. Derive project lock identities from canonical absolute project identities, cache-key locks from logical keys, and one exclusive manager-home lock from the configured Curator home. Define and enforce lock ordering: canonical project locks in unsigned UTF-8 order, at most one optional key lock released before commit, then the manager-home lock; recovery and GC use only the home lock. Lock files live in manager-created state, not package or project paths. Dry-run creates and acquires no lock. TASK-260720-31nl14 alone owns internal/transaction journals and target swaps.

## Acceptance Criteria
Independent processes cannot concurrently hold the same project or manager-home lock and can hold different project locks while private work proceeds; multi-project acquisition is deterministic and inversion attempts fail before deadlock; cache-key locks never nest with the home lock; cancellation and abnormal child exit release operating-system locks without deleting another owner state; Windows and Unix helpers pass subprocess contention tests; race tests show no data race; a dry-run code path test proves no lock directory or file is created.
