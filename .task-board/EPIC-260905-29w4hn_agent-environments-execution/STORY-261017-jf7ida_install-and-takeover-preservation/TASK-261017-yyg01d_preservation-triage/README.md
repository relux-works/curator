# TASK-261017-yyg01d: preservation-triage

## Description
Triage five field reports reproduced in an isolated sandbox by another operator on curator c2966302 and on 26d8ddb plus the migration fix. (A) curator install . silently deletes unmanaged entries under .agents/skills/ (unmanaged skill folders and a loose file), exits 0 with only up-to-date lines, and leaves a dangling unmanaged link in a native skills directory. (B) a skill package without a manifest passes curator skill check, but install fails with 'install marker is invalid for schema 5' from internal/marker. (C) a takeover of .agents/ or of native entrypoints could destroy hand-authored instruction files kept only in managed output. (D) a first profile install without --use activates the profile. (E) multi-runtime switching ends partial (profile_use_partial): some native homes switched while others refused on unmanaged entrypoints.

## Scope
read-only research: sandbox reproduction with a private scratch HOME and project, spec mapping, classification and fix leaves; no product change

## Acceptance Criteria
1. Each of A-E reproduced or not on current main in a private scratch HOME and project, with exact commands and exit codes; nothing runs against a real home. 2. Each mapped to the governing curator-spec clauses (core, environments, skillfile) with a verdict: implementation bug, spec gap, or specified behaviour, and if specified, whether diagnostics and docs say so plainly. 3. For each bug, a fix leaf: required behaviour, negative and positive controls, touched packages, risk. 4. A priority order, with A first if it reproduces (silent deletion of user content).
