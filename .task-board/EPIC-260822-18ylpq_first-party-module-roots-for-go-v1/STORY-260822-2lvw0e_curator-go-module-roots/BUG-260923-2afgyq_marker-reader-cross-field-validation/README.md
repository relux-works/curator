# BUG-260923-2afgyq: marker-reader-cross-field-validation

## Description
Implement the MARKER V3/V4 external repository cross-field checks missed by validV3Build. The five install-marker-v4 invalid cases currently admitted by marker.Read are declared/effective identity mismatch, local and network identity-kind mismatch, and SHA-1/SHA-256 effective revision-width mismatch. This is the implementation owner for the corresponding conformance gap ledger rows.

## Scope
internal/marker reader validation and conformance tests

## Acceptance Criteria
marker.Read rejects the five named invalid install-marker-v4 cases at the production reader; valid fixtures still accepted; their gap-ledger rows removed with exact counts; per-check mutants.
