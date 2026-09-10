# STORY-260910-stz5f0: import-upstream-high-water

## Description
Finding P4 (Low): import_bundle verifies signatures, chain and Merkle root but never compares the upstream snapshot against a persisted high-water for that upstream, so an old but validly signed bundle imports as new.

## Scope
curator-skill-registry bundle.py/store.py

## Acceptance Criteria
Each upstream key retains its accepted high-water; a bundle below it is rejected or warned; test covers the rollback bundle
