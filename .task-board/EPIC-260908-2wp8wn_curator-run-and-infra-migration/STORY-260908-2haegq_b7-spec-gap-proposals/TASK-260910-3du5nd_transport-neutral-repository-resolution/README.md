# TASK-260910-3du5nd: transport-neutral-repository-resolution

## Description
Draft a separate Curator specification amendment for repository declarations independent of SSH/HTTPS. Normalize supported URLs or explicit logical identities and resolve them through machine-local endpoint and authentication policy. This task records the user request; normative changes and implementation are not yet accepted.

## Scope
Repository identity, proposed Skillfile sources, permitted Git acquisition lanes, per-machine endpoint and authentication configuration, lock/audit identity, compatibility, bounded fallback and strict external-build transport integration.

## Acceptance Criteria
Examples show the same declaration and locked content acquired through SSH on a workstation and HTTPS in CI. Define logical identity syntax, machine policy, supported endpoint mappings, failure classes, portable lock invariants and secret handling. Preserve the existing closed external-build executor and broker contract. Record open decisions before normative edits.
