# Plan: STORY-260720-1uv5gi: csk Go build driver

Generated: 2026-07-20T06:17:07+04:00
Mode: children
Elements: 17
Phases: 12

## Phase 1 (no dependencies)
- TASK-260720-z2z795: install-transaction-engine
- TASK-260720-z9j4c9: schema-v6-build-model

## Phase 2
- TASK-260720-3c0ss2: build-source-context-boundary (blocked by: TASK-260720-z9j4c9)
- TASK-260720-3j8pp5: go-toolchain-identity (blocked by: TASK-260720-z9j4c9)

## Phase 3
- TASK-260720-2dnqw2: canonical-build-metadata (blocked by: TASK-260720-3c0ss2, TASK-260720-3j8pp5)
- TASK-260720-2g21eg: go-v1-compile-driver (blocked by: TASK-260720-3c0ss2, TASK-260720-3j8pp5)

## Phase 4
- TASK-260720-2jfnz6: protected-build-cache-posix (blocked by: TASK-260720-2dnqw2)

## Phase 5
- TASK-260720-8nxlgx: protected-build-cache-windows (blocked by: TASK-260720-2jfnz6)

## Phase 6
- TASK-260720-11yhth: command-runtime-activation (blocked by: TASK-260720-8nxlgx)
- TASK-260720-2x6mjn: side-effect-free-build-planner (blocked by: TASK-260720-2g21eg, TASK-260720-8nxlgx, TASK-260720-z2z795)

## Phase 7
- TASK-260720-3t8nr3: transactional-project-hybrid-install (blocked by: TASK-260720-11yhth, TASK-260720-2x6mjn)

## Phase 8
- TASK-260720-g7kgox: transactional-global-install (blocked by: TASK-260720-3t8nr3)

## Phase 9
- TASK-260720-th0jdi: build-currentness-repair-gc (blocked by: TASK-260720-g7kgox)

## Phase 10
- TASK-260720-12r55p: shared-v6-vector-consumer (blocked by: TASK-260720-th0jdi, TASK-260720-3ag6pi)
- TASK-260720-akf5kh: schema-v6-user-docs (blocked by: TASK-260720-th0jdi, TASK-260720-3lo9jc)

## Phase 11
- TASK-260720-3pemm6: cross-platform-go-build-e2e (blocked by: TASK-260720-12r55p)

## Phase 12
- TASK-260720-3s27te: integrated-csk-v6-verification (blocked by: TASK-260720-3pemm6, TASK-260720-akf5kh)

## Critical Path
TASK-260720-z9j4c9 -> TASK-260720-3c0ss2 -> TASK-260720-2dnqw2 -> TASK-260720-2jfnz6 -> TASK-260720-8nxlgx -> TASK-260720-11yhth -> TASK-260720-3t8nr3 -> TASK-260720-g7kgox -> TASK-260720-th0jdi -> TASK-260720-12r55p -> TASK-260720-3pemm6 -> TASK-260720-3s27te (12 phases)
