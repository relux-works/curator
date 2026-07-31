# Implement protected POSIX build cache

## Description
Define the csk protected-cache backend interface and implement immutable receipt and artifact storage on POSIX under a dedicated csk manager-home builds namespace.

## Scope
Own src/csk/builds/cache.py, src/csk/builds/cache_posix.py, and focused tests. Use a csk-specific top-level manager-home builds area so source snapshots under cache remain collision-free; protocol vectors address entries only by logical key and artifact-relative path. Verify the cache boundary before parsing candidates, use rooted no-follow access, stage outside the live namespace, and publish atomically. Do not implement Windows ACL logic, installer transactions, status, or GC.

## Acceptance Criteria
The manager creates a private owner-controlled boundary and validates effective-UID ownership, no group or other write access, private directories, regular singly linked receipt and artifact files, link-free containment, derived artifact path, size, hash, canonical receipt bytes, and complete expected input on every lookup and again before publication. An untrusted or corrupt candidate is never adopted even when internally self-consistent; dry-run reports a forced rebuild without mutation; real work builds into new protected state. Publication handles identical concurrent winners by discarding the loser and treats different bytes for one key as corruption or nondeterminism. Entries become immutable in ordinary use while locked GC remains possible. POSIX-focused pytest and strict mypy pass.
