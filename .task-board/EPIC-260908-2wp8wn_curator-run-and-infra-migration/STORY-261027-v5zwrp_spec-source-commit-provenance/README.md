# STORY-261027-v5zwrp: spec-source-commit-provenance

## Description
Restore truthful protected-main provenance for reviewed human-signed fast-forward source delivery while preserving distinct release-tag signing authority. Trigger: curator-spec post-landing Specification run40569031707 at74a7dcc503c0fd7acd491299ab71290ae40e708e rejects the authorized source signing key because only the release signer trust file is configured. No published history rewrite, tag authority expansion or gate bypass.

## Scope
Curator-spec source-commit versus release-tag trust policy, decision record, verifier/workflow/docs and targeted regression controls.

## Acceptance Criteria
Independent security-depth review accepts exact signed source; required PR and post-main provenance checks pass; existing release-tag trust bytes and negative guarantees remain intact.
