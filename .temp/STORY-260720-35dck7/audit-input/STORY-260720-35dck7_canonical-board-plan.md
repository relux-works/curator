# Plan: STORY-260720-35dck7: Protocol schema v6

Generated: 2026-07-20T05:53:00+04:00
Mode: children
Elements: 12
Phases: 10

## Phase 1 (no dependencies)
- TASK-260720-1nvomm: protocol-v6-core-security-contract

## Phase 2
- TASK-260720-17llva: go-v1-manager-profile (blocked by: TASK-260720-1nvomm)
- TASK-260720-wajgn8: manifest-v6-schemas (blocked by: TASK-260720-1nvomm)

## Phase 3
- TASK-260720-12iigs: compiled-artifact-schemas (blocked by: TASK-260720-17llva, TASK-260720-wajgn8)

## Phase 4
- TASK-260720-2zc6k1: conformance-claim-v2-schema (blocked by: TASK-260720-12iigs)

## Phase 5
- TASK-260720-37ei85: legacy-schema-compatibility-guards (blocked by: TASK-260720-2zc6k1)

## Phase 6
- TASK-260720-1s1vr6: build-driver-conformance-vectors (blocked by: TASK-260720-37ei85)

## Phase 7
- TASK-260720-cw39jh: manager-lifecycle-build-vectors (blocked by: TASK-260720-1s1vr6)

## Phase 8
- TASK-260720-1u7hes: validation-and-release-gates (blocked by: TASK-260720-cw39jh)
- TASK-260720-3lo9jc: schema-v6-authoring-and-cli-docs (blocked by: TASK-260720-cw39jh)

## Phase 9
- TASK-260720-q5oy3o: protocol-rc4-release-metadata (blocked by: TASK-260720-1u7hes, TASK-260720-3lo9jc)

## Phase 10
- TASK-260720-3ag6pi: protocol-v6-conformance-verification (blocked by: TASK-260720-q5oy3o)

## Critical Path
TASK-260720-1nvomm -> TASK-260720-17llva -> TASK-260720-12iigs -> TASK-260720-2zc6k1 -> TASK-260720-37ei85 -> TASK-260720-1s1vr6 -> TASK-260720-cw39jh -> TASK-260720-1u7hes -> TASK-260720-q5oy3o -> TASK-260720-3ag6pi (10 phases)
