# TASK-260910-24cuys: parse-opt-in-skillfile-v2-sources

## Description
Parse opt in skillfile v2 sources. Extend current implementation after inspecting existing outcomes; this task is not authorization to start work.

## Scope
internal/manifest, skillspec, identity, protocoljson; capability admission and source union.

## Acceptance Criteria
Accept relative/absolute path, git URL and explicit repository sources with valid refs; preserve every v1 field including legacy source meaning; reject mixed arms, unsupported versions and malformed directories before I/O. Add schema-backed positive/negative and v1 regressions.
