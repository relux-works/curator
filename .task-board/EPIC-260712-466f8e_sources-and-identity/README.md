# EPIC-260712-466f8e: sources-and-identity

## Description
Phase 2 of docs/implementation-plan.md. Canonical source identity and segment-aware allowlist (Spec 8.2), git operations through system git with the transport allowlist, ref resolution semantics, archive extraction with path-escape and link rejection, submodule detection, commit-keyed snapshot cache, and the content hash byte layout (Spec 8.5).

## Scope
See docs/implementation-plan.md, Phase 2. Stories and tasks are created under this epic as work starts.

## Acceptance Criteria
- Identity normalization matches Spec 8.2 for ssh, https, scp-style, and local forms
- Content hash equals golden fixture hashes byte for byte
- Suspicious URL and transport cases refused
