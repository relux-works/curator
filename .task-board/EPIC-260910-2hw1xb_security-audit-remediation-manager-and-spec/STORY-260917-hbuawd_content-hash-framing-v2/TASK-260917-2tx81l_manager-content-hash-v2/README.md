# TASK-260917-2tx81l: manager-content-hash-v2

## Description
curator: implement the v2 framing in internal/hashing behind a hash version; markers, context locks, verdict/pin state and registry matching (internal/registry/registry.go:298) carry and compare the version; v1 identities are recognised and never equal v2.

## Scope
internal/hashing, internal/marker, internal/contextlock, internal/registry, conformance vectors

## Acceptance Criteria
Spec vectors pass; v1/v2 mismatch tests; migration note in CHANGELOG
