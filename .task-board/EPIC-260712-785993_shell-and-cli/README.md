# EPIC-260712-785993: shell-and-cli

## Description
Phase 9 of docs/implementation-plan.md. Shell hooks for zsh, bash, and powershell with upward env search and PATH restore (Spec 14), and the full CLI surface of Spec 15 with exit code discipline.

## Scope
See docs/implementation-plan.md, Phase 9. Stories and tasks are created under this epic as work starts.

## Acceptance Criteria
- Hook behavior tested for enter, leave, and nested checkouts
- status --check drift produces a non-zero exit
- CLI help enumerates every documented command
