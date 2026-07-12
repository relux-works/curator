# EPIC-260712-c8ac0f: interop-golden-suite

## Description
Phase 10 of docs/implementation-plan.md. The conformance gate: golden fixtures for context file lists, marker JSON (normalized timestamps), content hashes, and signed registry objects with expected verification outcomes; a CI job asserting byte equality; a documented, reviewable regeneration script.

## Scope
See docs/implementation-plan.md, Phase 10. Stories and tasks are created under this epic as work starts.

## Acceptance Criteria
- CI job interop is green
- Fixtures carry no tool branding
- Regeneration script produces stable diffs
