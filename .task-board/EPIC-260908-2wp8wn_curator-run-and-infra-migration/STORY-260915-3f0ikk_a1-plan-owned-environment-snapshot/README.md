# STORY-260915-3f0ikk: a1-plan-owned-environment-snapshot

## Description
Upstream prerequisite for the launcher production wiring (TASK-260908-1o7i8y stop-the-line): skill-agents-management must expose the owned-environment snapshot (System.ChildEnv(nil, effectiveRequest)) of an admitted plan so the launcher composer can satisfy SPEC 4.5 without reconstructing the private request. Landed through the agents-management repository canon and released as a signed tag consumed by the launcher.

## Scope
(define story scope)

## Acceptance Criteria
Snapshot exposed with the admitted plan for all launch modes and the three native systems; existing goldens unchanged; tests and narrowing mutant; CHANGELOG entry; landed on main and tagged v0.5.13 (signed) by the orchestrator; launcher go.mod bumped in a follow-up.
