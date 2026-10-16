# SNAP-ID — Snapshot metadata and actual-tree verification

Owner repository: **curator**. Size: **3**. Required in the rc.7 materialize cut.

Normative clauses: `protocol/managed-home-snapshots.md` §§2 (prior snapshot), 4 (all metadata/identity rules), 6 (descriptor boundary), 7 (conformance bounds); CIP-0014 implementation leaf 2. Spec baseline: `5509499`.

Attack surface: confusing a canonical manifest, an observed live profile, and the bytes actually present in a snapshot.

Acceptance criteria:

- Implement closed manifest/descriptor codecs and full actual-tree verification. CCJ-1 plus exactly one LF frames both metadata files; source fragment hash excludes presentation LF, descriptor and manifest hashes include their LF.
- Verify the complete inventory, bytes, sizes, semantic modes, path/component collisions, channel references, adapter/profile equality, unknown versions/members and derived directories. Reject symlinks, hard-link aliases and special files in exported content; metadata is not sufficient evidence.
- Keep exact plan identity and portable comparison distinct. Portable comparison excludes only `fragment_sha256`; producer version, profile pin, descriptor and every other field remain binding. Never replace a persisted baseline after mismatch.
- Reuse the fixed separate schema/corpus; do not extend released fragments, introduce archive encoding, or treat `manifest-only` as verification of an existing directory. Keep verifier logic reusable by destination idempotence and the later published hosted verification API.

Mandatory tests with positive controls:

| Positive control | Negative case / required invariant |
| --- | --- |
| Canonical corpus tree verifies; exact metadata bytes match pinned vectors | Duplicate/unknown member, wrong LF/hash domain, unsorted or colliding path, wrong channel target, adapter or profile mismatch refuses |
| Byte-identical independent copy with correct modes verifies | Missing/extra entry, changed executable mode, symlink/hard-link alias, unreadable member or forged unchanged manifest refuses |
| Two source homes with identical portable output compare portably equal | Exact manifest hash differs with source fragment; changed producer version, descriptor, profile or file cannot compare equal |
| Original baseline verifies unchanged tree | Mutated tree remains refused after attempted recapture; no baseline rewrite |

Bounded checker controls: ignore execute mode/extra entries (CIP-0014 M5), or trust the saved manifest (M6), and demonstrate the intended verifier test fails. These are planned controls, not executed evidence.

Dependencies and paths: starts from accepted spec; parallel with SNAP-LOCK and SNAP-PUBLISH in separate new files. Proposed package `internal/envsnapshot/` (new, distinct from source-package `internal/snapshot`); pin corpus under `internal/crossconformance/testdata/`. Coordinate the package interface before other leaves consume it. Public CLI/receipt spelling belongs to HOST-HOME-SPEC, not this helper leaf.
