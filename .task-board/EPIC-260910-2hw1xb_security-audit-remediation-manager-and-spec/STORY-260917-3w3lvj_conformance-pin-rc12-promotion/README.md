# STORY-260917-3w3lvj: conformance-pin-rc12-promotion

## Description
Advance the curator hosted-gate conformance pin (SPEC_PIN in .github/workflows/ci.yml) from curator-spec 87a0d00 (v1.0.0-rc.11) to the qualified v1.0.0-rc.12 release at dced9b8, landing it together with the wave-1 manager implementations whose vectors the new root publishes (E2 TASK-260916-55g9dg, E4 TASK-260916-3oh0u8, S4 TASK-260910-gocke2): internal/config TestManagerConfigV2Vectors renders every published manager-config-v2 case, so the pin can only move on a tree that implements every knob at once. Operator decision 2026-09-17: tag rc.12 at dced9b8, qualify, promote; E1 and wave 3 follow in rc.13.

## Scope
curator .github/workflows/ci.yml SPEC_PIN; internal/config, internal/contextresolve/contextmaterialize (E2), cmd/curator umbrella + internal/shell (E4), internal/envprofile MCP passthrough (S4) as the union of the three held candidates

## Acceptance Criteria
SPEC_PIN equals the full commit of the signed v1.0.0-rc.12 tag; qualification evidence (tag signature, make validate at the tag, suite digests) recorded before the pin moves; the union tree passes the hosted gate on all lanes at the new root with no skipped or weakened vector case; the three manager stories are delivered by this landing and closed on the board with evidence
