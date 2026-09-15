# TASK-260910-19w2aj: resolve-source-closure-and-explicit-refresh

## Description
Resolve source closure and explicit refresh. Extend current implementation after inspecting existing outcomes; this task is not authorization to start work.

## Scope
internal/closure, closuregraph, closureexec, install, snapshot and Git acquisition integration.

## Acceptance Criteria
Resolve selected local/Git packages and transitive dependencies into a deterministic locked plan. Install/launch use pinned membership and immutable bytes; only explicit resolve/refresh reselects. Missing locked snapshot fails; refresh catches runtime-only and build-only changes. Retain existing dependency conflict and root-only branch rules.
