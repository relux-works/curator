# STORY-260922-188t6n: 0017-0018-spec-follow-ups

## Description
Named spec follow-ups of the adopted decisions 0017 and 0018 (curator-spec decisions/0017-environment-credential-modes.md choices 5 and 6; decisions/0018-curator-run-permission-interface.md choice 7 and Compatibility section). Control root: curator-spec (separate board owner: leaves land by PR + signed ff push + bound worktree complete; local gate = spec-gate.sh override). Three leaves: F-S1 marker credential record, F-S2 launch-env fragment permission members + minimum transport token, F-S3 fleet isolated policy. Each is a reviewed spec revision with schema cases, vectors where the section demands them, CHANGELOG, and no frozen v1 protocol schema change.

## Scope
curator-spec: environments.md, manager.md, launcher SPEC mirrors named per leaf, schemas (marker, manager-config-v2, system-config-v2), conformance cases/vectors, CHANGELOG. No curator or launcher code.

## Acceptance Criteria
All three leaves landed on curator-spec main with green make validate; each names the curator or launcher follow-up leaf its enactment requires; frozen v1 protocol schemas untouched.
