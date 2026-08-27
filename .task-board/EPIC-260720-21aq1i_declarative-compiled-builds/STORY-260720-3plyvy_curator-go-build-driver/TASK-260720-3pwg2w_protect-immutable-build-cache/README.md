# Protect and publish immutable build cache entries

## Description
Implement Curator local storage for logical go-v1 keys as immutable manager-protected entries with exact receipt and artifact validation, explicit miss states, and atomic winner publication.

## Scope
Own internal/buildcache filesystem storage and platform-specific protection helpers. Use the Curator implementation layout under manager home cache/build/go-v1/<64-hex-key> with a Curator-local receipt filename and bin/<command>[.exe]; portable tests address only logical keys and relative artifact paths. On POSIX verify effective ownership, private non-writable-by-group-or-other components, no-follow containment, regular singly linked receipt and artifact files, and executable artifact permissions. On Windows verify owner and DACL mutation rights, reject reparse escapes and multiply linked or special files. Expose hit, miss, corrupt, untrusted-provenance, and unsupported outcomes. Publication and quarantine require a caller-held manager-home lock; dry-run is strictly read-only.

## Acceptance Criteria
An exact protected receipt and artifact is a reusable hit; key, input, target, toolchain, artifact path, hash, size, canonical bytes, link safety, owner, mode or DACL, or root-boundary mismatch is never reused; the self-consistent forged-hit vector outside protected state reports would-rebuild-untrusted-cache in dry-run and forces a real rebuild rather than adoption; real publication uses private staging and one atomic directory winner, discards an identical concurrent loser, and errors on different bytes for the same key; corrupt entries are quarantined or replaced only under the home lock; unsupported platform protection disables persistent reuse fail closed; Unix and Windows platform tests plus race tests pass.
