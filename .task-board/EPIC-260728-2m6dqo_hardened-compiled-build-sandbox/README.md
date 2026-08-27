# EPIC-260728-2m6dqo: hardened-compiled-build-sandbox

## Description
Deliver a separately versioned fail-closed host sandbox for compiled skill builds after the portable manager-owned build profile ships. This follow-up must prove kernel- or hypervisor-enforced isolation instead of treating best-effort controls as equivalent guarantees.

## Scope
Cross-platform hardened execution profile, protocol and conformance amendments, native Curator and csk enforcement, macOS/Windows/Linux validation, receipts and cache separation. It is explicitly not a dependency or release gate for EPIC-260720-21aq1i portable compiled builds or external repository support.

## Acceptance Criteria
A reviewed hardened profile completely denies network access, exposes source and toolchain read-only, permits writes only under a private build root, bounds the full descendant process tree plus memory, disk, time and output, permits only an exact executable allowlist, and fails closed before compilation when any guarantee cannot be established. macOS, Windows and Linux claims are backed by native adversarial tests; unsupported hosts reject the hardened profile. Portable and hardened cache, receipt, marker and claim identities cannot alias.
