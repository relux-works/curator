# TASK-260819-ywsg92: package-sign-and-update-linux-provider

## Description
Package and operate the Linux provider as a separately installed trusted host component.

## Scope
Signed packages and repositories for the declared distributions, service manager integration, least-privilege split, Secure Boot or kernel-component signing where needed, authenticated Unix-domain IPC, policy ownership, upgrades, rollback prevention, key rotation, revocation, uninstall, recovery, and audit logs.

## Acceptance Criteria
Install and update verify the full signing chain and compatibility before activation; wrong, downgraded, revoked, partially installed, or unhealthy providers cannot satisfy verified mode; portable mode remains usable when policy permits; lifecycle tests pass.
