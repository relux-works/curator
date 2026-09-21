# STORY-260921-atwkfi: draft-sources-v1-corpus-follow-ups

## Description
Corpus/fixture defects in curator-spec conformance/draft-sources-v1 found by the curator implementation harness (TASK-260910-1xya7x, BUG-260920-2eg8nv). Work happens in the curator-spec repository (control root curator-spec, config .temp/launcher-migration/task-board.config.json); each leaf lands in curator-spec and is followed by a curator-side pin bump + driven row where a bound existed.

## Scope
curator-spec conformance/draft-sources-v1 corpus (schema-cases, index, vectors); curator SPEC_PIN promotion and bound→driven conversions

## Acceptance Criteria
every listed corpus defect fixed in curator-spec with the normative text unchanged or explicitly amended; the curator harness drives the affected cases (no bound left for them) after the pin bump
