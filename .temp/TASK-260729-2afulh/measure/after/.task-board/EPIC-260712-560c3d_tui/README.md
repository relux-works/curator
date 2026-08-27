# EPIC-260712-560c3d: tui

## Description
Phase 11 of docs/implementation-plan.md. A curator ui terminal view over installed state: projects, skills, activation, attestations. Built on bubbletea and lipgloss, tested closed-loop with tuitestkit (reducer tests, golden screens). Starts only after the interop gate is green.

## Scope
See docs/implementation-plan.md, Phase 11. Stories and tasks are created under this epic as work starts.

## Acceptance Criteria
- Reducer tests cover navigation and state transitions
- Golden screen tests pass on CI
- No core logic lives in the UI layer
