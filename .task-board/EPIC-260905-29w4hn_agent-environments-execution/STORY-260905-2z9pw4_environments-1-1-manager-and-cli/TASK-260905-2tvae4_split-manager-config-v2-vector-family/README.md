# TASK-260905-2tvae4: split-manager-config-v2-vector-family

## Description
Rework of the batch-3 head 9af8af8 (PR #41): the pinned Go manager fails TestManagerConfigVectors on the schema2-* cases added to vectors/manager-config.json. Move the schema-2 cases into a new vector family vectors/manager-config-v2.json, keep the schema-1 file byte-identical to main, wire generator/validator/README/manifest, reproduce the Implementations lane locally, force-with-lease the draft branch, watch PR #41. Brief: producer-brief-manager-split-vectors.md.

## Scope
(define task scope)

## Acceptance Criteria
vectors/manager-config.json identical to a68559b; manager-config-v2.json generated and validated; Implementations lane green on PR #41; one signed commit past a68559b; report attached.
