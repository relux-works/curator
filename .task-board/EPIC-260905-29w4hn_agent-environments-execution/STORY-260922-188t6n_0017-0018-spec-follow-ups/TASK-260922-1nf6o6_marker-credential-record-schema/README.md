# TASK-260922-1nf6o6: marker-credential-record-schema

## Description
F-S1 (0017 choice 5): extend the passthrough marker schema so each credential record carries isolation (effective mode shared|isolated), strategy, source_role (native = bytes owned by the native home, managed = bytes owned by the managed home), backend (file|keychain|ambient), backend_version (verified tool release) and provenance (provisioned|repaired|migrated); linkless strategies (macOS per-home keychain under isolated, ambient) record without a path; a schema-1 marker records path and strategy only and is never rewritten to add the record; publication under the manager-home mutation lock via same-directory temp + atomic rename with journal protection, rollback restores the preceding marker; backups and discovery never follow auth symlinks (lstat) and never archive credential bytes. The curator enactment (manager publishes the record) is a follow-up curator leaf named in results.md.

## Scope
curator-spec: environments §7.4/§8.4.1 and the marker schema (new schema version), schema cases (valid/invalid), vectors if the section demands, CHANGELOG. No frozen v1 protocol schema change.

## Acceptance Criteria
1) new marker schema version with the six record fields and the linkless-no-path rule, plus schema cases: valid record per strategy, invalid missing field, invalid unknown backend, schema-1 marker still valid and never carries the record; 2) environments text names the publication/rollback/lstat rules normatively; 3) make validate green (spec-gate.sh); 4) CHANGELOG entry; 5) results.md names the curator follow-up leaf and the exact fields it must publish.
