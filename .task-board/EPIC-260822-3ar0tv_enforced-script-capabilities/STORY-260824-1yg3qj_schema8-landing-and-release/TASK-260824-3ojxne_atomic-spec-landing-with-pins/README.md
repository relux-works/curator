# TASK-260824-3ojxne: atomic-spec-landing-with-pins

## Description
curator-spec: land schema 8 on main atomically per the impact-analysis order (steps 6-7). Worktree from origin/main; merge the candidate branch candidate/schema-8-rc.9 (at 6001dc33281b94a4ec7442ab15278550dd0f51d9) into a landing branch; advance the pinned implementation references and coverage assertions in the same PR so no main interval pairs new normative bytes with non-consuming pins — pin curator at its qualified commit (main containing PR 37, run 32689488293) and cocoaskills at its qualified commit (main containing PR 43, run 32756144649); update CHANGELOG.md and COMPATIBILITY.md; keep release/1.0.0-rc.8.json byte-immutable; regenerate vectors twice. Open the PR to main, wait for ALL required checks (Specification x3, Implementations x3, Formatting, Links) verified green pre-merge, squash merge — maintainer pre-authorized 2026-08-22, no human gate. Executor: claude only.

## Scope
(define task scope)

## Acceptance Criteria
Squash merge on main with every lane green; pins point at the two qualified commits; rc.8 untouched; double regeneration proven.
