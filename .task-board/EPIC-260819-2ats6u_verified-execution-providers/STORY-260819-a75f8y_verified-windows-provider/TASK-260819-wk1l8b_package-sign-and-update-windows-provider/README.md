# TASK-260819-wk1l8b: package-sign-and-update-windows-provider

## Description
Package and operate the Windows provider as a separately installed trusted component.

## Scope
Signed service and driver packages, installer authorization, certificate and catalog verification, least privilege, authenticated named-pipe or equivalent IPC, service recovery, upgrades, rollback prevention, key rotation, revocation, uninstall, recovery, audit logs, and compatibility with Secure Boot, VBS, and HVCI where supported.

## Acceptance Criteria
Install and update verify signatures and compatibility before activation; wrong, downgraded, revoked, test-signed, partially installed, or unhealthy providers cannot satisfy verified mode; portable mode remains usable when policy permits; lifecycle tests pass.
