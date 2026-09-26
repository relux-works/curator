# TASK-260923-2gt5f6: publish-passthrough-credential-record

## Description
When curator provisions, repairs or migrates a passthrough credential link, write the agent-environment-marker-v2 credential record (isolation, strategy, source_role, backend, backend_version, provenance; path only for linkable strategies) under the manager-home lock via same-directory temp + atomic rename with journal protection; a schema-1 marker is never rewritten just to add the record. Builds on the 0017 implementation (STORY-260922-1cenbr) once it is integrated.

## Scope
curator: internal/envprofile marker writer and the provision/repair/migrate call sites, tests, docs, CHANGELOG.

## Acceptance Criteria
1) each of provisioned/repaired/migrated writes the correct record at the production entry; 2) linkless strategies record without a path; 3) schema-1 markers untouched unless rewritten for another reason; 4) output validates against agent-environment-marker-v2 from curator-spec main; 5) mutants killed; CHANGELOG
