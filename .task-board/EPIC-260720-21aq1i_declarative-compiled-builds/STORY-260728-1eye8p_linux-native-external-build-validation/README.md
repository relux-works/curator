# STORY-260728-1eye8p: linux-native-compiled-build-validation

## Description
Run a separate later native Linux qualification for compiled local and external build flows after the corresponding protocol and manager surfaces are accepted and a dedicated Linux host is available.

## Scope
Native Linux toolchain preflight, installation and lifecycle evidence for Curator and csk across local and external sources using a clean non-root account, pinned toolchains, project/global activation, PATH shims, permissions, protected caches, offline reinstall, repair, rollback and adversarial source cases. This story does not block macOS/Windows rc.5 claims and Linux remains absent from claim v3 until acceptance.

## Acceptance Criteria
The provisioned Linux host and toolchain are recorded; both managers pass the accepted rc.5 shared and native suites under non-root operation; root-owned/system installation is tested only where explicitly required; evidence covers access failure, exact tag, audit-before-build, cache corruption, repair, rollback, and PATH activation; only after independent review may Linux be added to a conformance claim.
