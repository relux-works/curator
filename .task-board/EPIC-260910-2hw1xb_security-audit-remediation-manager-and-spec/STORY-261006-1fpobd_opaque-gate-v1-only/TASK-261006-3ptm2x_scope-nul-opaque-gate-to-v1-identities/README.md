# Scope the NUL opaque gate to v1 identities (core.md §8 interim rule)

## Description
curator-spec core.md §8 makes the NUL-byte opaque block an interim rule for v1 readers only; v2 hashes 0x00 as ordinary data. Curator main writes v2 since the rc.4 core (fae2ff9c), but internal/audit gate(), auditSubject/detect and internal/contextaudit still block NUL-bearing snapshots unconditionally. That blocks real installs, for example a vendored zoneinfo.zip in the spm tree on ivmbp. Make the block apply exactly when a v1 identity is computed or trusted (legacy markers, v1 readers, frozen v1 shapes) and never when a v2 identity is computed or verified. Keep every v1 protection and the verdict-cache exclusion for v1.

## Scope
internal/audit, internal/contextaudit, internal/opaquescan call sites, install/status paths that decide the hash version; tests; CHANGELOG entry under Unreleased

## Acceptance Criteria
A NUL-bearing skill installs and audits under v2 with a v2 identity; the same tree under a v1 reader or legacy v1 marker still gets the blocking opaque finding; there is no path where a v1 identity is computed over NUL bytes; the verdict cache never stores a v1 result for a NUL tree; production-entry tests for both versions; hosted gate green.
