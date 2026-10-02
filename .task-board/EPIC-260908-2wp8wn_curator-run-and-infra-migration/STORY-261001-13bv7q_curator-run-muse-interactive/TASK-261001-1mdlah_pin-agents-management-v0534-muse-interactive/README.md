# TASK-261001-1mdlah: pin-agents-management-v0534-muse-interactive

## Description
Bump agents-management to v0.5.34 in curator-agent-launcher, flip the Muse interactive bound to admission, prove the root-session plan with fake binaries.

## Scope
(define task scope)

## Acceptance Criteria
agents-management v0.5.34 pinned; curator run muse builds an interactive plan with XDG-only env, native=no flag, yolo=one --yolo, unlisted release refused; goldens updated incl. prompt-suggestion env; go test + windows vet evidence.
