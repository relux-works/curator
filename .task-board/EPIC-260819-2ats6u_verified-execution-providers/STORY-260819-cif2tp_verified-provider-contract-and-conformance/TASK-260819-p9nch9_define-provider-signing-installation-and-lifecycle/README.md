# TASK-260819-p9nch9: define-provider-signing-installation-and-lifecycle

## Description
Define the cross-platform trust and operations model for separately installed verified providers.

## Scope
Root of trust, artifact signing and notarization where applicable, installer authorization, least privilege, service identity, authenticated IPC bootstrap, updates, rollback prevention, key rotation, revocation, uninstall, audit logging, recovery, and fleet policy. Keep providers outside skill vendoring and binary admission.

## Acceptance Criteria
A reviewed lifecycle specification and reference validation library reject unsigned, revoked, downgraded, wrong-platform, wrong-owner, or incompatible providers; define recoverable installation and update behavior; and preserve portable operation when verified is not required.
