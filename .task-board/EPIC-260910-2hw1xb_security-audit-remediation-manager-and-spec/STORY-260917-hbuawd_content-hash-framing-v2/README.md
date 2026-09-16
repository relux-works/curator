# STORY-260917-hbuawd: content-hash-framing-v2

## Description
Finding K1 of the cocoaskills audit (High, specification-level; curator-spec issue #59): core.md section 8 frames the content hash as UTF8(path) || 0x00 || file_bytes with records joined by 0x00 and no length prefix or domain separation. File bytes may contain 0x00, so two different trees hash identically (reproduced: a single-file SKILL.md whose bytes encode path/0x00/content collides with the two-file tree it encodes). The hash keys every trust decision: verdict caches and pins, revocation and registry record matching on the hash alone (registry.md:103-105; curator internal/registry/registry.go:298), install markers and locks. The Go manager frames identically (internal/hashing/hashing.go:26). The spec section 8.1 build-source identity already shows the correct shape (domain prefix, record tag, 8-byte lengths).

## Scope
curator-spec protocol/core.md section 8 and registry.md matching rule, conformance vectors; curator internal/hashing, install markers, context locks, registry matching; curator-skill-registry record schema (hash version)

## Acceptance Criteria
core.md section 8 revised to a length-framed, domain-separated, versioned content hash with the two colliding trees as vectors that MUST differ and the empty tree as a vector; registry.md requires the framing version to match; curator implements v2 behind a hash version carried by markers, locks and registry matching, v1 identities never equal v2; the interim rule that a regular file containing 0x00 is a blocking opaque finding regardless of directory ships first; the registry service carries the hash version on records; cocoaskills tracks its half as STORY-260917-1o3esy on its own board
