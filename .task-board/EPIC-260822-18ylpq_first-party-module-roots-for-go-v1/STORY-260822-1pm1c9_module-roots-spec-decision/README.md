# STORY-260822-1pm1c9: module-roots-spec-decision

## Description
curator-spec: decision + normative prose for first-party module roots. Surface: build command gains a declared modules list; manager validates a bijection between declared module dirs and go.mod replace directives (directory-form only, strictly below the snapshot, link-free — snapshot validation already guarantees no links —, disjoint from build and runtime roots, no versions, no module-to-module redirects); declared dirs join the directive/cgo/assembly scan surface; external deps stay vendor-only versioned. Schema bump (coordinate with script-worker-v1 schema 8 — same bump or next). Conformance vectors: acceptance, escape attempts, redirect replaces, undeclared replace, unused declaration, nested modules, runtime-root overlap, Windows path collisions.

## Scope
(define story scope)

## Acceptance Criteria
Decision accepted; prose+schema+vectors merged with clean double regeneration; CHANGELOG/COMPATIBILITY updated.
