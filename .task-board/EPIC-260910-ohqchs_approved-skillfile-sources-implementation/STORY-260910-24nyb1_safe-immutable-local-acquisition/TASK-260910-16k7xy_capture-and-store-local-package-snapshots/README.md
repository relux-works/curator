# TASK-260910-16k7xy: capture-and-store-local-package-snapshots

## Description
Capture and store local package snapshots. Extend current implementation after inspecting existing outcomes; this task is not authorization to start work.

## Scope
internal/snapshot, hashing, contextstore or appropriate existing protected store.

## Acceptance Criteria
Capture admitted dirty/untracked filesystem bytes even inside Git; hash exact byte inventory and executable flags using all three normative vectors; freeze runtime/build/dependency inputs; fail capture races and missing snapshots; never synthesize a Git commit or use a live directory after capture.
