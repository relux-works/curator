# Implement immutable build-source identity

## Description
Introduce a raw-snapshot validator and the curator-build-source-v1 identity used before every build-cache lookup. Keep this identity separate from the existing marker-excluding installed-tree content hash.

## Scope
Own a new internal/buildsource package plus narrowly required snapshot-cache validation in internal/snapshot and tests. Walk without following links, validate portable collision-free protocol paths, accept only directories and regular files, sort unsigned UTF-8 path bytes, and hash every regular file including root .csk-install.json with the exact domain prefix and uint64 big-endian length framing. Freeze one validated snapshot instance through the last build child using an explicit validation token or recheck contract. Repair or rebuild an incomplete or tampered commit-keyed snapshot from the repository instead of trusting its directory presence. Do not modify internal/hashing ContentSHA256 semantics or define cache receipts.

## Acceptance Criteria
The authoritative empty, binary, ordering, marker, mode and timestamp, structural-collision, invalid path, link, special-file, duplicate-path, and mutation vectors pass; two snapshots differing only in root .csk-install.json retain legacy hash compatibility but produce different build-source digests; validation and digesting precede any cache callback in tests; a changed file, tree, or file type invalidates the frozen snapshot before artifact acceptance; corrupt or incomplete snapshot cache hits are not reused and are recreated atomically when the immutable repository source is available; legacy hashing tests remain unchanged.
