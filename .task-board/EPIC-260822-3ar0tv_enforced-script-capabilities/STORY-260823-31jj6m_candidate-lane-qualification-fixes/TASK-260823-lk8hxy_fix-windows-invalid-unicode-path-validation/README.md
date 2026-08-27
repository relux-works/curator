# TASK-260823-lk8hxy: fix-windows-invalid-unicode-path-validation

## Description
Windows candidate lane: invalid Unicode paths are accepted where the candidate requires rejection — internal/buildsource TestBuildSourceIdentityVectors/invalid-unicode-build-source-path and internal/godriver TestToolchainIdentityVectors/invalid-unicode-toolchain-path. Diagnose why Windows path handling admits invalid Unicode scalar sequences (likely UTF-16 conversion laundering invalid scalars before validation), fix fail-closed with tests, land via PR green CI. Evidence in .temp/TASK-260823-1l1p8q/windows-evidence-32638424105/. Worktree from origin/main. Fully autonomous per the 2026-08-22 pre-authorization.

## Scope
(define task scope)

## Acceptance Criteria
Both invalid-unicode candidate cases pass on windows; merged to main with green CI.
