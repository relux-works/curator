# TASK-260924-m28s6b: skillfile-lock-replay-on-fresh-machine

## Description
Implement the accepted skillfile-sources section 3 lock-replay rule (operator addendum 2026-09-24): with an existing Skillfile.lock.json and a missing local snapshot, re-materialize from the declared source (git/repository: fetch exactly the locked revision; path: read current bytes), accept only when package identity and content_sha256 equal the lock, else source_snapshot_changed; source_snapshot_unavailable only when the source cannot be reached; lock stays byte-identical; no tag/branch re-resolution. Skillfile.lock.json is a committed file: Curator must never add it to .gitignore or treat it as machine-private; docs say to commit it.

## Scope
(define task scope)

## Acceptance Criteria
fresh-machine rows per source kind (git tag, repository, path) through install and update with an existing lock; mismatch -> source_snapshot_changed; unreachable -> source_snapshot_unavailable; lock bytes unchanged; no re-resolution proven (moved tag does not change result); gitignore never lists the lock; mutants killed; docs + CHANGELOG
