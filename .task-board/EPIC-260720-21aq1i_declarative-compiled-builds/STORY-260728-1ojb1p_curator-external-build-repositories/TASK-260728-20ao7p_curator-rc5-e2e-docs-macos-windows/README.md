# TASK-260728-20ao7p: curator-rc5-e2e-docs-macos-windows

## Description
Re-scoped 2026-09-16 after closing draft PR #17 (superseded: the CI pieces landed on main in their own form, spec at v1.0.0-rc.11, SPEC_PIN promoted). The two pieces main still lacks from that branch: the native lifecycle blackbox test (cmd/curator/native_blackbox_test.go) and the external-build author guide (docs/external-build-repositories.md). Cut fresh from origin/main; do not revive the branch.

## Scope
cmd/curator native blackbox test; docs/external-build-repositories.md; the CI files are out of scope (already landed)

## Acceptance Criteria
Native lifecycle blackbox test runs on the hosted macOS and Windows lanes; the author guide is on main and linked from README; nothing else from PR #17 is carried over
