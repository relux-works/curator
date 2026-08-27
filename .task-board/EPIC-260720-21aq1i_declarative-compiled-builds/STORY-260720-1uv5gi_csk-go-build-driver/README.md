# csk Go build driver

## Description
Independently implement the accepted schema v6 contract in Python csk from current origin/main, matching protocol behavior and shared vectors while retaining csk-specific manager-home layout and seamless command activation.

## Scope
csk manifest model and skill check, installer lifecycle, runtime/build cache and receipt handling, shims on Unix and Windows, audit and dry-run ordering, status/currentness, documentation and tests. Fast-forward the clean local clone to origin/main before creating a task worktree.

## Acceptance Criteria
csk independently passes shared schema/build vectors; valid Go commands build and launch; fixed Go environment, cache identity, rebuild rules, dry-run purity and rollback match the protocol; unsafe or unsupported declarations fail closed; existing Python test suite and typing/lint gates pass.
