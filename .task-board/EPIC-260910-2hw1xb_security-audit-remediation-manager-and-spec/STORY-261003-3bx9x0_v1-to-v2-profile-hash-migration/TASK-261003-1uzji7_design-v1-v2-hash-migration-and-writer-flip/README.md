# TASK-261003-1uzji7: design-v1-v2-hash-migration-and-writer-flip

## Description
Spec + curator design for atomic v1->v2 identity migration; then writer flip; includes the TestRC14MigrationRehashesLegacyIdentities regression from the 1foyf3 draft.

## Scope
(define task scope)

## Acceptance Criteria
Atomic v1->v2 identity migration that rehashes and never relabels; EnableV2Writers=true; TestRC14MigrationRehashesLegacyIdentities restored; real-entry and legacy/NUL/mismatch negatives; rc.14 byte-exact-snapshot gap row removed only if the case passes; lands after the rc.3 tag.
