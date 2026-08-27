# Plan: STORY-260720-3plyvy: Curator Go build driver

Generated: 2026-07-20T06:53:14+04:00
Mode: children
Elements: 18
Phases: 10

## Phase 1 (no dependencies)
- TASK-260720-1zl1cj: add-cross-platform-manager-locks (blocked by: TASK-260720-3ag6pi)
- TASK-260720-256kj1: implement-build-source-identity (blocked by: TASK-260720-3ag6pi)
- TASK-260720-2g0e3b: parse-schema-v6-build-manifests (blocked by: TASK-260720-3ag6pi)

## Phase 2
- TASK-260720-3mrm4z: canonicalize-build-inputs-and-receipts (blocked by: TASK-260720-256kj1)
- TASK-260720-11pfex: activate-build-commands-and-hide-build-roots (blocked by: TASK-260720-2g0e3b)

## Phase 3
- TASK-260720-31nl14: add-durable-install-transaction-journal (blocked by: TASK-260720-1zl1cj, TASK-260720-3mrm4z)
- TASK-260720-3pwg2w: protect-immutable-build-cache (blocked by: TASK-260720-3mrm4z)
- TASK-260720-6i3cya: establish-trusted-go-toolchain-session (blocked by: TASK-260720-3mrm4z)

## Phase 4
- TASK-260720-29hi1h: launch-compiled-commands-through-shims (blocked by: TASK-260720-11pfex, TASK-260720-3pwg2w)
- TASK-260720-4bd0it: implement-install-marker-v2 (blocked by: TASK-260720-2g0e3b, TASK-260720-256kj1, TASK-260720-3mrm4z, TASK-260720-3pwg2w)
- TASK-260720-1zntv0: implement-go-v1-preflight-and-build (blocked by: TASK-260720-2g0e3b, TASK-260720-256kj1, TASK-260720-6i3cya)

## Phase 5
- TASK-260720-3itlly: stage-builds-without-install-mutation (blocked by: TASK-260720-11pfex, TASK-260720-256kj1, TASK-260720-6i3cya, TASK-260720-1zntv0, TASK-260720-3pwg2w, TASK-260720-4bd0it, TASK-260720-29hi1h)

## Phase 6
- TASK-260720-2284br: commit-installations-atomically (blocked by: TASK-260720-31nl14, TASK-260720-4bd0it, TASK-260720-29hi1h, TASK-260720-3itlly)

## Phase 7
- TASK-260720-1ljev5: collect-build-cache-safely (blocked by: TASK-260720-3pwg2w, TASK-260720-31nl14, TASK-260720-4bd0it, TASK-260720-2284br)

## Phase 8
- TASK-260720-1nlmvv: expose-build-currentness-and-repair (blocked by: TASK-260720-3itlly, TASK-260720-2284br, TASK-260720-1ljev5)

## Phase 9
- TASK-260720-2qqq0w: document-curator-compiled-builds (blocked by: TASK-260720-1nlmvv)
- TASK-260720-jrrgw9: verify-rc4-build-conformance (blocked by: TASK-260720-2284br, TASK-260720-1ljev5, TASK-260720-1nlmvv)

## Phase 10
- TASK-260720-1pvfj5: enforce-cross-platform-ci-gates (blocked by: TASK-260720-2qqq0w, TASK-260720-jrrgw9)

## Critical Path
TASK-260720-256kj1 -> TASK-260720-3mrm4z -> TASK-260720-3pwg2w -> TASK-260720-29hi1h -> TASK-260720-3itlly -> TASK-260720-2284br -> TASK-260720-1ljev5 -> TASK-260720-1nlmvv -> TASK-260720-2qqq0w -> TASK-260720-1pvfj5 (10 phases)
