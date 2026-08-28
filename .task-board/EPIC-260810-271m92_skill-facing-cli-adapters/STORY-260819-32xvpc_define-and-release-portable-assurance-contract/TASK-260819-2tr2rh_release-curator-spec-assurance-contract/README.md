# release-curator-spec-assurance-contract-rc8

## Description
Supersede the failed immutable rc.7 publication with a reviewed rc.8 release. Fix the release workflow so Python validation cannot dirty the tagged checkout, update all normative rc.8 version and release metadata surfaces, preserve rc.7 history, then merge, sign, publish, and verify rc.8 without claiming any verified platform implementation.

## Scope
curator-spec workflow, normative version surfaces, release metadata, PR merge, signed rc.8 tag, GitHub prerelease assets, and board evidence. Do not rewrite or delete rc.7 and do not add implementation conformance claims.

## Acceptance Criteria
A reviewed workflow regression fix prevents Python bytecode or equivalent generated files from failing the clean-checkout release gate; every authoritative version and release metadata surface consistently identifies v1.0.0-rc.8; PR checks and post-merge checks pass; the merge commit and annotated rc.8 tag are signature-verified; the GitHub prerelease and canonical artifacts exist and match recorded digests; rc.7 remains immutable; verified implementation and platform claim sets remain empty.
