# TASK-260728-1itx7a: verify-hardened-build-sandbox-conformance

## Description
Run independent black-box and adversarial cross-manager qualification for every claimed hardened host profile.

## Scope
Shared suite on Linux, ssh relux and ssh win; escape attempts for network, filesystem, executable graph and descendants; memory/disk/time/output exhaustion; cache/receipt alias rejection; documentation and claim audit.

## Acceptance Criteria
Curator and csk produce matching stable outcomes for every vector, every claimed profile passes all adversarial gates, unsupported profiles reject before compiler execution, portable artifacts never satisfy hardened currentness, and independent review accepts the evidence.
