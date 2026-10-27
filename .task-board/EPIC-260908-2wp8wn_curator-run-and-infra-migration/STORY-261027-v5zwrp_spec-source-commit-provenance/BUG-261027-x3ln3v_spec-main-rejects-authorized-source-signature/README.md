# BUG-261027-x3ln3v: spec-main-rejects-authorized-source-signature

## Description
Protected-main provenance fails after exact reviewed signed fast-forward: local maintainer_signature_valid is configured only with release-tag signers and falls through to GitHub-created merge-only validation. Reproduce on landed74a7dcc503c0fd7acd491299ab71290ae40e708e/run40569031707/job122417945263. Separate authorized source-commit trust from unchanged release-tag trust; preserve exact-head signed delivery, default-branch containment, strict unknown/unsigned rejection and immutable history.

## Scope
Small curator-spec trust-policy decision and complete affected verifier/workflow/docs change. No tag creation, release, host install, credentials or profile operation.

## Acceptance Criteria
Real current source signature is accepted by the intended commit-only policy; unsigned/unknown/wrong-target/uncontained and unauthorized release-tag signatures reject as intended; release tag allowlist remains unchanged; independent review and exact-head checks plus post-main job pass.
