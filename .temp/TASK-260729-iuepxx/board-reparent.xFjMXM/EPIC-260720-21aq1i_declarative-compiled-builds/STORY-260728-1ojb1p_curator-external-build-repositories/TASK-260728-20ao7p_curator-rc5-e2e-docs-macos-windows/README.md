# TASK-260728-20ao7p: curator-rc5-e2e-docs-macos-windows

## Description
Complete Curator rc.5 external-repository user/operator documentation and end-to-end qualification on local/SSH macOS and Windows hosts. Consume released shared vectors and exercise real project/global install, activation, lifecycle, offline, and rollback behavior.

## Scope
Curator docs and examples, shared-suite consumer, black-box and native tests using the available macOS host alias relux and Windows host alias win, release evidence, and exact curator-spec pin. Linux is explicitly excluded and remains a later story.

## Acceptance Criteria
Docs explain repository declarations, descriptor targets, lock/tag policy, access failures, operator substitutions, audit behavior, cache/offline semantics, PATH activation, and signing boundary without exposing unsafe workarounds; shared rc.5 vectors and clean local tests pass; macOS and Windows native evidence covers network and protected-offline install, exact-tag moved/missing failures, project/global activation, cache corruption, repair, crash/rollback, and uninstall; schema-6 regression suites pass; outcome records exact OS/toolchain/spec revisions and does not claim Linux.
