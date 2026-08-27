# Implement build-source identity and context isolation

## Description
Implement the exact curator-build-source-v1 raw-snapshot identity and make every declared build root statically invisible to prompt context and runtime copying on real installs, cache hits, and dry-runs.

## Scope
Own new source-boundary code under src/csk/builds, additive build-source support in src/csk/hashing.py without changing legacy content_sha256, src/csk/whitelist.py, and focused hashing, whitelist, and skill-check tests. Validate all snapshot descendants without following links, hash every regular file including a package-provided root .csk-install.json, and expose a frozen-snapshot identity used by later tasks. Do not run Go or implement receipts, cache publication, or installer commit logic.

## Acceptance Criteria
The digest uses the curator-build-source-v1 NUL domain prefix, unsigned UTF-8 path ordering, F records, and uint64 big-endian path and content lengths exactly as the shared vectors specify. Links, special files, invalid or duplicate portable paths, platform collisions, and mutation before the last build child exits fail closed. Root .csk-install.json changes alter build-source identity while legacy installed-tree content_sha256 remains unchanged. Nested build roots under eligible context roots are fully excluded while SKILL.md and unrelated eligible assets remain visible, and build roots are never runtime-copied. Focused vectors, pytest, and strict mypy pass.
