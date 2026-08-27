# Plan: STORY-260720-21bsr2: Compiled-build interoperability

Generated: 2026-07-20T06:53:11+04:00
Mode: children
Elements: 12
Phases: 8

## Phase 1 (no dependencies)
- TASK-260720-2g7avf: shared-compiled-build-interop-cases (blocked by: TASK-260720-cw39jh)
- TASK-260720-p7sdhg: language-driver-admission-matrix (blocked by: TASK-260720-3lo9jc)

## Phase 2
- TASK-260720-14jjgt: compiled-skill-authoring-guide (blocked by: TASK-260720-2g7avf, TASK-260720-3lo9jc)
- TASK-260720-1673lr: curator-shared-build-suite-consumer (blocked by: TASK-260720-2g7avf, TASK-260720-jrrgw9)
- TASK-260720-31zeo2: csk-shared-build-suite-consumer (blocked by: TASK-260720-2g7avf, TASK-260720-12r55p, TASK-260720-3pemm6)

## Phase 3
- TASK-260720-3nj1r6: cross-manager-interop-runner (blocked by: TASK-260720-1673lr, TASK-260720-31zeo2)

## Phase 4
- TASK-260720-3pvihp: qualify-manager-release-evidence (blocked by: TASK-260720-3nj1r6, TASK-260720-14jjgt, TASK-260720-p7sdhg, TASK-260720-3ag6pi, TASK-260720-1pvfj5, TASK-260720-3s27te)

## Phase 5
- TASK-260720-vs6den: promote-spec-implementation-pins (blocked by: TASK-260720-3pvihp)

## Phase 6
- TASK-260720-25d05o: qualify-protocol-release-evidence (blocked by: TASK-260720-vs6den)

## Phase 7
- TASK-260720-1utsx8: audit-csk-released-suite-pin (blocked by: TASK-260720-25d05o, TASK-260720-3s27te)
- TASK-260720-38l1sy: audit-curator-released-suite-pin (blocked by: TASK-260720-25d05o, TASK-260720-1pvfj5)

## Phase 8
- TASK-260720-22ynoi: verify-cross-manager-build-interop (blocked by: TASK-260720-38l1sy, TASK-260720-1utsx8, TASK-260720-vs6den, TASK-260720-14jjgt, TASK-260720-p7sdhg)

## Critical Path
TASK-260720-2g7avf -> TASK-260720-1673lr -> TASK-260720-3nj1r6 -> TASK-260720-3pvihp -> TASK-260720-vs6den -> TASK-260720-25d05o -> TASK-260720-1utsx8 -> TASK-260720-22ynoi (8 phases)
