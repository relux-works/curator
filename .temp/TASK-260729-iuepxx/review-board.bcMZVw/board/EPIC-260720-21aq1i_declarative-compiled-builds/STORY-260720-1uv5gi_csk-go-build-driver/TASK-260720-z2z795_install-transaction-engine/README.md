# Implement install transaction engine

## Description
Implement the project-lock and manager-home-lock hierarchy plus durable target journals, recovery, deterministic commit ordering, and reverse rollback as reusable infrastructure.

## Scope
Own src/csk/locking.py, a new src/csk/transactions.py module, and focused locking, recovery, and concurrency tests. Add canonical per-project operation locks held from planning through handoff and a single manager-home mutation lock used for shared recovery, publication, target commit, rollback, and GC. Journal generic mutable targets and generation digests without integrating compiler or installer policy. Dry-run lock routing is integrated by a later task.

## Acceptance Criteria
Project locks are acquired by canonical project identity in unsigned UTF-8 byte order. Optional per-key build locks are released before the home lock. No project or cache lock is acquired while the home lock is held. Journals durably record transaction id, project identity, ordered target classes and identifiers, expected preimages or generations, backups, desired digests, and commit state. Commit sorts target classes and identifiers deterministically, keeps backups until consumer-last durability, and reverse rollback restores only when the current target still equals the journal desired digest. Recovery under the home lock completes or rolls back interrupted work regardless of initiating project. Concurrent success preserves both projects, and one project rollback cannot overwrite another success. Focused pytest and strict mypy pass.
