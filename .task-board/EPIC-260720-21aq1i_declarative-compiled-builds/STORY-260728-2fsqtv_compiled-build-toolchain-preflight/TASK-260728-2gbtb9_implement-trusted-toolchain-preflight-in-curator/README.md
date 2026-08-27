# TASK-260728-2gbtb9: implement-trusted-toolchain-preflight-in-curator

## Description
Implement common fail-fast toolchain requirement resolution and guidance diagnostics in Go Curator for local and external compiled commands.

## Scope
Manifest planning, trusted locator registry, executable/version/fingerprint probes, metadata cross-checks, no-PATH fallback unless operator policy allows it, cache input integration, dry-run/status behavior, platform guidance rendering and macOS/Windows tests.

## Acceptance Criteria
Curator rejects missing, incompatible, untrusted or contradictory toolchains before source/compiler/persistent mutation; valid toolchains bind exact fingerprints into build identity; diagnostics contain stable codes and manager-owned guidance without executing installers; existing schema-6 Go behavior remains compatible.
