# STORY-260910-25yc0h: records-boundary-binding

## Description
Finding R1/P1 (High, cross-repo): /v1/records pages carry no snapshot boundary and the manager client never replays /v1/log, so a key-holding registry can hide a revocation while serving an honest advancing snapshot. Bind record pages to a verifiable boundary and make the client check it against persisted high-water state.

## Scope
curator-spec protocol/registry.md + internal/registry client

## Acceptance Criteria
Spec defines the boundary-carrying records response (or equivalent inclusion evidence); the client rejects pages evaluated below its persisted high-water; conformance vectors cover the stale-page case
