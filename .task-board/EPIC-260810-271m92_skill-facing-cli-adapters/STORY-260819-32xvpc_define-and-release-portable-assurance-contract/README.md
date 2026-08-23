# STORY-260819-32xvpc: define-and-release-portable-assurance-contract

## Description
Define the normative portable versus verified assurance boundary in curator-spec, publish the required release candidate, and align Curator configuration, receipts, diagnostics, and conformance consumption.

## Scope
Authoritative curator-spec and Curator integration surfaces only. Preserve published release bytes. Define portable as the default CLI-only mode; reserve explicit verified provider selection for the separate epic; prohibit silent downgrade, assurance claim inflation, and vendored compiled artifacts.

## Acceptance Criteria
The normative spec, schemas, vectors, decisions, release metadata, and operator guidance distinguish portable and verified. Portable receipts enumerate actual capabilities and cannot qualify as verified. Verified selection fails closed when a compatible healthy provider is absent. A reviewed release candidate is published when protocol identity changes, and Curator consumes its exact immutable pin.
