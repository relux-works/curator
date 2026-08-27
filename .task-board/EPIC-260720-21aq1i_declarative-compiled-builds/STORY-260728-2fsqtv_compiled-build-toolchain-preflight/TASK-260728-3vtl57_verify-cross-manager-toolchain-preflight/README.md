# TASK-260728-3vtl57: verify-cross-manager-toolchain-preflight

## Description
Verify toolchain preflight and guidance parity across Curator and csk on macOS and Windows, with Linux qualification when the host is available.

## Scope
Present, absent, too old, too new, prerelease, untrusted path, metadata mismatch, platform mismatch and changed-after-probe fixtures for Go, Rust, Swift and selected Kotlin toolchains; diagnostic and guidance snapshots; cache identity and no-mutation assertions.

## Acceptance Criteria
Both managers return matching stable codes and semantic fields, fail before forbidden work, accept only trusted compatible fingerprints, and emit accurate platform guidance; native evidence is attached per supported tuple with no fabricated claims.
