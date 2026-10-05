# STORY-261004-7fglii: shell-hook-path-and-backend-env-hardening

## Description
DESIGN PENDING — NOT ACCEPTED FOR EXECUTION. Operator decision 2026-10-04: decide the design first, then prioritise. Do not spawn producers and do not schedule until the operator accepts a design. Fast-track security asks relayed by the cocoaskills orchestrator (a2a #curator topics shell-hook-k3-path-order and launch-project-roots-idea, 2026-10-03). Verified on curator main: internal/envfiles/envfiles.go:36 PREPENDS project .agents/bin to PATH; internal/shell/shell.go:58 ships DefaultTrustProfile=A-warning, so an unapproved or changed .agents/env.sh is warned about and still sourced on cd.

## Scope
(define story scope)

## Acceptance Criteria
(define acceptance criteria)
